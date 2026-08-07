package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"chenze-faka/internal/model"
	"chenze-faka/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupLicenseTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	return r
}

func setLicenseVerified(svc *service.LicenseService, verified bool) {
	v := reflect.ValueOf(svc).Elem()
	f := v.FieldByName("verified")
	*(*bool)(unsafe.Pointer(f.UnsafeAddr())) = verified
}

func writeCacheFile(dir string, lastSuccessTime int64) (string, func()) {
	cacheFile := filepath.Join(dir, "license.cache")
	cache := service.LicenseCache{
		LastSuccessTime: lastSuccessTime,
	}
	data, _ := json.Marshal(cache)
	os.WriteFile(cacheFile, data, 0644)
	return cacheFile, func() { os.Remove(cacheFile) }
}

func TestLicenseMiddleware_NilSvc(t *testing.T) {
	r := setupLicenseTestRouter()
	mw := NewLicenseMiddleware(nil)

	r.GET("/any-path", mw.Handle(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/any-path", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLicenseMiddleware_Verified(t *testing.T) {
	cfg := &model.LicenseConfig{
		Enabled:     true,
		CacheFile:   "test_license_cache_verified",
		GracePeriod: 86400,
	}
	svc := service.NewLicenseService(cfg)
	setLicenseVerified(svc, true)

	r := setupLicenseTestRouter()
	mw := NewLicenseMiddleware(svc)

	r.GET("/protected", mw.Handle(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLicenseMiddleware_GracePeriodValid(t *testing.T) {
	dir := t.TempDir()
	cacheFile, cleanup := writeCacheFile(dir, time.Now().Unix())
	defer cleanup()

	cfg := &model.LicenseConfig{
		Enabled:     true,
		CacheFile:   cacheFile,
		GracePeriod: 86400,
	}
	svc := service.NewLicenseService(cfg)
	setLicenseVerified(svc, false)

	r := setupLicenseTestRouter()
	mw := NewLicenseMiddleware(svc)

	r.GET("/protected", mw.Handle(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLicenseMiddleware_PublicPathAllowed(t *testing.T) {
	cfg := &model.LicenseConfig{
		Enabled:     true,
		CacheFile:   "test_license_cache_public",
		GracePeriod: 1,
	}
	svc := service.NewLicenseService(cfg)
	setLicenseVerified(svc, false)

	publicPaths := []string{
		"/",
		"/install",
		"/install-page",
		"/static/app.js",
		"/api/site/config",
		"/api/products",
		"/api/products/on-shelf",
		"/api/products/on-shelf-grouped",
		"/api/products/1",
		"/api/orders/query",
		"/api/orders/123",
		"/api/auth/login",
		"/api/auth/register",
		"/api/cards/product/1",
	}

	for _, path := range publicPaths {
		t.Run(path, func(t *testing.T) {
			r := setupLicenseTestRouter()
			mw := NewLicenseMiddleware(svc)

			r.GET(path, mw.Handle(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "public path %s should be allowed", path)
		})
	}
}

func TestLicenseMiddleware_APIBlocked(t *testing.T) {
	cfg := &model.LicenseConfig{
		Enabled:     true,
		CacheFile:   "test_license_cache_blocked",
		GracePeriod: 1,
	}
	svc := service.NewLicenseService(cfg)
	setLicenseVerified(svc, false)

	r := setupLicenseTestRouter()
	mw := NewLicenseMiddleware(svc)

	r.GET("/api/admin/users", mw.Handle(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "系统授权验证失败")
}

func TestLicenseMiddleware_NonAPIRedirected(t *testing.T) {
	cfg := &model.LicenseConfig{
		Enabled:     true,
		CacheFile:   "test_license_cache_redirect",
		GracePeriod: 1,
	}
	svc := service.NewLicenseService(cfg)
	setLicenseVerified(svc, false)

	r := setupLicenseTestRouter()
	mw := NewLicenseMiddleware(svc)

	r.GET("/admin/dashboard", mw.Handle(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/install", w.Header().Get("Location"))
}