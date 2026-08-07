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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthCoverageTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_fk=1"), &gorm.Config{})
	require.NoError(t, err)
	dropAllCoverageTables(db)
	database.DB = db
	require.NoError(t, db.AutoMigrate(
		&model.User{}, &model.Category{}, &model.Product{},
		&model.Card{}, &model.Order{}, &model.PaymentChannel{},
		&model.EmailConfig{}, &model.Node{},
		&model.OperationLog{}, &model.OrderLog{}, &model.EmailLog{},
		&model.UpgradeLog{}, &model.FileUpload{},
	))
	return db
}

func TestParseTokenExpiredCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	user, err := svc.Register("token_expired_user", "token_pass")
	require.NoError(t, err)

	secret := "test-secret"
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(-1 * time.Hour).Unix(),
		"iat":      time.Now().Add(-2 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	result, err := svc.ParseToken(tokenString, secret)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestParseTokenTamperedCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	user, err := svc.Register("tampered_user", "tampered_pass")
	require.NoError(t, err)

	secret := "test-secret"
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	tampered := tokenString + "extra"

	result, err := svc.ParseToken(tampered, secret)
	assert.Error(t, err)
	assert.Nil(t, result)

	result2, err := svc.ParseToken("completely-invalid-jwt-token", secret)
	assert.Error(t, err)
	assert.Nil(t, result2)
}

func TestParseTokenWrongSecretCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	user, err := svc.Register("wrong_secret_user", "wrong_secret_pass")
	require.NoError(t, err)

	token, _, err := svc.Login("wrong_secret_user", "wrong_secret_pass", "correct-secret", 72)
	require.NoError(t, err)

	result, err := svc.ParseToken(token, "wrong-secret")
	assert.Error(t, err)
	assert.Nil(t, result)

	claims, err := svc.ParseToken(token, "correct-secret")
	require.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, float64(user.ID), (*claims)["user_id"])
}

func TestLoginWrongSaltPasswordCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	_, err := svc.Register("salt_user", "salt_pass")
	require.NoError(t, err)

	t.Run("login with correct password succeeds", func(t *testing.T) {
		token, user, err := svc.Login("salt_user", "salt_pass", "test-secret", 72)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotNil(t, user)
	})

	t.Run("login with wrong password fails", func(t *testing.T) {
		token, user, err := svc.Login("salt_user", "wrong_salt_pass", "test-secret", 72)
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "用户名或密码错误")
	})

	t.Run("login with nonexistent user fails", func(t *testing.T) {
		token, user, err := svc.Login("ghost_user", "anypass", "test-secret", 72)
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "用户名或密码错误")
	})
}

func TestLoginDBNilNotAdminFallbackCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	t.Run("login with DB nil and not admin returns error", func(t *testing.T) {
		token, user, err := svc.Login("regular_user", "password", "test-secret", 72)
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "用户名或密码错误")
	})

	t.Run("login with DB nil and admin credentials succeeds", func(t *testing.T) {
		token, user, err := svc.Login("admin", "admin123", "admin-secret", 72)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotNil(t, user)
		assert.Equal(t, "admin", user.Username)
		assert.Equal(t, model.RoleAdmin, user.Role)
	})

	t.Run("login with DB nil and admin wrong password fails", func(t *testing.T) {
		token, user, err := svc.Login("admin", "wrongpass", "admin-secret", 72)
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Nil(t, user)
	})
}

func TestGetUserByIDCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	user, err := svc.Register("getuser_coverage", "getuser_pass")
	require.NoError(t, err)

	t.Run("get user by valid id", func(t *testing.T) {
		found, err := svc.GetUserByID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Username, found.Username)
		assert.Equal(t, user.Role, found.Role)
	})

	t.Run("get user by invalid id returns error", func(t *testing.T) {
		found, err := svc.GetUserByID(9999)
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "用户不存在")
	})
}

func TestCheckInstalledCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	t.Run("check installed with no users returns false", func(t *testing.T) {
		installed := svc.CheckInstalled()
		assert.False(t, installed)
	})

	t.Run("check installed with users returns true", func(t *testing.T) {
		_, err := svc.Register("check_user", "check_pass")
		require.NoError(t, err)

		installed := svc.CheckInstalled()
		assert.True(t, installed)
	})

	t.Run("check installed when DB is nil returns false", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		installed := svc.CheckInstalled()
		assert.False(t, installed)
	})
}

func TestRegisterCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	t.Run("register with valid data", func(t *testing.T) {
		user, err := svc.Register("register_valid", "valid_pass")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "register_valid", user.Username)
		assert.Equal(t, model.RoleAdmin, user.Role)
		assert.NotEmpty(t, user.PasswordHash)
		assert.NotEmpty(t, user.Salt)
	})

	t.Run("register with empty username fails", func(t *testing.T) {
		user, err := svc.Register("", "password")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "用户名和密码不能为空")
	})

	t.Run("register with empty password fails", func(t *testing.T) {
		user, err := svc.Register("username", "")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "用户名和密码不能为空")
	})

	t.Run("register duplicate user fails", func(t *testing.T) {
		user, err := svc.Register("register_valid", "newpass")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "用户名已存在")
	})
}

func TestGenerateTokenCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	user := &model.User{
		ID:       1,
		Username: "token_user",
		Role:     model.RoleAdmin,
	}

	t.Run("generate token with default expire time", func(t *testing.T) {
		token, err := generateToken(user, "test-secret", 0)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("generate token with negative expire time uses default", func(t *testing.T) {
		token, err := generateToken(user, "test-secret", -1)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("generated token can be parsed", func(t *testing.T) {
		svc := NewAuthService()
		tokenStr, err := generateToken(user, "parse-secret", 72)
		require.NoError(t, err)

		claims, err := svc.ParseToken(tokenStr, "parse-secret")
		require.NoError(t, err)
		assert.Equal(t, float64(1), (*claims)["user_id"])
		assert.Equal(t, "token_user", (*claims)["username"])
	})
}

func TestLoginTokenFlowCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	user, err := svc.Register("flow_user", "flow_pass")
	require.NoError(t, err)

	token, loggedInUser, err := svc.Login("flow_user", "flow_pass", "flow-secret", 72)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, loggedInUser)

	claims, err := svc.ParseToken(token, "flow-secret")
	require.NoError(t, err)
	assert.Equal(t, float64(user.ID), (*claims)["user_id"])
	assert.Equal(t, "flow_user", (*claims)["username"])

	found, err := svc.GetUserByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, "flow_user", found.Username)
}

func TestPasswordHashVerificationCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	user, err := svc.Register("hash_user", "hash_pass_123")
	require.NoError(t, err)

	salt := user.Salt
	correctHash := utils.HashPassword("hash_pass_123", salt)
	assert.Equal(t, user.PasswordHash, correctHash)

	wrongHash := utils.HashPassword("wrong_password", salt)
	assert.NotEqual(t, user.PasswordHash, wrongHash)
}

func TestParseTokenInvalidSignMethodCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	secret := "test-secret"
	claims := jwt.MapClaims{
		"user_id":  1,
		"username": "testuser",
		"role":     "user",
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	t.Run("token with none signing method fails", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		result, err := svc.ParseToken(tokenString, secret)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("empty token string fails", func(t *testing.T) {
		result, err := svc.ParseToken("", secret)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("garbage token string fails", func(t *testing.T) {
		result, err := svc.ParseToken("not-a-valid-jwt-token", secret)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestRegisterEdgeCasesCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	t.Run("register with very long username succeeds", func(t *testing.T) {
		longName := "very_long_username_that_is_within_limits_abcdefghijklmnop"
		user, err := svc.Register(longName, "validpass123")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, longName, user.Username)
	})

	t.Run("register with empty password fails", func(t *testing.T) {
		user, err := svc.Register("user_no_pass", "")
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("register with short password succeeds", func(t *testing.T) {
		user, err := svc.Register("user_short_pass", "12")
		assert.NoError(t, err)
		assert.NotNil(t, user)
	})

	t.Run("register with special chars in password succeeds", func(t *testing.T) {
		user, err := svc.Register("user_special_pass", "p@$$w0rd!#%")
		assert.NoError(t, err)
		assert.NotNil(t, user)
	})
}

func TestGetUserByIDDBCoverage(t *testing.T) {
	setupAuthCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	t.Run("get user by existing ID", func(t *testing.T) {
		user, err := svc.Register("getbyid_user", "getbyid_pass")
		require.NoError(t, err)

		result, err := svc.GetUserByID(user.ID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, user.ID, result.ID)
	})

	t.Run("get user by ID when DB is nil", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		result, err := svc.GetUserByID(1)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}