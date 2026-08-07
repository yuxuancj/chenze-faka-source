package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupInstallTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	return r
}

func TestInstallMiddleware_WhenInstalled(t *testing.T) {
	os.WriteFile("install.lock", []byte("installed"), 0644)
	defer os.Remove("install.lock")

	r := setupInstallTestRouter()
	mw := NewInstallMiddleware()

	r.GET("/some-path", mw.Handle(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/some-path", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInstallMiddleware_NotInstalled_InstallPathAllowed(t *testing.T) {
	os.Remove("install.lock")

	r := setupInstallTestRouter()
	mw := NewInstallMiddleware()

	r.GET("/install", mw.Handle(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "install page"})
	})

	req := httptest.NewRequest(http.MethodGet, "/install", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInstallMiddleware_NotInstalled_NonAPIRedirected(t *testing.T) {
	os.Remove("install.lock")

	r := setupInstallTestRouter()
	mw := NewInstallMiddleware()

	r.GET("/dashboard", mw.Handle(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "dashboard"})
	})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/install", w.Header().Get("Location"))
}

func TestInstallMiddleware_NotInstalled_APIJSONResponse(t *testing.T) {
	os.Remove("install.lock")

	r := setupInstallTestRouter()
	mw := NewInstallMiddleware()

	r.GET("/api/users", mw.Handle(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "users"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "系统未安装")
	assert.Contains(t, w.Body.String(), `"code":1000`)
}

func TestInstallMiddleware_NotInstalled_APISiteConfigSpecialCode(t *testing.T) {
	os.Remove("install.lock")

	r := setupInstallTestRouter()
	mw := NewInstallMiddleware()

	r.GET("/api/site/config", mw.Handle(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "config"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/site/config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":1000`)
}

func TestIsInstalled(t *testing.T) {
	os.Remove("install.lock")
	assert.False(t, isInstalled())

	os.WriteFile("install.lock", []byte("installed"), 0644)
	assert.True(t, isInstalled())
	os.Remove("install.lock")

	assert.False(t, isInstalled())
}

func TestIsAllowedWhenNotInstalled(t *testing.T) {
	tests := []struct {
		path    string
		allowed bool
	}{
		{"/install", true},
		{"/install/step1", true},
		{"/static/app.js", true},
		{"/assets/logo.png", true},
		{"/favicon.ico", true},
		{"/api/install/env", true},
		{"/dashboard", false},
		{"/api/users", false},
		{"/login", false},
		{"/", false},
		{"/products", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.allowed, isAllowedWhenNotInstalled(tt.path))
		})
	}
}

func TestIsAPIRequest(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		accept      string
		contentType string
		want        bool
	}{
		{"api path", "/api/users", "", "", true},
		{"api path sub", "/api/users/1", "", "", true},
		{"accept json", "/web/page", "application/json", "", true},
		{"content type json", "/web/page", "", "application/json", true},
		{"web page", "/web/page", "text/html", "", false},
		{"no headers no api", "/dashboard", "", "", false},
		{"partial api prefix not match", "/apix/users", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.GET(tt.path, func(c *gin.Context) {})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			assert.Equal(t, tt.want, isAPIRequest(c))
		})
	}
}