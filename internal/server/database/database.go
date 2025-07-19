package database

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"eitec-vpn/internal/server/models"
	"eitec-vpn/internal/shared/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase(dbPath string) error {
	return InitDatabaseWithOptions(dbPath, false)
}

// InitDatabaseWithOptions 初始化数据库连接，可选择是否强制初始化
func InitDatabaseWithOptions(dbPath string, forceInit bool) error {
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

	// 自动迁移数据库结构
	if err := AutoMigrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 只有新数据库或强制初始化时才执行
	if isNewDB || forceInit {
		if forceInit {
			log.Println("强制初始化数据库...")
		} else {
			log.Println("检测到新数据库，正在初始化默认数据...")
		}

		if err := InitDefaultData(); err != nil {
			return fmt.Errorf("初始化默认数据失败: %w", err)
		}
		log.Println("默认数据初始化完成")
	}

	return nil
}

// AutoMigrate 自动迁移数据库结构
func AutoMigrate() error {
	err := DB.AutoMigrate(
		&models.WireGuardInterface{},
		&models.Module{},
		&models.User{},
		&models.UserVPN{}, // 添加UserVPN模型
		&models.SystemConfig{},
		&models.IPPool{},
	)
	if err != nil {
		return fmt.Errorf("自动迁移失败: %w", err)
	}

	return nil
}

// InitDefaultData 初始化默认数据
func InitDefaultData() error {
	// 初始化系统配置
	if err := initSystemConfig(); err != nil {
		return fmt.Errorf("初始化系统配置失败: %w", err)
	}

	// 初始化默认用户
	if err := initDefaultAdmin(); err != nil {
		return fmt.Errorf("初始化默认用户失败: %w", err)
	}

	// 初始化默认WireGuard接口
	if err := initDefaultInterfaces(); err != nil {
		return fmt.Errorf("初始化默认接口失败: %w", err)
	}

	return nil
}

// initDefaultAdmin 初始化默认管理员
func initDefaultAdmin() error {
	var count int64
	DB.Model(&models.User{}).Count(&count)

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
	admin := &models.User{
		Username: "admin",
		Password: hashedPassword, // 使用哈希后的密码
		IsActive: true,
	}

	if err := DB.Create(admin).Error; err != nil {
		return fmt.Errorf("创建默认管理员失败: %w", err)
	}

	log.Println("创建默认管理员: admin/admin123 (密码已加密)")
	return nil
}

// initIPPool 初始化IP地址池
func initIPPool() error {
	var count int64
	DB.Model(&models.IPPool{}).Count(&count)

	// 如果已有IP池，跳过
	if count > 0 {
		return nil
	}

	// 生成 10.10.0.2-10.10.0.254 的IP地址池
	network := "10.10.0.0/24"
	for i := 2; i <= 254; i++ {
		ip := fmt.Sprintf("10.10.0.%d", i)
		ipPool := &models.IPPool{
			Network:   network,
			IPAddress: ip,
			IsUsed:    false,
		}

		if err := DB.Create(ipPool).Error; err != nil {
			return fmt.Errorf("创建IP池失败 %s: %w", ip, err)
		}
	}

	log.Println("初始化IP地址池: 10.10.0.2-10.10.0.254")
	return nil
}

// initSystemConfig 初始化系统配置
func initSystemConfig() error {
	configs := []models.SystemConfig{
		{
			Key:     "server.public_key",
			Value:   "",
			Type:    "string",
			Comment: "服务器WireGuard公钥",
		},
		{
			Key:     "server.private_key",
			Value:   "",
			Type:    "string",
			Comment: "服务器WireGuard私钥",
		},
		{
			Key:     "server.endpoint",
			Value:   "",
			Type:    "string",
			Comment: "服务器公网地址:端口",
		},
		{
			Key:     "wg.network",
			Value:   "10.10.0.0/24",
			Type:    "string",
			Comment: "WireGuard网络段",
		},
		{
			Key:     "wg.dns",
			Value:   "8.8.8.8,8.8.4.4",
			Type:    "string",
			Comment: "DNS服务器",
		},
	}

	for _, config := range configs {
		var existing models.SystemConfig
		if err := DB.Where("key = ?", config.Key).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 配置不存在，创建
				if err := DB.Create(&config).Error; err != nil {
					return fmt.Errorf("创建系统配置失败 %s: %w", config.Key, err)
				}
			} else {
				return fmt.Errorf("查询系统配置失败 %s: %w", config.Key, err)
			}
		}
	}

	return nil
}

// initDefaultInterfaces 初始化默认WireGuard接口
func initDefaultInterfaces() error {
	// 检查是否已有接口
	var count int64
	if err := DB.Model(&models.WireGuardInterface{}).Count(&count).Error; err != nil {
		return fmt.Errorf("检查接口数量失败: %w", err)
	}

	// 如果已有接口，跳过初始化
	if count > 0 {
		log.Println("检测到已存在WireGuard接口，跳过默认接口初始化")
		return nil
	}

	log.Println("正在初始化默认WireGuard接口...")

	// 获取默认模板
	templates := models.GetDefaultTemplates()

	for _, template := range templates {
		// 生成密钥对
		keyPair, err := generateKeyPair()
		if err != nil {
			log.Printf("生成接口 %s 密钥对失败: %v", template.Name, err)
			continue
		}

		// 计算服务器IP
		serverIP, err := calculateServerIP(template.Network)
		if err != nil {
			log.Printf("计算接口 %s 服务器IP失败: %v", template.Name, err)
			continue
		}

		// 创建接口记录
		wgInterface := &models.WireGuardInterface{
			Name:        template.Name,
			Description: template.Description,
			Network:     template.Network,
			ServerIP:    serverIP,
			ListenPort:  template.ListenPort,
			PublicKey:   keyPair.PublicKey,
			PrivateKey:  keyPair.PrivateKey,
			Status:      models.InterfaceStatusDown,
			MaxPeers:    template.MaxPeers,
			DNS:         template.DNS,
			PostUp:      template.PostUp,
			PostDown:    template.PostDown,
			MTU:         1420,
			SaveConfig:  true,
		}

		if err := DB.Create(wgInterface).Error; err != nil {
			log.Printf("创建接口 %s 失败: %v", template.Name, err)
			continue
		}

		// 为接口创建IP池
		if err := createIPPoolForInterface(wgInterface); err != nil {
			log.Printf("为接口 %s 创建IP池失败: %v", template.Name, err)
			// 不删除接口，只记录错误
		}

		log.Printf("成功创建接口: %s (%s)", template.Name, template.Description)
	}

	log.Println("默认WireGuard接口初始化完成")
	return nil
}

// generateKeyPair 生成WireGuard密钥对（简化版本）
func generateKeyPair() (*models.WireGuardKey, error) {
	// 这里应该调用WireGuard的密钥生成函数
	// 为了演示，我们生成一些占位符密钥
	publicKey, err := utils.GenerateRandomString(43)
	if err != nil {
		return nil, fmt.Errorf("生成公钥失败: %w", err)
	}

	privateKey, err := utils.GenerateRandomString(43)
	if err != nil {
		return nil, fmt.Errorf("生成私钥失败: %w", err)
	}

	return &models.WireGuardKey{
		PublicKey:  publicKey + "=",  // WireGuard公钥44字符
		PrivateKey: privateKey + "=", // WireGuard私钥44字符
	}, nil
}

// calculateServerIP 计算服务器IP地址
func calculateServerIP(network string) (string, error) {
	_, ipNet, err := net.ParseCIDR(network)
	if err != nil {
		return "", err
	}

	// 使用网络段的第一个可用IP作为服务器IP
	ip := ipNet.IP.To4()
	if ip == nil {
		return "", fmt.Errorf("不支持IPv6")
	}

	// 第一个IP通常是网络地址，第二个IP作为服务器IP
	ip[3] = ip[3] + 1
	return ip.String(), nil
}

// createIPPoolForInterface 为接口创建IP池
func createIPPoolForInterface(wgInterface *models.WireGuardInterface) error {
	_, ipNet, err := net.ParseCIDR(wgInterface.Network)
	if err != nil {
		return err
	}

	// 生成IP池
	ip := ipNet.IP.To4()
	if ip == nil {
		return fmt.Errorf("不支持IPv6")
	}

	// 跳过网络地址和服务器IP，从第三个IP开始
	startIP := make(net.IP, 4)
	copy(startIP, ip)
	startIP[3] = startIP[3] + 2

	// 计算可用IP数量
	ones, bits := ipNet.Mask.Size()
	availableIPs := 1<<(bits-ones) - 3 // 减去网络地址、广播地址和服务器IP

	// 限制IP数量避免过多
	if availableIPs > 200 {
		availableIPs = 200
	}

	// 批量创建IP池记录
	var ipPools []models.IPPool
	for i := 0; i < availableIPs; i++ {
		currentIP := make(net.IP, 4)
		copy(currentIP, startIP)
		currentIP[3] = currentIP[3] + byte(i)

		// 检查是否超出网络范围
		if !ipNet.Contains(currentIP) {
			break
		}

		ipPools = append(ipPools, models.IPPool{
			Network:   wgInterface.Network,
			IPAddress: currentIP.String(),
			IsUsed:    false,
		})
	}

	// 批量插入
	if len(ipPools) > 0 {
		if err := DB.CreateInBatches(ipPools, 100).Error; err != nil {
			return fmt.Errorf("批量创建IP池失败: %w", err)
		}
	}

	return nil
}

// GetAvailableIP 获取可用IP地址
func GetAvailableIP() (string, error) {
	var ipPool models.IPPool
	if err := DB.Where("is_used = ?", false).First(&ipPool).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("没有可用的IP地址")
		}
		return "", fmt.Errorf("查询可用IP失败: %w", err)
	}

	return ipPool.IPAddress, nil
}

// AllocateIP 分配IP地址给模块
func AllocateIP(ip string, moduleID uint) error {
	result := DB.Model(&models.IPPool{}).
		Where("ip_address = ? AND is_used = ?", ip, false).
		Updates(map[string]interface{}{
			"is_used":   true,
			"module_id": moduleID,
		})

	if result.Error != nil {
		return fmt.Errorf("分配IP地址失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("IP地址 %s 不可用", ip)
	}

	return nil
}

// ReleaseIP 释放IP地址
func ReleaseIP(ip string) error {
	result := DB.Model(&models.IPPool{}).
		Where("ip_address = ?", ip).
		Updates(map[string]interface{}{
			"is_used":   false,
			"module_id": nil,
		})

	if result.Error != nil {
		return fmt.Errorf("释放IP地址失败: %w", result.Error)
	}

	return nil
}

// GetSystemConfig 获取系统配置
func GetSystemConfig(key string) (string, error) {
	var config models.SystemConfig
	if err := DB.Where("key = ?", key).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("配置项不存在: %s", key)
		}
		return "", fmt.Errorf("查询配置失败: %w", err)
	}

	return config.Value, nil
}

// SetSystemConfig 设置系统配置
func SetSystemConfig(key, value string) error {
	var config models.SystemConfig
	if err := DB.Where("key = ?", key).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 配置不存在，创建新配置
			config = models.SystemConfig{
				Key:   key,
				Value: value,
				Type:  "string",
			}
			return DB.Create(&config).Error
		}
		return fmt.Errorf("查询系统配置失败: %w", err)
	}

	// 配置存在，更新值
	config.Value = value
	return DB.Save(&config).Error
}

// Close 关闭数据库连接
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
