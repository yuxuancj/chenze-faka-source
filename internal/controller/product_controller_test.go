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
	"chenze-faka/internal/pkg/utils"
	"chenze-faka/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupProductTestServer() (*gin.Engine, string) {
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
			}
		}
	}

	adminToken := createAdminUserAndToken("admin", "admin123")

	return r, adminToken
}

func createAdminUserAndToken(username, password string) string {
	salt := utils.GenerateSalt()
	hash := utils.HashPassword(password, salt)
	user := &model.User{
		Username:     username,
		PasswordHash: hash,
		Salt:         salt,
		Role:         model.RoleAdmin,
	}
	database.DB.Create(user)

	authSvc := service.NewAuthService()
	token, _, _ := authSvc.Login(username, password, "test-secret", 72)
	return token
}

func createTestProduct(name string, price float64, status int) *model.Product {
	product := &model.Product{
		Name:     name,
		Category: "测试分类",
		Price:    price,
		Stock:    100,
		Status:   model.ProductStatusOnShelf,
		Sort:     0,
	}
	database.DB.Create(product)
	if status == model.ProductStatusOffShelf {
		database.DB.Model(&model.Product{}).Where("id = ?", product.ID).Update("status", model.ProductStatusOffShelf)
		product.Status = model.ProductStatusOffShelf
	}
	return product
}

func TestListProducts(t *testing.T) {
	r, _ := setupProductTestServer()
	createTestProduct("商品A", 10.00, model.ProductStatusOnShelf)
	createTestProduct("商品B", 20.00, model.ProductStatusOnShelf)

	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data)
	assert.NotEmpty(t, data["products"])
}

func TestListOnShelfProducts(t *testing.T) {
	r, _ := setupProductTestServer()
	createTestProduct("在售商品", 10.00, model.ProductStatusOnShelf)
	createTestProduct("下架商品", 20.00, model.ProductStatusOffShelf)

	req := httptest.NewRequest(http.MethodGet, "/api/products/on-shelf", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
}

func TestGetProductByID(t *testing.T) {
	r, _ := setupProductTestServer()
	product := createTestProduct("测试商品", 25.00, model.ProductStatusOnShelf)

	req := httptest.NewRequest(http.MethodGet, "/api/products/"+itoa(product.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "测试商品", data["name"])
	assert.Equal(t, 25.00, data["price"])
}

func TestGetProductByIDNotFound(t *testing.T) {
	r, _ := setupProductTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/products/99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(404), resp["code"])
}

func TestAdminCreateProduct(t *testing.T) {
	r, token := setupProductTestServer()

	body := map[string]interface{}{
		"name":     "管理员创建商品",
		"category": "测试分类",
		"price":    99.99,
		"stock":    50,
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
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "管理员创建商品", data["name"])
	assert.Equal(t, 99.99, data["price"])
}

func TestAdminCreateProductUnauthorized(t *testing.T) {
	r, _ := setupProductTestServer()

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
}

func TestAdminUpdateProduct(t *testing.T) {
	r, token := setupProductTestServer()
	product := createTestProduct("原始名称", 10.00, model.ProductStatusOnShelf)

	body := map[string]interface{}{
		"id":   product.ID,
		"name": "更新后名称",
		"price": 30.00,
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
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "更新后名称", data["name"])
	assert.Equal(t, 30.00, data["price"])
}

func TestAdminDeleteProduct(t *testing.T) {
	r, token := setupProductTestServer()
	product := createTestProduct("待删除商品", 15.00, model.ProductStatusOnShelf)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/products/"+itoa(product.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	var count int64
	database.DB.Model(&model.Product{}).Where("id = ?", product.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestAdminListProducts(t *testing.T) {
	r, token := setupProductTestServer()
	createTestProduct("商品X", 10.00, model.ProductStatusOnShelf)
	createTestProduct("商品Y", 20.00, model.ProductStatusOnShelf)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/products", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
}

func itoa(i uint) string {
	if i == 0 {
		return "0"
	}
	const digits = "0123456789"
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = digits[i%10]
		i /= 10
	}
	return string(buf[pos:])
}