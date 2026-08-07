package service

import (
	"testing"
	"time"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
	"chenze-faka/internal/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthDBNilSupplement(t *testing.T) {
	svc := NewAuthService()

	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	t.Run("register with nil db returns error", func(t *testing.T) {
		user, err := svc.Register("testuser", "testpass")
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("login with nil db and wrong credentials", func(t *testing.T) {
		token, user, err := svc.Login("wrong", "wrong", "secret", 72)
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Nil(t, user)
	})

	t.Run("login with nil db and admin fallback works", func(t *testing.T) {
		token, user, err := svc.Login("admin", "admin123", "secret", 72)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotNil(t, user)
		assert.Equal(t, "admin", user.Username)
	})

	t.Run("parse token with nil db still works", func(t *testing.T) {
		secret := "test-secret"
		claims := jwt.MapClaims{
			"user_id":  uint(1),
			"username": "test",
			"role":     model.RoleAdmin,
			"exp":      time.Now().Add(time.Hour).Unix(),
			"iat":      time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
		require.NoError(t, err)

		result, err := svc.ParseToken(tokenString, secret)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestAuthRegisterEmptyFields(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	t.Run("register with empty username fails", func(t *testing.T) {
		user, err := svc.Register("", "password")
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("register with empty password fails", func(t *testing.T) {
		user, err := svc.Register("username", "")
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("register with both empty fails", func(t *testing.T) {
		user, err := svc.Register("", "")
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestAuthRegisterDuplicate(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	user, err := svc.Register("duplicate_user", "password1")
	require.NoError(t, err)
	assert.NotNil(t, user)

	t.Run("register with same username fails", func(t *testing.T) {
		user2, err := svc.Register("duplicate_user", "password2")
		assert.Error(t, err)
		assert.Nil(t, user2)
		assert.Contains(t, err.Error(), "用户名已存在")
	})
}

func TestAuthLoginWrongPassword(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	_, err := svc.Register("login_user", "correct_pass")
	require.NoError(t, err)

	t.Run("login with wrong password fails", func(t *testing.T) {
		token, user, err := svc.Login("login_user", "wrong_pass", "secret", 72)
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "用户名或密码错误")
	})

	t.Run("login with wrong username fails", func(t *testing.T) {
		token, user, err := svc.Login("nonexistent", "password", "secret", 72)
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Nil(t, user)
	})
}

func TestAuthLoginSuccess(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	_, err := svc.Register("login_success_user", "password123")
	require.NoError(t, err)

	t.Run("login with correct credentials succeeds", func(t *testing.T) {
		token, user, err := svc.Login("login_success_user", "password123", "test-secret", 72)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotNil(t, user)
		assert.Equal(t, "login_success_user", user.Username)
	})
}

func TestParseTokenTampered(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	user, err := svc.Register("tamper_user", "password")
	require.NoError(t, err)

	secret := "test-secret"
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	t.Run("tampered token fails verification", func(t *testing.T) {
		tamperedToken := tokenString + "tampered"
		result, err := svc.ParseToken(tamperedToken, secret)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("token signed with different secret fails", func(t *testing.T) {
		result, err := svc.ParseToken(tokenString, "wrong-secret")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestParseTokenInvalidFormat(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	t.Run("completely invalid token string fails", func(t *testing.T) {
		result, err := svc.ParseToken("invalid-token-string", "secret")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("empty token fails", func(t *testing.T) {
		result, err := svc.ParseToken("", "secret")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGenerateTokenSuccess(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	user, err := svc.Register("token_user", "password")
	require.NoError(t, err)

	t.Run("parse valid token returns user info", func(t *testing.T) {
		secret := "test-secret"
		token, _, err := svc.Login("token_user", "password", secret, 72)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		result, err := svc.ParseToken(token, secret)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, user.ID, uint((*result)["user_id"].(float64)))
		assert.Equal(t, "token_user", (*result)["username"])
	})
}

func TestHashPasswordAndSalt(t *testing.T) {
	t.Run("hash password with salt produces consistent result", func(t *testing.T) {
		salt := "test_salt_123"
		hash1 := utils.HashPassword("mypassword", salt)
		hash2 := utils.HashPassword("mypassword", salt)
		assert.Equal(t, hash1, hash2)
	})

	t.Run("hash password with different salt produces different result", func(t *testing.T) {
		hash1 := utils.HashPassword("mypassword", "salt1")
		hash2 := utils.HashPassword("mypassword", "salt2")
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("different password with same salt produces different hash", func(t *testing.T) {
		hash1 := utils.HashPassword("password1", "somesalt")
		hash2 := utils.HashPassword("password2", "somesalt")
		assert.NotEqual(t, hash1, hash2)
	})
}