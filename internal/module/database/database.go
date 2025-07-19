package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"eitec-vpn/internal/module/models"
	"eitec-vpn/internal/shared/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitModuleDatabase 初始化模块端数据库
func InitModuleDatabase(dbPath string) error {
	return InitModuleDatabaseWithOptions(dbPath, false)
}

// InitModuleDatabaseWithOptions 初始化模块端数据库，可选择是否强制初始化
func InitModuleDatabaseWithOptions(dbPath string, forceInit bool) error {
	var err error

	// 确保数据库目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("创建数据库目录失败: %w", err)
	}

	// 检查是否为新数据库
	_, err = os.Stat(dbPath)
	isNewDB := os.IsNotExist(err)

	// 连接SQLite数据库 - 默认使用Silent日志级别
	logLevel := logger.Silent
	if os.Getenv("DB_DEBUG") == "true" {
		logLevel = logger.Info
	}

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 自动迁移数据库结构（使用模块端专用的models）
	if err := models.AutoMigrate(DB); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 只有新数据库或强制初始化才执行数据初始化
	if isNewDB || forceInit {
		log.Println("检测到新数据库或强制初始化，正在初始化默认数据...")
		if err := InitDefaultData(); err != nil {
			return fmt.Errorf("初始化默认数据失败: %w", err)
		}
		log.Println("模块端数据库初始化完成")
	}

	return nil
}

// InitDefaultData 初始化默认数据
func InitDefaultData() error {
	// 初始化默认管理员用户
	if err := initDefaultAdmin(); err != nil {
		return fmt.Errorf("初始化默认管理员失败: %w", err)
	}

	return nil
}

// initDefaultAdmin 初始化默认管理员
func initDefaultAdmin() error {
	var count int64
	if err := DB.Model(&models.LocalUser{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果已有用户，跳过
	if count > 0 {
		return nil
	}

	// 对密码进行哈希处理
	hashedPassword, err := utils.HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("密码哈希失败: %w", err)
	}

	// 创建默认管理员
	admin := &models.LocalUser{
		Username: "admin",
		Password: hashedPassword,
		Role:     models.LocalUserRoleAdmin,
		IsActive: true,
	}

	if err := DB.Create(admin).Error; err != nil {
		return fmt.Errorf("创建默认管理员失败: %w", err)
	}

	log.Println("创建模块端默认管理员: admin/admin123 (密码已加密)")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
