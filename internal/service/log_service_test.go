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

func dropAllLogTables(db *gorm.DB) {
	tables := []string{"users", "categories", "products", "cards", "orders", "payment_channels", "email_configs", "nodes", "operation_logs", "order_logs", "email_logs", "upgrade_logs", "file_uploads"}
	for _, tb := range tables {
		db.Migrator().DropTable(tb)
	}
}

func setupLogTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_fk=1"), &gorm.Config{})
	require.NoError(t, err)
	dropAllLogTables(db)
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

func TestLogService_WriteOperation(t *testing.T) {
	setupLogTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewLogService()

	t.Run("creates operation log record", func(t *testing.T) {
		svc.WriteOperation(1, "admin", "login", "user", "1", "用户登录", "127.0.0.1", "Mozilla/5.0", model.LogStatusSuccess)

		var logs []model.OperationLog
		require.NoError(t, database.DB.Find(&logs).Error)
		require.Len(t, logs, 1)
		assert.Equal(t, uint(1), logs[0].UserID)
		assert.Equal(t, "admin", logs[0].Username)
		assert.Equal(t, "login", logs[0].Action)
		assert.Equal(t, "user", logs[0].TargetType)
		assert.Equal(t, "1", logs[0].TargetID)
		assert.Equal(t, "用户登录", logs[0].Detail)
		assert.Equal(t, "127.0.0.1", logs[0].IP)
		assert.Equal(t, "Mozilla/5.0", logs[0].UserAgent)
		assert.Equal(t, model.LogStatusSuccess, logs[0].Status)
	})

	t.Run("writes multiple records", func(t *testing.T) {
		svc.WriteOperation(2, "user1", "create", "product", "5", "创建产品", "192.168.1.1", "curl", model.LogStatusSuccess)
		svc.WriteOperation(2, "user1", "update", "product", "5", "更新产品", "192.168.1.1", "curl", model.LogStatusSuccess)

		var logs []model.OperationLog
		require.NoError(t, database.DB.Order("id ASC").Find(&logs).Error)
		assert.Len(t, logs, 3)
	})
}

func TestLogService_WriteOperationDBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewLogService()
	assert.NotPanics(t, func() {
		svc.WriteOperation(1, "admin", "login", "user", "1", "detail", "ip", "ua", 1)
	})
}

func TestLogService_WriteOrder(t *testing.T) {
	setupLogTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewLogService()

	t.Run("creates order log record", func(t *testing.T) {
		svc.WriteOrder("ORD2024001", "create", "创建订单", "admin")

		var logs []model.OrderLog
		require.NoError(t, database.DB.Find(&logs).Error)
		require.Len(t, logs, 1)
		assert.Equal(t, "ORD2024001", logs[0].OrderNo)
		assert.Equal(t, "create", logs[0].Action)
		assert.Equal(t, "创建订单", logs[0].Detail)
		assert.Equal(t, "admin", logs[0].Operator)
	})
}

func TestLogService_WriteOrderDBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewLogService()
	assert.NotPanics(t, func() {
		svc.WriteOrder("ORD001", "create", "detail", "admin")
	})
}

func TestLogService_ListOperation(t *testing.T) {
	setupLogTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewLogService()

	svc.WriteOperation(1, "admin", "login", "user", "1", "登录", "127.0.0.1", "ua", 1)
	svc.WriteOperation(1, "admin", "create", "product", "2", "创建产品", "127.0.0.1", "ua", 1)
	svc.WriteOperation(2, "user1", "update", "order", "3", "更新订单", "192.168.1.1", "ua", 1)
	svc.WriteOperation(2, "user1", "delete", "order", "4", "删除订单", "192.168.1.1", "ua", 1)

	t.Run("list with no filters", func(t *testing.T) {
		result, err := svc.ListOperation(1, 10, "", "")
		require.NoError(t, err)
		assert.Equal(t, int64(4), result.Total)
		assert.Len(t, result.Items, 4)
	})

	t.Run("list with username filter", func(t *testing.T) {
		result, err := svc.ListOperation(1, 10, "admin", "")
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
		assert.Len(t, result.Items, 2)
	})

	t.Run("list with action filter", func(t *testing.T) {
		result, err := svc.ListOperation(1, 10, "", "login")
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Len(t, result.Items, 1)
	})

	t.Run("list with pagination", func(t *testing.T) {
		result, err := svc.ListOperation(1, 2, "", "")
		require.NoError(t, err)
		assert.Equal(t, int64(4), result.Total)
		assert.Len(t, result.Items, 2)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 2, result.PageSize)
	})

	t.Run("list second page", func(t *testing.T) {
		result, err := svc.ListOperation(2, 2, "", "")
		require.NoError(t, err)
		assert.Len(t, result.Items, 2)
	})

	t.Run("list with default page", func(t *testing.T) {
		result, err := svc.ListOperation(0, 10, "", "")
		require.NoError(t, err)
		assert.Equal(t, 1, result.Page)
	})

	t.Run("list with default pageSize", func(t *testing.T) {
		result, err := svc.ListOperation(1, 0, "", "")
		require.NoError(t, err)
		assert.Equal(t, 10, result.PageSize)
	})
}

func TestLogService_ListOperationDBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewLogService()
	result, err := svc.ListOperation(1, 10, "", "")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.Total)
	assert.Len(t, result.Items, 0)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.PageSize)
}

func TestLogService_ListOrder(t *testing.T) {
	setupLogTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewLogService()

	svc.WriteOrder("ORD001", "create", "创建", "admin")
	svc.WriteOrder("ORD001", "pay", "支付", "admin")
	svc.WriteOrder("ORD002", "create", "创建", "user1")

	t.Run("list all order logs", func(t *testing.T) {
		result, err := svc.ListOrder("", 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 3)
	})

	t.Run("list by order number", func(t *testing.T) {
		result, err := svc.ListOrder("ORD001", 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
		assert.Len(t, result.Items, 2)
	})

	t.Run("list with pagination", func(t *testing.T) {
		result, err := svc.ListOrder("", 1, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 2)
	})
}

func TestLogService_ListOrderDBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewLogService()
	result, err := svc.ListOrder("", 1, 10)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.Total)
	assert.Len(t, result.Items, 0)
}

func TestLogService_ListEmail(t *testing.T) {
	setupLogTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewLogService()

	database.DB.Create(&model.EmailLog{ToEmail: "user1@test.com", Subject: "Test", Status: 1})
	database.DB.Create(&model.EmailLog{ToEmail: "user2@test.com", Subject: "Test2", Status: 1})
	database.DB.Create(&model.EmailLog{ToEmail: "user1@test.com", Subject: "Test3", Status: 2})

	t.Run("list all email logs", func(t *testing.T) {
		result, err := svc.ListEmail(1, 10, "")
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 3)
	})

	t.Run("list by email filter", func(t *testing.T) {
		result, err := svc.ListEmail(1, 10, "user1")
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
		assert.Len(t, result.Items, 2)
	})

	t.Run("list with pagination", func(t *testing.T) {
		result, err := svc.ListEmail(1, 2, "")
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 2)
	})
}

func TestLogService_ListEmailDBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewLogService()
	result, err := svc.ListEmail(1, 10, "")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.Total)
	assert.Len(t, result.Items, 0)
}

func TestLogService_ListLogin(t *testing.T) {
	setupLogTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewLogService()

	svc.WriteOperation(1, "admin", "login", "user", "1", "登录成功", "127.0.0.1", "ua", 1)
	svc.WriteOperation(2, "user1", "login", "user", "2", "登录成功", "192.168.1.1", "ua", 1)
	svc.WriteOperation(1, "admin", "create", "product", "3", "创建产品", "127.0.0.1", "ua", 1)

	t.Run("list login records", func(t *testing.T) {
		result, err := svc.ListLogin(1, 10, "")
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
		assert.Len(t, result.Items, 2)
		for _, item := range result.Items {
			assert.Equal(t, "login", item.Action)
		}
	})

	t.Run("list login with username filter", func(t *testing.T) {
		result, err := svc.ListLogin(1, 10, "admin")
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Len(t, result.Items, 1)
	})

	t.Run("list login with pagination", func(t *testing.T) {
		result, err := svc.ListLogin(1, 1, "")
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
		assert.Len(t, result.Items, 1)
	})
}

func TestLogService_ListLoginDBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewLogService()
	result, err := svc.ListLogin(1, 10, "")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "数据库未连接")
}