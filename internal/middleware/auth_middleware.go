package middleware

import (
	"strings"

	"chenze-faka/internal/config"
	"chenze-faka/internal/service"
	"chenze-faka/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	authService *service.AuthService
	jwtConfig   config.JWTConfig
}

func NewAuthMiddleware(jwtConfig config.JWTConfig) *AuthMiddleware {
	return &AuthMiddleware{
		authService: service.NewAuthService(jwtConfig),
		jwtConfig:   jwtConfig,
	}
}

func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			authHeader = c.Query("token")
		}

		if authHeader == "" {
			utils.Unauthorized(c, "authorization required")
			c.Abort()
			return
		}

		tokenString := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:]
		}

		claims, err := m.authService.ParseToken(tokenString)
		if err != nil {
			utils.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		userID := uint((*claims)["user_id"].(float64))
		user, err := m.authService.GetUserByID(userID)
		if err != nil {
			utils.Unauthorized(c, "user not found")
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Next()
	}
}

func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			authHeader = c.Query("token")
		}

		if authHeader == "" {
			c.Next()
			return
		}

		tokenString := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:]
		}

		claims, err := m.authService.ParseToken(tokenString)
		if err == nil {
			userID := uint((*claims)["user_id"].(float64))
			user, err := m.authService.GetUserByID(userID)
			if err == nil {
				c.Set("user", user)
				c.Set("user_id", user.ID)
			}
		}

		c.Next()
	}
}
