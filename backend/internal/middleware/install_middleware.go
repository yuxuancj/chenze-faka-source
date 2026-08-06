package middleware

import (
	"net/http"
	"os"
	"strings"

	"chenze-faka/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type InstallMiddleware struct{}

func NewInstallMiddleware() *InstallMiddleware {
	return &InstallMiddleware{}
}

func (m *InstallMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isInstalled() {
			c.Next()
			return
		}

		path := c.Request.URL.Path

		if isAllowedWhenNotInstalled(path) {
			c.Next()
			return
		}

		if isAPIRequest(c) {
			c.JSON(http.StatusOK, response.Response{
				Code:    1000,
				Message: "系统未安装",
			})
			c.Abort()
			return
		}

		c.Redirect(http.StatusFound, "/install")
		c.Abort()
	}
}

func isInstalled() bool {
	_, err := os.Stat("install.lock")
	return err == nil
}

func isAllowedWhenNotInstalled(path string) bool {
	allowedPrefixes := []string{
		"/install",
		"/static/",
		"/assets/",
		"/favicon.ico",
	}

	for _, prefix := range allowedPrefixes {
		if path == prefix || strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func isAPIRequest(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "application/json") {
		return true
	}

	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/api/") {
		return true
	}

	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		return true
	}

	return false
}