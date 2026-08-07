package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chenze-faka/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const testJWTSecret = "test-secret"

func generateTestToken(userID uint, username, role string, secret string) string {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func setupAuthTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	return r
}

func TestAuthRequired_ValidBearerToken(t *testing.T) {
	r := setupAuthTestRouter()
	mw := NewAuthMiddleware(testJWTSecret, 72)

	r.GET("/protected", mw.AuthRequired(), func(c *gin.Context) {
		user, _ := c.Get("user")
		u := user.(*model.User)
		assert.Equal(t, uint(1), u.ID)
		assert.Equal(t, "admin", u.Username)
		assert.Equal(t, "admin", u.Role)
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	token := generateTestToken(1, "admin", "admin", testJWTSecret)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthRequired_NoToken(t *testing.T) {
	r := setupAuthTestRouter()
	mw := NewAuthMiddleware(testJWTSecret, 72)

	r.GET("/protected", mw.AuthRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRequired_InvalidToken(t *testing.T) {
	r := setupAuthTestRouter()
	mw := NewAuthMiddleware(testJWTSecret, 72)

	r.GET("/protected", mw.AuthRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-here")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRequired_QueryParamToken(t *testing.T) {
	r := setupAuthTestRouter()
	mw := NewAuthMiddleware(testJWTSecret, 72)

	r.GET("/protected", mw.AuthRequired(), func(c *gin.Context) {
		user, _ := c.Get("user")
		u := user.(*model.User)
		assert.Equal(t, uint(1), u.ID)
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	token := generateTestToken(1, "admin", "admin", testJWTSecret)
	req := httptest.NewRequest(http.MethodGet, "/protected?token="+token, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthRequired_NoDB_Success(t *testing.T) {
	r := setupAuthTestRouter()
	mw := NewAuthMiddleware(testJWTSecret, 72)

	r.GET("/protected", mw.AuthRequired(), func(c *gin.Context) {
		user, exists := c.Get("user")
		assert.True(t, exists)
		u := user.(*model.User)
		assert.Equal(t, uint(1), u.ID)
		assert.Equal(t, "admin", u.Username)
		assert.Equal(t, "admin", u.Role)

		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, uint(1), userID)

		username, exists := c.Get("username")
		assert.True(t, exists)
		assert.Equal(t, "admin", username)

		role, exists := c.Get("role")
		assert.True(t, exists)
		assert.Equal(t, "admin", role)

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	token := generateTestToken(1, "admin", "admin", testJWTSecret)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminRequired_AdminRole(t *testing.T) {
	r := setupAuthTestRouter()
	mw := NewAuthMiddleware(testJWTSecret, 72)

	r.GET("/admin", mw.AuthRequired(), mw.AdminRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin ok"})
	})

	token := generateTestToken(1, "admin", "admin", testJWTSecret)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminRequired_NonAdminRole(t *testing.T) {
	r := setupAuthTestRouter()
	mw := NewAuthMiddleware(testJWTSecret, 72)

	r.GET("/admin", mw.AuthRequired(), mw.AdminRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin ok"})
	})

	token := generateTestToken(2, "user1", "user", testJWTSecret)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminRequired_NoUserInContext(t *testing.T) {
	r := setupAuthTestRouter()
	mw := NewAuthMiddleware(testJWTSecret, 72)

	r.GET("/admin", mw.AdminRequired(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}