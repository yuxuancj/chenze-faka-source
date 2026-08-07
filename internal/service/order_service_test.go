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

func dropAllOrderTables(db *gorm.DB) {
	tables := []string{"users", "categories", "products", "cards", "orders", "payment_channels", "email_configs", "nodes", "operation_logs", "order_logs", "email_logs", "upgrade_logs", "file_uploads"}
	for _, t := range tables {
		db.Migrator().DropTable(t)
	}
}

func setupOrderTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	dropAllOrderTables(db)
	database.DB = db
	require.NoError(t, db.AutoMigrate(
		&model.User{}, &model.Category{}, &model.Product{},
		&model.Card{}, &model.Order{}, &model.PaymentChannel{},
		&model.EmailConfig{}, &model.Node{},
		&model.OperationLog{}, &model.OrderLog{}, &model.EmailLog{},
		&model.UpgradeLog{}, &model.FileUpload{},
	))
}

func createTestProductForOrder(t *testing.T, name string, price float64, stock int) *model.Product {
	t.Helper()
	product := &model.Product{
		Name:        name,
		Category:    "测试分类",
		Price:       price,
		Description: "测试产品描述",
		Stock:       stock,
		Status:      model.ProductStatusOnShelf,
		Sort:        1,
	}
	require.NoError(t, database.DB.Create(product).Error)
	return product
}

func createTestCardsForOrder(t *testing.T, productID uint, count int) []model.Card {
	t.Helper()
	cards := make([]model.Card, count)
	for i := 0; i < count; i++ {
		cards[i] = model.Card{
			ProductID:  productID,
			CardNoHash: "order_test_card_hash_" + string(rune('a'+i)) + "_" + time.Now().Format("150405.000"),
			Status:     model.CardStatusUnsold,
		}
		require.NoError(t, database.DB.Create(&cards[i]).Error)
	}
	return cards
}

func defaultPayConfigForOrder() model.PayConfig {
	return model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "test_merchant",
		Key:      "test_key",
	}
}

func TestOrderService_CreateOrder(t *testing.T) {
	setupOrderTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := defaultPayConfigForOrder()

	t.Run("create order with valid data", func(t *testing.T) {
		product := createTestProductForOrder(t, "测试商品A", 10.00, 100)

		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  2,
			Contact:   "test@example.com",
			PayMethod: "alipay",
		}

		result, err := svc.CreateOrder(req, payConfig)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.OrderNo)
		assert.Equal(t, 20.00, result.Amount)
		assert.NotEmpty(t, result.PayURL)
		assert.NotEmpty(t, result.ExpireAt)
	})

	t.Run("create order with quantity <= 0", func(t *testing.T) {
		product := createTestProductForOrder(t, "测试商品B", 10.00, 100)

		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  0,
			Contact:   "test@example.com",
			PayMethod: "alipay",
		}

		result, err := svc.CreateOrder(req, payConfig)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "数量必须大于0")
	})

	t.Run("create order with quantity > 99", func(t *testing.T) {
		product := createTestProductForOrder(t, "测试商品C", 10.00, 200)

		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  100,
			Contact:   "test@example.com",
			PayMethod: "alipay",
		}

		result, err := svc.CreateOrder(req, payConfig)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "单次购买数量不能超过99")
	})

	t.Run("create order with empty contact", func(t *testing.T) {
		product := createTestProductForOrder(t, "测试商品D", 10.00, 100)

		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  1,
			Contact:   "",
			PayMethod: "alipay",
		}

		result, err := svc.CreateOrder(req, payConfig)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "联系方式不能为空")
	})

	t.Run("create order with invalid pay method", func(t *testing.T) {
		product := createTestProductForOrder(t, "测试商品E", 10.00, 100)

		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  1,
			Contact:   "test@example.com",
			PayMethod: "bitcoin",
		}

		result, err := svc.CreateOrder(req, payConfig)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "无效的支付方式")
	})

	t.Run("create order with nonexistent product", func(t *testing.T) {
		req := &CreateOrderRequest{
			ProductID: 9999,
			Quantity:  1,
			Contact:   "test@example.com",
			PayMethod: "alipay",
		}

		result, err := svc.CreateOrder(req, payConfig)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "产品不存在")
	})

	t.Run("create order with insufficient stock", func(t *testing.T) {
		product := createTestProductForOrder(t, "测试商品F", 10.00, 1)

		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  5,
			Contact:   "test@example.com",
			PayMethod: "alipay",
		}

		result, err := svc.CreateOrder(req, payConfig)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "库存不足")
	})
}

func TestOrderService_HandlePaymentCallback(t *testing.T) {
	setupOrderTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := defaultPayConfigForOrder()

	product := createTestProductForOrder(t, "回调测试商品", 15.00, 50)
	createTestCardsForOrder(t, product.ID, 10)

	t.Run("normal flow", func(t *testing.T) {
		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  2,
			Contact:   "callback@example.com",
			PayMethod: "alipay",
		}
		result, err := svc.CreateOrder(req, payConfig)
		require.NoError(t, err)
		require.NotNil(t, result)

		ok, err := svc.HandlePaymentCallback("pay_001", result.OrderNo, result.Amount)
		require.NoError(t, err)
		assert.True(t, ok)

		var order model.Order
		require.NoError(t, database.DB.Where("order_no = ?", result.OrderNo).First(&order).Error)
		assert.Equal(t, model.OrderStatusComplete, order.Status)
	})

	t.Run("already completed order returns false", func(t *testing.T) {
		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  1,
			Contact:   "completed@example.com",
			PayMethod: "alipay",
		}
		result, err := svc.CreateOrder(req, payConfig)
		require.NoError(t, err)

		ok, err := svc.HandlePaymentCallback("pay_002", result.OrderNo, result.Amount)
		require.NoError(t, err)
		assert.True(t, ok)

		ok, err = svc.HandlePaymentCallback("pay_003", result.OrderNo, result.Amount)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("amount mismatch", func(t *testing.T) {
		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  1,
			Contact:   "mismatch@example.com",
			PayMethod: "alipay",
		}
		result, err := svc.CreateOrder(req, payConfig)
		require.NoError(t, err)

		ok, err := svc.HandlePaymentCallback("pay_004", result.OrderNo, 999.99)
		assert.Error(t, err)
		assert.False(t, ok)
		assert.Contains(t, err.Error(), "金额不匹配")
	})

	t.Run("insufficient cards", func(t *testing.T) {
		smallProduct := createTestProductForOrder(t, "无卡密商品", 20.00, 5)

		req := &CreateOrderRequest{
			ProductID: smallProduct.ID,
			Quantity:  3,
			Contact:   "nocard@example.com",
			PayMethod: "alipay",
		}
		result, err := svc.CreateOrder(req, payConfig)
		require.NoError(t, err)

		ok, err := svc.HandlePaymentCallback("pay_005", result.OrderNo, result.Amount)
		assert.Error(t, err)
		assert.False(t, ok)
		assert.Contains(t, err.Error(), "可用卡密不足")
	})
}

func TestOrderService_QueryOrder(t *testing.T) {
	setupOrderTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := defaultPayConfigForOrder()

	product := createTestProductForOrder(t, "查询测试商品", 10.00, 100)

	req := &CreateOrderRequest{
		ProductID: product.ID,
		Quantity:  2,
		Contact:   "query@example.com",
		PayMethod: "alipay",
	}
	createResult, err := svc.CreateOrder(req, payConfig)
	require.NoError(t, err)

	t.Run("query by orderNo", func(t *testing.T) {
		result, err := svc.QueryOrder(createResult.OrderNo, "")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, createResult.OrderNo, result.OrderNo)
		assert.Equal(t, 2, result.Quantity)
		assert.Equal(t, 20.00, result.Amount)
	})

	t.Run("query by contact", func(t *testing.T) {
		result, err := svc.QueryOrder("", "query@example.com")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "query@example.com", result.Contact)
	})

	t.Run("query with no params", func(t *testing.T) {
		result, err := svc.QueryOrder("", "")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "订单号或联系方式不能为空")
	})

	t.Run("query nonexistent order", func(t *testing.T) {
		result, err := svc.QueryOrder("NONEXISTENT123", "")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestOrderService_List(t *testing.T) {
	setupOrderTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := defaultPayConfigForOrder()

	product := createTestProductForOrder(t, "列表测试商品", 10.00, 100)

	for i := 0; i < 5; i++ {
		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  1,
			Contact:   "list@example.com",
			PayMethod: "alipay",
		}
		_, err := svc.CreateOrder(req, payConfig)
		require.NoError(t, err)
	}

	t.Run("list with default params", func(t *testing.T) {
		result, err := svc.List(1, 10, -1, "")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(5), result.Total)
		assert.Len(t, result.Orders, 5)
	})

	t.Run("list with pagination", func(t *testing.T) {
		result, err := svc.List(1, 3, -1, "")
		require.NoError(t, err)
		assert.Len(t, result.Orders, 3)

		result2, err := svc.List(2, 3, -1, "")
		require.NoError(t, err)
		assert.Len(t, result2.Orders, 2)
	})

	t.Run("list with status filter", func(t *testing.T) {
		result, err := svc.List(1, 10, model.OrderStatusPending, "")
		require.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)

		result, err = svc.List(1, 10, model.OrderStatusComplete, "")
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.Total)
	})

	t.Run("list with keyword filter", func(t *testing.T) {
		result, err := svc.List(1, 10, -1, "list@example")
		require.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)

		result, err = svc.List(1, 10, -1, "nonexistent")
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.Total)
	})
}

func TestOrderService_AutoCloseExpiredOrders(t *testing.T) {
	setupOrderTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := defaultPayConfigForOrder()

	product := createTestProductForOrder(t, "过期测试商品", 10.00, 100)

	req := &CreateOrderRequest{
		ProductID: product.ID,
		Quantity:  1,
		Contact:   "expire@example.com",
		PayMethod: "alipay",
	}
	result, err := svc.CreateOrder(req, payConfig)
	require.NoError(t, err)

	require.NoError(t, database.DB.Model(&model.Order{}).
		Where("order_no = ?", result.OrderNo).
		Update("expired_at", time.Now().Add(-1*time.Hour)).Error)

	err = svc.AutoCloseExpiredOrders()
	require.NoError(t, err)

	var order model.Order
	require.NoError(t, database.DB.Where("order_no = ?", result.OrderNo).First(&order).Error)
	assert.Equal(t, model.OrderStatusCancel, order.Status)
}