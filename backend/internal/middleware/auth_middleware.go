package middleware

import (
	"strings"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/response"
	"chenze-faka/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	authService *service.AuthService
	jwtSecret   string
	expireHours int
}

func NewAuthMiddleware(jwtSecret string, expireHours int) *AuthMiddleware {
	return &AuthMiddleware{
		authService: service.NewAuthService(),
		jwtSecret:   jwtSecret,
		expireHours: expireHours,
	}
}

func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			authHeader = c.Query("token")
		}

		if authHeader == "" {
			response.Unauthorized(c, "需要登录")
			c.Abort()
			return
		}

		tokenString := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:]
		}

		claims, err := m.authService.ParseToken(tokenString, m.jwtSecret)
		if err != nil {
			response.Unauthorized(c, "登录已过期,请重新登录")
			c.Abort()
			return
		}

		userID := uint((*claims)["user_id"].(float64))
		user, err := m.authService.GetUserByID(userID)
		if err != nil {
			username := ""
			if uname, ok := (*claims)["username"].(string); ok {
				username = uname
			}
			role := "admin"
			if r, ok := (*claims)["role"].(string); ok {
				role = r
			}
			user = &model.User{
				ID:       userID,
				Username: username,
				Role:     role,
			}
		}

		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("role", user.Role)
		c.Next()
	}
}

func (m *AuthMiddleware) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			response.Unauthorized(c, "需要登录")
			c.Abort()
			return
		}

		u := user.(*model.User)
		if u.Role != model.RoleAdmin {
			response.Forbidden(c, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}