package service

import (
	"testing"
	"time"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderDBNilSupplement(t *testing.T) {
	svc := NewOrderService()

	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	payConfig := coveragePayConfig()

	t.Run("create order with nil db", func(t *testing.T) {
		result, err := svc.CreateOrder(&CreateOrderRequest{
			ProductID: 1, Quantity: 1, Contact: "test@test.com", PayMethod: "alipay",
		}, payConfig)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("handle payment callback with nil db", func(t *testing.T) {
		ok, err := svc.HandlePaymentCallback("pay", "order", 10.0)
		assert.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("query order with nil db", func(t *testing.T) {
		result, err := svc.QueryOrder("order_no", "")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("list with nil db returns empty", func(t *testing.T) {
		result, err := svc.List(1, 10, -1, "")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(0), result.Total)
	})

	t.Run("auto close expired with nil db returns nil", func(t *testing.T) {
		err := svc.AutoCloseExpiredOrders()
		assert.NoError(t, err)
	})
}

func TestOrderCreateValidationFields(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()
	product := createOrderCoverageProduct(t, "Test Product", 10.00, 100)

	t.Run("create with quantity 0 fails", func(t *testing.T) {
		result, err := svc.CreateOrder(&CreateOrderRequest{
			ProductID: product.ID, Quantity: 0, Contact: "test@test.com", PayMethod: "alipay",
		}, payConfig)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("create with quantity > 99 fails", func(t *testing.T) {
		result, err := svc.CreateOrder(&CreateOrderRequest{
			ProductID: product.ID, Quantity: 100, Contact: "test@test.com", PayMethod: "alipay",
		}, payConfig)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("create with empty contact fails", func(t *testing.T) {
		result, err := svc.CreateOrder(&CreateOrderRequest{
			ProductID: product.ID, Quantity: 1, Contact: "", PayMethod: "alipay",
		}, payConfig)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("create with invalid pay method fails", func(t *testing.T) {
		result, err := svc.CreateOrder(&CreateOrderRequest{
			ProductID: product.ID, Quantity: 1, Contact: "test@test.com", PayMethod: "bitcoin",
		}, payConfig)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("create with nonexistent product fails", func(t *testing.T) {
		result, err := svc.CreateOrder(&CreateOrderRequest{
			ProductID: 9999, Quantity: 1, Contact: "test@test.com", PayMethod: "alipay",
		}, payConfig)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("create with insufficient stock fails", func(t *testing.T) {
		smallProduct := createOrderCoverageProduct(t, "LowStock Product", 5.00, 1)
		result, err := svc.CreateOrder(&CreateOrderRequest{
			ProductID: smallProduct.ID, Quantity: 5, Contact: "test@test.com", PayMethod: "alipay",
		}, payConfig)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestOrderQueryByContact(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()
	product := createOrderCoverageProduct(t, "Query Product", 10.00, 100)

	createOrderCoverageCards(t, product.ID, 5)

	result, err := svc.CreateOrder(&CreateOrderRequest{
		ProductID: product.ID,
		Quantity:  2,
		Contact:   "contact_query@test.com",
		PayMethod: "alipay",
	}, payConfig)
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Run("query by contact finds latest order", func(t *testing.T) {
		queryResult, err := svc.QueryOrder("", "contact_query@test.com")
		require.NoError(t, err)
		assert.NotNil(t, queryResult)
		assert.Equal(t, result.OrderNo, queryResult.OrderNo)
	})

	t.Run("query by both empty returns error", func(t *testing.T) {
		queryResult, err := svc.QueryOrder("", "")
		assert.Error(t, err)
		assert.Nil(t, queryResult)
	})

	t.Run("query by nonexistent contact returns not found", func(t *testing.T) {
		queryResult, err := svc.QueryOrder("", "ghost@test.com")
		assert.Error(t, err)
		assert.Nil(t, queryResult)
	})
}

func TestOrderQueryByOrderNoNotFound(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()

	t.Run("query with nonexistent order no returns error", func(t *testing.T) {
		result, err := svc.QueryOrder("NONEXISTENT-ORDER-NO", "")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestOrderListFilters(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()

	for i := 0; i < 3; i++ {
		product := createOrderCoverageProduct(t, "List Product", 10.00, 100)
		_, err := svc.CreateOrder(&CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  1,
			Contact:   "list_test@test.com",
			PayMethod: "alipay",
		}, payConfig)
		require.NoError(t, err)
	}

	t.Run("list with default params returns all orders", func(t *testing.T) {
		result, err := svc.List(0, 0, -1, "")
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Orders, 3)
	})

	t.Run("list with status filter", func(t *testing.T) {
		result, err := svc.List(1, 10, model.OrderStatusPending, "")
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
	})

	t.Run("list with completed status returns 0", func(t *testing.T) {
		result, err := svc.List(1, 10, model.OrderStatusComplete, "")
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.Total)
	})

	t.Run("list with keyword filter", func(t *testing.T) {
		result, err := svc.List(1, 10, -1, "list_test")
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
	})
}

func TestOrderVerifyNotifySignature(t *testing.T) {
	svc := NewOrderService()

	t.Run("verify notify with invalid signature returns false", func(t *testing.T) {
		params := map[string]string{
			"order_no": "ORDER-001",
			"amount":   "10.00",
			"sign":     "invalid_signature",
		}
		assert.False(t, svc.VerifyNotify(params, "test_key"))
	})

	t.Run("verify notify with empty params returns false", func(t *testing.T) {
		assert.False(t, svc.VerifyNotify(map[string]string{}, "test_key"))
	})

	t.Run("verify notify with mismatched key returns false", func(t *testing.T) {
		params := map[string]string{
			"order_no": "ORDER-001",
			"amount":   "10.00",
		}
		assert.False(t, svc.VerifyNotify(params, "wrong_key"))
	})
}

func TestOrderAutoCloseExpired(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()

	product := createOrderCoverageProduct(t, "Expire Product", 10.00, 100)

	result, err := svc.CreateOrder(&CreateOrderRequest{
		ProductID: product.ID,
		Quantity:  1,
		Contact:   "expire@test.com",
		PayMethod: "alipay",
	}, payConfig)
	require.NoError(t, err)

	t.Run("auto close with no expired orders does nothing", func(t *testing.T) {
		err := svc.AutoCloseExpiredOrders()
		require.NoError(t, err)

		queryResult, err := svc.QueryOrder(result.OrderNo, "")
		require.NoError(t, err)
		assert.Equal(t, model.OrderStatusPending, queryResult.Status)
	})

	t.Run("auto close after setting expired_at to past", func(t *testing.T) {
		require.NoError(t, database.DB.Model(&model.Order{}).
			Where("order_no = ?", result.OrderNo).
			Update("expired_at", time.Now().Add(-1*time.Hour)).Error)

		err := svc.AutoCloseExpiredOrders()
		require.NoError(t, err)

		queryResult, err := svc.QueryOrder(result.OrderNo, "")
		require.NoError(t, err)
		assert.Equal(t, model.OrderStatusCancel, queryResult.Status)
	})
}

func TestOrderHandlePaymentCallbackEdgeCases(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()

	product := createOrderCoverageProduct(t, "Callback Product", 10.00, 100)
	createOrderCoverageCards(t, product.ID, 5)

	t.Run("callback with nonexistent order fails", func(t *testing.T) {
		ok, err := svc.HandlePaymentCallback("pay_no", "NONEXISTENT", 10.0)
		assert.Error(t, err)
		assert.False(t, ok)
	})

	order, err := svc.CreateOrder(&CreateOrderRequest{
		ProductID: product.ID,
		Quantity:  3,
		Contact:   "callback@test.com",
		PayMethod: "alipay",
	}, payConfig)
	require.NoError(t, err)

	t.Run("callback with wrong amount fails", func(t *testing.T) {
		ok, err := svc.HandlePaymentCallback("pay_no", order.OrderNo, 999.0)
		assert.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("callback with correct amount but not enough cards fails", func(t *testing.T) {
		order2, err := svc.CreateOrder(&CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  10,
			Contact:   "callback2@test.com",
			PayMethod: "alipay",
		}, payConfig)
		require.NoError(t, err)

		ok, err := svc.HandlePaymentCallback("pay_no2", order2.OrderNo, order2.Amount)
		assert.Error(t, err)
		assert.False(t, ok)
	})
}

func TestOrderHandlePaymentCallbackSuccess(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()

	product := createOrderCoverageProduct(t, "CallbackSuccess Product", 10.00, 100)
	createOrderCoverageCards(t, product.ID, 5)

	order, err := svc.CreateOrder(&CreateOrderRequest{
		ProductID: product.ID,
		Quantity:  2,
		Contact:   "callback_success@test.com",
		PayMethod: "alipay",
	}, payConfig)
	require.NoError(t, err)

	t.Run("callback with correct params succeeds", func(t *testing.T) {
		ok, err := svc.HandlePaymentCallback("PAY-001", order.OrderNo, order.Amount)
		require.NoError(t, err)
		assert.True(t, ok)

		queryResult, err := svc.QueryOrder(order.OrderNo, "")
		require.NoError(t, err)
		assert.Equal(t, model.OrderStatusComplete, queryResult.Status)
		assert.NotEmpty(t, queryResult.Cards)
	})
}

func TestOrderListPagination(t *testing.T) {
	setupOrderCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewOrderService()
	payConfig := coveragePayConfig()
	product := createOrderCoverageProduct(t, "Pagination Product", 10.00, 100)

	for i := 0; i < 15; i++ {
		_, err := svc.CreateOrder(&CreateOrderRequest{
			ProductID: product.ID,
			Quantity:  1,
			Contact:   "pagination@test.com",
			PayMethod: "alipay",
		}, payConfig)
		require.NoError(t, err)
	}

	t.Run("list first page with 5 items", func(t *testing.T) {
		result, err := svc.List(1, 5, -1, "")
		require.NoError(t, err)
		assert.Equal(t, int64(15), result.Total)
		assert.Len(t, result.Orders, 5)
	})

	t.Run("list second page", func(t *testing.T) {
		result, err := svc.List(2, 5, -1, "")
		require.NoError(t, err)
		assert.Len(t, result.Orders, 5)
	})

	t.Run("list last page with remaining items", func(t *testing.T) {
		result, err := svc.List(4, 5, -1, "")
		require.NoError(t, err)
		assert.Len(t, result.Orders, 0)
	})
}