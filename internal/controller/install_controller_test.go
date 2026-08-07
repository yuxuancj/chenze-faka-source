package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupInstallTestServer() *gin.Engine {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: ":memory:",
	}
	database.Init(cfg)
	database.AutoMigrate()

	authCtrl := NewAuthController("test-secret", 72, "Test Site", &model.LicenseConfig{}, cfg)

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		install := api.Group("/install")
		{
			install.POST("", authCtrl.Install)
			install.POST("/test-database", authCtrl.TestDatabase)
			install.GET("/license", authCtrl.GetLicenseStatus)
		}
	}

	return r
}

func TestInstallFullFlow(t *testing.T) {
	r := setupInstallTestServer()

	os.Remove("install.lock")

	body := map[string]interface{}{
		"site_name": "测试站点",
		"license_key": "test-license-key-12345",
		"database": map[string]interface{}{
			"driver":  "sqlite",
			"sqlite_path": ":memory:",
		},
		"jwt": map[string]interface{}{
			"secret":      "test-secret",
			"expire_time": 72,
		},
		"username": "admin",
		"password": "admin123",
		"force":    true,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])

	os.Remove("install.lock")
	os.Remove("config.yaml")
}

func TestInstallSkip(t *testing.T) {
	r := setupInstallTestServer()

	os.Remove("install.lock")

	body := map[string]interface{}{
		"skip": true,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	os.Remove("install.lock")
}

func TestInstallWithoutLicenseKey(t *testing.T) {
	r := setupInstallTestServer()

	body := map[string]interface{}{
		"site_name": "测试站点",
		"database": map[string]interface{}{
			"driver":  "sqlite",
			"sqlite_path": ":memory:",
		},
		"username": "admin",
		"password": "admin123",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
	assert.Contains(t, resp["message"], "授权密钥")
}

func TestInstallWithoutUsernamePassword(t *testing.T) {
	r := setupInstallTestServer()

	body := map[string]interface{}{
		"site_name":   "测试站点",
		"license_key": "test-license-key",
		"database": map[string]interface{}{
			"driver":  "sqlite",
			"sqlite_path": ":memory:",
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
	assert.Contains(t, resp["message"], "管理员用户名和密码")
}

func TestInstallTestDatabaseSqlite(t *testing.T) {
	r := setupInstallTestServer()

	body := map[string]interface{}{
		"host":     "localhost",
		"port":     3306,
		"database": "test_db",
		"username": "root",
		"password": "",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/install/test-database", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
}

func TestInstallInvalidBody(t *testing.T) {
	r := setupInstallTestServer()

	req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
}

func TestGetLicenseStatus(t *testing.T) {
	r := setupInstallTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/install/license", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, false, data["installed"])
}
