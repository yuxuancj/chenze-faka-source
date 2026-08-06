package model

import (
	"time"

	"gorm.io/gorm"
)

type Card struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	ProductID uint           `json:"product_id" gorm:"index;not null"`
	CardNo    string         `json:"card_no" gorm:"uniqueIndex;size:100;not null"`
	Status    int            `json:"status" gorm:"not null;default:0"`
	OrderNo   string         `json:"order_no" gorm:"index;size:50"`
	SoldAt    *time.Time     `json:"sold_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

const (
	CardStatusUnsold = 0
	CardStatusSold   = 1
)

func (Card) TableName() string {
	return "cards"
}