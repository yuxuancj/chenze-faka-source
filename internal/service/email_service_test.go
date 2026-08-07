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

func dropAllEmailTables(db *gorm.DB) {
	tables := []string{"users", "categories", "products", "cards", "orders", "payment_channels", "email_configs", "nodes", "operation_logs", "order_logs", "email_logs", "upgrade_logs", "file_uploads"}
	for _, tb := range tables {
		db.Migrator().DropTable(tb)
	}
}

func setupEmailTestDB(t *testing.T) *gorm.DB {
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

func TestEmailService_Create(t *testing.T) {
	setupEmailTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewEmailService()

	t.Run("create email config successfully", func(t *testing.T) {
		cfg := &model.EmailConfig{
			SMTPHost: "smtp.test.com",
			SMTPPort: 465,
			Username: "test@test.com",
			Password: "password",
			Sender:   "test@test.com",
			UseSSL:   true,
			Status:   model.ChannelEnabled,
		}
		result, err := svc.Create(cfg)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "smtp.test.com", result.SMTPHost)
		assert.Equal(t, 465, result.SMTPPort)
		assert.Equal(t, "test@test.com", result.Username)
		assert.Equal(t, "password", result.Password)
		assert.Equal(t, "test@test.com", result.Sender)
		assert.True(t, result.UseSSL)
		assert.Equal(t, model.ChannelEnabled, result.Status)
	})

	t.Run("create with default port", func(t *testing.T) {
		cfg := &model.EmailConfig{
			SMTPHost: "smtp.test.com",
			Username: "test2@test.com",
			Password: "password",
			Sender:   "test2@test.com",
			Status:   model.ChannelEnabled,
		}
		result, err := svc.Create(cfg)
		require.NoError(t, err)
		assert.Equal(t, 465, result.SMTPPort)
	})

	t.Run("create with empty SMTP host fails", func(t *testing.T) {
		cfg := &model.EmailConfig{
			Username: "test@test.com",
			Password: "password",
		}
		result, err := svc.Create(cfg)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "SMTP配置不完整")
	})

	t.Run("create with empty username fails", func(t *testing.T) {
		cfg := &model.EmailConfig{
			SMTPHost: "smtp.test.com",
			Password: "password",
		}
		result, err := svc.Create(cfg)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "SMTP配置不完整")
	})
}

func TestEmailService_Update(t *testing.T) {
	setupEmailTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewEmailService()

	cfg := &model.EmailConfig{
		SMTPHost: "smtp.test.com",
		SMTPPort: 465,
		Username: "test@test.com",
		Password: "password",
		Sender:   "test@test.com",
		UseSSL:   true,
		Status:   model.ChannelEnabled,
	}
	created, err := svc.Create(cfg)
	require.NoError(t, err)

	t.Run("update email config", func(t *testing.T) {
		updateCfg := &model.EmailConfig{
			SMTPHost: "smtp-new.test.com",
			SMTPPort: 587,
			Username: "new@test.com",
			Password: "newpass",
			Sender:   "new@test.com",
			UseSSL:   false,
			Status:   model.ChannelDisabled,
		}
		updated, err := svc.Update(created.ID, updateCfg)
		require.NoError(t, err)
		assert.Equal(t, "smtp-new.test.com", updated.SMTPHost)
		assert.Equal(t, 587, updated.SMTPPort)
		assert.Equal(t, "new@test.com", updated.Username)
		assert.Equal(t, "newpass", updated.Password)
		assert.Equal(t, "new@test.com", updated.Sender)
		assert.False(t, updated.UseSSL)
		assert.Equal(t, model.ChannelDisabled, updated.Status)
	})

	t.Run("update nonexistent config", func(t *testing.T) {
		updateCfg := &model.EmailConfig{SMTPHost: "smtp.test.com", Username: "x@x.com"}
		updated, err := svc.Update(9999, updateCfg)
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.Contains(t, err.Error(), "配置不存在")
	})

	t.Run("update with partial fields", func(t *testing.T) {
		updateCfg := &model.EmailConfig{SMTPHost: "", SMTPPort: 0, Username: "", Password: "", Sender: "", UseSSL: true, Status: model.ChannelEnabled}
		updated, err := svc.Update(created.ID, updateCfg)
		require.NoError(t, err)
		assert.Equal(t, "smtp-new.test.com", updated.SMTPHost)
		assert.Equal(t, model.ChannelEnabled, updated.Status)
	})
}

func TestEmailService_Delete(t *testing.T) {
	setupEmailTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewEmailService()

	cfg := &model.EmailConfig{
		SMTPHost: "smtp.test.com",
		Username: "del@test.com",
		Password: "password",
		Sender:   "del@test.com",
		Status:   model.ChannelEnabled,
	}
	created, err := svc.Create(cfg)
	require.NoError(t, err)

	err = svc.Delete(created.ID)
	require.NoError(t, err)

	_, err = svc.Update(created.ID, &model.EmailConfig{SMTPHost: "x", Username: "y"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "配置不存在")
}

func TestEmailService_List(t *testing.T) {
	setupEmailTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewEmailService()

	for i := 0; i < 3; i++ {
		cfg := &model.EmailConfig{
			SMTPHost: "smtp" + string(rune('A'+i)) + ".test.com",
			Username: "user" + string(rune('0'+i)) + "@test.com",
			Password: "password",
			Sender:   "user" + string(rune('0'+i)) + "@test.com",
			Status:   model.ChannelEnabled,
		}
		_, err := svc.Create(cfg)
		require.NoError(t, err)
	}

	t.Run("list all configs", func(t *testing.T) {
		result, err := svc.List(1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 3)
	})

	t.Run("list with pagination", func(t *testing.T) {
		result, err := svc.List(1, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 2)

		result2, err := svc.List(2, 2)
		require.NoError(t, err)
		assert.Len(t, result2.Items, 1)
	})

	t.Run("list with default page", func(t *testing.T) {
		result, err := svc.List(0, 10)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Page)
	})

	t.Run("list with default pageSize", func(t *testing.T) {
		result, err := svc.List(1, 0)
		require.NoError(t, err)
		assert.Equal(t, 10, result.PageSize)
	})
}

func TestEmailService_SendCardEmailErrorPaths(t *testing.T) {
	setupEmailTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewEmailService()

	t.Run("send with no active config", func(t *testing.T) {
		err := svc.SendCardEmail([]string{"user@test.com"}, "Test", "Body", "ORD001")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "没有可用的邮件配置")
	})

	cfg := &model.EmailConfig{
		SMTPHost: "invalid-smtp.nonexistent.com",
		SMTPPort: 465,
		Username: "test@test.com",
		Password: "invalidpass",
		Sender:   "test@test.com",
		UseSSL:   false,
		Status:   model.ChannelEnabled,
	}
	_, err := svc.Create(cfg)
	require.NoError(t, err)

	t.Run("send with invalid SMTP config returns error but logs it", func(t *testing.T) {
		err := svc.SendCardEmail([]string{"user@test.com"}, "Test Subject", "Test Body", "ORD002")
		assert.Error(t, err)

		var logs []model.EmailLog
		require.NoError(t, database.DB.Find(&logs).Error)
		require.Len(t, logs, 1)
		assert.Equal(t, model.EmailLogFailed, logs[0].Status)
		assert.Contains(t, logs[0].ErrorMsg, "")
		assert.Equal(t, "user@test.com", logs[0].ToEmail)
		assert.Equal(t, "Test Subject", logs[0].Subject)
	})
}

func TestEmailService_DBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewEmailService()

	t.Run("create with nil DB", func(t *testing.T) {
		cfg := &model.EmailConfig{SMTPHost: "smtp.test.com", Username: "test@test.com", Password: "p", Sender: "s"}
		result, err := svc.Create(cfg)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "数据库未连接")
	})

	t.Run("update with nil DB", func(t *testing.T) {
		result, err := svc.Update(1, &model.EmailConfig{SMTPHost: "x", Username: "y"})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "数据库未连接")
	})

	t.Run("delete with nil DB", func(t *testing.T) {
		err := svc.Delete(1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "数据库未连接")
	})

	t.Run("list with nil DB", func(t *testing.T) {
		result, err := svc.List(1, 10)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(0), result.Total)
		assert.Len(t, result.Items, 0)
	})

	t.Run("sendCardEmail with nil DB", func(t *testing.T) {
		err := svc.SendCardEmail([]string{"user@test.com"}, "Test", "Body", "ORD001")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "数据库未连接")
	})
}