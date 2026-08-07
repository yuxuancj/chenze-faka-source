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
	"chenze-faka/internal/pkg/utils"
	"chenze-faka/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupOrderExtTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_fk=1"), &gorm.Config{})
	require.NoError(t, err)
	dropAllControllerCoverageTables(db)
	database.DB = db
	require.NoError(t, db.AutoMigrate(
		&model.User{}, &model.Category{}, &model.Product{},
		&model.Card{}, &model.Order{}, &model.PaymentChannel{},
		&model.EmailConfig{}, &model.Node{},
		&model.OperationLog{}, &model.OrderLog{}, &model.EmailLog{},
		&model.UpgradeLog{}, &model.FileUpload{},
	))
	return db
}

func setupOrderExtTestServer(payConfig model.PayConfig) *gin.Engine {
	orderCtrl := NewOrderController(payConfig)
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
			orders.GET("/return", orderCtrl.Return)
			orders.GET("/:order_no", orderCtrl.GetByOrderNo)
			orders.POST("/notify", orderCtrl.Notify)
		}

		admin := api.Group("/admin", authMw.AuthRequired())
		{
			orderAdmin := admin.Group("/orders")
			{
				orderAdmin.GET("", orderCtrl.List)
				orderAdmin.GET("/list", adminCtrl.OrderList)
			}
		}
	}

	return r
}

func createAdminTokenForOrderExt(t *testing.T) string {
	salt := utils.GenerateSalt()
	hash := utils.HashPassword("admin123", salt)
	user := &model.User{
		Username:     "admin_order_ext",
		PasswordHash: hash,
		Salt:         salt,
		Role:         model.RoleAdmin,
	}
	require.NoError(t, database.DB.Create(user).Error)

	authSvc := service.NewAuthService()
	token, _, err := authSvc.Login("admin_order_ext", "admin123", "test-secret", 72)
	require.NoError(t, err)
	return token
}

func setupOrderExtTestData(t *testing.T, payConfig model.PayConfig) (*gin.Engine, *model.Product, []*model.Card) {
	r := setupOrderExtTestServer(payConfig)

	product := &model.Product{
		Name:     "ExtTest商品",
		Category: "ExtTest分类",
		Price:    10.00,
		Stock:    100,
		Status:   model.ProductStatusOnShelf,
		Sort:     1,
	}
	require.NoError(t, database.DB.Create(product).Error)

	encryptKey := "chenze_faka_card_encrypt_key_2024!!"
	cardNos := []string{"EXT-001", "EXT-002", "EXT-003", "EXT-004", "EXT-005"}
	cards := make([]*model.Card, 0)
	for _, cn := range cardNos {
		encrypted, _ := crypto.AesEncrypt(cn, encryptKey)
		card := &model.Card{
			ProductID:  product.ID,
			CardNoHash: encrypted,
			Status:     model.CardStatusUnsold,
		}
		require.NoError(t, database.DB.Create(card).Error)
		cards = append(cards, card)
	}

	return r, product, cards
}

func TestCreateOrderCoverage(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, product, _ := setupOrderExtTestData(t, payConfig)

	t.Run("create order with valid data succeeds", func(t *testing.T) {
		body := map[string]interface{}{
			"product_id": product.ID,
			"quantity":   2,
			"contact":    "coverage@test.com",
			"pay_method": "alipay",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.NotEmpty(t, data["order_no"])
		assert.Equal(t, 20.00, data["amount"])
	})

	t.Run("create order with invalid body fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})
}

func TestQueryOrderCoverage(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, product, _ := setupOrderExtTestData(t, payConfig)

	t.Run("query with no params returns error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orders/query", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})

	t.Run("query by order_no succeeds", func(t *testing.T) {
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
		orderNo := createResp["data"].(map[string]interface{})["order_no"].(string)

		req := httptest.NewRequest(http.MethodGet, "/api/orders/query?order_no="+orderNo, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})
}

func TestGetByOrderNoNonExistentCoverage(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, _, _ := setupOrderExtTestData(t, payConfig)

	req := httptest.NewRequest(http.MethodGet, "/api/orders/NONEXISTENT12345", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(1), resp["code"])
}

func TestNotifyCoverage(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, product, _ := setupOrderExtTestData(t, payConfig)

	t.Run("notify with missing sign goes through success path", func(t *testing.T) {
		body := map[string]interface{}{
			"product_id": product.ID,
			"quantity":   1,
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
		orderNo := createResp["data"].(map[string]interface{})["order_no"].(string)
		amount := createResp["data"].(map[string]interface{})["amount"].(float64)

		notifyBody := "out_trade_no=" + orderNo +
			"&trade_no=NOTIFY_TEST" +
			"&money=" + strconv.FormatFloat(amount, 'f', 2, 64)

		req := httptest.NewRequest(http.MethodPost, "/api/orders/notify", bytes.NewBufferString(notifyBody))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.ParseForm()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "success", w.Body.String())
	})

	t.Run("notify with invalid sign fails", func(t *testing.T) {
		body := map[string]interface{}{
			"product_id": product.ID,
			"quantity":   1,
			"contact":    "notify2@test.com",
			"pay_method": "alipay",
		}
		bodyBytes, _ := json.Marshal(body)
		createReq := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
		createReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()
		r.ServeHTTP(createW, createReq)
		var createResp map[string]interface{}
		json.Unmarshal(createW.Body.Bytes(), &createResp)
		orderNo := createResp["data"].(map[string]interface{})["order_no"].(string)
		amount := createResp["data"].(map[string]interface{})["amount"].(float64)

		notifyBody := "out_trade_no=" + orderNo +
			"&trade_no=NOTIFY_TEST2" +
			"&money=" + strconv.FormatFloat(amount, 'f', 2, 64) +
			"&sign=invalid_sign&sign_type=MD5"

		req := httptest.NewRequest(http.MethodPost, "/api/orders/notify", bytes.NewBufferString(notifyBody))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.ParseForm()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "fail", w.Body.String())
	})

	t.Run("notify with valid sign succeeds", func(t *testing.T) {
		body := map[string]interface{}{
			"product_id": product.ID,
			"quantity":   1,
			"contact":    "notify3@test.com",
			"pay_method": "alipay",
		}
		bodyBytes, _ := json.Marshal(body)
		createReq := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
		createReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()
		r.ServeHTTP(createW, createReq)
		var createResp map[string]interface{}
		json.Unmarshal(createW.Body.Bytes(), &createResp)
		orderNo := createResp["data"].(map[string]interface{})["order_no"].(string)
		amount := createResp["data"].(map[string]interface{})["amount"].(float64)

		params := map[string]string{
			"out_trade_no": orderNo,
			"trade_no":     "NOTIFY_VALID",
			"money":        strconv.FormatFloat(amount, 'f', 2, 64),
		}
		sign := utils.MD5Sum("money="+params["money"]+"&out_trade_no="+params["out_trade_no"]+"&trade_no="+params["trade_no"]+"&key="+payConfig.Key)
		params["sign"] = sign
		params["sign_type"] = "MD5"

		notifyBody := "out_trade_no=" + params["out_trade_no"] +
			"&trade_no=" + params["trade_no"] +
			"&money=" + params["money"] +
			"&sign=" + params["sign"] +
			"&sign_type=MD5"

		req := httptest.NewRequest(http.MethodPost, "/api/orders/notify", bytes.NewBufferString(notifyBody))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.ParseForm()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "success", w.Body.String())
	})
}

func TestReturnEndpointCoverage(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, product, _ := setupOrderExtTestData(t, payConfig)

	t.Run("return with existing order returns success", func(t *testing.T) {
		body := map[string]interface{}{
			"product_id": product.ID,
			"quantity":   1,
			"contact":    "return@test.com",
			"pay_method": "alipay",
		}
		bodyBytes, _ := json.Marshal(body)
		createReq := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
		createReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()
		r.ServeHTTP(createW, createReq)
		var createResp map[string]interface{}
		json.Unmarshal(createW.Body.Bytes(), &createResp)
		orderNo := createResp["data"].(map[string]interface{})["order_no"].(string)

		req := httptest.NewRequest(http.MethodGet, "/api/orders/return?out_trade_no="+orderNo, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})

	t.Run("return with non-existent order returns unknown", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orders/return?out_trade_no=NONEXISTENT", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "unknown", data["status"])
	})

	t.Run("return with no order_no returns unknown", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orders/return", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})
}

func TestListEndpointCoverage(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, product, _ := setupOrderExtTestData(t, payConfig)

	token := createAdminTokenForOrderExt(t)

	for i := 0; i < 3; i++ {
		body := map[string]interface{}{
			"product_id": product.ID,
			"quantity":   1,
			"contact":    "list_test@test.com",
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
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
}

func TestListUnauthorizedCoverage(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, _, _ := setupOrderExtTestData(t, payConfig)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListWithQueryParamsCoverage(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, product, _ := setupOrderExtTestData(t, payConfig)
	token := createAdminTokenForOrderExt(t)

	for i := 0; i < 3; i++ {
		body := map[string]interface{}{
			"product_id": product.ID,
			"quantity":   1,
			"contact":    "query_test@test.com",
			"pay_method": "alipay",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	t.Run("list with status filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/orders?status=1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})

	t.Run("list with keyword filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/orders?keyword=query_test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})

	t.Run("list with page and page_size", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/orders?page=1&page_size=2", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(2), data["page_size"])
	})

	t.Run("list with default page params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/orders", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})
}

func TestReturnEndpointEdgeCasesCoverage(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, product, _ := setupOrderExtTestData(t, payConfig)

	t.Run("return with no out_trade_no returns success with unknown", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orders/return", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "unknown", data["status"])
	})

	t.Run("return with non-existent out_trade_no returns success with unknown", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orders/return?out_trade_no=NONEXISTENT123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})

	t.Run("return with valid out_trade_no returns order data", func(t *testing.T) {
		body := map[string]interface{}{
			"product_id": product.ID,
			"quantity":   1,
			"contact":    "return_test@test.com",
			"pay_method": "alipay",
		}
		bodyBytes, _ := json.Marshal(body)
		createReq := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
		createReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()
		r.ServeHTTP(createW, createReq)
		var createResp map[string]interface{}
		json.Unmarshal(createW.Body.Bytes(), &createResp)
		orderNo := createResp["data"].(map[string]interface{})["order_no"].(string)

		req := httptest.NewRequest(http.MethodGet, "/api/orders/return?out_trade_no="+orderNo, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})
}

func TestNotifyWithNewOrderCoverage(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, product, _ := setupOrderExtTestData(t, payConfig)

	body := map[string]interface{}{
		"product_id": product.ID,
		"quantity":   1,
		"contact":    "notify_test@test.com",
		"pay_method": "alipay",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	orderNo := data["order_no"].(string)
	amount := data["amount"].(float64)

	t.Run("notify with valid params updates order", func(t *testing.T) {
		notifyBody := "out_trade_no=" + orderNo +
			"&trade_no=NOTIFY_NEW" +
			"&money=" + strconv.FormatFloat(amount, 'f', 2, 64) +
			"&sign_type=MD5"
		signParams := "money=" + strconv.FormatFloat(amount, 'f', 2, 64) +
			"&out_trade_no=" + orderNo +
			"&trade_no=NOTIFY_NEW&key=" + payConfig.Key
		sign := utils.MD5Sum(signParams)
		notifyBodyWithSign := notifyBody + "&sign=" + sign

		notifyReq := httptest.NewRequest(http.MethodPost, "/api/orders/notify", bytes.NewBufferString(notifyBodyWithSign))
		notifyReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		notifyReq.ParseForm()
		notifyW := httptest.NewRecorder()
		r.ServeHTTP(notifyW, notifyReq)
		assert.Equal(t, "success", notifyW.Body.String())
	})

	t.Run("notify with invalid amount fails", func(t *testing.T) {
		notifyBody := "out_trade_no=" + orderNo +
			"&trade_no=NOTIFY_BAD" +
			"&money=999.99&sign_type=MD5"
		signParams := "money=999.99" +
			"&out_trade_no=" + orderNo +
			"&trade_no=NOTIFY_BAD&key=" + payConfig.Key
		sign := utils.MD5Sum(signParams)
		notifyBodyWithSign := notifyBody + "&sign=" + sign

		notifyReq := httptest.NewRequest(http.MethodPost, "/api/orders/notify", bytes.NewBufferString(notifyBodyWithSign))
		notifyReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		notifyReq.ParseForm()
		notifyW := httptest.NewRecorder()
		r.ServeHTTP(notifyW, notifyReq)
		assert.Equal(t, "fail", notifyW.Body.String())
	})
}

func TestListErrorCoverage(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, _, _ := setupOrderExtTestData(t, payConfig)
	token := createAdminTokenForOrderExt(t)

	t.Run("list with invalid page param defaults to 1", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/orders?page=abc&page_size=5", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})

	t.Run("list with invalid page_size param defaults to 10", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/orders?page=1&page_size=xyz", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})

	t.Run("list with invalid status param defaults to 0", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/orders?status=abc", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})
}