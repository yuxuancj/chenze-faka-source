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

func setupCardExtTestServer() (*gin.Engine, string) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: ":memory:",
	}
	database.Init(cfg)
	database.AutoMigrate()

	cardCtrl := NewCardController()
	authMw := middleware.NewAuthMiddleware("test-secret", 72)

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		cards := api.Group("/cards")
		{
			cards.POST("/import", cardCtrl.Import)
			cards.GET("/:id", cardCtrl.GetByID)
			cards.GET("", cardCtrl.List)
			cards.DELETE("/:id", cardCtrl.Delete)
			cards.GET("/product/:id/count", cardCtrl.CountByProduct)
		}

		admin := api.Group("/admin", authMw.AuthRequired())
		{
			cardAdmin := admin.Group("/cards")
			{
				cardAdmin.POST("/import", NewAdminController(nil).CardImport)
				cardAdmin.GET("", NewAdminController(nil).CardList)
			}
		}
	}

	adminToken := createAdminUserAndToken("admin_ext", "admin123")
	return r, adminToken
}

func TestCardController_Import(t *testing.T) {
	r, _ := setupCardExtTestServer()
	product := createTestProductForCard()

	cardText := "EXT-001\nEXT-002\nEXT-003\n"

	body := map[string]interface{}{
		"product_id": product.ID,
		"card_text":  cardText,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/cards/import", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(3), data["total_count"])
	assert.Equal(t, float64(3), data["imported"])
}

func TestCardController_ImportWithoutBody(t *testing.T) {
	r, _ := setupCardExtTestServer()

	req := httptest.NewRequest(http.MethodPost, "/api/cards/import", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
}

func TestCardController_ImportEmptyCardText(t *testing.T) {
	r, _ := setupCardExtTestServer()
	product := createTestProductForCard()

	body := map[string]interface{}{
		"product_id": product.ID,
		"card_text":  "   \n  \n  ",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/cards/import", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
}

func TestCardController_GetByID(t *testing.T) {
	r, _ := setupCardExtTestServer()
	product := createTestProductForCard()
	createTestCards(product.ID, 3)

	req := httptest.NewRequest(http.MethodGet, "/api/cards/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["id"])
}

func TestCardController_GetByIDInvalid(t *testing.T) {
	r, _ := setupCardExtTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/cards/9999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCardController_List(t *testing.T) {
	r, _ := setupCardExtTestServer()
	product := createTestProductForCard()
	createTestCards(product.ID, 5)

	req := httptest.NewRequest(http.MethodGet, "/api/cards?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data)
	assert.NotEmpty(t, data["cards"])
	assert.Equal(t, float64(5), data["total"])
}

func TestCardController_ListPagination(t *testing.T) {
	r, _ := setupCardExtTestServer()
	product := createTestProductForCard()
	createTestCards(product.ID, 5)

	req := httptest.NewRequest(http.MethodGet, "/api/cards?page=1&page_size=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(5), data["total"])
	cards := data["cards"].([]interface{})
	assert.Len(t, cards, 2)
}

func TestCardController_Delete(t *testing.T) {
	r, _ := setupCardExtTestServer()
	product := createTestProductForCard()
	cards := createTestCards(product.ID, 3)

	req := httptest.NewRequest(http.MethodDelete, "/api/cards/"+itoa(cards[0].ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	var count int64
	database.DB.Model(&model.Card{}).Where("id = ?", cards[0].ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestCardController_DeleteInvalid(t *testing.T) {
	r, _ := setupCardExtTestServer()

	req := httptest.NewRequest(http.MethodDelete, "/api/cards/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
}

func TestCardController_CountByProduct(t *testing.T) {
	r, _ := setupCardExtTestServer()
	product := createTestProductForCard()
	createTestCards(product.ID, 7)

	req := httptest.NewRequest(http.MethodGet, "/api/cards/product/"+itoa(product.ID)+"/count", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(7), data["count"])
}

func TestCardController_ImportDuplicateCards(t *testing.T) {
	r, token := setupCardExtTestServer()
	product := createTestProductForCard()

	body := map[string]interface{}{
		"product_id": product.ID,
		"card_text":  "DUP-EXT-001\nDUP-EXT-001\nDUP-EXT-002\n",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/cards/import", bytes.NewReader(bodyBytes))
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
	assert.Equal(t, float64(3), data["total_count"])
	assert.Equal(t, float64(3), data["imported"])
}
