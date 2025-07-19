package models

import (
	"gorm.io/gorm"
)

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&LocalModule{},
		&LocalUser{},
	)
}
