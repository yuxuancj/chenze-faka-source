package model

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	Icon      string         `json:"icon" gorm:"size:200"`
	Sort      int            `json:"sort" gorm:"not null;default:0"`
	Status    int            `json:"status" gorm:"not null;default:1"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

const (
	CategoryDisabled = 0
	CategoryEnabled  = 1
)

func (Category) TableName() string {
	return "categories"
}
