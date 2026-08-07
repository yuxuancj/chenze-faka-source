package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_TableName(t *testing.T) {
	assert.Equal(t, "users", User{}.TableName())
}

func TestProduct_TableName(t *testing.T) {
	assert.Equal(t, "products", Product{}.TableName())
}

func TestCard_TableName(t *testing.T) {
	assert.Equal(t, "cards", Card{}.TableName())
}

func TestOrder_TableName(t *testing.T) {
	assert.Equal(t, "orders", Order{}.TableName())
}

func TestCategory_TableName(t *testing.T) {
	assert.Equal(t, "categories", Category{}.TableName())
}

func TestPaymentChannel_TableName(t *testing.T) {
	assert.Equal(t, "payment_channels", PaymentChannel{}.TableName())
}

func TestEmailConfig_TableName(t *testing.T) {
	assert.Equal(t, "email_configs", EmailConfig{}.TableName())
}

func TestEmailLog_TableName(t *testing.T) {
	assert.Equal(t, "email_logs", EmailLog{}.TableName())
}

func TestNode_TableName(t *testing.T) {
	assert.Equal(t, "nodes", Node{}.TableName())
}

func TestOperationLog_TableName(t *testing.T) {
	assert.Equal(t, "operation_logs", OperationLog{}.TableName())
}

func TestOrderLog_TableName(t *testing.T) {
	assert.Equal(t, "order_logs", OrderLog{}.TableName())
}

func TestUpgradeLog_TableName(t *testing.T) {
	assert.Equal(t, "upgrade_logs", UpgradeLog{}.TableName())
}

func TestOrderStatusText(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{OrderStatusPending, "待支付"},
		{OrderStatusPaid, "已支付"},
		{OrderStatusComplete, "已完成"},
		{OrderStatusCancel, "已取消"},
		{99, "未知"},
		{-1, "未知"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, OrderStatusText(tt.status))
		})
	}
}

func TestProduct_ToVO(t *testing.T) {
	p := &Product{
		ID:                1,
		Name:              "Test Product",
		Category:          "Electronics",
		Price:             99.99,
		Description:       "A test product",
		Image:             "/img/test.png",
		Stock:             100,
		LowStockThreshold: 10,
		StockAlert:        true,
		Status:            ProductStatusOnShelf,
		Sort:              5,
	}

	vo := p.ToVO()

	assert.Equal(t, uint(1), vo.ID)
	assert.Equal(t, "Test Product", vo.Name)
	assert.Equal(t, "Electronics", vo.Category)
	assert.Equal(t, 99.99, vo.Price)
	assert.Equal(t, "A test product", vo.Description)
	assert.Equal(t, "/img/test.png", vo.Image)
	assert.Equal(t, 100, vo.Stock)
	assert.Equal(t, 10, vo.LowStockThreshold)
	assert.True(t, vo.StockAlert)
	assert.Equal(t, ProductStatusOnShelf, vo.Status)
	assert.Equal(t, 5, vo.Sort)
}

func TestProduct_ToVO_ZeroValues(t *testing.T) {
	p := &Product{}
	vo := p.ToVO()

	assert.Equal(t, uint(0), vo.ID)
	assert.Equal(t, "", vo.Name)
	assert.Equal(t, "", vo.Category)
	assert.Equal(t, 0.0, vo.Price)
	assert.Equal(t, "", vo.Description)
	assert.Equal(t, "", vo.Image)
	assert.Equal(t, 0, vo.Stock)
	assert.Equal(t, 0, vo.LowStockThreshold)
	assert.False(t, vo.StockAlert)
	assert.Equal(t, 0, vo.Status)
	assert.Equal(t, 0, vo.Sort)
}