package service

import (
	"testing"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func dropAllAuthTables(db *gorm.DB) {
	tables := []string{"users", "categories", "products", "cards", "orders", "payment_channels", "email_configs", "nodes", "operation_logs", "order_logs", "email_logs", "upgrade_logs", "file_uploads"}
	for _, t := range tables {
		db.Migrator().DropTable(t)
	}
}

func setupAuthTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	dropAllAuthTables(db)
	database.DB = db
	require.NoError(t, db.AutoMigrate(
		&model.User{}, &model.Category{}, &model.Product{},
		&model.Card{}, &model.Order{}, &model.PaymentChannel{},
		&model.EmailConfig{}, &model.Node{},
		&model.OperationLog{}, &model.OrderLog{}, &model.EmailLog{},
		&model.UpgradeLog{}, &model.FileUpload{},
	))
}

func TestAuthService_Register(t *testing.T) {
	setupAuthTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	t.Run("register with valid params", func(t *testing.T) {
		user, err := svc.Register("testuser", "testpass123")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, model.RoleAdmin, user.Role)
		assert.NotEmpty(t, user.PasswordHash)
		assert.NotEmpty(t, user.Salt)
	})

	t.Run("register with empty username", func(t *testing.T) {
		user, err := svc.Register("", "password")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "用户名和密码不能为空")
	})

	t.Run("register with empty password", func(t *testing.T) {
		user, err := svc.Register("username", "")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "用户名和密码不能为空")
	})

	t.Run("register duplicate user", func(t *testing.T) {
		user, err := svc.Register("testuser", "testpass123")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "用户名已存在")
	})
}

func TestAuthService_Login(t *testing.T) {
	setupAuthTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	_, err := svc.Register("loginuser", "loginpass123")
	require.NoError(t, err)

	t.Run("login with correct credentials", func(t *testing.T) {
		token, user, err := svc.Login("loginuser", "loginpass123", "test-secret", 72)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotNil(t, user)
		assert.Equal(t, "loginuser", user.Username)
	})

	t.Run("login with wrong password", func(t *testing.T) {
		token, user, err := svc.Login("loginuser", "wrongpass", "test-secret", 72)
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Nil(t, user)
	})

	t.Run("login with nonexistent user", func(t *testing.T) {
		token, user, err := svc.Login("nonexistent", "password", "test-secret", 72)
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Nil(t, user)
	})

	t.Run("login when DB is nil - admin fallback", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		token, user, err := svc.Login("admin", "admin123", "admin-secret", 72)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotNil(t, user)
		assert.Equal(t, "admin", user.Username)
		assert.Equal(t, model.RoleAdmin, user.Role)
	})

	t.Run("login when DB is nil - wrong credentials", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		token, user, err := svc.Login("admin", "wrongpass", "admin-secret", 72)
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Nil(t, user)
	})
}

func TestAuthService_ParseToken(t *testing.T) {
	setupAuthTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	user, err := svc.Register("tokenuser", "tokenpass")
	require.NoError(t, err)

	t.Run("parse valid token", func(t *testing.T) {
		token, _, err := svc.Login("tokenuser", "tokenpass", "test-secret", 72)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		claims, err := svc.ParseToken(token, "test-secret")
		require.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, float64(user.ID), (*claims)["user_id"])
		assert.Equal(t, "tokenuser", (*claims)["username"])
	})

	t.Run("parse invalid token", func(t *testing.T) {
		claims, err := svc.ParseToken("invalid-token-string", "test-secret")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("parse token with wrong secret", func(t *testing.T) {
		token, _, err := svc.Login("tokenuser", "tokenpass", "test-secret", 72)
		require.NoError(t, err)

		claims, err := svc.ParseToken(token, "wrong-secret")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestAuthService_GetUserByID(t *testing.T) {
	setupAuthTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	user, err := svc.Register("getuser", "getpass")
	require.NoError(t, err)

	t.Run("get user by valid id", func(t *testing.T) {
		found, err := svc.GetUserByID(user.ID)
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.Username, found.Username)
	})

	t.Run("get user by invalid id", func(t *testing.T) {
		found, err := svc.GetUserByID(9999)
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "用户不存在")
	})
}

func TestAuthService_CheckInstalled(t *testing.T) {
	setupAuthTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewAuthService()

	t.Run("check installed with no users", func(t *testing.T) {
		installed := svc.CheckInstalled()
		assert.False(t, installed)
	})

	t.Run("check installed with users", func(t *testing.T) {
		_, err := svc.Register("checkuser", "checkpass")
		require.NoError(t, err)

		installed := svc.CheckInstalled()
		assert.True(t, installed)
	})
}