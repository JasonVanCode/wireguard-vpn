package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户信息 (简化版，只支持管理员)
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"not null;size:50;uniqueIndex"`
	Password  string         `json:"-" gorm:"not null;size:255"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	LastLogin *time.Time     `json:"last_login"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
