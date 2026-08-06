package model

import (
	"time"

	"gorm.io/gorm"
)

type EmailConfig struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	SMTPHost  string         `json:"smtp_host" gorm:"size:200;not null"`
	SMTPPort  int            `json:"smtp_port" gorm:"not null;default:465"`
	Username  string         `json:"username" gorm:"size:200;not null"`
	Password  string         `json:"password" gorm:"size:500;not null"`
	Sender    string         `json:"sender" gorm:"size:200;not null"`
	UseSSL    bool           `json:"use_ssl" gorm:"default:true"`
	Status    int            `json:"status" gorm:"not null;default:1"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (EmailConfig) TableName() string {
	return "email_configs"
}

type EmailLog struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	ToEmail   string    `json:"to_email" gorm:"size:200;index"`
	Subject   string    `json:"subject" gorm:"size:500"`
	Content   string    `json:"content" gorm:"type:text"`
	Status    int       `json:"status" gorm:"not null;default:0"`
	ErrorMsg  string    `json:"error_msg" gorm:"size:500"`
	RelatedNo string    `json:"related_no" gorm:"size:50;index"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	EmailLogPending = 0
	EmailLogSent    = 1
	EmailLogFailed  = 2
)

func (EmailLog) TableName() string {
	return "email_logs"
}
