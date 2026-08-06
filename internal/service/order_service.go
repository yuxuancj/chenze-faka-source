package service

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"chenze-faka/internal/config"
	"chenze-faka/internal/model"
	"chenze-faka/internal/repository"
)

type OrderService struct {
	orderRepo   *repository.OrderRepository
	productRepo *repository.ProductRepository
	cardRepo    *repository.CardRepository
	payConfig   config.PayConfig
}

func NewOrderService(payConfig config.PayConfig) *OrderService {
	return &OrderService{
		orderRepo:   repository.NewOrderRepository(),
		productRepo: repository.NewProductRepository(),
		cardRepo:    repository.NewCardRepository(),
		payConfig:   payConfig,
	}
}

type CreateOrderRequest struct {
	ProductID uint   `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Contact   string `json:"contact"`
	Contact2  string `json:"contact2"`
	PayMethod string `json:"pay_method"`
}

type CreateOrderResponse struct {
	OrderNo  string  `json:"order_no"`
	PayURL   string  `json:"pay_url"`
	Amount   float64 `json:"amount"`
	ExpireAt string  `json:"expire_at"`
}

type QueryOrderResponse struct {
	OrderNo  string   `json:"order_no"`
	Status   int      `json:"status"`
	StatusText string `json:"status_text"`
	Quantity int      `json:"quantity"`
	Amount   float64  `json:"amount"`
	Contact  string   `json:"contact"`
	Cards    []string `json:"cards,omitempty"`
	PayAt    string   `json:"pay_at,omitempty"`
	CreatedAt string  `json:"created_at"`
}

func (s *OrderService) CreateOrder(req *CreateOrderRequest) (*CreateOrderResponse, error) {
	if req.Quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}
	if req.Quantity > 99 {
		return nil, errors.New("quantity cannot exceed 99")
	}
	if req.Contact == "" {
		return nil, errors.New("contact cannot be empty")
	}
	if req.PayMethod != "alipay" && req.PayMethod != "wechat" {
		return nil, errors.New("invalid pay method")
	}

	product, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if product.Stock < req.Quantity {
		return nil, errors.New("insufficient stock")
	}

	orderNo := generateOrderNo()
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

	if err := s.orderRepo.Create(order); err != nil {
		return nil, err
	}

	payURL := s.buildPayURL(orderNo, totalAmount, req.PayMethod, product.Name)

	return &CreateOrderResponse{
		OrderNo:  orderNo,
		PayURL:   payURL,
		Amount:   totalAmount,
		ExpireAt: expireAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *OrderService) HandlePaymentCallback(payNo, orderNo string, amount float64) (bool, error) {
	order, err := s.orderRepo.GetByOrderNo(orderNo)
	if err != nil {
		return false, errors.New("order not found")
	}

	if order.Status != model.OrderStatusPending {
		return false, nil
	}

	if order.TotalAmount != amount {
		return false, errors.New("amount mismatch")
	}

	cards, err := s.cardRepo.GetAvailableCards(order.ProductID, order.Quantity)
	if err != nil {
		return false, err
	}
	if len(cards) < order.Quantity {
		return false, errors.New("insufficient cards")
	}

	now := time.Now()
	order.Status = model.OrderStatusComplete
	order.PayNo = payNo
	order.PaidAt = &now

	if err := s.orderRepo.Update(order); err != nil {
		return false, err
	}

	for _, card := range cards {
		s.cardRepo.MarkAsSold(card.ID, orderNo)
	}

	s.productRepo.UpdateStock(order.ProductID, -order.Quantity)

	return true, nil
}

func (s *OrderService) QueryOrder(orderNo, contact string) (*QueryOrderResponse, error) {
	var order *model.Order
	var err error

	if orderNo != "" {
		order, err = s.orderRepo.GetByOrderNo(orderNo)
	} else if contact != "" {
		orders, err := s.orderRepo.GetByContact(contact)
		if err != nil || len(orders) == 0 {
			return nil, errors.New("order not found")
		}
		order = orders[0]
	} else {
		return nil, errors.New("order number or contact cannot be empty")
	}

	if err != nil {
		return nil, errors.New("order not found")
	}

	response := &QueryOrderResponse{
		OrderNo:   order.OrderNo,
		Status:    order.Status,
		StatusText: getStatusText(order.Status),
		Quantity:  order.Quantity,
		Amount:    order.TotalAmount,
		Contact:   order.Contact,
		CreatedAt: order.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if order.Status == model.OrderStatusComplete {
		cards, _, _ := s.cardRepo.List(1, order.Quantity, order.ProductID, model.CardStatusSold)
		for _, card := range cards {
			if card.OrderNo == order.OrderNo {
				response.Cards = append(response.Cards, card.CardNo)
			}
		}
		response.PayAt = order.PaidAt.Format("2006-01-02 15:04:05")
	}

	return response, nil
}

func (s *OrderService) List(page, pageSize, status int, keyword string) ([]model.Order, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	if status < 0 {
		status = -1
	}

	return s.orderRepo.List(page, pageSize, status, keyword)
}

func (s *OrderService) AutoCloseExpiredOrders() error {
	orders, err := s.orderRepo.GetExpiredOrders()
	if err != nil {
		return err
	}

	for _, order := range orders {
		order.Status = model.OrderStatusCancel
		s.orderRepo.Update(order)
	}

	return nil
}

func (s *OrderService) buildPayURL(orderNo string, amount float64, payMethod string, productName string) string {
	if s.payConfig.URL == "" || s.payConfig.Merchant == "" || s.payConfig.Key == "" {
		return ""
	}

	params := map[string]string{
		"pid":          s.payConfig.Merchant,
		"type":         payMethod,
		"out_trade_no": orderNo,
		"notify_url":   "/api/pay/notify",
		"return_url":   "/api/pay/return",
		"name":         productName,
		"money":        fmt.Sprintf("%.2f", amount),
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}

	signStr := strings.Join(parts, "&") + "&key=" + s.payConfig.Key
	sign := md5Sum(signStr)
	params["sign"] = sign
	params["sign_type"] = "MD5"

	var queryParts []string
	for k, v := range params {
		queryParts = append(queryParts, k+"="+v)
	}

	return s.payConfig.URL + "/submit.php?" + strings.Join(queryParts, "&")
}

func generateOrderNo() string {
	now := time.Now()
	return fmt.Sprintf("CZ%d%06d", now.Format("20060102150405"), now.Nanosecond()/1000)
}

func getStatusText(status int) string {
	switch status {
	case model.OrderStatusPending:
		return "pending"
	case model.OrderStatusPaid:
		return "paid"
	case model.OrderStatusComplete:
		return "completed"
	case model.OrderStatusCancel:
		return "cancelled"
	default:
		return "unknown"
	}
}

func md5Sum(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

func (s *OrderService) verifySign(params map[string]string) bool {
	if s.payConfig.Key == "" {
		return false
	}

	sign := params["sign"]
	signType := params["sign_type"]

	cleaned := make(map[string]string)
	for k, v := range params {
		if k == "sign" || k == "sign_type" {
			continue
		}
		cleaned[k] = v
	}

	keys := make([]string, 0, len(cleaned))
	for k := range cleaned {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		if cleaned[k] != "" {
			parts = append(parts, k+"="+cleaned[k])
		}
	}

	signStr := strings.Join(parts, "&") + "&key=" + s.payConfig.Key
	calcSign := md5Sum(signStr)

	if signType == "MD5" {
		return sign == calcSign
	}
	return false
}

func (s *OrderService) VerifyNotify(params map[string]string) bool {
	return s.verifySign(params)
}
