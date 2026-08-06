package model

import (
	"time"
)

type UpgradeLog struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	Version     string    `json:"version" gorm:"size:50;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Status      int       `json:"status" gorm:"not null;default:0"`
	FileName    string    `json:"file_name" gorm:"size:255"`
	Size        int64     `json:"size" gorm:"default:0"`
	Operator    string    `json:"operator" gorm:"size:50"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const (
	UpgradeStatusPending = 0
	UpgradeStatusSuccess = 1
	UpgradeStatusFailed  = 2
)

func (UpgradeLog) TableName() string {
	return "upgrade_logs"
}