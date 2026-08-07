package router

import (
	"chenze-faka/internal/model"
	"chenze-faka/internal/service"
	"embed"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed assets
var testFS embed.FS

func createTestConfig() *model.Config {
	return &model.Config{
		System: model.SystemConfig{
			SiteName: "Test Site",
			Port:     8080,
			Mode:     "debug",
		},
		Database: model.DatabaseConfig{},
		JWT: model.JWTConfig{
			Secret:     "test-secret",
			ExpireTime: 72,
		},
		Pay: model.PayConfig{
			URL:      "https://pay.example.com",
			Merchant: "test-merchant",
			Key:      "test-key",
		},
		License: model.LicenseConfig{
			Enabled: false,
		},
	}
}

func setupInstallLock(t *testing.T) {
	t.Helper()
	err := os.WriteFile("install.lock", []byte("test"), 0644)
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Remove("install.lock")
	})
}

func setupVerifiedLicenseSvc(cfg *model.Config) *service.LicenseService {
	licenseSvc := service.NewLicenseService(&cfg.License)
	licenseSvc.Verify()
	return licenseSvc
}

func TestSetup_InstallRoute(t *testing.T) {
	setupInstallLock(t)

	cfg := createTestConfig()
	licenseSvc := setupVerifiedLicenseSvc(cfg)
	r := Setup(cfg, licenseSvc, testFS)

	req := httptest.NewRequest(http.MethodGet, "/install", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetup_ApiInstallEnv(t *testing.T) {
	setupInstallLock(t)

	cfg := createTestConfig()
	licenseSvc := setupVerifiedLicenseSvc(cfg)
	r := Setup(cfg, licenseSvc, testFS)

	req := httptest.NewRequest(http.MethodGet, "/api/install/env", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetup_ApiSiteConfig(t *testing.T) {
	setupInstallLock(t)

	cfg := createTestConfig()
	licenseSvc := setupVerifiedLicenseSvc(cfg)
	r := Setup(cfg, licenseSvc, testFS)

	req := httptest.NewRequest(http.MethodGet, "/api/site/config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetup_HealthEndpoint(t *testing.T) {
	setupInstallLock(t)

	cfg := createTestConfig()
	licenseSvc := setupVerifiedLicenseSvc(cfg)
	r := Setup(cfg, licenseSvc, testFS)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetup_AssetServing(t *testing.T) {
	setupInstallLock(t)

	cfg := createTestConfig()
	licenseSvc := setupVerifiedLicenseSvc(cfg)
	r := Setup(cfg, licenseSvc, testFS)

	req := httptest.NewRequest(http.MethodGet, "/assets/test.js", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "console.log")
}

func TestSetup_AssetServingCSS(t *testing.T) {
	setupInstallLock(t)

	cfg := createTestConfig()
	licenseSvc := setupVerifiedLicenseSvc(cfg)
	r := Setup(cfg, licenseSvc, testFS)

	req := httptest.NewRequest(http.MethodGet, "/assets/test.css", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "margin")
}

func TestSetup_AssetNotFound(t *testing.T) {
	setupInstallLock(t)

	cfg := createTestConfig()
	licenseSvc := setupVerifiedLicenseSvc(cfg)
	r := Setup(cfg, licenseSvc, testFS)

	req := httptest.NewRequest(http.MethodGet, "/assets/nonexistent.js", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"app.js", "application/javascript; charset=utf-8"},
		{"style.css", "text/css; charset=utf-8"},
		{"image.png", "image/png"},
		{"photo.jpg", "image/jpeg"},
		{"photo.jpeg", "image/jpeg"},
		{"icon.svg", "image/svg+xml"},
		{"favicon.ico", "image/x-icon"},
		{"data.map", "application/json"},
		{"data.json", "application/json"},
		{"file.unknown", "application/octet-stream"},
		{"file", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectContentType(tt.name)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestServeEmbeddedFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("known file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/index.html", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		serveEmbeddedFile(c, testFS, "assets/index.html")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "<!DOCTYPE html>")
	})

	t.Run("js file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test.js", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		serveEmbeddedFile(c, testFS, "assets/assets/test.js")

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("css file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test.css", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		serveEmbeddedFile(c, testFS, "assets/assets/test.css")

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("png file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test.png", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		serveEmbeddedFile(c, testFS, "assets/assets/test.png")

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unknown file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/missing.html", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		serveEmbeddedFile(c, testFS, "assets/nonexistent.html")

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}