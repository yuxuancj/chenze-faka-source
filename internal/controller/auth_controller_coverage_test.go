package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"chenze-faka/internal/middleware"
	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
	"chenze-faka/internal/pkg/utils"
	"chenze-faka/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func dropAllControllerCoverageTables(db *gorm.DB) {
	tables := []string{"users", "categories", "products", "cards", "orders", "payment_channels", "email_configs", "nodes", "operation_logs", "order_logs", "email_logs", "upgrade_logs", "file_uploads"}
	for _, t := range tables {
		db.Migrator().DropTable(t)
	}
}

func setupAuthCoverageTestDB(t *testing.T) *gorm.DB {
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

func setupAuthCoverageTestServer(dbCfg *model.DatabaseConfig, licenseCfg *model.LicenseConfig) *gin.Engine {
	authCtrl := NewAuthController("test-secret", 72, "Coverage Test Site", licenseCfg, dbCfg)

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		install := api.Group("/install")
		{
			install.GET("/env", authCtrl.CheckEnv)
			install.POST("/test-database", authCtrl.TestDatabase)
			install.GET("/license", authCtrl.GetLicenseStatus)
			install.POST("", authCtrl.Install)
		}
		auth := api.Group("/auth")
		{
			auth.POST("/login", authCtrl.Login)
			auth.POST("/register", authCtrl.Register)
			auth.GET("/captcha", authCtrl.GetCaptcha)
		}
		site := api.Group("/site")
		{
			site.GET("/config", authCtrl.GetSiteConfig)
		}
		license := api.Group("/license")
		{
			license.POST("/verify", authCtrl.VerifyLicense)
		}
	}

	return r
}

func TestCheckEnvCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	t.Run("sqlite driver returns normal and 3.x", func(t *testing.T) {
		dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

		req := httptest.NewRequest(http.MethodGet, "/api/install/env", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "normal", data["mysql_status"])
		assert.Equal(t, "SQLite 3.x", data["mysql_version"])
	})

	t.Run("nil dbCfg returns not-installed", func(t *testing.T) {
		r := setupAuthCoverageTestServer(nil, &model.LicenseConfig{})

		req := httptest.NewRequest(http.MethodGet, "/api/install/env", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "not-installed", data["mysql_status"])
	})

	t.Run("mysql driver fails connection", func(t *testing.T) {
		dbCfg := &model.DatabaseConfig{
			Driver:   "mysql",
			Host:     "127.0.0.1",
			Port:     19999,
			User:     "test",
			Password: "test",
			DBName:   "test",
		}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

		req := httptest.NewRequest(http.MethodGet, "/api/install/env", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "not-installed", data["mysql_status"])
	})
}

func TestInstallCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	t.Run("install with force=true succeeds", func(t *testing.T) {
		os.Remove("install.lock")
		os.Remove("config.yaml")

		dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

		body := map[string]interface{}{
			"site_name":   "Force Install Site",
			"license_key": "test-license-key-force",
			"database": map[string]interface{}{
				"driver":      "sqlite",
				"sqlite_path": ":memory:",
			},
			"jwt": map[string]interface{}{
				"secret":      "test-secret",
				"expire_time": 72,
			},
			"username": "forceadmin",
			"password": "forcepass123",
			"force":    true,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])

		os.Remove("install.lock")
		os.Remove("config.yaml")
	})

	t.Run("install with skip=true creates lock file", func(t *testing.T) {
		os.Remove("install.lock")

		dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

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
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])

		_, err := os.Stat("install.lock")
		assert.NoError(t, err)
		os.Remove("install.lock")
	})

	t.Run("install missing license key fails", func(t *testing.T) {
		dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

		body := map[string]interface{}{
			"site_name": "No License Site",
			"database": map[string]interface{}{
				"driver":      "sqlite",
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
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
		assert.Contains(t, resp["message"], "授权密钥")
	})

	t.Run("install missing username fails", func(t *testing.T) {
		dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

		body := map[string]interface{}{
			"site_name":   "No User Site",
			"license_key": "test-key",
			"database": map[string]interface{}{
				"driver":      "sqlite",
				"sqlite_path": ":memory:",
			},
			"password": "admin123",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
		assert.Contains(t, resp["message"], "管理员用户名")
	})

	t.Run("install missing password fails", func(t *testing.T) {
		dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

		body := map[string]interface{}{
			"site_name":   "No Pass Site",
			"license_key": "test-key",
			"database": map[string]interface{}{
				"driver":      "sqlite",
				"sqlite_path": ":memory:",
			},
			"username": "admin",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
		assert.Contains(t, resp["message"], "管理员用户名")
	})
}

func TestTestDatabaseCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	t.Run("test database with sqlite succeeds", func(t *testing.T) {
		dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

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
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})

	t.Run("test database with empty db name fails", func(t *testing.T) {
		dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

		body := map[string]interface{}{
			"host":     "localhost",
			"port":     3306,
			"database": "",
			"username": "root",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/install/test-database", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
		assert.Contains(t, resp["message"], "数据库名不能为空")
	})

	t.Run("test database with empty username fails", func(t *testing.T) {
		dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

		body := map[string]interface{}{
			"host":     "localhost",
			"port":     3306,
			"database": "test_db",
			"username": "",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/install/test-database", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
		assert.Contains(t, resp["message"], "用户名不能为空")
	})

	t.Run("test database with invalid mysql fails", func(t *testing.T) {
		dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

		body := map[string]interface{}{
			"host":     "127.0.0.1",
			"port":     19999,
			"database": "nonexistent_db",
			"username": "root",
			"password": "wrongpass",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/install/test-database", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})
}

func TestGetLicenseStatusCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	t.Run("no install lock returns not installed", func(t *testing.T) {
		os.Remove("install.lock")
		dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

		req := httptest.NewRequest(http.MethodGet, "/api/install/license", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, false, data["installed"])
		assert.Equal(t, "Coverage Test Site", data["site_name"])
	})

	t.Run("with install lock returns installed", func(t *testing.T) {
		err := os.WriteFile("install.lock", []byte("installed"), 0644)
		require.NoError(t, err)
		defer os.Remove("install.lock")

		dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

		req := httptest.NewRequest(http.MethodGet, "/api/install/license", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, true, data["installed"])
	})
}

func TestGetSiteConfigCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
	r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/site/config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Coverage Test Site", data["site_name"])
}

func TestGetCaptchaCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
	r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/auth/captcha", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	data := resp["data"].(map[string]interface{})
	assert.NotEmpty(t, data["id"])
	assert.NotEmpty(t, data["img"])
}

func TestVerifyLicenseCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	t.Run("empty license key fails", func(t *testing.T) {
		dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

		body := map[string]interface{}{
			"license_key": "",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/license/verify", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
		assert.Contains(t, resp["message"], "授权密钥不能为空")
	})

	t.Run("invalid license key with no base URL still returns success", func(t *testing.T) {
		dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

		body := map[string]interface{}{
			"license_key": "invalid-key-12345",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/license/verify", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})
}

func TestRegisterCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
	r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

	t.Run("register with valid data", func(t *testing.T) {
		body := map[string]interface{}{
			"username": "coverage_user",
			"password": "coverage_pass",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "coverage_user", data["username"])
	})

	t.Run("register with empty username fails", func(t *testing.T) {
		body := map[string]interface{}{
			"username": "",
			"password": "coverage_pass",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})

	t.Run("register with empty password fails", func(t *testing.T) {
		body := map[string]interface{}{
			"username": "user_no_pass",
			"password": "",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})

	t.Run("register duplicate user fails", func(t *testing.T) {
		body := map[string]interface{}{
			"username": "coverage_user",
			"password": "new_pass",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
		assert.Contains(t, resp["message"], "已存在")
	})

	t.Run("register with invalid body fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte("not-json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})
}

func TestLoginCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
	r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

	service.NewAuthService().Register("coverage_login_user", "coverage_login_pass")

	t.Run("login with correct credentials succeeds", func(t *testing.T) {
		body := map[string]interface{}{
			"username": "coverage_login_user",
			"password": "coverage_login_pass",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.NotEmpty(t, data["token"])
		user := data["user"].(map[string]interface{})
		assert.Equal(t, "coverage_login_user", user["username"])
	})

	t.Run("login with wrong password fails", func(t *testing.T) {
		body := map[string]interface{}{
			"username": "coverage_login_user",
			"password": "wrong_password",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})

	t.Run("login with nonexistent user fails", func(t *testing.T) {
		body := map[string]interface{}{
			"username": "ghost_user",
			"password": "somepass",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})

	t.Run("login with invalid body fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})
}

func TestLoginWithExpiredTokenCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	salt := utils.GenerateSalt()
	hash := utils.HashPassword("testpass", salt)
	user := &model.User{
		Username:     "expiry_user",
		PasswordHash: hash,
		Salt:         salt,
		Role:         model.RoleAdmin,
	}
	require.NoError(t, database.DB.Create(user).Error)

	secret := "test-secret"
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(-1 * time.Hour).Unix(),
		"iat":      time.Now().Add(-2 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	authMw := middleware.NewAuthMiddleware(secret, 72)

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/protected", authMw.AuthRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLoginWithTamperedTokenCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	secret := "test-secret"
	claims := jwt.MapClaims{
		"user_id":  float64(1),
		"username": "tampered",
		"role":     model.RoleAdmin,
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	tampered := tokenString + "tampered"

	authMw := middleware.NewAuthMiddleware(secret, 72)

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/protected", authMw.AuthRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tampered)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCheckEnvMySQLCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	t.Run("mysql driver without connection returns not-installed", func(t *testing.T) {
		dbCfg := &model.DatabaseConfig{
			Driver:   "mysql",
			Host:     "127.0.0.1",
			Port:     3306,
			User:     "root",
			Password: "wrong_password",
			DBName:   "nonexistent",
		}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})
		req := httptest.NewRequest(http.MethodGet, "/api/install/env", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "not-installed", data["mysql_status"])
	})

	t.Run("sqlite driver success path", func(t *testing.T) {
		dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})
		req := httptest.NewRequest(http.MethodGet, "/api/install/env", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "SQLite 3.x", data["mysql_version"])
	})
}

func TestCheckEnvInvalidDriverCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	t.Run("invalid driver defaults to not-installed", func(t *testing.T) {
		dbCfg := &model.DatabaseConfig{Driver: "postgres"}
		r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})
		req := httptest.NewRequest(http.MethodGet, "/api/install/env", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "not-installed", data["mysql_status"])
	})
}

func TestLoginEdgeCasesCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
	r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

	svc := service.NewAuthService()
	_, err := svc.Register("edge_login_user", "correct_pass")
	require.NoError(t, err)

	t.Run("login with wrong password returns error code", func(t *testing.T) {
		body := map[string]interface{}{
			"username": "edge_login_user",
			"password": "wrong_pass",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})

	t.Run("login with non-existent user returns error code", func(t *testing.T) {
		body := map[string]interface{}{
			"username": "no_such_user_xyz",
			"password": "whatever",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})

	t.Run("login with empty credentials returns error", func(t *testing.T) {
		body := map[string]interface{}{
			"username": "",
			"password": "",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})
}

func TestVerifyLicenseExpiredCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	r := setupAuthCoverageTestServer(nil, &model.LicenseConfig{
		LicenseKey: "EXP-LICENSE-KEY",
	})

	t.Run("verify license with expired license returns not verified", func(t *testing.T) {
		body := map[string]interface{}{
			"license_key": "EXP-LICENSE-KEY",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/license/verify", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, false, data["verified"])
	})
}

func TestTestDatabaseMysqlFailCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	r := setupAuthCoverageTestServer(nil, &model.LicenseConfig{})

	t.Run("test database with invalid mysql config fails", func(t *testing.T) {
		body := map[string]interface{}{
			"host":     "127.0.0.1",
			"port":     3306,
			"username": "root",
			"password": "invalid",
			"database": "nonexistent",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/install/test-database", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})
}

func TestGetProfileCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := service.NewAuthService()
	user, err := svc.Register("profile_user", "profile_pass")
	require.NoError(t, err)

	secret := "test-secret"
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/api/auth/profile", middleware.NewAuthMiddleware(secret, 72).AuthRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInstallSkipCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	os.Remove("install.lock")
	defer os.Remove("install.lock")

	dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
	r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

	t.Run("install with skip=true creates install.lock", func(t *testing.T) {
		body := map[string]interface{}{
			"site_name":   "Skip Install Site",
			"license_key": "test-license-skip",
			"database": map[string]interface{}{
				"driver":      "sqlite",
				"sqlite_path": ":memory:",
			},
			"jwt": map[string]interface{}{
				"secret":      "test-secret",
				"expire_time": 72,
			},
			"username": "skipadmin",
			"password": "skippass123",
			"skip":     true,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		_, err := os.Stat("install.lock")
		assert.NoError(t, err)
	})
}

func TestInstallMissingFieldsCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
	r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

	t.Run("install with missing license key fails", func(t *testing.T) {
		body := map[string]interface{}{
			"site_name": "No License Site",
			"database": map[string]interface{}{
				"driver":      "sqlite",
				"sqlite_path": ":memory:",
			},
			"jwt": map[string]interface{}{
				"secret":      "test-secret",
				"expire_time": 72,
			},
			"username": "admin",
			"password": "pass123",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})

	t.Run("install with missing username fails", func(t *testing.T) {
		body := map[string]interface{}{
			"site_name":   "No User Site",
			"license_key": "test-license-no-user",
			"database": map[string]interface{}{
				"driver":      "sqlite",
				"sqlite_path": ":memory:",
			},
			"jwt": map[string]interface{}{
				"secret":      "test-secret",
				"expire_time": 72,
			},
			"password": "pass123",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})

	t.Run("install with missing password fails", func(t *testing.T) {
		body := map[string]interface{}{
			"site_name":   "No Pass Site",
			"license_key": "test-license-no-pass",
			"database": map[string]interface{}{
				"driver":      "sqlite",
				"sqlite_path": ":memory:",
			},
			"jwt": map[string]interface{}{
				"secret":      "test-secret",
				"expire_time": 72,
			},
			"username": "admin",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})
}

func TestInstallAlreadyInstalledCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := service.NewAuthService()
	svc.Register("existing_admin", "existing_pass")

	dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
	r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

	t.Run("install when already installed fails", func(t *testing.T) {
		body := map[string]interface{}{
			"site_name":   "Already Installed Site",
			"license_key": "test-license-already",
			"database": map[string]interface{}{
				"driver":      "sqlite",
				"sqlite_path": ":memory:",
			},
			"jwt": map[string]interface{}{
				"secret":      "test-secret",
				"expire_time": 72,
			},
			"username": "newadmin",
			"password": "newpass123",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["code"])
	})
}

func TestBuildPayURLCoverage(t *testing.T) {
	orderCtrl := NewOrderController(model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	})

	t.Run("build pay URL with valid config returns URL", func(t *testing.T) {
		url := orderCtrl.BuildPayURL("ORDER123", 10.00, "alipay", "Test Product")
		assert.Contains(t, url, "https://pay.example.com")
		assert.Contains(t, url, "out_trade_no=ORDER123")
		assert.Contains(t, url, "money=10.00")
		assert.Contains(t, url, "sign=")
	})

	t.Run("build pay URL with empty config returns empty", func(t *testing.T) {
		ctrl := NewOrderController(model.PayConfig{})
		url := ctrl.BuildPayURL("ORDER123", 10.00, "alipay", "Test Product")
		assert.Equal(t, "", url)
	})

	t.Run("build pay URL with partial config returns empty", func(t *testing.T) {
		ctrl := NewOrderController(model.PayConfig{
			URL: "https://pay.example.com",
		})
		url := ctrl.BuildPayURL("ORDER123", 10.00, "alipay", "Test Product")
		assert.Equal(t, "", url)
	})
}