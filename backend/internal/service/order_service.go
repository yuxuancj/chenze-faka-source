package service

import (
	"errors"
	"time"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
	"chenze-faka/internal/pkg/utils"

	"gorm.io/gorm"
)

type OrderService struct{}

func NewOrderService() *OrderService {
	return &OrderService{}
}

type CreateOrderRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"`
	Contact   string `json:"contact" binding:"required"`
	Contact2  string `json:"contact2"`
	PayMethod string `json:"pay_method" binding:"required"`
}

type CreateOrderResult struct {
	OrderNo  string  `json:"order_no"`
	PayURL   string  `json:"pay_url"`
	Amount   float64 `json:"amount"`
	ExpireAt string  `json:"expire_at"`
}

type OrderListResult struct {
	Orders []model.Order `json:"orders"`
	Total  int64         `json:"total"`
	Page   int           `json:"page"`
	PageSize int          `json:"page_size"`
}

func (s *OrderService) CreateOrder(req *CreateOrderRequest, payConfig model.PayConfig) (*CreateOrderResult, error) {
	if req.Quantity <= 0 {
		return nil, errors.New("数量必须大于0")
	}
	if req.Quantity > 99 {
		return nil, errors.New("单次购买数量不能超过99")
	}
	if req.Contact == "" {
		return nil, errors.New("联系方式不能为空")
	}
	if req.PayMethod != "alipay" && req.PayMethod != "wechat" {
		return nil, errors.New("无效的支付方式")
	}

	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}

	var product model.Product
	if err := database.DB.First(&product, req.ProductID).Error; err != nil {
		return nil, errors.New("产品不存在")
	}

	if product.Stock < req.Quantity {
		return nil, errors.New("库存不足")
	}

	orderNo := utils.GenerateOrderNo()
	totalAmount := product.Price * float64(req.Quantity)
	expireAt := time.Now().Add(30 * time.Minute)

	order := &model.Order{
		OrderNo:     orderNo,
		ProductID:   product.ID,
		ProductName: product.Name,
		Quantity:    req.Quantity,
		Price:       product.Price,
		TotalAmount: totalAmount,
		Contact:     req.Contact,
		Contact2:    req.Contact2,
		PayMethod:   req.PayMethod,
		Status:      model.OrderStatusPending,
		ExpiredAt:   expireAt,
	}

	if err := database.DB.Create(order).Error; err != nil {
		return nil, err
	}

	payURL := utils.BuildPayURL(orderNo, totalAmount, req.PayMethod, product.Name, payConfig)

	return &CreateOrderResult{
		OrderNo:  orderNo,
		PayURL:   payURL,
		Amount:   totalAmount,
		ExpireAt: expireAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *OrderService) HandlePaymentCallback(payNo, orderNo string, amount float64) (bool, error) {
	if database.DB == nil {
		return false, errors.New("数据库未连接")
	}

	var order model.Order
	if err := database.DB.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return false, errors.New("订单不存在")
	}

	if order.Status != model.OrderStatusPending {
		return false, nil
	}

	if order.TotalAmount != amount {
		return false, errors.New("金额不匹配")
	}

	var cards []model.Card
	err := database.DB.Where("product_id = ? AND status = ?", order.ProductID, model.CardStatusUnsold).
		Limit(order.Quantity).Find(&cards).Error
	if err != nil || len(cards) < order.Quantity {
		return false, errors.New("可用卡密不足")
	}

	now := time.Now()
	order.Status = model.OrderStatusComplete
	order.PayNo = payNo
	order.PaidAt = &now

	if err := database.DB.Save(&order).Error; err != nil {
		return false, err
	}

	for _, card := range cards {
		database.DB.Model(&model.Card{}).Where("id = ?", card.ID).Updates(map[string]interface{}{
			"status":   model.CardStatusSold,
			"order_no": orderNo,
			"sold_at":  now,
		})
	}

	database.DB.Model(&model.Product{}).Where("id = ?", order.ProductID).
		UpdateColumn("stock", gorm.Expr("stock - ?", order.Quantity))

	return true, nil
}

func (s *OrderService) QueryOrder(orderNo, contact string) (*model.OrderQueryResult, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}

	var order model.Order
	var err error

	if orderNo != "" {
		err = database.DB.Where("order_no = ?", orderNo).First(&order).Error
	} else if contact != "" {
		err = database.DB.Where("contact = ?", contact).Order("id DESC").First(&order).Error
	} else {
		return nil, errors.New("订单号或联系方式不能为空")
	}

	if err != nil {
		return nil, errors.New("订单不存在")
	}

	result := &model.OrderQueryResult{
		OrderNo:    order.OrderNo,
		Status:     order.Status,
		StatusText: model.OrderStatusText(order.Status),
		Quantity:   order.Quantity,
		Amount:     order.TotalAmount,
		Contact:    order.Contact,
		CreatedAt:  order.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if order.Status == model.OrderStatusComplete {
		var cards []model.Card
		database.DB.Where("order_no = ?", order.OrderNo).Find(&cards)
		for _, c := range cards {
			result.Cards = append(result.Cards, c.CardNo)
		}
		if order.PaidAt != nil {
			result.PaidAt = order.PaidAt.Format("2006-01-02 15:04:05")
		}
	}

	return result, nil
}

func (s *OrderService) List(page, pageSize, status int, keyword string) (*OrderListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	if database.DB == nil {
		return &OrderListResult{Orders: []model.Order{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}

	var orders []model.Order
	var total int64

	query := database.DB.Model(&model.Order{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("order_no LIKE ? OR contact LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, err
	}

	return &OrderListResult{
		Orders:   orders,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *OrderService) VerifyNotify(params map[string]string, payKey string) bool {
	return utils.VerifyPaySign(params, payKey)
}

func (s *OrderService) AutoCloseExpiredOrders() error {
	if database.DB == nil {
		return nil
	}

	var orders []*model.Order
	if err := database.DB.Where("status = ? AND expired_at < ?", model.OrderStatusPending, time.Now()).Find(&orders).Error; err != nil {
		return err
	}

	for _, order := range orders {
		order.Status = model.OrderStatusCancel
		database.DB.Save(order)
	}
	return nil
}