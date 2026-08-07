package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
	"chenze-faka/internal/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAuthTestServer() *gin.Engine {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: ":memory:",
	}
	database.Init(cfg)
	database.AutoMigrate()

	authCtrl := NewAuthController("test-secret", 72, "Test Site", &model.LicenseConfig{}, cfg)

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		install := api.Group("/install")
		{
			install.GET("/env", authCtrl.CheckEnv)
		}
		auth := api.Group("/auth")
		{
			auth.POST("/login", authCtrl.Login)
			auth.POST("/register", authCtrl.Register)
		}
	}

	return r
}

func createTestUser(username, password string) {
	salt := utils.GenerateSalt()
	hash := utils.HashPassword(password, salt)
	user := &model.User{
		Username:     username,
		PasswordHash: hash,
		Salt:         salt,
		Role:         model.RoleAdmin,
	}
	database.DB.Create(user)
}

func TestCheckEnv(t *testing.T) {
	r := setupAuthTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/install/env", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotEmpty(t, data["os"])
	assert.NotEmpty(t, data["go_version"])
	assert.Equal(t, "normal", data["mysql_status"])
	assert.Equal(t, "SQLite 3.x", data["mysql_version"])
}

func TestLoginSuccess(t *testing.T) {
	r := setupAuthTestServer()
	createTestUser("testuser", "testpass123")

	body := map[string]string{
		"username": "testuser",
		"password": "testpass123",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.NotEmpty(t, data["token"])
	assert.NotNil(t, data["user"])
	user := data["user"].(map[string]interface{})
	assert.Equal(t, "testuser", user["username"])
	assert.Equal(t, "admin", user["role"])
}

func TestLoginInvalidCredentials(t *testing.T) {
	r := setupAuthTestServer()
	createTestUser("testuser", "testpass123")

	body := map[string]string{
		"username": "testuser",
		"password": "wrongpassword",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
	assert.NotEmpty(t, resp["message"])
}

func TestRegisterSuccess(t *testing.T) {
	r := setupAuthTestServer()

	body := map[string]string{
		"username": "newuser",
		"password": "newpass123",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), resp["code"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "newuser", data["username"])
	assert.NotNil(t, data["id"])
}

func TestRegisterDuplicate(t *testing.T) {
	r := setupAuthTestServer()
	createTestUser("existinguser", "pass123")

	body := map[string]string{
		"username": "existinguser",
		"password": "newpass123",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), resp["code"])
	assert.NotEmpty(t, resp["message"])
}