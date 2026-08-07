package controller

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"chenze-faka/internal/middleware"
	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAdminSupplementTestServer() (*gin.Engine, string) {
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
			admin.GET("/system/status", adminCtrl.SystemStatus)
			admin.GET("/system/license", adminCtrl.LicenseStatus)
			admin.POST("/system/license/verify", adminCtrl.VerifyLicense)

			admin.GET("/dashboard/order-status", adminCtrl.OrderStatusCounts)

			payments := admin.Group("/payments")
			{
				payments.GET("", adminCtrl.PaymentList)
				payments.GET("/all", adminCtrl.PaymentAll)
				payments.POST("", adminCtrl.PaymentCreate)
				payments.PUT("", adminCtrl.PaymentUpdate)
				payments.DELETE("/:id", adminCtrl.PaymentDelete)
			}

			emails := admin.Group("/emails")
			{
				emails.GET("", adminCtrl.EmailList)
				emails.POST("", adminCtrl.EmailCreate)
				emails.PUT("", adminCtrl.EmailUpdate)
				emails.DELETE("/:id", adminCtrl.EmailDelete)
				emails.POST("/test/:id", adminCtrl.EmailTest)
				emails.GET("/logs", adminCtrl.EmailLogs)
			}

			nodes := admin.Group("/nodes")
			{
				nodes.GET("", adminCtrl.NodeList)
				nodes.POST("", adminCtrl.NodeCreate)
				nodes.PUT("", adminCtrl.NodeUpdate)
				nodes.DELETE("/:id", adminCtrl.NodeDelete)
				nodes.POST("/ping/:id", adminCtrl.NodePing)
			}

			admin.GET("/settings", adminCtrl.GetSettings)

			upgrade := admin.Group("/upgrade")
			{
				upgrade.GET("/version", adminCtrl.GetVersion)
				upgrade.GET("/logs", adminCtrl.UpgradeLogs)
			}

			upload := admin.Group("/upload")
			{
				upload.POST("", adminCtrl.UploadFile)
				upload.GET("", adminCtrl.ListFiles)
				upload.DELETE("/:id", adminCtrl.DeleteFile)
			}
		}
	}

	adminToken := createAdminUserAndToken("admin_supp", "admin123")
	return r, adminToken
}

func TestAdminSystemStatus(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/system/status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "1.0.0", data["version"])
}

func TestAdminLicenseStatus(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/system/license", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "unknown", data["status"])
}

func TestAdminVerifyLicense(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/system/license/verify", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
	assert.Equal(t, "授权服务未初始化", resp["message"])
}

func TestAdminOrderStatusCounts(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/dashboard/order-status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Contains(t, data, "pending")
	assert.Contains(t, data, "paid")
	assert.Contains(t, data, "complete")
	assert.Contains(t, data, "cancel")
}

func TestAdminPaymentList(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/payments", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data["items"])
	assert.NotNil(t, data["total"])
}

func TestAdminPaymentAll(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/payments/all", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data["items"])
}

func TestAdminPaymentCreate(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	body := map[string]interface{}{
		"name":     "支付宝",
		"type":     "alipay",
		"icon":     "💰",
		"config":   `{"app_id":"test"}`,
		"fee_rate": 0.01,
		"sort":     1,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/payments", bytes.NewReader(bodyBytes))
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
	assert.Equal(t, "支付宝", data["name"])
	assert.NotNil(t, data["id"])
}

func TestAdminPaymentUpdate(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	createBody := map[string]interface{}{
		"name": "原始支付",
		"type": "alipay",
		"sort": 1,
	}
	createBytes, _ := json.Marshal(createBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/payments", bytes.NewReader(createBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	data := createResp["data"].(map[string]interface{})
	chID := data["id"].(float64)

	updateBody := map[string]interface{}{
		"id":       chID,
		"name":     "更新支付名",
		"icon":     "💳",
		"config":   `{"key":"val"}`,
		"fee_rate": 0.02,
		"status":   1,
		"sort":     2,
	}
	updateBytes, _ := json.Marshal(updateBody)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/payments", bytes.NewReader(updateBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	respData := resp["data"].(map[string]interface{})
	assert.Equal(t, "更新支付名", respData["name"])
}

func TestAdminPaymentDelete(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	createBody := map[string]interface{}{
		"name": "待删除支付",
		"type": "wechat",
		"sort": 1,
	}
	createBytes, _ := json.Marshal(createBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/payments", bytes.NewReader(createBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	data := createResp["data"].(map[string]interface{})
	chID := data["id"].(float64)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/payments/"+itoa(uint(chID)), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
}

func TestAdminEmailList(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/emails", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data["items"])
	assert.NotNil(t, data["total"])
}

func TestAdminEmailCreate(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	body := map[string]interface{}{
		"smtp_host": "smtp.test.com",
		"smtp_port": 465,
		"username":  "test@test.com",
		"password":  "secret123",
		"sender":    "test@test.com",
		"use_ssl":   true,
		"status":    1,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/emails", bytes.NewReader(bodyBytes))
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
	assert.Equal(t, "smtp.test.com", data["smtp_host"])
	assert.NotNil(t, data["id"])
}

func TestAdminEmailUpdate(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	createBody := map[string]interface{}{
		"smtp_host": "smtp.test.com",
		"smtp_port": 465,
		"username":  "test@test.com",
		"password":  "secret123",
		"sender":    "test@test.com",
		"use_ssl":   true,
		"status":    1,
	}
	createBytes, _ := json.Marshal(createBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/emails", bytes.NewReader(createBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	data := createResp["data"].(map[string]interface{})
	emailID := data["id"].(float64)

	updateBody := map[string]interface{}{
		"id":         emailID,
		"smtp_host":  "smtp-new.test.com",
		"smtp_port":  587,
		"username":   "new@test.com",
		"password":   "newpass",
		"sender":     "new@test.com",
		"use_ssl":    false,
		"status":     1,
	}
	updateBytes, _ := json.Marshal(updateBody)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/emails", bytes.NewReader(updateBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	respData := resp["data"].(map[string]interface{})
	assert.Equal(t, "smtp-new.test.com", respData["smtp_host"])
}

func TestAdminEmailDelete(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	createBody := map[string]interface{}{
		"smtp_host": "smtp.test.com",
		"smtp_port": 465,
		"username":  "test@test.com",
		"password":  "secret123",
		"sender":    "test@test.com",
		"use_ssl":   true,
		"status":    1,
	}
	createBytes, _ := json.Marshal(createBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/emails", bytes.NewReader(createBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	data := createResp["data"].(map[string]interface{})
	emailID := data["id"].(float64)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/emails/"+itoa(uint(emailID)), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
}

func TestAdminEmailTest(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/emails/test/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
	assert.Equal(t, "配置不存在", resp["message"])
}

func TestAdminEmailLogs(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/emails/logs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data["items"])
	assert.NotNil(t, data["total"])
}

func TestAdminNodeList(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/nodes", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data["items"])
	assert.NotNil(t, data["total"])
}

func TestAdminNodeCreate(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	body := map[string]interface{}{
		"name":   "节点A",
		"url":    "https://node-a.example.com",
		"weight": 10,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/nodes", bytes.NewReader(bodyBytes))
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
	assert.Equal(t, "节点A", data["name"])
	assert.NotNil(t, data["id"])
}

func TestAdminNodeUpdate(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	createBody := map[string]interface{}{
		"name":   "原始节点",
		"url":    "https://node.example.com",
		"weight": 5,
	}
	createBytes, _ := json.Marshal(createBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/nodes", bytes.NewReader(createBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	data := createResp["data"].(map[string]interface{})
	nodeID := data["id"].(float64)

	updateBody := map[string]interface{}{
		"id":     nodeID,
		"name":   "更新节点",
		"url":    "https://node-new.example.com",
		"weight": 20,
		"status": 1,
	}
	updateBytes, _ := json.Marshal(updateBody)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/nodes", bytes.NewReader(updateBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	respData := resp["data"].(map[string]interface{})
	assert.Equal(t, "更新节点", respData["name"])
}

func TestAdminNodeDelete(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	createBody := map[string]interface{}{
		"name":   "待删除节点",
		"url":    "https://node-del.example.com",
		"weight": 1,
	}
	createBytes, _ := json.Marshal(createBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/nodes", bytes.NewReader(createBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	var createResp map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	data := createResp["data"].(map[string]interface{})
	nodeID := data["id"].(float64)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/nodes/"+itoa(uint(nodeID)), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
}

func TestAdminNodePing(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/nodes/ping/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
	assert.Equal(t, "节点不存在", resp["message"])
}

func TestAdminGetSettings(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data)
}

func TestAdminGetVersion(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/upgrade/version", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data["version"])
	assert.NotNil(t, data["name"])
}

func TestAdminUpgradeLogs(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/upgrade/logs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data["logs"])
	assert.NotNil(t, data["total"])
}

func TestAdminUploadFile(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("hello world from test file"))
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data["id"])
	assert.NotNil(t, data["url"])
	assert.NotNil(t, data["size"])
	assert.Equal(t, "file", data["type"])
}

func TestAdminUploadFileUnsupportedType(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.exe")
	require.NoError(t, err)
	_, err = part.Write([]byte("binary content"))
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
	assert.Equal(t, "不支持的文件类型", resp["message"])
}

func TestAdminListFiles(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/upload", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data["files"])
	assert.NotNil(t, data["total"])
}

func TestAdminDeleteFile(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "delete_me.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("to be deleted"))
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	uploadReq := httptest.NewRequest(http.MethodPost, "/api/admin/upload", body)
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())
	uploadReq.Header.Set("Authorization", "Bearer "+token)
	uploadW := httptest.NewRecorder()
	r.ServeHTTP(uploadW, uploadReq)

	var uploadResp map[string]interface{}
	json.Unmarshal(uploadW.Body.Bytes(), &uploadResp)
	uploadData := uploadResp["data"].(map[string]interface{})
	fileID := uploadData["id"].(float64)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/upload/"+itoa(uint(fileID)), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])
}

func TestAdminUploadFileImage(t *testing.T) {
	r, token := setupAdminSupplementTestServer()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.png")
	require.NoError(t, err)
	_, err = part.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "image", data["type"])
}