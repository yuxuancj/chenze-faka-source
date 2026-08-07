package controller

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"chenze-faka/internal/middleware"
	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAdminGapTestServer() (*gin.Engine, string) {
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
			admin.GET("/categories/all", adminCtrl.CategoryAll)

			cards := admin.Group("/cards")
			{
				cards.GET("/export", adminCtrl.CardExport)
			}

			logs := admin.Group("/logs")
			{
				logs.GET("/orders", adminCtrl.OrderLogs)
				logs.GET("/login", adminCtrl.LoginLogs)
			}

			upgrade := admin.Group("/upgrade")
			{
				upgrade.GET("/check", adminCtrl.CheckUpdate)
				upgrade.POST("/upload", adminCtrl.UploadPackage)
				upgrade.POST("/apply", adminCtrl.ApplyUpgrade)
			}

			upload := admin.Group("/upload")
			{
				upload.POST("", adminCtrl.UploadFile)
				upload.GET("/:id", adminCtrl.GetFile)
			}
		}
	}

	adminToken := createAdminUserAndToken("admin_gap", "admin123")
	return r, adminToken
}

func TestAdminCategoryAll(t *testing.T) {
	r, token := setupAdminGapTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/categories/all", nil)
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

func TestAdminCardExport(t *testing.T) {
	r, token := setupAdminGapTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/cards/export", nil)
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

func TestAdminOrderLogs(t *testing.T) {
	r, token := setupAdminGapTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/logs/orders", nil)
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

func TestAdminLoginLogs(t *testing.T) {
	r, token := setupAdminGapTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/logs/login", nil)
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

func TestAdminCheckUpdate(t *testing.T) {
	r, token := setupAdminGapTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/upgrade/check", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotNil(t, data["has_update"])
	assert.NotNil(t, data["current"])
	assert.NotNil(t, data["latest"])
}

func TestAdminUploadPackageError(t *testing.T) {
	r, token := setupAdminGapTestServer()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/upgrade/upload", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
}

func TestAdminApplyUpgradeError(t *testing.T) {
	r, token := setupAdminGapTestServer()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/upgrade/apply", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
	assert.NotEmpty(t, resp["message"])
}

func TestAdminGetFileNotFound(t *testing.T) {
	r, token := setupAdminGapTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/upload/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
	assert.Equal(t, "文件不存在", resp["message"])
}

func TestAdminGetFileSuccess(t *testing.T) {
	r, token := setupAdminGapTestServer()
	t.Cleanup(func() { os.RemoveAll("./uploads") })

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "getfile_test.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("test file content for get"))
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	uploadReq := httptest.NewRequest(http.MethodPost, "/api/admin/upload", body)
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())
	uploadReq.Header.Set("Authorization", "Bearer "+token)
	uploadW := httptest.NewRecorder()
	r.ServeHTTP(uploadW, uploadReq)

	var uploadResp map[string]interface{}
	err = json.Unmarshal(uploadW.Body.Bytes(), &uploadResp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), uploadResp["code"])

	uploadData := uploadResp["data"].(map[string]interface{})
	fileID := uploadData["id"].(float64)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/upload/"+itoa(uint(fileID)), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test file content for get", w.Body.String())
}

func TestAdminGetFileInvalidID(t *testing.T) {
	r, token := setupAdminGapTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/upload/abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
	assert.Equal(t, "无效的ID", resp["message"])
}