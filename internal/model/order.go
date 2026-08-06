package model

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	OrderNo     string         `json:"order_no" gorm:"uniqueIndex;size:50;not null"`
	ProductID   uint           `json:"product_id" gorm:"index;not null"`
	ProductName string         `json:"product_name" gorm:"size:200"`
	Quantity    int            `json:"quantity" gorm:"not null;default:1"`
	Price       float64        `json:"price" gorm:"type:decimal(10,2);not null"`
	TotalAmount float64        `json:"total_amount" gorm:"type:decimal(10,2);not null"`
	Contact     string         `json:"contact" gorm:"size:100"`
	Contact2    string         `json:"contact2" gorm:"size:100"`
	PayMethod   string         `json:"pay_method" gorm:"size:20;not null"`
	Status      int            `json:"status" gorm:"not null;default:0"`
	PayNo       string         `json:"pay_no" gorm:"size:100"`
	PaidAt      *time.Time     `json:"paid_at"`
	ExpiredAt   time.Time      `json:"expired_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

const (
	OrderStatusPending  = 0
	OrderStatusPaid     = 1
	OrderStatusComplete = 2
	OrderStatusCancel   = 3
)

func (Order) TableName() string {
	return "orders"
}
