package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/stretchr/testify/assert"
)

func TestOrderListErrorPath(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, _, _ := setupOrderExtTestData(t, payConfig)

	token := createAdminTokenForOrderExt(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/orders?status=abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
}

func TestOrderListMultipleFilters(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, product, _ := setupOrderExtTestData(t, payConfig)
	token := createAdminTokenForOrderExt(t)

	for i := 0; i < 5; i++ {
		body := map[string]interface{}{
			"product_id": product.ID,
			"quantity":   1,
			"contact":    "multi_filter@test.com",
			"pay_method": "alipay",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	t.Run("list with all query params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/orders?page=1&page_size=3&status=0&keyword=multi_filter", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})

	t.Run("list with large page size", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/orders?page=1&page_size=100", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})

	t.Run("list with invalid page params defaults to 1", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/orders?page=-1&page_size=-5", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["page"])
		assert.Equal(t, float64(10), data["page_size"])
	})
}

func TestGetByOrderNoWithContact(t *testing.T) {
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
		"contact":    "contact_lookup@test.com",
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

	t.Run("get by order no with contact filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orders/"+orderNo+"?contact=contact_lookup@test.com", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})

	t.Run("get by order no with wrong contact still works (contact ignored when orderNo present)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orders/"+orderNo+"?contact=wrong@test.com", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})
}

func TestNotifyEdgeCases(t *testing.T) {
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
		"contact":    "notify_edge@test.com",
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

	t.Run("notify with empty params goes through fail", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/orders/notify", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, "fail", w.Body.String())
	})

	t.Run("notify with non-existent order no fails", func(t *testing.T) {
		notifyBody := "out_trade_no=NONEXISTENT_ORDER" +
			"&trade_no=TEST" +
			"&money=" + strconv.FormatFloat(amount, 'f', 2, 64)
		req := httptest.NewRequest(http.MethodPost, "/api/orders/notify", bytes.NewBufferString(notifyBody))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.ParseForm()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, "fail", w.Body.String())
	})

	_ = orderNo
}

func TestQueryByContactEndpoint(t *testing.T) {
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
		"contact":    "contact_query_ep@test.com",
		"pay_method": "alipay",
	}
	bodyBytes, _ := json.Marshal(body)
	createReq := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	assert.Equal(t, http.StatusOK, createW.Code)

	t.Run("query by contact finds order", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orders/query?contact=contact_query_ep@test.com", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})

	t.Run("query by nonexistent contact returns error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orders/query?contact=ghost@test.com", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})
}

func TestListAdminUnauthorized(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, _, _ := setupOrderExtTestData(t, payConfig)

	t.Run("list without token returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/orders", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("list with invalid token returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/orders", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestCreateOrderInvalidProductSupplement(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, _, _ := setupOrderExtTestData(t, payConfig)

	t.Run("create order with nonexistent product fails", func(t *testing.T) {
		body := map[string]interface{}{
			"product_id": 99999,
			"quantity":   1,
			"contact":    "nobody@test.com",
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
		assert.Equal(t, float64(1), resp["code"])
	})
}

func TestReturnEndpointWithContact(t *testing.T) {
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
		"contact":    "return_contact@test.com",
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
}

func TestOrderListKeywordMatchesOrderNo(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, product, _ := setupOrderExtTestData(t, payConfig)
	token := createAdminTokenForOrderExt(t)

	body := map[string]interface{}{
		"product_id": product.ID,
		"quantity":   1,
		"contact":    "kw_match@test.com",
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

	req := httptest.NewRequest(http.MethodGet, "/api/admin/orders?keyword="+orderNo, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
}

func TestBuildPayURLEdgeCases(t *testing.T) {
	t.Run("build pay URL with special characters in product name", func(t *testing.T) {
		orderCtrl := NewOrderController(model.PayConfig{
			URL:      "https://pay.example.com",
			Merchant: "test_merchant",
			Key:      "test_key",
		})
		url := orderCtrl.BuildPayURL("ORDER-001", 999.99, "alipay", "商品/特殊\\字符")
		assert.NotEmpty(t, url)
		assert.Contains(t, url, "https://pay.example.com")
	})

	t.Run("build pay URL with zero amount", func(t *testing.T) {
		orderCtrl := NewOrderController(model.PayConfig{
			URL:      "https://pay.example.com",
			Merchant: "test_merchant",
			Key:      "test_key",
		})
		url := orderCtrl.BuildPayURL("ORDER-002", 0, "wechat", "TestProduct")
		assert.NotEmpty(t, url)
	})

	t.Run("build pay URL with long order no", func(t *testing.T) {
		orderCtrl := NewOrderController(model.PayConfig{
			URL:      "https://pay.example.com",
			Merchant: "test_merchant",
			Key:      "test_key",
		})
		longOrderNo := "ORDER-" + string(make([]byte, 100))
		url := orderCtrl.BuildPayURL(longOrderNo, 10.00, "alipay", "Product")
		assert.NotEmpty(t, url)
	})
}

func TestCreateOrderQuantityEdgeCases(t *testing.T) {
	setupOrderExtTestDB(t)
	defer func() { database.DB = nil }()

	payConfig := model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
	r, product, _ := setupOrderExtTestData(t, payConfig)

	t.Run("create order with quantity 0 fails", func(t *testing.T) {
		body := map[string]interface{}{
			"product_id": product.ID,
			"quantity":   0,
			"contact":    "q0@test.com",
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
		assert.Equal(t, float64(1), resp["code"])
	})

	t.Run("create order with quantity > 99 fails", func(t *testing.T) {
		body := map[string]interface{}{
			"product_id": product.ID,
			"quantity":   100,
			"contact":    "q100@test.com",
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
		assert.Equal(t, float64(1), resp["code"])
	})

	t.Run("create order with invalid pay method fails", func(t *testing.T) {
		body := map[string]interface{}{
			"product_id": product.ID,
			"quantity":   1,
			"contact":    "invalidpay@test.com",
			"pay_method": "bitcoin",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})

	t.Run("create order with missing contact fails", func(t *testing.T) {
		body := map[string]interface{}{
			"product_id": product.ID,
			"quantity":   1,
			"contact":    "",
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
		assert.Equal(t, float64(1), resp["code"])
	})
}