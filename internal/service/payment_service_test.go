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

func dropAllPaymentTables(db *gorm.DB) {
	tables := []string{"users", "categories", "products", "cards", "orders", "payment_channels", "email_configs", "nodes", "operation_logs", "order_logs", "email_logs", "upgrade_logs", "file_uploads"}
	for _, tb := range tables {
		db.Migrator().DropTable(tb)
	}
}

func setupPaymentTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_fk=1"), &gorm.Config{})
	require.NoError(t, err)
	dropAllPaymentTables(db)
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

func TestPaymentService_Create(t *testing.T) {
	setupPaymentTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewPaymentService()

	t.Run("create payment channel successfully", func(t *testing.T) {
		ch, err := svc.Create("支付宝", model.PayTypeAlipay, "alipay.png", `{"app_id":"123"}`, 0.01, 1)
		require.NoError(t, err)
		assert.NotNil(t, ch)
		assert.Equal(t, "支付宝", ch.Name)
		assert.Equal(t, model.PayTypeAlipay, ch.Type)
		assert.Equal(t, "alipay.png", ch.Icon)
		assert.Equal(t, `{"app_id":"123"}`, ch.Config)
		assert.Equal(t, 0.01, ch.FeeRate)
		assert.Equal(t, model.ChannelEnabled, ch.Status)
		assert.Equal(t, 1, ch.Sort)
	})

	t.Run("create with empty name fails", func(t *testing.T) {
		ch, err := svc.Create("", model.PayTypeAlipay, "", "", 0, 1)
		assert.Error(t, err)
		assert.Nil(t, ch)
		assert.Contains(t, err.Error(), "支付名称不能为空")
	})

	t.Run("create with empty icon and config succeeds", func(t *testing.T) {
		ch, err := svc.Create("微信支付", model.PayTypeWechat, "", "", 0, 2)
		require.NoError(t, err)
		assert.NotNil(t, ch)
		assert.Equal(t, "", ch.Icon)
		assert.Equal(t, "", ch.Config)
	})

	t.Run("create with zero fee rate", func(t *testing.T) {
		ch, err := svc.Create("Stripe", model.PayTypeStripe, "", "", 0, 3)
		require.NoError(t, err)
		assert.Equal(t, 0.0, ch.FeeRate)
	})
}

func TestPaymentService_Update(t *testing.T) {
	setupPaymentTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewPaymentService()

	ch, err := svc.Create("支付宝", model.PayTypeAlipay, "alipay.png", `{"app_id":"123"}`, 0.01, 1)
	require.NoError(t, err)

	t.Run("update name and config", func(t *testing.T) {
		updated, err := svc.Update(ch.ID, "支付宝V2", "new_icon.png", `{"app_id":"456"}`, 0.02, model.ChannelEnabled, 5)
		require.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, "支付宝V2", updated.Name)
		assert.Equal(t, "new_icon.png", updated.Icon)
		assert.Equal(t, `{"app_id":"456"}`, updated.Config)
		assert.Equal(t, 0.02, updated.FeeRate)
		assert.Equal(t, 5, updated.Sort)
	})

	t.Run("update with empty name keeps old name", func(t *testing.T) {
		updated, err := svc.Update(ch.ID, "", "", "", 0.05, model.ChannelEnabled, 3)
		require.NoError(t, err)
		assert.Equal(t, "支付宝V2", updated.Name)
	})

	t.Run("update nonexistent channel", func(t *testing.T) {
		updated, err := svc.Update(9999, "不存在", "", "", 0, model.ChannelEnabled, 1)
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.Contains(t, err.Error(), "支付接口不存在")
	})

	t.Run("update status to disabled", func(t *testing.T) {
		updated, err := svc.Update(ch.ID, "", "", "", 0, model.ChannelDisabled, 1)
		require.NoError(t, err)
		assert.Equal(t, model.ChannelDisabled, updated.Status)
	})

	t.Run("update status to enabled", func(t *testing.T) {
		updated, err := svc.Update(ch.ID, "", "", "", 0, model.ChannelEnabled, 1)
		require.NoError(t, err)
		assert.Equal(t, model.ChannelEnabled, updated.Status)
	})
}

func TestPaymentService_Delete(t *testing.T) {
	setupPaymentTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewPaymentService()

	ch, err := svc.Create("待删除支付", model.PayTypeCustom, "", "", 0, 1)
	require.NoError(t, err)

	err = svc.Delete(ch.ID)
	require.NoError(t, err)

	_, err = svc.Update(ch.ID, "test", "", "", 0, model.ChannelEnabled, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "支付接口不存在")
}

func TestPaymentService_GetActive(t *testing.T) {
	setupPaymentTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewPaymentService()

	ch1, err := svc.Create("支付宝", model.PayTypeAlipay, "", "", 0, 1)
	require.NoError(t, err)
	_, err = svc.Create("微信", model.PayTypeWechat, "", "", 0, 2)
	require.NoError(t, err)
	ch3, err := svc.Create("Stripe", model.PayTypeStripe, "", "", 0, 3)
	require.NoError(t, err)

	_, err = svc.Update(ch3.ID, "", "", "", 0, model.ChannelDisabled, 3)
	require.NoError(t, err)

	active, err := svc.GetActive()
	require.NoError(t, err)
	assert.Len(t, active, 2)
	assert.Equal(t, "支付宝", active[0].Name)
	assert.Equal(t, "微信", active[1].Name)
	_ = ch1
}

func TestPaymentService_List(t *testing.T) {
	setupPaymentTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewPaymentService()

	_, err := svc.Create("支付宝", model.PayTypeAlipay, "", "", 0, 1)
	require.NoError(t, err)
	_, err = svc.Create("微信", model.PayTypeWechat, "", "", 0, 2)
	require.NoError(t, err)
	_, err = svc.Create("Stripe", model.PayTypeStripe, "", "", 0, 3)
	require.NoError(t, err)

	t.Run("list all", func(t *testing.T) {
		result, err := svc.List(1, 10, "")
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 3)
	})

	t.Run("list with keyword", func(t *testing.T) {
		result, err := svc.List(1, 10, "支付宝")
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Len(t, result.Items, 1)
	})

	t.Run("list with pagination", func(t *testing.T) {
		result, err := svc.List(1, 2, "")
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 2)

		result2, err := svc.List(2, 2, "")
		require.NoError(t, err)
		assert.Len(t, result2.Items, 1)
	})

	t.Run("list with default page", func(t *testing.T) {
		result, err := svc.List(0, 10, "")
		require.NoError(t, err)
		assert.Equal(t, 1, result.Page)
	})

	t.Run("list with default pageSize", func(t *testing.T) {
		result, err := svc.List(1, 0, "")
		require.NoError(t, err)
		assert.Equal(t, 10, result.PageSize)
	})
}

func TestPaymentService_DBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewPaymentService()

	t.Run("create with nil DB", func(t *testing.T) {
		ch, err := svc.Create("test", "alipay", "", "", 0, 1)
		assert.Error(t, err)
		assert.Nil(t, ch)
		assert.Contains(t, err.Error(), "数据库未连接")
	})

	t.Run("update with nil DB", func(t *testing.T) {
		ch, err := svc.Update(1, "test", "", "", 0, 1, 1)
		assert.Error(t, err)
		assert.Nil(t, ch)
		assert.Contains(t, err.Error(), "数据库未连接")
	})

	t.Run("delete with nil DB", func(t *testing.T) {
		err := svc.Delete(1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "数据库未连接")
	})

	t.Run("getActive with nil DB", func(t *testing.T) {
		channels, err := svc.GetActive()
		require.NoError(t, err)
		assert.Len(t, channels, 0)
	})

	t.Run("list with nil DB", func(t *testing.T) {
		result, err := svc.List(1, 10, "")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(0), result.Total)
		assert.Len(t, result.Items, 0)
	})
}