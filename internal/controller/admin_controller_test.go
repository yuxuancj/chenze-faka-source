package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chenze-faka/internal/middleware"
	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAdminTestServer() (*gin.Engine, string) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: ":memory:",
	}
	database.Init(cfg)
	database.AutoMigrate()

	adminCtrl := NewAdminController(nil)
	authMw := middleware.NewAuthMiddleware("test-secret", 72)

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		admin := api.Group("/admin", authMw.AuthRequired())
		{
			admin.GET("/dashboard", adminCtrl.Dashboard)

			cats := admin.Group("/categories")
			{
				cats.GET("", adminCtrl.CategoryList)
				cats.POST("", adminCtrl.CategoryCreate)
				cats.PUT("", adminCtrl.CategoryUpdate)
				cats.DELETE("/:id", adminCtrl.CategoryDelete)
			}

			products := admin.Group("/products")
			{
				products.POST("", adminCtrl.ProductCreate)
				products.PUT("", adminCtrl.ProductUpdate)
				products.DELETE("/:id", adminCtrl.ProductDelete)
			}

			cards := admin.Group("/cards")
			{
				cards.POST("/import", adminCtrl.CardImport)
				cards.GET("", adminCtrl.CardList)
			}

			logs := admin.Group("/logs")
			{
				logs.GET("/operations", adminCtrl.OperationLogs)
			}

			admin.PUT("/settings", adminCtrl.UpdateSettings)
		}
	}

	adminToken := createAdminUserAndToken("admin_test", "admin123")
	return r, adminToken
}

func TestAdminDashboard(t *testing.T) {
	r, token := setupAdminTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
	assert.NotNil(t, resp["data"])
}

func TestAdminCategoryList(t *testing.T) {
	r, token := setupAdminTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/categories", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
}

func TestAdminCategoryCreate(t *testing.T) {
	r, token := setupAdminTestServer()

	body := map[string]interface{}{
		"name": "测试分类",
		"icon": "📁",
		"sort": 1,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/categories", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "测试分类", data["name"])
	assert.NotNil(t, data["id"])
}

func TestAdminCategoryUpdate(t *testing.T) {
	r, token := setupAdminTestServer()

	createBody := map[string]interface{}{
		"name": "原始分类",
		"icon": "📁",
		"sort": 1,
	}
	createBytes, _ := json.Marshal(createBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/categories", bytes.NewReader(createBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	data := createResp["data"].(map[string]interface{})
	catID := data["id"].(float64)

	updateBody := map[string]interface{}{
		"id":   catID,
		"name": "更新分类名",
		"icon": "📂",
		"sort": 2,
	}
	updateBytes, _ := json.Marshal(updateBody)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/categories", bytes.NewReader(updateBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	respData := resp["data"].(map[string]interface{})
	assert.Equal(t, "更新分类名", respData["name"])
}

func TestAdminCategoryDelete(t *testing.T) {
	r, token := setupAdminTestServer()

	createBody := map[string]interface{}{
		"name": "待删除分类",
		"icon": "🗑",
		"sort": 1,
	}
	createBytes, _ := json.Marshal(createBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/categories", bytes.NewReader(createBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	data := createResp["data"].(map[string]interface{})
	catID := data["id"].(float64)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/categories/"+itoa(uint(catID)), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
}

func TestAdminProductCreate(t *testing.T) {
	r, token := setupAdminTestServer()

	body := map[string]interface{}{
		"name":     "测试商品",
		"category": "测试分类",
		"price":    9.99,
		"stock":    100,
		"status":   1,
		"sort":     1,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/products", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "测试商品", data["name"])
}

func TestAdminProductUpdate(t *testing.T) {
	r, token := setupAdminTestServer()

	createBody := map[string]interface{}{
		"name":     "原始商品",
		"category": "测试分类",
		"price":    10.00,
		"stock":    50,
		"sort":     1,
	}
	createBytes, _ := json.Marshal(createBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/products", bytes.NewReader(createBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	data := createResp["data"].(map[string]interface{})
	prodID := data["id"].(float64)

	updateBody := map[string]interface{}{
		"id":          prodID,
		"name":        "更新商品名",
		"category":    "新分类",
		"price":       20.00,
		"description": "更新描述",
		"sort":        2,
	}
	updateBytes, _ := json.Marshal(updateBody)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/products", bytes.NewReader(updateBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	respData := resp["data"].(map[string]interface{})
	assert.Equal(t, "更新商品名", respData["name"])
}

func TestAdminProductDelete(t *testing.T) {
	r, token := setupAdminTestServer()

	createBody := map[string]interface{}{
		"name":     "待删除商品",
		"category": "测试分类",
		"price":    5.00,
		"stock":    10,
		"sort":     1,
	}
	createBytes, _ := json.Marshal(createBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/products", bytes.NewReader(createBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	data := createResp["data"].(map[string]interface{})
	prodID := data["id"].(float64)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/products/"+itoa(uint(prodID)), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
}

func TestAdminCardImport(t *testing.T) {
	r, token := setupAdminTestServer()

	createBody := map[string]interface{}{
		"name":     "卡密商品",
		"category": "测试分类",
		"price":    10.00,
		"stock":    0,
		"sort":     1,
	}
	createBytes, _ := json.Marshal(createBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/products", bytes.NewReader(createBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	prodData := createResp["data"].(map[string]interface{})
	prodID := prodData["id"].(float64)

	importBody := map[string]interface{}{
		"product_id": prodID,
		"card_text":  "ADMIN-001\nADMIN-002\nADMIN-003\n",
	}
	importBytes, _ := json.Marshal(importBody)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/cards/import", bytes.NewReader(importBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(3), data["imported"])
}

func TestAdminCardList(t *testing.T) {
	r, token := setupAdminTestServer()

	createBody := map[string]interface{}{
		"name":     "卡密商品",
		"category": "测试分类",
		"price":    10.00,
		"stock":    0,
		"sort":     1,
	}
	createBytes, _ := json.Marshal(createBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/products", bytes.NewReader(createBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	prodData := createResp["data"].(map[string]interface{})
	prodID := prodData["id"].(float64)

	importBody := map[string]interface{}{
		"product_id": prodID,
		"card_text":  "LIST-001\nLIST-002\n",
	}
	importBytes, _ := json.Marshal(importBody)

	importReq := httptest.NewRequest(http.MethodPost, "/api/admin/cards/import", bytes.NewReader(importBytes))
	importReq.Header.Set("Content-Type", "application/json")
	importReq.Header.Set("Authorization", "Bearer "+token)
	importW := httptest.NewRecorder()
	r.ServeHTTP(importW, importReq)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/cards", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data["cards"])
}

func TestAdminOperationLogs(t *testing.T) {
	r, token := setupAdminTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/logs/operations", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
}

func TestAdminUpdateSettings(t *testing.T) {
	r, token := setupAdminTestServer()

	body := map[string]interface{}{
		"site_name":    "测试站点",
		"order_expire": 30,
		"card_prefix":  "CARD",
		"maintenance":  false,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, true, data["updated"])
}

func TestAdminUnauthorized(t *testing.T) {
	r, _ := setupAdminTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
