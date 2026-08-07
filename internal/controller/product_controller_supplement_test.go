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
	"chenze-faka/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupProductSupplementTestServer() (*gin.Engine, string) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: ":memory:",
	}
	database.Init(cfg)
	database.AutoMigrate()

	productCtrl := NewProductController()
	adminCtrl := NewAdminController(nil)
	authMw := middleware.NewAuthMiddleware("test-secret", 72)

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		products := api.Group("/products")
		{
			products.GET("", productCtrl.List)
			products.GET("/on-shelf", productCtrl.OnShelf)
			products.GET("/on-shelf-grouped", productCtrl.OnShelfGrouped)
			products.GET("/:id", productCtrl.GetByID)
		}

		admin := api.Group("/admin", authMw.AuthRequired())
		{
			productAdmin := admin.Group("/products")
			{
				productAdmin.GET("", adminCtrl.ProductList)
				productAdmin.POST("", adminCtrl.ProductCreate)
				productAdmin.PUT("", adminCtrl.ProductUpdate)
				productAdmin.DELETE("/:id", adminCtrl.ProductDelete)
				productAdmin.GET("/on-shelf", productCtrl.OnShelf)
				productAdmin.GET("/on-shelf-grouped", productCtrl.OnShelfGrouped)
			}
		}
	}

	adminToken := createAdminUserAndToken("admin_supplement", "admin123")
	return r, adminToken
}

func TestProductSupplement_Create(t *testing.T) {
	r, token := setupProductSupplementTestServer()

	t.Run("create product successfully via admin", func(t *testing.T) {
		body := map[string]interface{}{
			"name":     "补充测试商品",
			"category": "测试分类",
			"price":    88.88,
			"stock":    200,
			"status":   model.ProductStatusOnShelf,
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
		require.NoError(t, err)
		assert.Equal(t, float64(0), resp["code"])

		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "补充测试商品", data["name"])
		assert.Equal(t, 88.88, data["price"])
		assert.Equal(t, float64(200), data["stock"])
	})

	t.Run("create product without auth returns 401", func(t *testing.T) {
		body := map[string]interface{}{
			"name":  "未授权商品",
			"price": 10.00,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/admin/products", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("create product with invalid json returns error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/admin/products", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, float64(1), resp["code"])
	})
}

func TestProductSupplement_Update(t *testing.T) {
	r, token := setupProductSupplementTestServer()

	product := createTestProduct("补充原始商品", 10.00, model.ProductStatusOnShelf)

	t.Run("update product successfully", func(t *testing.T) {
		body := map[string]interface{}{
			"id":     product.ID,
			"name":   "补充更新后商品",
			"price":  66.66,
			"sort":   10,
			"status": model.ProductStatusOnShelf,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/admin/products", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, float64(0), resp["code"])

		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "补充更新后商品", data["name"])
		assert.Equal(t, 66.66, data["price"])
	})

	t.Run("update nonexistent product returns error", func(t *testing.T) {
		body := map[string]interface{}{
			"id":   99999,
			"name": "幽灵商品",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/admin/products", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, float64(1), resp["code"])
	})
}

func TestProductSupplement_Delete(t *testing.T) {
	r, token := setupProductSupplementTestServer()

	product := createTestProduct("补充待删除商品", 15.00, model.ProductStatusOnShelf)

	t.Run("delete product successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/products/"+itoa(product.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, float64(0), resp["code"])

		var count int64
		database.DB.Model(&model.Product{}).Where("id = ?", product.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("delete nonexistent product returns success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/products/99999", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, float64(0), resp["code"])
	})
}

func TestProductSupplement_OnShelfGrouped(t *testing.T) {
	r, token := setupProductSupplementTestServer()

	createTestProduct("QQ币-10个", 10.00, model.ProductStatusOnShelf)
	createTestProduct("QQ币-50个", 50.00, model.ProductStatusOnShelf)
	createTestProduct("爱奇艺月卡", 15.00, model.ProductStatusOnShelf)
	createTestProduct("下架商品", 5.00, model.ProductStatusOffShelf)

	t.Run("get on-shelf grouped products via admin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/products/on-shelf-grouped", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, float64(0), resp["code"])

		data := resp["data"].([]interface{})
		require.Len(t, data, 1)

		group := data[0].(map[string]interface{})
		assert.NotEmpty(t, group["category"])
		products := group["products"].([]interface{})
		assert.Len(t, products, 3)
	})

	t.Run("get on-shelf grouped via public route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/products/on-shelf-grouped", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, float64(0), resp["code"])

		data := resp["data"].([]interface{})
		require.Len(t, data, 1)
	})
}

func TestProductSupplement_OnShelf(t *testing.T) {
	r, token := setupProductSupplementTestServer()

	createTestProduct("在售商品1", 10.00, model.ProductStatusOnShelf)
	createTestProduct("在售商品2", 20.00, model.ProductStatusOnShelf)
	createTestProduct("在售商品3", 30.00, model.ProductStatusOnShelf)
	createTestProduct("下架商品", 5.00, model.ProductStatusOffShelf)

	t.Run("get on-shelf products via admin route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/products/on-shelf", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, float64(0), resp["code"])

		data := resp["data"].([]interface{})
		assert.Len(t, data, 3)
	})

	t.Run("get on-shelf products via public route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/products/on-shelf", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, float64(0), resp["code"])

		data := resp["data"].([]interface{})
		assert.Len(t, data, 3)

		for _, p := range data {
			product := p.(map[string]interface{})
			assert.NotEqual(t, model.ProductStatusOffShelf, int(product["status"].(float64)))
		}
	})
}

func TestProductSupplement_FullCRUDWorkflow(t *testing.T) {
	r, token := setupProductSupplementTestServer()

	t.Run("full create-read-update-delete workflow", func(t *testing.T) {
		createBody := map[string]interface{}{
			"name":     "工作流商品",
			"category": "工作流分类",
			"price":    99.99,
			"stock":    100,
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
		assert.Equal(t, float64(0), createResp["code"])
		createData := createResp["data"].(map[string]interface{})
		productID := uint(createData["id"].(float64))

		updateBody := map[string]interface{}{
			"id":    productID,
			"name":  "工作流商品-已更新",
			"price": 199.99,
		}
		updateBytes, _ := json.Marshal(updateBody)

		updateReq := httptest.NewRequest(http.MethodPut, "/api/admin/products", bytes.NewReader(updateBytes))
		updateReq.Header.Set("Content-Type", "application/json")
		updateReq.Header.Set("Authorization", "Bearer "+token)
		updateW := httptest.NewRecorder()
		r.ServeHTTP(updateW, updateReq)

		var updateResp map[string]interface{}
		json.Unmarshal(updateW.Body.Bytes(), &updateResp)
		assert.Equal(t, float64(0), updateResp["code"])
		updateData := updateResp["data"].(map[string]interface{})
		assert.Equal(t, "工作流商品-已更新", updateData["name"])

		deleteReq := httptest.NewRequest(http.MethodDelete, "/api/admin/products/"+itoa(productID), nil)
		deleteReq.Header.Set("Authorization", "Bearer "+token)
		deleteW := httptest.NewRecorder()
		r.ServeHTTP(deleteW, deleteReq)

		var deleteResp map[string]interface{}
		json.Unmarshal(deleteW.Body.Bytes(), &deleteResp)
		assert.Equal(t, float64(0), deleteResp["code"])

		var count int64
		database.DB.Model(&model.Product{}).Where("id = ?", productID).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}

func TestProductSupplement_DBSideEffects(t *testing.T) {
	r, _ := setupProductSupplementTestServer()

	t.Run("verify product service uses correct DB", func(t *testing.T) {
		svc := service.NewProductService()
		products, err := svc.ListOnShelf()
		require.NoError(t, err)
		assert.NotNil(t, products)
	})

	t.Run("list after supplement setup returns data", func(t *testing.T) {
		createTestProduct("额外商品", 25.00, model.ProductStatusOnShelf)

		req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, float64(0), resp["code"])
	})
}