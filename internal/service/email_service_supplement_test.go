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

func setupEmailSupplementTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_fk=1"), &gorm.Config{})
	require.NoError(t, err)
	dropAllEmailTables(db)
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

func TestEmailService_TestConnection(t *testing.T) {
	setupEmailSupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewEmailService()

	t.Run("test connection with bad SMTP creds returns error", func(t *testing.T) {
		cfg := &model.EmailConfig{
			SMTPHost: "127.0.0.1",
			SMTPPort: 19999,
			Username: "bad-creds@test.com",
			Password: "bad-password",
			Sender:   "bad-creds@test.com",
			UseSSL:   false,
			Status:   model.ChannelEnabled,
		}
		created, err := svc.Create(cfg)
		require.NoError(t, err)

		err = svc.TestConnection(created.ID)
		assert.Error(t, err)
	})

	t.Run("test connection with nonexistent config id", func(t *testing.T) {
		err := svc.TestConnection(9999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "配置不存在")
	})

	t.Run("test connection with nil DB", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		err := svc.TestConnection(1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "数据库未连接")
	})
}

func TestEmailService_SendErrorPaths(t *testing.T) {
	setupEmailSupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewEmailService()

	t.Run("send with empty recipient list fails", func(t *testing.T) {
		cfg := &model.EmailConfig{
			SMTPHost: "127.0.0.1",
			SMTPPort: 19999,
			Username: "test@test.com",
			Password: "password",
			Sender:   "test@test.com",
			UseSSL:   true,
			Status:   model.ChannelEnabled,
		}
		_, err := svc.Create(cfg)
		require.NoError(t, err)

		err = svc.SendCardEmail([]string{}, "Subject", "Body", "ORD-NIL")
		assert.Error(t, err)
	})

	t.Run("send with invalid SMTP host fails and logs", func(t *testing.T) {
		cfg := &model.EmailConfig{
			SMTPHost: "127.0.0.1",
			SMTPPort: 19998,
			Username: "no@test.com",
			Password: "nopass",
			Sender:   "no@test.com",
			UseSSL:   false,
			Status:   model.ChannelEnabled,
		}
		_, err := svc.Create(cfg)
		require.NoError(t, err)

		err = svc.SendCardEmail([]string{"dest@test.com"}, "Test", "Body", "ORD-INVALID")
		assert.Error(t, err)

		var logs []model.EmailLog
		require.NoError(t, database.DB.Find(&logs).Error)
		require.NotEmpty(t, logs)
		assert.Equal(t, model.EmailLogFailed, logs[len(logs)-1].Status)
	})

	t.Run("send with nil DB returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		err := svc.SendCardEmail([]string{"user@test.com"}, "Test", "Body", "ORD-NODB")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "数据库未连接")
	})
}

func TestEmailService_SendWithEmptyFields(t *testing.T) {
	setupEmailSupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewEmailService()

	t.Run("send with empty sender falls back to username", func(t *testing.T) {
		cfg := &model.EmailConfig{
			SMTPHost: "127.0.0.1",
			SMTPPort: 19997,
			Username: "fallback@test.com",
			Password: "pass",
			Sender:   "",
			UseSSL:   false,
			Status:   model.ChannelEnabled,
		}
		_, err := svc.Create(cfg)
		require.NoError(t, err)

		err = svc.SendCardEmail([]string{"dest@test.com"}, "Test", "Body", "ORD-EMPTY")
		assert.Error(t, err)
	})

	t.Run("send with zero port uses default 465", func(t *testing.T) {
		cfg := &model.EmailConfig{
			SMTPHost: "smtp.test.com",
			SMTPPort: 0,
			Username: "test@test.com",
			Password: "pass",
			Sender:   "test@test.com",
			Status:   model.ChannelEnabled,
		}
		created, err := svc.Create(cfg)
		require.NoError(t, err)
		assert.Equal(t, 465, created.SMTPPort)
	})
}