package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"chenze-faka/internal/middleware"
	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/crypto"
	"chenze-faka/internal/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupOrderTestServer() (*gin.Engine, string) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: ":memory:",
	}
	database.Init(cfg)
	database.AutoMigrate()

	orderCtrl := NewOrderController(model.PayConfig{})
	adminCtrl := NewAdminController(nil)
	authMw := middleware.NewAuthMiddleware("test-secret", 72)

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		orders := api.Group("/orders")
		{
			orders.POST("", orderCtrl.Create)
			orders.GET("/query", orderCtrl.Query)
			orders.GET("/:order_no", orderCtrl.GetByOrderNo)
			orders.POST("/notify", orderCtrl.Notify)
		}

		admin := api.Group("/admin", authMw.AuthRequired())
		{
			orderAdmin := admin.Group("/orders")
			{
				orderAdmin.GET("", adminCtrl.OrderList)
			}
		}
	}

	adminToken := createAdminUserAndToken("admin", "admin123")
	return r, adminToken
}

func setupOrderTestServerWithData() (*gin.Engine, string, *model.Product, []*model.Card) {
	r, token := setupOrderTestServer()

	product := &model.Product{
		Name:     "测试商品",
		Category: "测试分类",
		Price:    10.00,
		Stock:    100,
		Status:   model.ProductStatusOnShelf,
		Sort:     1,
	}
	database.DB.Create(product)

	encryptKey := "chenze_faka_card_encrypt_key_2024!!"
	cardNos := []string{"CARD001", "CARD002", "CARD003", "CARD004", "CARD005"}
	cards := make([]*model.Card, 0)
	for _, cn := range cardNos {
		encrypted, _ := crypto.AesEncrypt(cn, encryptKey)
		card := &model.Card{
			ProductID:  product.ID,
			CardNoHash: encrypted,
			Status:     model.CardStatusUnsold,
		}
		database.DB.Create(card)
		cards = append(cards, card)
	}

	database.DB.Model(&model.Product{}).Where("id = ?", product.ID).UpdateColumn("stock", 100)

	return r, token, product, cards
}

func TestCreateOrder(t *testing.T) {
	r, _, product, _ := setupOrderTestServerWithData()

	body := map[string]interface{}{
		"product_id": product.ID,
		"quantity":   2,
		"contact":    "test@example.com",
		"pay_method": "alipay",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotEmpty(t, data["order_no"])
	assert.Equal(t, 20.00, data["amount"])
	assert.NotEmpty(t, data["expire_at"])
}

func TestCreateOrderInvalidProduct(t *testing.T) {
	r, _, _, _ := setupOrderTestServerWithData()

	body := map[string]interface{}{
		"product_id": 99999,
		"quantity":   1,
		"contact":    "test@example.com",
		"pay_method": "alipay",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
}

func TestQueryOrder(t *testing.T) {
	r, _, product, _ := setupOrderTestServerWithData()

	body := map[string]interface{}{
		"product_id": product.ID,
		"quantity":   1,
		"contact":    "query@test.com",
		"pay_method": "alipay",
	}
	bodyBytes, _ := json.Marshal(body)

	createReq := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	createData := createResp["data"].(map[string]interface{})
	orderNo := createData["order_no"].(string)

	req := httptest.NewRequest(http.MethodGet, "/api/orders/query?order_no="+orderNo, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, orderNo, data["order_no"])
	assert.Equal(t, float64(0), data["status"])
}

func TestGetOrderByOrderNo(t *testing.T) {
	r, _, product, _ := setupOrderTestServerWithData()

	body := map[string]interface{}{
		"product_id": product.ID,
		"quantity":   1,
		"contact":    "getby@test.com",
		"pay_method": "wechat",
	}
	bodyBytes, _ := json.Marshal(body)

	createReq := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	createData := createResp["data"].(map[string]interface{})
	orderNo := createData["order_no"].(string)

	req := httptest.NewRequest(http.MethodGet, "/api/orders/"+orderNo, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, orderNo, data["order_no"])
	assert.Equal(t, float64(0), data["status"])
}

func TestPaymentNotify(t *testing.T) {
	r, _, product, _ := setupOrderTestServerWithData()

	body := map[string]interface{}{
		"product_id": product.ID,
		"quantity":   2,
		"contact":    "notify@test.com",
		"pay_method": "alipay",
	}
	bodyBytes, _ := json.Marshal(body)

	createReq := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	createData := createResp["data"].(map[string]interface{})
	orderNo := createData["order_no"].(string)
	amount := createData["amount"].(float64)

	notifyBody := "out_trade_no=" + orderNo +
		"&trade_no=TESTPAY123" +
		"&money=" + strconv.FormatFloat(amount, 'f', 2, 64)

	req := httptest.NewRequest(http.MethodPost, "/api/orders/notify", bytes.NewBufferString(notifyBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ParseForm()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())

	queryReq := httptest.NewRequest(http.MethodGet, "/api/orders/"+orderNo, nil)
	queryW := httptest.NewRecorder()
	r.ServeHTTP(queryW, queryReq)

	var queryResp map[string]interface{}
	json.Unmarshal(queryW.Body.Bytes(), &queryResp)
	queryData := queryResp["data"].(map[string]interface{})
	assert.Equal(t, float64(2), queryData["status"])
	assert.NotEmpty(t, queryData["cards"])
}

func TestAdminListOrders(t *testing.T) {
	r, token, product, _ := setupOrderTestServerWithData()

	for i := 0; i < 3; i++ {
		body := map[string]interface{}{
			"product_id": product.ID,
			"quantity":   1,
			"contact":    "adminlist@test.com",
			"pay_method": "alipay",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data)
	assert.NotEmpty(t, data["orders"])
}

func TestAdminListOrdersUnauthorized(t *testing.T) {
	r, _, _, _ := setupOrderTestServerWithData()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}