package model

import (
	"time"

	"gorm.io/gorm"
)

type PaymentChannel struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	Type      string         `json:"type" gorm:"size:50;not null"`
	Icon      string         `json:"icon" gorm:"size:200"`
	Config    string         `json:"config" gorm:"type:text"`
	FeeRate   float64        `json:"fee_rate" gorm:"type:decimal(5,4);default:0"`
	Status    int            `json:"status" gorm:"not null;default:1"`
	Sort      int            `json:"sort" gorm:"not null;default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

const (
	PayTypeAlipay = "alipay"
	PayTypeWechat = "wechat"
	PayTypeStripe = "stripe"
	PayTypeCustom = "custom"

	ChannelDisabled = 0
	ChannelEnabled  = 1
)

func (PaymentChannel) TableName() string {
	return "payment_channels"
}
