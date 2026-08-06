package middleware

import (
	"chenze-faka/internal/license"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type LicenseMiddleware struct {
	svc *license.Service
}

func NewLicenseMiddleware(svc *license.Service) *LicenseMiddleware {
	return &LicenseMiddleware{svc: svc}
}

func (m *LicenseMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.svc.IsVerified() && !m.svc.IsGracePeriodValid() {
			path := c.Request.URL.Path

			if isPublicPath(path) {
				c.Next()
				return
			}

			if path == "/install-page" || path == "/install" || strings.HasPrefix(path, "/static/") {
				c.Next()
				return
			}

			if strings.HasPrefix(path, "/admin/") {
				c.Redirect(http.StatusFound, "/admin/login.html")
				return
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "system license verification failed, please contact admin",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func isPublicPath(path string) bool {
	publicPaths := []string{
		"/",
		"/index.html",
		"/static/",
		"/api/products",
		"/api/products/on-shelf",
		"/api/orders/query",
		"/api/site/config",
		"/admin/login.html",
		"/install-page",
		"/install",
	}

	for _, p := range publicPaths {
		if p == "/" && (path == "/" || path == "/index.html") {
			return true
		}
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	return false
}

func isInstallPath(path string) bool {
	return path == "/install-page" || path == "/install" || strings.HasPrefix(path, "/install/")
}
