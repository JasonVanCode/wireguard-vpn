package services

import (
	"errors"
	"fmt"
	"time"

	"eitec-vpn/internal/server/database"
	"eitec-vpn/internal/server/models"
	"eitec-vpn/internal/shared/config"
	"eitec-vpn/internal/shared/wireguard"

	"gorm.io/gorm"
)

// 获取全局配置
func getGlobalConfigForUserVPN() *config.ServerConfig {
	// 直接从config包获取全局配置
	return config.GetGlobalServerConfig()
}

// UserVPNService 用户VPN管理服务
type UserVPNService struct {
	db *gorm.DB
}

// NewUserVPNService 创建用户VPN管理服务
func NewUserVPNService() *UserVPNService {
	return &UserVPNService{
		db: database.DB,
	}
}

// CreateUserVPN 为指定模块创建用户VPN配置
func (uvs *UserVPNService) CreateUserVPN(config *models.UserVPNConfig) (*models.UserVPN, error) {
	// 验证模块是否存在
	var module models.Module
	if err := uvs.db.First(&module, config.ModuleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("指定的模块不存在")
		}
		return nil, fmt.Errorf("查询模块失败: %w", err)
	}

	// 检查用户名是否在该模块下已存在
	var existingUser models.UserVPN
	if err := uvs.db.Where("module_id = ? AND username = ?", config.ModuleID, config.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("该模块下用户名已存在")
	}

	// 生成WireGuard密钥对
	keyPair, err := wireguard.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("生成密钥对失败: %w", err)
	}

	// 生成预共享密钥
	presharedKey, err := wireguard.GeneratePresharedKey()
	if err != nil {
		return nil, fmt.Errorf("生成预共享密钥失败: %w", err)
	}

	// 获取模块关联的接口配置用于AllowedIPs生成
	var wgInterface models.WireGuardInterface
	if err := uvs.db.First(&wgInterface, module.InterfaceID).Error; err != nil {
		return nil, fmt.Errorf("获取WireGuard接口配置失败: %w", err)
	}

	// 从模块关联的接口获取可用IP
	moduleService := NewModuleService()
	ipAddress, err := moduleService.getAvailableIPForInterface(module.InterfaceID)
	if err != nil {
		return nil, fmt.Errorf("分配IP地址失败: %w", err)
	}

	// 设置默认值
	allowedIPs := config.AllowedIPs
	fmt.Printf("🔍 [AllowedIPs生成] 开始生成用户VPN AllowedIPs\n")
	fmt.Printf("🔍 [AllowedIPs生成] 用户传入的AllowedIPs: '%s'\n", config.AllowedIPs)
	fmt.Printf("🔍 [AllowedIPs生成] 接口网段(wgInterface.Network): '%s'\n", wgInterface.Network)
	fmt.Printf("🔍 [AllowedIPs生成] 模块内网段(module.AllowedIPs): '%s'\n", module.AllowedIPs)

	if allowedIPs == "" || allowedIPs == "0.0.0.0/0" {
		fmt.Printf("🔍 [AllowedIPs生成] 用户未指定AllowedIPs或使用默认全网访问，开始智能生成\n")
		// 智能生成AllowedIPs：当前接口的VPN网段 + 模块配置的内网段
		allowedIPs = wgInterface.Network // 首先添加VPN网段，如10.0.8.0/24
		fmt.Printf("🔍 [AllowedIPs生成] 添加VPN网段: '%s'\n", allowedIPs)

		// 添加模块配置的内网段（从数据库中读取）
		// 排除数据库默认值 "192.168.1.0/24"，因为这通常不是实际的内网段
		if module.AllowedIPs != "" && module.AllowedIPs != "192.168.1.0/24" {
			allowedIPs += fmt.Sprintf(", %s", module.AllowedIPs)
			fmt.Printf("🔍 [AllowedIPs生成] 添加模块内网段，最终结果: '%s'\n", allowedIPs)
		} else {
			fmt.Printf("🔍 [AllowedIPs生成] 模块内网段为空或为默认值，跳过添加\n")
		}
		// 注意：如果模块没有配置有效的内网段，则只允许访问VPN网段
	} else {
		fmt.Printf("🔍 [AllowedIPs生成] 使用用户指定的AllowedIPs: '%s'\n", allowedIPs)
	}

	fmt.Printf("🎯 [AllowedIPs生成] 最终生成的AllowedIPs: '%s'\n", allowedIPs)

	maxDevices := config.MaxDevices
	if maxDevices <= 0 {
		maxDevices = 1 // 默认1个设备
	}

	// 创建用户VPN记录
	userVPN := &models.UserVPN{
		ModuleID:     config.ModuleID,
		Username:     config.Username,
		Email:        config.Email,
		Description:  config.Description,
		PublicKey:    keyPair.PublicKey,
		PrivateKey:   keyPair.PrivateKey,
		PresharedKey: presharedKey,
		IPAddress:    ipAddress,
		Status:       models.UserVPNStatusOffline,
		AllowedIPs:   allowedIPs,
		PersistentKA: 25,
		ExpiresAt:    config.ExpiresAt,
		IsActive:     true,
		MaxDevices:   maxDevices,
	}

	if err := uvs.db.Create(userVPN).Error; err != nil {
		// 如果创建失败，释放IP地址
		moduleService.releaseIPForInterface(module.InterfaceID, ipAddress)
		return nil, fmt.Errorf("创建用户VPN失败: %w", err)
	}

	// 分配IP地址
	if err := moduleService.allocateIPForInterface(module.InterfaceID, ipAddress, userVPN.ID); err != nil {
		uvs.db.Delete(userVPN)
		return nil, fmt.Errorf("分配IP地址失败: %w", err)
	}

	// 自动更新WireGuard接口配置
	if err := moduleService.updateInterfaceConfig(module.InterfaceID); err != nil {
		// 记录错误但不影响用户VPN创建成功
		fmt.Printf("警告：更新WireGuard配置失败 - 接口ID: %d, 错误: %v\n", module.InterfaceID, err)
	}

	fmt.Printf("用户VPN创建成功 - 模块ID: %d, 用户: %s, IP: %s\n", config.ModuleID, config.Username, ipAddress)

	return userVPN, nil
}

// GetUserVPNsByModule 获取指定模块的用户VPN列表
func (uvs *UserVPNService) GetUserVPNsByModule(moduleID uint, page, pageSize int) ([]models.UserVPN, int64, error) {
	var userVPNs []models.UserVPN
	var total int64

	query := uvs.db.Model(&models.UserVPN{}).Where("module_id = ?", moduleID)

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计用户VPN数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&userVPNs).Error; err != nil {
		return nil, 0, fmt.Errorf("查询用户VPN列表失败: %w", err)
	}

	return userVPNs, total, nil
}

// GetUserVPN 获取单个用户VPN信息
func (uvs *UserVPNService) GetUserVPN(id uint) (*models.UserVPN, error) {
	var userVPN models.UserVPN
	if err := uvs.db.Preload("Module").First(&userVPN, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户VPN不存在")
		}
		return nil, fmt.Errorf("查询用户VPN失败: %w", err)
	}

	return &userVPN, nil
}

// GenerateUserVPNConfig 生成用户VPN配置文件
func (uvs *UserVPNService) GenerateUserVPNConfig(id uint) (string, error) {
	userVPN, err := uvs.GetUserVPN(id)
	if err != nil {
		return "", err
	}

	// 获取模块关联的接口信息
	var wgInterface models.WireGuardInterface
	if err := uvs.db.First(&wgInterface, userVPN.Module.InterfaceID).Error; err != nil {
		return "", fmt.Errorf("获取WireGuard接口配置失败: %w", err)
	}

	// 获取端点配置，使用与模块配置相同的智能选择逻辑：
	var serverEndpoint string

	// 1. 优先使用配置文件中的服务器IP + 接口端口
	if cfg := getGlobalConfigForUserVPN(); cfg != nil && cfg.App.ServerIP != "" {
		serverEndpoint = fmt.Sprintf("%s:%d", cfg.App.ServerIP, wgInterface.ListenPort)
	} else {
		// 2. 然后使用系统配置的endpoint
		if systemEndpoint, err := database.GetSystemConfig("server.endpoint"); err == nil && systemEndpoint != "" {
			serverEndpoint = systemEndpoint
		} else {
			// 3. 最后兜底：动态构建端点 vpn.eitec.com + 接口端口
			serverEndpoint = fmt.Sprintf("vpn.eitec.com:%d", wgInterface.ListenPort)
		}
	}

	// 生成用户配置文件
	fmt.Printf("📄 [配置生成] 开始生成用户VPN配置文件\n")
	fmt.Printf("📄 [配置生成] 用户ID: %d, 用户名: %s\n", id, userVPN.Username)
	fmt.Printf("📄 [配置生成] 数据库中存储的AllowedIPs: '%s'\n", userVPN.AllowedIPs)
	fmt.Printf("📄 [配置生成] 服务端点: %s\n", serverEndpoint)

	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/32
DNS = %s

[Peer]
PublicKey = %s
PresharedKey = %s
Endpoint = %s
AllowedIPs = %s
PersistentKeepalive = %d`,
		userVPN.PrivateKey,
		userVPN.IPAddress,
		wgInterface.DNS,
		wgInterface.PublicKey,
		userVPN.PresharedKey,
		serverEndpoint,
		userVPN.AllowedIPs,
		userVPN.PersistentKA)

	fmt.Printf("✅ [配置生成] 配置文件生成完成 - 用户ID: %d, 用户名: %s, AllowedIPs: %s\n", id, userVPN.Username, userVPN.AllowedIPs)

	return config, nil
}

// UpdateUserVPN 更新用户VPN信息
func (uvs *UserVPNService) UpdateUserVPN(id uint, updates map[string]interface{}) error {
	result := uvs.db.Model(&models.UserVPN{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("更新用户VPN失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("用户VPN不存在")
	}

	fmt.Printf("用户VPN信息更新 - 用户VPN ID: %d\n", id)

	return nil
}

// DeleteUserVPN 删除用户VPN
func (uvs *UserVPNService) DeleteUserVPN(id uint) error {
	userVPN, err := uvs.GetUserVPN(id)
	if err != nil {
		return err
	}

	// 保存接口ID用于后续配置更新
	interfaceID := userVPN.Module.InterfaceID

	// 释放IP地址
	moduleService := NewModuleService()
	if err := moduleService.releaseIPForInterface(interfaceID, userVPN.IPAddress); err != nil {
		return fmt.Errorf("释放IP地址失败: %w", err)
	}

	// 删除用户VPN记录（硬删除）
	if err := uvs.db.Unscoped().Delete(&models.UserVPN{}, id).Error; err != nil {
		return fmt.Errorf("删除用户VPN失败: %w", err)
	}

	// 自动更新WireGuard接口配置
	if err := moduleService.updateInterfaceConfig(interfaceID); err != nil {
		// 记录错误但不影响删除成功
		fmt.Printf("警告：更新WireGuard配置失败 - 接口ID: %d, 错误: %v\n", interfaceID, err)
	}

	fmt.Printf("用户VPN删除 - 用户VPN ID: %d, 用户名: %s\n", id, userVPN.Username)

	return nil
}

// UpdateUserVPNStatus 更新用户VPN状态
func (uvs *UserVPNService) UpdateUserVPNStatus(id uint, status models.UserVPNStatus) error {
	result := uvs.db.Model(&models.UserVPN{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    status,
		"last_seen": time.Now(),
	})

	if result.Error != nil {
		return fmt.Errorf("更新用户VPN状态失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("用户VPN不存在")
	}

	fmt.Printf("状态变更为: %s - 用户VPN ID: %d\n", status.String(), id)

	return nil
}

// GetUserVPNStats 获取模块的用户VPN统计信息
func (uvs *UserVPNService) GetUserVPNStats(moduleID uint) (map[string]int64, error) {
	stats := make(map[string]int64)

	// 总用户数
	var total int64
	if err := uvs.db.Model(&models.UserVPN{}).Where("module_id = ?", moduleID).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("统计总用户数失败: %w", err)
	}
	stats["total"] = total

	// 在线用户数
	var online int64
	if err := uvs.db.Model(&models.UserVPN{}).Where("module_id = ? AND status = ?", moduleID, models.UserVPNStatusOnline).Count(&online).Error; err != nil {
		return nil, fmt.Errorf("统计在线用户数失败: %w", err)
	}
	stats["online"] = online

	// 离线用户数
	var offline int64
	if err := uvs.db.Model(&models.UserVPN{}).Where("module_id = ? AND status = ?", moduleID, models.UserVPNStatusOffline).Count(&offline).Error; err != nil {
		return nil, fmt.Errorf("统计离线用户数失败: %w", err)
	}
	stats["offline"] = offline

	// 活跃用户数
	var active int64
	if err := uvs.db.Model(&models.UserVPN{}).Where("module_id = ? AND is_active = ?", moduleID, true).Count(&active).Error; err != nil {
		return nil, fmt.Errorf("统计活跃用户数失败: %w", err)
	}
	stats["active"] = active

	return stats, nil
}
