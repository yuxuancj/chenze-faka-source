package middleware

import (
	"net/http"
	"strings"

	"chenze-faka/internal/pkg/response"
	"chenze-faka/internal/service"

	"github.com/gin-gonic/gin"
)

type LicenseMiddleware struct {
	svc *service.LicenseService
}

func NewLicenseMiddleware(svc *service.LicenseService) *LicenseMiddleware {
	return &LicenseMiddleware{svc: svc}
}

func (m *LicenseMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.svc == nil || m.svc.IsVerified() || m.svc.IsGracePeriodValid() {
			c.Next()
			return
		}

		path := c.Request.URL.Path

		if isLicensePublicPath(path) {
			c.Next()
			return
		}

		if strings.HasPrefix(path, "/api/") {
			response.FailWithCode(c, http.StatusUnauthorized, "系统授权验证失败,请联系管理员")
			c.Abort()
			return
		}

		c.Redirect(http.StatusFound, "/install")
		c.Abort()
	}
}

func isLicensePublicPath(path string) bool {
	publicPaths := []string{
		"/",
		"/install",
		"/install-page",
		"/static/",
		"/assets/",
		"/api/site/config",
		"/api/products",
		"/api/products/on-shelf",
		"/api/products/on-shelf-grouped",
		"/api/products/",
		"/api/orders/query",
		"/api/orders/",
		"/api/auth/login",
		"/api/auth/register",
		"/api/cards/product/",
	}

	for _, p := range publicPaths {
		if p == "/" {
			if path == "/" {
				return true
			}
			continue
		}
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}