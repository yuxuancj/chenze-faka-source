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
	"chenze-faka/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthSupplementTestDB(t *testing.T) *gorm.DB {
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

func writeLicenseCacheFile(t *testing.T) {
	t.Helper()
	cache := map[string]interface{}{
		"last_verify_time":  time.Now().Unix(),
		"last_success_time": time.Now().Unix(),
		"last_result":       "pass",
		"last_message":      "ok",
		"expire_at":         "2099-12-31",
		"app_name":          "TestApp",
	}
	data, err := json.Marshal(cache)
	require.NoError(t, err)
	err = os.WriteFile("license.cache", data, 0644)
	require.NoError(t, err)
}

func TestCheckMysqlVersionStatusEnDirect(t *testing.T) {
	t.Run("mysql 8.0 returns normal", func(t *testing.T) {
		assert.Equal(t, "normal", checkMysqlVersionStatusEn("8.0.32"))
	})
	t.Run("mysql 5.7 returns normal", func(t *testing.T) {
		assert.Equal(t, "normal", checkMysqlVersionStatusEn("5.7.44"))
	})
	t.Run("mysql 5.6 returns error", func(t *testing.T) {
		assert.Equal(t, "error", checkMysqlVersionStatusEn("5.6.51"))
	})
	t.Run("mysql 10.0 returns normal", func(t *testing.T) {
		assert.Equal(t, "normal", checkMysqlVersionStatusEn("10.0.1"))
	})
	t.Run("empty string returns error", func(t *testing.T) {
		assert.Equal(t, "error", checkMysqlVersionStatusEn(""))
	})
	t.Run("no match returns error", func(t *testing.T) {
		assert.Equal(t, "error", checkMysqlVersionStatusEn("unknown-version"))
	})
	t.Run("mariadb 10.11 returns normal", func(t *testing.T) {
		assert.Equal(t, "normal", checkMysqlVersionStatusEn("10.11.2-MariaDB"))
	})
}

func TestCheckMysqlVersionStatusDirect(t *testing.T) {
	t.Run("mysql 8.0 returns 正常", func(t *testing.T) {
		assert.Equal(t, "正常", checkMysqlVersionStatus("8.0.32"))
	})
	t.Run("mysql 5.7 returns 正常", func(t *testing.T) {
		assert.Equal(t, "正常", checkMysqlVersionStatus("5.7.44"))
	})
	t.Run("mysql 5.6 returns 异常", func(t *testing.T) {
		assert.Equal(t, "异常", checkMysqlVersionStatus("5.6.51"))
	})
	t.Run("empty string returns 异常", func(t *testing.T) {
		assert.Equal(t, "异常", checkMysqlVersionStatus(""))
	})
	t.Run("no match returns 异常", func(t *testing.T) {
		assert.Equal(t, "异常", checkMysqlVersionStatus("garbage"))
	})
}

func TestGenerateSaltDirect(t *testing.T) {
	t.Run("generateSalt returns 32 char hex string", func(t *testing.T) {
		salt := generateSalt()
		assert.Len(t, salt, 32)
	})
	t.Run("generateSalt returns different values each time", func(t *testing.T) {
		salt1 := generateSalt()
		salt2 := generateSalt()
		assert.NotEqual(t, salt1, salt2)
	})
}

func TestHashPasswordDirect(t *testing.T) {
	t.Run("hashPassword returns consistent 64 char hex", func(t *testing.T) {
		hash := hashPassword("password", "salt")
		assert.Len(t, hash, 64)
		hash2 := hashPassword("password", "salt")
		assert.Equal(t, hash, hash2)
	})
	t.Run("hashPassword with different password returns different hash", func(t *testing.T) {
		hash1 := hashPassword("pass1", "salt")
		hash2 := hashPassword("pass2", "salt")
		assert.NotEqual(t, hash1, hash2)
	})
	t.Run("hashPassword with different salt returns different hash", func(t *testing.T) {
		hash1 := hashPassword("password", "salt1")
		hash2 := hashPassword("password", "salt2")
		assert.NotEqual(t, hash1, hash2)
	})
}

func TestSaveConfigDirect(t *testing.T) {
	t.Run("saveConfig writes file successfully", func(t *testing.T) {
		cfg := &model.Config{}
		cfg.System.SiteName = "test"
		cfg.System.Port = 8080
		err := saveConfig("test_config.yaml", cfg)
		require.NoError(t, err)

		data, err := os.ReadFile("test_config.yaml")
		require.NoError(t, err)
		assert.Contains(t, string(data), "site_name: test")
		assert.Contains(t, string(data), "port: 8080")
		os.Remove("test_config.yaml")
	})
}

func TestYamlMarshalDirect(t *testing.T) {
	t.Run("yamlMarshal marshals struct correctly", func(t *testing.T) {
		type Simple struct {
			Name  string `yaml:"name"`
			Value int    `yaml:"value"`
		}
		data, err := yamlMarshal(&Simple{Name: "hello", Value: 42})
		require.NoError(t, err)
		assert.Contains(t, string(data), "name: hello")
		assert.Contains(t, string(data), "value: 42")
	})
}

func setupProfileTestServer(secret string) *gin.Engine {
	authMw := middleware.NewAuthMiddleware(secret, 72)
	r := gin.New()
	r.Use(gin.Recovery())
	auth := r.Group("/api/auth")
	{
		auth.GET("/profile", authMw.AuthRequired(), func(c *gin.Context) {
			userObj, exists := c.Get("user")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "message": "未登录"})
				return
			}
			u := userObj.(*model.User)
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"data": gin.H{
					"id":       u.ID,
					"username": u.Username,
					"role":     u.Role,
				},
			})
		})
	}
	return r
}

func TestGetProfileNoUser(t *testing.T) {
	setupAuthSupplementTestDB(t)
	defer func() { database.DB = nil }()

	r := setupProfileTestServer("test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/auth/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetProfileWithUser(t *testing.T) {
	setupAuthSupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := service.NewAuthService()
	user, err := svc.Register("profile_user_supp", "profile_pass")
	require.NoError(t, err)

	secret := "test-secret"
	token, _, err := svc.Login("profile_user_supp", "profile_pass", secret, 72)
	require.NoError(t, err)

	r := setupProfileTestServer(secret)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "profile_user_supp", data["username"])
	assert.Equal(t, user.ID, uint(data["id"].(float64)))
}

func TestInstallSuccessWithCache(t *testing.T) {
	setupAuthSupplementTestDB(t)
	defer func() { database.DB = nil }()

	writeLicenseCacheFile(t)
	defer os.Remove("license.cache")
	defer os.Remove("install.lock")
	defer os.Remove("config.yaml")

	dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
	r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

	body := map[string]interface{}{
		"site_name":   "Cache Install Site",
		"license_key": "test-license-cache",
		"database": map[string]interface{}{
			"driver":      "sqlite",
			"sqlite_path": ":memory:",
		},
		"jwt": map[string]interface{}{
			"secret":      "test-secret",
			"expire_time": 72,
		},
		"username": "cacheadmin",
		"password": "cachepass123",
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
	_, err = os.Stat("config.yaml")
	assert.NoError(t, err)

	var userCount int64
	database.DB.Model(&model.User{}).Count(&userCount)
	assert.Equal(t, int64(1), userCount)
}

func TestInstallAlreadyInstalledWithCache(t *testing.T) {
	setupAuthSupplementTestDB(t)
	defer func() { database.DB = nil }()

	writeLicenseCacheFile(t)
	defer os.Remove("license.cache")
	defer os.Remove("install.lock")
	defer os.Remove("config.yaml")

	svc := service.NewAuthService()
	svc.Register("existing_user_cache", "existing_pass")

	dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
	r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

	body := map[string]interface{}{
		"site_name":   "Already Installed Cache Site",
		"license_key": "test-license-already-cache",
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
	t.Logf("Install response: %v", resp)
}

func TestInstallForceWithCache(t *testing.T) {
	setupAuthSupplementTestDB(t)
	defer func() { database.DB = nil }()

	writeLicenseCacheFile(t)
	defer os.Remove("license.cache")
	defer os.Remove("install.lock")
	defer os.Remove("config.yaml")

	svc := service.NewAuthService()
	svc.Register("preexisting_user", "preexisting_pass")

	dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
	r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

	body := map[string]interface{}{
		"site_name":   "Force Cache Site",
		"license_key": "test-license-force-cache",
		"database": map[string]interface{}{
			"driver":      "sqlite",
			"sqlite_path": ":memory:",
		},
		"jwt": map[string]interface{}{
			"secret":      "test-secret",
			"expire_time": 72,
		},
		"username": "forceadmin2",
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
	assert.Equal(t, float64(0), resp["code"])

	var userCount int64
	database.DB.Model(&model.User{}).Count(&userCount)
	assert.Equal(t, int64(1), userCount)
}

func TestInstallDefaultJWTConfig(t *testing.T) {
	setupAuthSupplementTestDB(t)
	defer func() { database.DB = nil }()

	writeLicenseCacheFile(t)
	defer os.Remove("license.cache")
	defer os.Remove("install.lock")
	defer os.Remove("config.yaml")

	dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
	r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

	body := map[string]interface{}{
		"site_name":   "Default JWT Site",
		"license_key": "test-license-default-jwt",
		"database": map[string]interface{}{
			"driver":      "sqlite",
			"sqlite_path": ":memory:",
		},
		"username": "defaultjwtadmin",
		"password": "defaultjwtpass",
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

	data, err := os.ReadFile("config.yaml")
	require.NoError(t, err)
	assert.Contains(t, string(data), "chenze-faka-secret")
}

func TestLoginWithCaptchaSuccess(t *testing.T) {
	setupAuthSupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := service.NewAuthService()
	svc.Register("captcha_user", "captcha_pass")

	dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
	r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

	captchaReq := httptest.NewRequest(http.MethodGet, "/api/auth/captcha", nil)
	captchaW := httptest.NewRecorder()
	r.ServeHTTP(captchaW, captchaReq)
	var captchaResp map[string]interface{}
	json.Unmarshal(captchaW.Body.Bytes(), &captchaResp)
	captchaData := captchaResp["data"].(map[string]interface{})
	captchaID := captchaData["id"].(string)

	body := map[string]interface{}{
		"username":  "captcha_user",
		"password":  "captcha_pass",
		"captcha_id": captchaID,
		"captcha":   "wrong_captcha",
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
	assert.Contains(t, resp["message"], "验证码错误")
}

func TestVerifyLicenseQuickVerifyError(t *testing.T) {
	setupAuthSupplementTestDB(t)
	defer func() { database.DB = nil }()

	r := setupAuthCoverageTestServer(nil, &model.LicenseConfig{
		Enabled:     true,
		BaseURL:     "http://127.0.0.1:19999",
		AppKey:      "test",
		AppSecret:   "test",
		LicenseKey:  "test-key",
		CacheFile:   "nonexistent_license.cache",
		GracePeriod: 1,
	})

	body := map[string]interface{}{
		"license_key": "some-key",
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
}

func TestGetProfileCoverageSupplement(t *testing.T) {
	setupAuthSupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := service.NewAuthService()
	user, err := svc.Register("profile_user", "profile_pass")
	require.NoError(t, err)

	t.Run("get profile without user in context returns unauthorized", func(t *testing.T) {
		r := gin.New()
		handler := NewAuthController("secret", 72, "TestSite", &model.LicenseConfig{}, nil).GetProfile
		r.GET("/test", handler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("get profile with valid user returns success", func(t *testing.T) {
		r2 := gin.New()
		handler := NewAuthController("secret", 72, "TestSite", &model.LicenseConfig{}, nil).GetProfile
		r2.GET("/test", func(c *gin.Context) {
			c.Set("user", user)
			handler(c)
		})
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r2.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, user.ID, uint(data["id"].(float64)))
		assert.Equal(t, "profile_user", data["username"])
	})
}

func TestInstallInvalidJSON(t *testing.T) {
	setupAuthSupplementTestDB(t)
	defer func() { database.DB = nil }()

	dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
	r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

	req := httptest.NewRequest(http.MethodPost, "/api/install", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(1), resp["code"])
}

func TestInstallSkipSuccess(t *testing.T) {
	setupAuthSupplementTestDB(t)
	defer func() { database.DB = nil }()

	os.Remove("install.lock")
	defer os.Remove("install.lock")

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
}

func TestVerifyLicenseNilCfg(t *testing.T) {
	setupAuthSupplementTestDB(t)
	defer func() { database.DB = nil }()

	r := setupAuthCoverageTestServer(nil, nil)

	body := map[string]interface{}{
		"license_key": "test-key-nil-cfg",
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
}

func TestVerifyLicenseInvalidJSON(t *testing.T) {
	setupAuthSupplementTestDB(t)
	defer func() { database.DB = nil }()

	r := setupAuthCoverageTestServer(nil, &model.LicenseConfig{})

	req := httptest.NewRequest(http.MethodPost, "/api/license/verify", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(1), resp["code"])
}

func TestLoginCaptchaNotVerified(t *testing.T) {
	setupAuthSupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := service.NewAuthService()
	svc.Register("captcha_user2", "captcha_pass2")

	dbCfg := &model.DatabaseConfig{Driver: "sqlite", SQLite: ":memory:"}
	r := setupAuthCoverageTestServer(dbCfg, &model.LicenseConfig{})

	body := map[string]interface{}{
		"username":  "captcha_user2",
		"password":  "captcha_pass2",
		"captcha_id": "fake-id",
		"captcha":   "fake-captcha",
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
	assert.Contains(t, resp["message"], "验证码错误")
}

func TestCheckEnvMySQLVersionPaths(t *testing.T) {
	setupAuthSupplementTestDB(t)
	defer func() { database.DB = nil }()

	t.Run("check mysql version 5.7.30 returns normal", func(t *testing.T) {
		status := checkMysqlVersionStatusEn("5.7.30-log")
		assert.Equal(t, "normal", status)
	})

	t.Run("check mysql version 5.6 returns error", func(t *testing.T) {
		status := checkMysqlVersionStatusEn("5.6.0")
		assert.Equal(t, "error", status)
	})

	t.Run("check mysql version 5.7.0 returns normal", func(t *testing.T) {
		status := checkMysqlVersionStatusEn("5.7.0")
		assert.Equal(t, "normal", status)
	})

	t.Run("check mariadb 10.5 returns normal", func(t *testing.T) {
		status := checkMysqlVersionStatusEn("10.5.12-MariaDB")
		assert.Equal(t, "normal", status)
	})
}