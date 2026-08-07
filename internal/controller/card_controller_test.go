package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chenze-faka/internal/middleware"
	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/crypto"
	"chenze-faka/internal/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupCardTestServer() (*gin.Engine, string) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: ":memory:",
	}
	database.Init(cfg)
	database.AutoMigrate()

	cardCtrl := NewCardController()
	adminCtrl := NewAdminController(nil)
	authMw := middleware.NewAuthMiddleware("test-secret", 72)

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		cards := api.Group("/cards")
		{
			cards.GET("/product/:id/count", cardCtrl.CountByProduct)
		}

		admin := api.Group("/admin", authMw.AuthRequired())
		{
			cardAdmin := admin.Group("/cards")
			{
				cardAdmin.POST("/import", adminCtrl.CardImport)
				cardAdmin.GET("", adminCtrl.CardList)
				cardAdmin.DELETE("/:id", adminCtrl.CardDelete)
			}
		}
	}

	adminToken := createAdminUserAndToken("admin", "admin123")
	return r, adminToken
}

func createTestProductForCard() *model.Product {
	product := &model.Product{
		Name:     "卡密商品",
		Category: "测试分类",
		Price:    10.00,
		Stock:    50,
		Status:   model.ProductStatusOnShelf,
		Sort:     1,
	}
	database.DB.Create(product)
	return product
}

func createTestCards(productID uint, count int) []*model.Card {
	encryptKey := "chenze_faka_card_encrypt_key_2024!!"
	cards := make([]*model.Card, 0)
	for i := 0; i < count; i++ {
		cardNo := "CARD" + itoa(uint(i+1))
		encrypted, _ := crypto.AesEncrypt(cardNo, encryptKey)
		card := &model.Card{
			ProductID:  productID,
			CardNoHash: encrypted,
			Status:     model.CardStatusUnsold,
		}
		database.DB.Create(card)
		cards = append(cards, card)
	}
	return cards
}

func TestImportCards(t *testing.T) {
	r, token := setupCardTestServer()
	product := createTestProductForCard()

	cardText := "CARD-001\nCARD-002\nCARD-003\nCARD-004\nCARD-005\n"

	body := map[string]interface{}{
		"product_id": product.ID,
		"card_text":  cardText,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/cards/import", bytes.NewReader(bodyBytes))
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
	assert.Equal(t, float64(5), data["total_count"])
	assert.Equal(t, float64(5), data["imported"])
}

func TestImportCardsUnauthorized(t *testing.T) {
	r, _ := setupCardTestServer()
	product := createTestProductForCard()

	body := map[string]interface{}{
		"product_id": product.ID,
		"card_text":  "CARD-001\n",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/cards/import", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListCards(t *testing.T) {
	r, token := setupCardTestServer()
	product := createTestProductForCard()
	createTestCards(product.ID, 5)

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
	assert.NotNil(t, data)
	assert.NotEmpty(t, data["cards"])
}

func TestListCardsByProduct(t *testing.T) {
	r, token := setupCardTestServer()
	product := createTestProductForCard()
	createTestCards(product.ID, 3)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/cards?product_id="+itoa(product.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	cards := data["cards"].([]interface{})
	assert.Len(t, cards, 3)
}

func TestDeleteCard(t *testing.T) {
	r, token := setupCardTestServer()
	product := createTestProductForCard()
	cards := createTestCards(product.ID, 3)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/cards/"+itoa(cards[0].ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
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

func TestCountCardsByProduct(t *testing.T) {
	r, _ := setupCardTestServer()
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
	assert.Equal(t, float64(product.ID), data["product_id"])
}

func TestCountCardsByProductEmpty(t *testing.T) {
	r, _ := setupCardTestServer()
	product := createTestProductForCard()

	req := httptest.NewRequest(http.MethodGet, "/api/cards/product/"+itoa(product.ID)+"/count", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(0), data["count"])
}