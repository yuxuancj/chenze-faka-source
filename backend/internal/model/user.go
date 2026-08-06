package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	Username     string         `json:"username" gorm:"uniqueIndex;size:50;not null"`
	PasswordHash string         `json:"-" gorm:"size:255;not null"`
	Salt         string         `json:"-" gorm:"size:10;not null"`
	Role         string         `json:"role" gorm:"size:20;not null;default:admin"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

func (User) TableName() string {
	return "users"
}