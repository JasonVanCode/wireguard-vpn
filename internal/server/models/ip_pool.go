package models

import (
	"time"

	"gorm.io/gorm"
)

// IPPool IP地址池管理
type IPPool struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Network   string         `json:"network" gorm:"not null;size:20"`
	IPAddress string         `json:"ip_address" gorm:"not null;size:15;uniqueIndex"`
	IsUsed    bool           `json:"is_used" gorm:"default:false"`
	ModuleID  *uint          `json:"module_id" gorm:"index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
