package service

import (
	"testing"
	"time"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
	"chenze-faka/internal/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupOrderCoverageTestDB(t *testing.T) *gorm.DB {
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

func createOrderCoverageProduct(t *testing.T, name string, price float64, stock int) *model.Product {
	t.Helper()
	product := &model.Product{
		Name:     name,
		Category: "Coverage分类",
		Price:    price,
		Stock:    stock,
		Status:   model.ProductStatusOnShelf,
		Sort:     1,
	}
	require.NoError(t, database.DB.Create(product).Error)
	return product
}

func createOrderCoverageCards(t *testing.T, productID uint, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		card := &model.Card{
			ProductID:  productID,
			CardNoHash: "order_cov_card_" + string(rune('a'+i)) + "_" + time.Now().Format("150405.000"),
			Status:     model.CardStatusUnsold,
		}
		require.NoError(t, database.DB.Create(card).Error)
	}
}

func coveragePayConfig() model.PayConfig {
	return model.PayConfig{
		URL:      "https://pay.example.com",
		Merchant: "cov_merchant",
		Key:      "cov_test_key",
	}
}

func TestQueryOrderByContactCoverage(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()

	product := createOrderCoverageProduct(t, "ContactCoverage商品", 10.00, 100)
	createOrderCoverageCards(t, product.ID, 5)

	req := &CreateOrderRequest{
		ProductID: product.ID,
		Quantity:  1,
		Contact:   "contact_query@test.com",
		PayMethod: "alipay",
	}
	createResult, err := svc.CreateOrder(req, payConfig)
	require.NoError(t, err)

	t.Run("query by contact finds the latest order", func(t *testing.T) {
		result, err := svc.QueryOrder("", "contact_query@test.com")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "contact_query@test.com", result.Contact)
		assert.Equal(t, createResult.OrderNo, result.OrderNo)
	})

	t.Run("query by nonexistent contact returns error", func(t *testing.T) {
		result, err := svc.QueryOrder("", "nobody@test.com")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "订单不存在")
	})
}

func TestQueryOrderBothEmptyCoverage(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()

	result, err := svc.QueryOrder("", "")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "订单号或联系方式不能为空")
}

func TestListDefaultParamsCoverage(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()

	product := createOrderCoverageProduct(t, "ListDefault商品", 10.00, 100)

	for i := 0; i < 5; i++ {
		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  1,
			Contact:   "list_default@test.com",
			PayMethod: "alipay",
		}
		_, err := svc.CreateOrder(req, payConfig)
		require.NoError(t, err)
	}

	t.Run("list with zero page and pageSize defaults to 1 and 10", func(t *testing.T) {
		result, err := svc.List(0, 0, -1, "")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(5), result.Total)
		assert.Len(t, result.Orders, 5)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 10, result.PageSize)
	})
}

func TestListStatusFilterCoverage(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()

	product := createOrderCoverageProduct(t, "StatusFilter商品", 10.00, 100)
	createOrderCoverageCards(t, product.ID, 10)

	for i := 0; i < 3; i++ {
		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  1,
			Contact:   "status_filter@test.com",
			PayMethod: "alipay",
		}
		_, err := svc.CreateOrder(req, payConfig)
		require.NoError(t, err)
	}

	t.Run("filter by pending status returns all", func(t *testing.T) {
		result, err := svc.List(1, 10, model.OrderStatusPending, "")
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
	})

	t.Run("filter by completed status returns none", func(t *testing.T) {
		result, err := svc.List(1, 10, model.OrderStatusComplete, "")
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.Total)
	})

	t.Run("filter by cancelled status returns none", func(t *testing.T) {
		result, err := svc.List(1, 10, model.OrderStatusCancel, "")
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.Total)
	})
}

func TestListKeywordFilterCoverage(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()

	product := createOrderCoverageProduct(t, "KeywordFilter商品", 10.00, 100)

	for i := 0; i < 3; i++ {
		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  1,
			Contact:   "keyword_user@test.com",
			PayMethod: "alipay",
		}
		_, err := svc.CreateOrder(req, payConfig)
		require.NoError(t, err)
	}

	t.Run("filter by keyword in contact", func(t *testing.T) {
		result, err := svc.List(1, 10, -1, "keyword")
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
	})

	t.Run("filter by nonexistent keyword returns empty", func(t *testing.T) {
		result, err := svc.List(1, 10, -1, "nonexistent_keyword")
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.Total)
	})
}

func TestVerifyNotifyCoverage(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()

	t.Run("verify notify with empty pay key returns false", func(t *testing.T) {
		params := map[string]string{
			"order_no": "test",
			"sign":     "abc",
		}
		result := svc.VerifyNotify(params, "")
		assert.False(t, result)
	})

	t.Run("verify notify with invalid sign returns false", func(t *testing.T) {
		params := map[string]string{
			"order_no": "ORDER-TEST-001",
			"money":    "10.00",
			"sign":     "invalid_sign",
			"sign_type": "MD5",
		}
		result := svc.VerifyNotify(params, "test_key_12345")
		assert.False(t, result)
	})

	t.Run("verify notify with valid sign returns true", func(t *testing.T) {
		params := map[string]string{
			"order_no":  "ORDER-VALID-001",
			"money":     "10.00",
			"sign_type": "MD5",
		}
		calcSign := utils.MD5Sum("money=10.00&order_no=ORDER-VALID-001&key=test_key_12345")
		params["sign"] = calcSign

		result := svc.VerifyNotify(params, "test_key_12345")
		assert.True(t, result)
	})
}

func TestAutoCloseExpiredOrdersCoverage(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()

	product := createOrderCoverageProduct(t, "ExpiredCoverage商品", 10.00, 100)
	createOrderCoverageCards(t, product.ID, 5)

	t.Run("auto close expired orders changes status to cancel", func(t *testing.T) {
		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  1,
			Contact:   "expire@test.com",
			PayMethod: "alipay",
		}
		createResult, err := svc.CreateOrder(req, payConfig)
		require.NoError(t, err)

		require.NoError(t, database.DB.Model(&model.Order{}).
			Where("order_no = ?", createResult.OrderNo).
			Update("expired_at", time.Now().Add(-1*time.Hour)).Error)

		err = svc.AutoCloseExpiredOrders()
		require.NoError(t, err)

		var order model.Order
		require.NoError(t, database.DB.Where("order_no = ?", createResult.OrderNo).First(&order).Error)
		assert.Equal(t, model.OrderStatusCancel, order.Status)
	})

	t.Run("auto close with no expired orders does nothing", func(t *testing.T) {
		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  1,
			Contact:   "active@test.com",
			PayMethod: "alipay",
		}
		createResult, err := svc.CreateOrder(req, payConfig)
		require.NoError(t, err)

		err = svc.AutoCloseExpiredOrders()
		require.NoError(t, err)

		var order model.Order
		require.NoError(t, database.DB.Where("order_no = ?", createResult.OrderNo).First(&order).Error)
		assert.Equal(t, model.OrderStatusPending, order.Status)
	})
}

func TestQueryOrderCompletedStatusCoverage(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()

	product := createOrderCoverageProduct(t, "CompletedStatus商品", 10.00, 100)
	createOrderCoverageCards(t, product.ID, 5)

	req := &CreateOrderRequest{
		ProductID: product.ID,
		Quantity:  2,
		Contact:   "completed@test.com",
		PayMethod: "alipay",
	}
	createResult, err := svc.CreateOrder(req, payConfig)
	require.NoError(t, err)

	_, err = svc.HandlePaymentCallback("pay_completed", createResult.OrderNo, createResult.Amount)
	require.NoError(t, err)

	result, err := svc.QueryOrder(createResult.OrderNo, "")
	require.NoError(t, err)
	assert.Equal(t, model.OrderStatusComplete, result.Status)
	assert.Equal(t, "已完成", result.StatusText)
	assert.NotEmpty(t, result.Cards)
	assert.NotEmpty(t, result.PaidAt)
}

func TestHandlePaymentCallbackAmountMismatchCoverage(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()

	product := createOrderCoverageProduct(t, "AmountMismatch商品", 10.00, 100)
	createOrderCoverageCards(t, product.ID, 5)

	req := &CreateOrderRequest{
		ProductID: product.ID,
		Quantity:  1,
		Contact:   "mismatch@test.com",
		PayMethod: "alipay",
	}
	createResult, err := svc.CreateOrder(req, payConfig)
	require.NoError(t, err)

	t.Run("amount mismatch returns error", func(t *testing.T) {
		ok, err := svc.HandlePaymentCallback("pay_mismatch", createResult.OrderNo, 999.99)
		assert.Error(t, err)
		assert.False(t, ok)
		assert.Contains(t, err.Error(), "金额不匹配")
	})

	t.Run("nonexistent order returns error", func(t *testing.T) {
		ok, err := svc.HandlePaymentCallback("pay_no_order", "NONEXISTENT_ORDER", 10.00)
		assert.Error(t, err)
		assert.False(t, ok)
		assert.Contains(t, err.Error(), "订单不存在")
	})
}

func TestOrderServiceDBNilCoverage(t *testing.T) {
	svc := NewOrderService()
	payConfig := model.PayConfig{URL: "https://pay.example.com", Merchant: "m", Key: "k"}

	t.Run("create order with DB nil returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		req := &CreateOrderRequest{ProductID: 1, Quantity: 1, Contact: "test@test.com", PayMethod: "alipay"}
		result, err := svc.CreateOrder(req, payConfig)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("query order with DB nil returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		result, err := svc.QueryOrder("ORDER-001", "")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("list with DB nil returns empty", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		result, err := svc.List(1, 10, -1, "")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("auto close with DB nil returns nil", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		err := svc.AutoCloseExpiredOrders()
		assert.NoError(t, err)
	})

	t.Run("handle payment callback with DB nil returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		ok, err := svc.HandlePaymentCallback("pay_no", "ORDER-001", 10.00)
		assert.Error(t, err)
		assert.False(t, ok)
	})
}

func TestOrderCreateInsufficientStockCoverage(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()

	product := createOrderCoverageProduct(t, "LowStock商品", 10.00, 1)
	createOrderCoverageCards(t, product.ID, 1)

	t.Run("create order with quantity exceeding available stock fails", func(t *testing.T) {
		req := &CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  5,
			Contact:   "lowstock@test.com",
			PayMethod: "alipay",
		}
		result, err := svc.CreateOrder(req, payConfig)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestOrderListStatusAndKeywordCoverage(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()

	product := createOrderCoverageProduct(t, "ListFilter商品", 10.00, 100)
	createOrderCoverageCards(t, product.ID, 10)

	req1 := &CreateOrderRequest{ProductID: product.ID, Quantity: 1, Contact: "filter1@test.com", PayMethod: "alipay"}
	_, err := svc.CreateOrder(req1, payConfig)
	require.NoError(t, err)

	req2 := &CreateOrderRequest{ProductID: product.ID, Quantity: 1, Contact: "filter2@test.com", PayMethod: "wechat"}
	createResult, err := svc.CreateOrder(req2, payConfig)
	require.NoError(t, err)

	_, err = svc.HandlePaymentCallback("pay_no", createResult.OrderNo, 10.00)
	require.NoError(t, err)

	t.Run("list with status filter", func(t *testing.T) {
		result, err := svc.List(1, 10, model.OrderStatusPending, "")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("list with completed status filter", func(t *testing.T) {
		result, err := svc.List(1, 10, model.OrderStatusComplete, "")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("list with keyword matching contact", func(t *testing.T) {
		result, err := svc.List(1, 10, -1, "filter1")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("list with no matches returns empty", func(t *testing.T) {
		result, err := svc.List(1, 10, -1, "nonexistent-keyword")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}