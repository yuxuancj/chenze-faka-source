package model

import (
	"time"

	"gorm.io/gorm"
)

type Node struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	URL       string         `json:"url" gorm:"size:500;not null"`
	Weight    int            `json:"weight" gorm:"not null;default:1"`
	Status    int            `json:"status" gorm:"not null;default:1"`
	LastPing  *time.Time     `json:"last_ping"`
	PingTime  int64          `json:"ping_time" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

const (
	NodeOffline = 0
	NodeOnline  = 1
)

func (Node) TableName() string {
	return "nodes"
}
