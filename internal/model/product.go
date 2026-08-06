package model

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	Name        string         `json:"name" gorm:"size:200;not null"`
	Category    string         `json:"category" gorm:"size:100;index"`
	Price       float64        `json:"price" gorm:"type:decimal(10,2);not null"`
	Description string         `json:"description" gorm:"type:text"`
	Image       string         `json:"image" gorm:"size:500"`
	Stock       int            `json:"stock" gorm:"not null;default:0"`
	Status      int            `json:"status" gorm:"not null;default:1"`
	Sort        int            `json:"sort" gorm:"not null;default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

const (
	ProductStatusOffShelf = 0
	ProductStatusOnShelf  = 1
)

func (Product) TableName() string {
	return "products"
}
