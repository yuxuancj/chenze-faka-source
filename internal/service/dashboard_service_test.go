package service

import (
	"testing"
	"time"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func dropAllDashboardTables(db *gorm.DB) {
	tables := []string{"users", "categories", "products", "cards", "orders", "payment_channels", "email_configs", "nodes", "operation_logs", "order_logs", "email_logs", "upgrade_logs", "file_uploads"}
	for _, tb := range tables {
		db.Migrator().DropTable(tb)
	}
}

func setupDashboardTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_fk=1"), &gorm.Config{})
	require.NoError(t, err)
	dropAllDashboardTables(db)
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

func TestDashboardService_GetStats(t *testing.T) {
	setupDashboardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewDashboardService()

	now := time.Now()

	user := model.User{Username: "admin", PasswordHash: "hash", Salt: "salt"}
	require.NoError(t, database.DB.Create(&user).Error)

	product := model.Product{Name: "测试产品", Price: 99.99, Stock: 100, Status: model.ProductStatusOnShelf}
	require.NoError(t, database.DB.Create(&product).Error)

	order1 := model.Order{
		OrderNo:     "ORD001",
		ProductID:   product.ID,
		ProductName: "测试产品",
		Quantity:    2,
		Price:       99.99,
		TotalAmount: 199.98,
		PayMethod:   "alipay",
		Status:      model.OrderStatusPaid,
		PaidAt:      &now,
		CreatedAt:   now,
	}
	require.NoError(t, database.DB.Create(&order1).Error)

	order2 := model.Order{
		OrderNo:     "ORD002",
		ProductID:   product.ID,
		ProductName: "测试产品",
		Quantity:    1,
		Price:       99.99,
		TotalAmount: 99.99,
		PayMethod:   "wechat",
		Status:      model.OrderStatusPaid,
		CreatedAt:   now,
	}
	require.NoError(t, database.DB.Create(&order2).Error)

	card := model.Card{ProductID: product.ID, Status: model.CardStatusSold, OrderNo: "ORD001", SoldAt: &now}
	require.NoError(t, database.DB.Create(&card).Error)

	t.Run("get stats with data", func(t *testing.T) {
		stats, err := svc.GetStats()
		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.Equal(t, int64(1), stats.ProductCount)
		assert.Equal(t, int64(2), stats.OrderCount)
		assert.Equal(t, int64(1), stats.CardSoldCount)
		assert.Equal(t, int64(1), stats.UserCount)

		assert.True(t, stats.TotalRevenue > 0)
		assert.True(t, stats.TodayOrders >= 0)
		assert.True(t, stats.TodayRevenue >= 0)
		assert.True(t, stats.MonthOrders >= 0)
		assert.True(t, stats.MonthRevenue >= 0)

		assert.Len(t, stats.RecentOrders, 2)
		assert.Len(t, stats.TopProducts, 1)
		assert.Len(t, stats.PayMethodStats, 2)
		assert.NotEmpty(t, stats.SalesTrend)

		for _, tp := range stats.TopProducts {
			if tp.ID == product.ID {
				assert.Equal(t, "测试产品", tp.Name)
				assert.True(t, tp.SoldQty > 0)
				assert.True(t, tp.Revenue > 0)
			}
		}

		for _, ps := range stats.PayMethodStats {
			assert.Contains(t, []string{"alipay", "wechat"}, ps.Method)
		}

		for _, sp := range stats.SalesTrend {
			assert.NotEmpty(t, sp.Date)
		}
	})

	t.Run("stats for today's orders verified", func(t *testing.T) {
		stats, err := svc.GetStats()
		require.NoError(t, err)
		assert.NotNil(t, stats)
	})
}

func TestDashboardService_GetStatsDBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewDashboardService()
	stats, err := svc.GetStats()
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.ProductCount)
	assert.Equal(t, int64(0), stats.OrderCount)
	assert.Equal(t, int64(0), stats.CardSoldCount)
	assert.Equal(t, int64(0), stats.UserCount)
}

func TestDashboardService_GetOrderStatusCounts(t *testing.T) {
	setupDashboardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewDashboardService()

	order1 := model.Order{OrderNo: "ORD001", ProductName: "P1", PayMethod: "alipay", TotalAmount: 10, Status: model.OrderStatusPending}
	order2 := model.Order{OrderNo: "ORD002", ProductName: "P1", PayMethod: "alipay", TotalAmount: 20, Status: model.OrderStatusPaid}
	order3 := model.Order{OrderNo: "ORD003", ProductName: "P1", PayMethod: "alipay", TotalAmount: 30, Status: model.OrderStatusComplete}
	order4 := model.Order{OrderNo: "ORD004", ProductName: "P1", PayMethod: "alipay", TotalAmount: 40, Status: model.OrderStatusCancel}

	require.NoError(t, database.DB.Create(&order1).Error)
	require.NoError(t, database.DB.Create(&order2).Error)
	require.NoError(t, database.DB.Create(&order3).Error)
	require.NoError(t, database.DB.Create(&order4).Error)

	counts, err := svc.GetOrderStatusCounts()
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts["pending"])
	assert.Equal(t, int64(1), counts["paid"])
	assert.Equal(t, int64(1), counts["complete"])
	assert.Equal(t, int64(1), counts["cancel"])
}

func TestDashboardService_GetOrderStatusCountsDBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewDashboardService()
	counts, err := svc.GetOrderStatusCounts()
	assert.Error(t, err)
	assert.Nil(t, counts)
	assert.Contains(t, err.Error(), "数据库未连接")
}

func TestDashboardService_GetSalesTrend(t *testing.T) {
	setupDashboardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewDashboardService()

	now := time.Now()
	for i := 0; i < 3; i++ {
		order := model.Order{
			OrderNo:     "ORD" + string(rune('0'+i)),
			ProductName: "产品" + string(rune('0'+i)),
			PayMethod:   "alipay",
			TotalAmount: float64((i + 1) * 100),
			Status:      model.OrderStatusPaid,
			CreatedAt:   now,
		}
		require.NoError(t, database.DB.Create(&order).Error)
	}

	stats, err := svc.GetStats()
	require.NoError(t, err)
	require.NotNil(t, stats)
	require.Len(t, stats.SalesTrend, 7)

	for _, tp := range stats.SalesTrend {
		assert.NotEmpty(t, tp.Date)
	}

	foundPaid := false
	for _, tp := range stats.SalesTrend {
		if tp.Orders > 0 {
			foundPaid = true
			assert.True(t, tp.Revenue > 0)
		}
	}
	assert.True(t, foundPaid)
}