package model

import (
	"time"
)

type OperationLog struct {
	ID         uint      `json:"id" gorm:"primarykey"`
	UserID     uint      `json:"user_id" gorm:"index"`
	Username   string    `json:"username" gorm:"size:50;index"`
	Action     string    `json:"action" gorm:"size:100;index"`
	TargetType string    `json:"target_type" gorm:"size:50"`
	TargetID   string    `json:"target_id" gorm:"size:100"`
	Detail     string    `json:"detail" gorm:"type:text"`
	IP         string    `json:"ip" gorm:"size:50"`
	UserAgent  string    `json:"user_agent" gorm:"size:500"`
	Status     int       `json:"status" gorm:"not null;default:1"`
	CreatedAt  time.Time `json:"created_at"`
}

const (
	LogStatusSuccess = 1
	LogStatusFailed  = 0
)

func (OperationLog) TableName() string {
	return "operation_logs"
}

type OrderLog struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	OrderNo   string    `json:"order_no" gorm:"size:50;index"`
	Action    string    `json:"action" gorm:"size:100"`
	Detail    string    `json:"detail" gorm:"type:text"`
	Operator  string    `json:"operator" gorm:"size:50"`
	CreatedAt time.Time `json:"created_at"`
}

func (OrderLog) TableName() string {
	return "order_logs"
}
