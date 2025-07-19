package models

import (
	"time"

	"gorm.io/gorm"
)

// SystemConfig 系统配置
type SystemConfig struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Key       string         `json:"key" gorm:"not null;size:100;uniqueIndex"`
	Value     string         `json:"value" gorm:"type:text"`
	Type      string         `json:"type" gorm:"size:20;default:'string'"`
	Comment   string         `json:"comment" gorm:"size:200"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
