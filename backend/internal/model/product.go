package model

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID               uint           `json:"id" gorm:"primarykey"`
	Name             string         `json:"name" gorm:"size:200;not null"`
	Category         string         `json:"category" gorm:"size:100;index"`
	Price            float64        `json:"price" gorm:"type:decimal(10,2);not null"`
	Description      string         `json:"description" gorm:"type:text"`
	Image            string         `json:"image" gorm:"size:500"`
	Stock            int            `json:"stock" gorm:"not null;default:0"`
	LowStockThreshold int           `json:"low_stock_threshold" gorm:"not null;default:10"`
	StockAlert       bool           `json:"stock_alert" gorm:"not null;default:false"`
	Status           int            `json:"status" gorm:"not null;default:1"`
	Sort             int            `json:"sort" gorm:"not null;default:0"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

const (
	ProductStatusOffShelf = 0
	ProductStatusOnShelf  = 1
)

func (Product) TableName() string {
	return "products"
}

type ProductVO struct {
	ID               uint    `json:"id"`
	Name             string  `json:"name"`
	Category         string  `json:"category"`
	Price            float64 `json:"price"`
	Description      string  `json:"description"`
	Image            string  `json:"image"`
	Stock            int     `json:"stock"`
	LowStockThreshold int     `json:"low_stock_threshold"`
	StockAlert       bool    `json:"stock_alert"`
	Status           int     `json:"status"`
	Sort             int     `json:"sort"`
}

func (p *Product) ToVO() ProductVO {
	return ProductVO{
		ID:               p.ID,
		Name:             p.Name,
		Category:         p.Category,
		Price:            p.Price,
		Description:      p.Description,
		Image:            p.Image,
		Stock:            p.Stock,
		LowStockThreshold: p.LowStockThreshold,
		StockAlert:       p.StockAlert,
		Status:           p.Status,
		Sort:             p.Sort,
	}
}

type ProductGroup struct {
	Category string      `json:"category"`
	Products []ProductVO `json:"products"`
}