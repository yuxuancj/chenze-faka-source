package repository

import (
	"chenze-faka/internal/model"
	"chenze-faka/pkg/database"

	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{db: database.DB}
}

func (r *OrderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepository) Update(order *model.Order) error {
	return r.db.Save(order).Error
}

func (r *OrderRepository) GetByID(id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) GetByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) List(page, pageSize int, status int, keyword string) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("order_no LIKE ? OR contact LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *OrderRepository) GetExpiredOrders() ([]*model.Order, error) {
	var orders []*model.Order
	err := r.db.Where("status = ? AND expired_at < NOW()", model.OrderStatusPending).Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) UpdateStatus(orderNo string, status int) error {
	return r.db.Model(&model.Order{}).Where("order_no = ?", orderNo).
		Update("status", status).Error
}

func (r *OrderRepository) GetByContact(contact string) ([]*model.Order, error) {
	var orders []*model.Order
	err := r.db.Where("contact = ?", contact).Order("id DESC").Find(&orders).Error
	return orders, err
}
