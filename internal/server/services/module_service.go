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
func getGlobalConfig() *config.ServerConfig {
	// 直接从config包获取全局配置
	return config.GetGlobalServerConfig()
}

// ModuleConfig 模块配置结构
type ModuleConfig struct {
	Name                string
	Location            string
	Description         string
	InterfaceID         uint // 指定的WireGuard接口ID
	AllowedIPs          string
	LocalIP             string // 模块在内网的IP地址
	PersistentKeepalive int
	DNS                 string
	AutoGenerateKeys    bool
	AutoAssignIP        bool
	ConfigTemplate      string
}

// ModuleService 模块管理服务
type ModuleService struct {
	db *gorm.DB
}

// NewModuleService 创建模块管理服务
func NewModuleService() *ModuleService {
	return &ModuleService{
		db: database.DB,
	}
}

// CreateModuleWithConfig 使用配置创建新模块
func (ms *ModuleService) CreateModuleWithConfig(config *ModuleConfig) (*models.Module, error) {
	// 检查模块名是否已存在
	var existingModule models.Module
	if err := ms.db.Where("name = ?", config.Name).First(&existingModule).Error; err == nil {
		return nil, errors.New("模块名称已存在")
	}

	// 验证接口是否存在
	var wgInterface models.WireGuardInterface
	if err := ms.db.First(&wgInterface, config.InterfaceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("指定的WireGuard接口不存在")
		}
		return nil, fmt.Errorf("查询接口失败: %w", err)
	}

	// 检查接口是否已达到最大连接数
	var moduleCount int64
	if err := ms.db.Model(&models.Module{}).Where("interface_id = ?", config.InterfaceID).Count(&moduleCount).Error; err != nil {
		return nil, fmt.Errorf("检查接口连接数失败: %w", err)
	}

	if moduleCount >= int64(wgInterface.MaxPeers) {
		return nil, fmt.Errorf("接口 %s 已达到最大连接数 %d", wgInterface.Name, wgInterface.MaxPeers)
	}

	var keyPair *models.WireGuardKey
	var presharedKey string
	var err error

	// 根据配置生成或不生成密钥对
	if config.AutoGenerateKeys {
		keyPair, err = wireguard.GenerateKeyPair()
		if err != nil {
			return nil, fmt.Errorf("生成密钥对失败: %w", err)
		}

		// 同时生成预共享密钥增强安全性
		presharedKey, err = wireguard.GeneratePresharedKey()
		if err != nil {
			return nil, fmt.Errorf("生成预共享密钥失败: %w", err)
		}
	} else {
		// 如果不自动生成，使用空密钥，后续可手动设置
		keyPair = &models.WireGuardKey{
			PublicKey:  "",
			PrivateKey: "",
		}
	}

	var ipAddress string
	// 根据配置分配或不分配IP地址
	if config.AutoAssignIP {
		// 从指定接口的IP池中分配IP
		ipAddress, err = ms.getAvailableIPForInterface(config.InterfaceID)
		if err != nil {
			return nil, fmt.Errorf("分配IP地址失败: %w", err)
		}
	}

	// 根据配置模板设置参数
	allowedIPs := config.AllowedIPs
	persistentKA := config.PersistentKeepalive

	switch config.ConfigTemplate {
	case "high-security":
		// 高安全配置：更严格的网络限制
		if allowedIPs == "192.168.1.0/24" {
			allowedIPs = "192.168.1.0/28" // 更小的网段
		}
		persistentKA = 60 // 更长的保活间隔
	case "low-latency":
		// 低延迟配置：更频繁的保活
		persistentKA = 10 // 更短的保活间隔
	case "custom":
		// 自定义配置：使用用户提供的参数
		// 保持原有参数不变
	default: // "default"
		// 默认配置：已经设置好的参数
	}

	// 创建模块记录
	module := &models.Module{
		Name:         config.Name,
		Location:     config.Location,
		Description:  config.Description,
		InterfaceID:  config.InterfaceID,
		PublicKey:    keyPair.PublicKey,
		PrivateKey:   keyPair.PrivateKey,
		PresharedKey: presharedKey, // 保存预共享密钥
		IPAddress:    ipAddress,
		LocalIP:      config.LocalIP, // 保存模块的内网IP地址
		Status:       models.ModuleStatusUnconfigured,
		AllowedIPs:   allowedIPs,
		PersistentKA: persistentKA,
	}

	if err := ms.db.Create(module).Error; err != nil {
		// 如果创建失败，释放IP地址
		if config.AutoAssignIP && ipAddress != "" {
			ms.releaseIPForInterface(config.InterfaceID, ipAddress)
		}
		return nil, fmt.Errorf("创建模块失败: %w", err)
	}

	// 分配IP地址给模块
	if config.AutoAssignIP && ipAddress != "" {
		if err := ms.allocateIPForInterface(config.InterfaceID, ipAddress, module.ID); err != nil {
			ms.db.Delete(module)
			return nil, fmt.Errorf("分配IP地址失败: %w", err)
		}
	}

	// 保存模块配置到系统配置表
	ms.saveModuleConfig(module.ID, config)

	// 自动更新WireGuard接口配置
	if err := ms.updateInterfaceConfig(config.InterfaceID); err != nil {
		// 记录错误但不影响模块创建成功
		fmt.Printf("警告：更新WireGuard配置失败 - 接口ID: %d, 错误: %v\n", config.InterfaceID, err)
	}

	// 记录操作日志
	// 简化：使用标准日志而不是数据库日志
	fmt.Printf("模块创建成功 - 接口: %s, 模板: %s\n", wgInterface.Name, config.ConfigTemplate)

	return module, nil
}

// saveModuleConfig 保存模块扩展配置
func (ms *ModuleService) saveModuleConfig(moduleID uint, config *ModuleConfig) {
	// 保存模块的扩展配置到系统配置表
	configs := map[string]string{
		fmt.Sprintf("module.%d.description", moduleID):     config.Description,
		fmt.Sprintf("module.%d.local_ip", moduleID):        config.LocalIP,
		fmt.Sprintf("module.%d.dns", moduleID):             config.DNS,
		fmt.Sprintf("module.%d.config_template", moduleID): config.ConfigTemplate,
	}

	for key, value := range configs {
		if value != "" {
			database.SetSystemConfig(key, value)
		}
	}
}

// CreateModule 创建新模块
func (ms *ModuleService) CreateModule(name, location string) (*models.Module, error) {
	// 检查模块名是否已存在
	var existingModule models.Module
	if err := ms.db.Where("name = ?", name).First(&existingModule).Error; err == nil {
		return nil, errors.New("模块名称已存在")
	}

	// 生成WireGuard密钥对
	keyPair, err := wireguard.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("生成密钥对失败: %w", err)
	}

	// 分配IP地址
	ipAddress, err := database.GetAvailableIP()
	if err != nil {
		return nil, fmt.Errorf("分配IP地址失败: %w", err)
	}

	// 创建模块记录
	module := &models.Module{
		Name:         name,
		Location:     location,
		PublicKey:    keyPair.PublicKey,
		PrivateKey:   keyPair.PrivateKey,
		IPAddress:    ipAddress,
		Status:       models.ModuleStatusUnconfigured,
		AllowedIPs:   "192.168.1.0/24",
		PersistentKA: 25,
	}

	if err := ms.db.Create(module).Error; err != nil {
		// 如果创建失败，释放IP地址
		database.ReleaseIP(ipAddress)
		return nil, fmt.Errorf("创建模块失败: %w", err)
	}

	// 分配IP地址给模块
	if err := database.AllocateIP(ipAddress, module.ID); err != nil {
		ms.db.Delete(module)
		return nil, fmt.Errorf("分配IP地址失败: %w", err)
	}

	// 记录操作日志
	// 简化：使用标准日志而不是数据库日志
	fmt.Printf("模块创建成功 - 接口: %s, 模板: %s\n", "N/A", "N/A")

	return module, nil
}

// GetModule 获取模块信息
func (ms *ModuleService) GetModule(id uint) (*models.Module, error) {
	var module models.Module
	if err := ms.db.First(&module, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("模块不存在")
		}
		return nil, fmt.Errorf("查询模块失败: %w", err)
	}

	return &module, nil
}

// GetModuleByName 根据名称获取模块
func (ms *ModuleService) GetModuleByName(name string) (*models.Module, error) {
	var module models.Module
	if err := ms.db.Where("name = ?", name).First(&module).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("模块不存在")
		}
		return nil, fmt.Errorf("查询模块失败: %w", err)
	}

	return &module, nil
}

// GetModules 获取模块列表 (分页)
func (ms *ModuleService) GetModules(page, pageSize int, filters map[string]interface{}) ([]models.Module, int64, error) {
	var modules []models.Module
	var total int64

	query := ms.db.Model(&models.Module{})

	// 应用过滤条件
	for key, value := range filters {
		switch key {
		case "name":
			query = query.Where("name LIKE ?", "%"+value.(string)+"%")
		case "location":
			query = query.Where("location LIKE ?", "%"+value.(string)+"%")
		case "status":
			query = query.Where("status = ?", value)
		case "ip_address":
			query = query.Where("ip_address = ?", value)
		}
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计模块数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&modules).Error; err != nil {
		return nil, 0, fmt.Errorf("查询模块列表失败: %w", err)
	}

	return modules, total, nil
}

// UpdateModule 更新模块信息
func (ms *ModuleService) UpdateModule(id uint, updates map[string]interface{}) error {
	// 如果更新名称，检查是否重复
	if newName, exists := updates["name"]; exists {
		var existingModule models.Module
		if err := ms.db.Where("name = ? AND id != ?", newName, id).First(&existingModule).Error; err == nil {
			return errors.New("模块名称已存在")
		}
	}

	result := ms.db.Model(&models.Module{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("更新模块失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("模块不存在")
	}

	// 简化：使用标准日志而不是数据库日志
	fmt.Printf("模块信息更新 - 模块ID: %d\n", id)

	return nil
}

// DeleteModule 删除模块
func (ms *ModuleService) DeleteModule(id uint) error {
	module, err := ms.GetModule(id)
	if err != nil {
		return err
	}

	// 保存接口ID用于后续配置更新
	interfaceID := module.InterfaceID

	// 释放IP地址
	if err := ms.releaseIPForInterface(interfaceID, module.IPAddress); err != nil {
		return fmt.Errorf("释放IP地址失败: %w", err)
	}

	// 删除模块记录（硬删除）
	if err := ms.db.Unscoped().Delete(&models.Module{}, id).Error; err != nil {
		return fmt.Errorf("删除模块失败: %w", err)
	}

	// 自动更新WireGuard接口配置
	if err := ms.updateInterfaceConfig(interfaceID); err != nil {
		// 记录错误但不影响删除成功
		fmt.Printf("警告：更新WireGuard配置失败 - 接口ID: %d, 错误: %v\n", interfaceID, err)
	}

	// 简化：使用标准日志而不是数据库日志
	fmt.Printf("模块删除 - 模块ID: %d, 名称: %s\n", id, module.Name)

	return nil
}

// UpdateModuleStatus 更新模块状态
func (ms *ModuleService) UpdateModuleStatus(id uint, status models.ModuleStatus) error {
	result := ms.db.Model(&models.Module{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    status,
		"last_seen": time.Now(),
	})

	if result.Error != nil {
		return fmt.Errorf("更新模块状态失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("模块不存在")
	}

	// 记录状态变更日志
	// 简化：使用标准日志而不是数据库日志
	fmt.Printf("状态变更为: %s - 模块ID: %d\n", status.String(), id)

	return nil
}

// UpdateModuleTraffic 更新模块流量统计
func (ms *ModuleService) UpdateModuleTraffic(id uint, rxBytes, txBytes uint64) error {
	result := ms.db.Model(&models.Module{}).Where("id = ?", id).Updates(map[string]interface{}{
		"total_rx_bytes": rxBytes,
		"total_tx_bytes": txBytes,
		"last_seen":      time.Now(),
	})

	if result.Error != nil {
		return fmt.Errorf("更新模块流量失败: %w", result.Error)
	}

	return nil
}

// UpdateModuleHandshake 更新模块握手时间
func (ms *ModuleService) UpdateModuleHandshake(id uint, handshakeTime time.Time) error {
	result := ms.db.Model(&models.Module{}).Where("id = ?", id).Updates(map[string]interface{}{
		"latest_handshake": handshakeTime,
		"last_seen":        time.Now(),
	})

	if result.Error != nil {
		return fmt.Errorf("更新模块握手时间失败: %w", result.Error)
	}

	return nil
}

// GenerateModuleConfig 生成模块配置
func (ms *ModuleService) GenerateModuleConfig(id uint) (string, error) {
	module, err := ms.GetModule(id)
	if err != nil {
		return "", err
	}

	// 获取模块关联的WireGuard接口配置
	var wgInterface models.WireGuardInterface
	if err := ms.db.First(&wgInterface, module.InterfaceID).Error; err != nil {
		return "", fmt.Errorf("获取WireGuard接口配置失败: %w", err)
	}

	// 获取端点配置，按优先级选择：
	var serverEndpoint string

	// 1. 优先使用配置文件中的服务器IP + 接口端口
	if cfg := getGlobalConfig(); cfg != nil && cfg.App.ServerIP != "" {
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

	// 使用接口的DNS配置，如果没有则使用系统默认
	dns := wgInterface.DNS
	if dns == "" {
		if systemDNS, err := database.GetSystemConfig("wg.dns"); err == nil {
			dns = systemDNS
		} else {
			dns = "8.8.8.8,8.8.4.4" // 默认DNS
		}
	}

	// 获取模块配置的local_ip
	moduleLocalIP, _ := database.GetSystemConfig(fmt.Sprintf("module.%d.local_ip", id))

	// 生成配置 - 传递接口信息和local_ip
	config := wireguard.GenerateModuleConfigWithLocalIP(module, &wgInterface, serverEndpoint, dns, moduleLocalIP)

	// 记录配置生成日志
	fmt.Printf("生成模块配置 - 模块ID: %d, 接口: %s, 网络: %s, 端点: %s, LocalIP: %s\n", id, wgInterface.Name, wgInterface.Network, serverEndpoint, moduleLocalIP)

	return config, nil
}

// GeneratePeerConfig 生成运维端Peer配置
func (ms *ModuleService) GeneratePeerConfig(id uint) (string, error) {
	module, err := ms.GetModule(id)
	if err != nil {
		return "", err
	}

	// 使用与模块下载配置相同的逻辑，确保内网穿透功能
	allowedIPs := "10.10.0.1/32" // 服务器管理IP

	// 添加模块配置的内网网段，实现双向内网穿透
	if module.AllowedIPs != "" && module.AllowedIPs != "192.168.1.0/24" {
		// 自定义内网段
		allowedIPs += fmt.Sprintf(", %s", module.AllowedIPs)
	} else if module.AllowedIPs == "192.168.1.0/24" {
		// 默认内网段
		allowedIPs += ", 192.168.1.0/24"
	}

	// 生成运维端Peer配置
	config := wireguard.GeneratePeerConfig(module)

	// 记录配置生成日志
	fmt.Printf("生成运维端Peer配置 - 模块ID: %d, 内网网段: %s\n", id, module.AllowedIPs)

	return config, nil
}

// GetModuleStats 获取模块统计信息
func (ms *ModuleService) GetModuleStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// 总模块数
	var total int64
	if err := ms.db.Model(&models.Module{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("统计总模块数失败: %w", err)
	}
	stats["total"] = total

	// 在线模块数
	var online int64
	if err := ms.db.Model(&models.Module{}).Where("status = ?", models.ModuleStatusOnline).Count(&online).Error; err != nil {
		return nil, fmt.Errorf("统计在线模块数失败: %w", err)
	}
	stats["online"] = online

	// 离线模块数
	var offline int64
	if err := ms.db.Model(&models.Module{}).Where("status = ?", models.ModuleStatusOffline).Count(&offline).Error; err != nil {
		return nil, fmt.Errorf("统计离线模块数失败: %w", err)
	}
	stats["offline"] = offline

	// 警告模块数
	var warning int64
	if err := ms.db.Model(&models.Module{}).Where("status = ?", models.ModuleStatusWarning).Count(&warning).Error; err != nil {
		return nil, fmt.Errorf("统计警告模块数失败: %w", err)
	}
	stats["warning"] = warning

	// 未配置模块数
	var unconfigured int64
	if err := ms.db.Model(&models.Module{}).Where("status = ?", models.ModuleStatusUnconfigured).Count(&unconfigured).Error; err != nil {
		return nil, fmt.Errorf("统计未配置模块数失败: %w", err)
	}
	stats["unconfigured"] = unconfigured

	return stats, nil
}

// GetTotalTraffic 获取总流量统计
func (ms *ModuleService) GetTotalTraffic() (map[string]uint64, error) {
	traffic := make(map[string]uint64)

	// 总接收流量
	var totalRx uint64
	if err := ms.db.Model(&models.Module{}).Select("COALESCE(SUM(total_rx_bytes), 0)").Scan(&totalRx).Error; err != nil {
		return nil, fmt.Errorf("统计总接收流量失败: %w", err)
	}
	traffic["total_rx"] = totalRx

	// 总发送流量
	var totalTx uint64
	if err := ms.db.Model(&models.Module{}).Select("COALESCE(SUM(total_tx_bytes), 0)").Scan(&totalTx).Error; err != nil {
		return nil, fmt.Errorf("统计总发送流量失败: %w", err)
	}
	traffic["total_tx"] = totalTx

	traffic["total"] = totalRx + totalTx

	return traffic, nil
}

// GetRecentlyActiveModules 获取最近活跃的模块
func (ms *ModuleService) GetRecentlyActiveModules(limit int) ([]models.Module, error) {
	var modules []models.Module
	if err := ms.db.Where("last_seen IS NOT NULL").
		Order("last_seen DESC").
		Limit(limit).
		Find(&modules).Error; err != nil {
		return nil, fmt.Errorf("查询最近活跃模块失败: %w", err)
	}

	return modules, nil
}

// SyncModuleStatus 同步模块状态 (从WireGuard获取实际状态)
func (ms *ModuleService) SyncModuleStatus() error {
	// 获取WireGuard状态
	wgStatus, err := wireguard.GetWireGuardStatus("wg0")
	if err != nil {
		return fmt.Errorf("获取WireGuard状态失败: %w", err)
	}

	// 获取所有模块
	var modules []models.Module
	if err := ms.db.Find(&modules).Error; err != nil {
		return fmt.Errorf("查询模块列表失败: %w", err)
	}

	// 更新模块状态
	for _, module := range modules {
		if peer, exists := wgStatus[module.PublicKey]; exists {
			// 模块在线
			status := models.ModuleStatusOnline

			// 检查是否长时间未握手 (超过5分钟认为有问题)
			if time.Since(peer.LatestHandshake) > 5*time.Minute {
				status = models.ModuleStatusWarning
			}

			// 更新模块信息
			ms.db.Model(&module).Updates(map[string]interface{}{
				"status":           status,
				"latest_handshake": peer.LatestHandshake,
				"total_rx_bytes":   peer.TransferRxBytes,
				"total_tx_bytes":   peer.TransferTxBytes,
				"last_seen":        time.Now(),
			})
		} else {
			// 模块离线
			ms.db.Model(&module).Updates(map[string]interface{}{
				"status": models.ModuleStatusOffline,
			})
		}
	}

	return nil
}

// RegenerateModuleKeys 重新生成模块密钥
func (ms *ModuleService) RegenerateModuleKeys(id uint) (*models.Module, error) {
	// 检查模块是否存在
	_, err := ms.GetModule(id)
	if err != nil {
		return nil, err
	}

	// 生成新的密钥对
	keyPair, err := wireguard.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("生成密钥对失败: %w", err)
	}

	// 更新模块密钥
	updates := map[string]interface{}{
		"public_key":  keyPair.PublicKey,
		"private_key": keyPair.PrivateKey,
		"status":      models.ModuleStatusUnconfigured, // 需要重新配置
	}

	if err := ms.UpdateModule(id, updates); err != nil {
		return nil, err
	}

	// 记录密钥重生成日志
	// 简化：使用标准日志而不是数据库日志
	fmt.Printf("重新生成密钥对 - 模块ID: %d\n", id)

	// 返回更新后的模块
	return ms.GetModule(id)
}

// getAvailableIPForInterface 为指定接口获取可用IP
func (ms *ModuleService) getAvailableIPForInterface(interfaceID uint) (string, error) {
	// 获取接口信息
	var wgInterface models.WireGuardInterface
	if err := ms.db.First(&wgInterface, interfaceID).Error; err != nil {
		return "", fmt.Errorf("查询接口失败: %w", err)
	}

	// 从接口对应的网络段中获取可用IP
	var ipPool models.IPPool
	if err := ms.db.Where("network = ? AND is_used = ?", wgInterface.Network, false).First(&ipPool).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("接口 %s 没有可用的IP地址", wgInterface.Name)
		}
		return "", fmt.Errorf("查询可用IP失败: %w", err)
	}

	return ipPool.IPAddress, nil
}

// allocateIPForInterface 为指定接口分配IP地址给模块
func (ms *ModuleService) allocateIPForInterface(interfaceID uint, ip string, moduleID uint) error {
	// 获取接口信息
	var wgInterface models.WireGuardInterface
	if err := ms.db.First(&wgInterface, interfaceID).Error; err != nil {
		return fmt.Errorf("查询接口失败: %w", err)
	}

	result := ms.db.Model(&models.IPPool{}).
		Where("network = ? AND ip_address = ? AND is_used = ?", wgInterface.Network, ip, false).
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

// releaseIPForInterface 释放指定接口的IP地址
func (ms *ModuleService) releaseIPForInterface(interfaceID uint, ip string) error {
	// 获取接口信息
	var wgInterface models.WireGuardInterface
	if err := ms.db.First(&wgInterface, interfaceID).Error; err != nil {
		return fmt.Errorf("查询接口失败: %w", err)
	}

	result := ms.db.Model(&models.IPPool{}).
		Where("network = ? AND ip_address = ?", wgInterface.Network, ip).
		Updates(map[string]interface{}{
			"is_used":   false,
			"module_id": nil,
		})

	if result.Error != nil {
		return fmt.Errorf("释放IP地址失败: %w", result.Error)
	}

	return nil
}

// updateInterfaceConfig 更新WireGuard接口配置
func (ms *ModuleService) updateInterfaceConfig(interfaceID uint) error {
	// 获取接口信息
	var wgInterface models.WireGuardInterface
	if err := ms.db.First(&wgInterface, interfaceID).Error; err != nil {
		return fmt.Errorf("查询接口失败: %w", err)
	}

	// 创建接口服务实例
	interfaceService := NewWireGuardInterfaceService()

	// 重新生成配置文件（无论接口状态如何都要更新）
	configContent := interfaceService.GenerateInterfaceConfig(&wgInterface)
	configPath := fmt.Sprintf("/etc/wireguard/%s.conf", wgInterface.Name)

	// 写入配置文件
	if err := wireguard.WriteConfigFile(configPath, configContent); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	fmt.Printf("✅ 成功更新接口 %s 的配置文件: %s\n", wgInterface.Name, configPath)

	// 只有当接口正在运行时才重新加载WireGuard
	if wgInterface.Status == models.InterfaceStatusUp {
		if err := wireguard.RestartWireGuard(wgInterface.Name); err != nil {
			fmt.Printf("⚠️  重新加载WireGuard配置失败（但配置文件已更新）: %v\n", err)
			// 不返回错误，因为配置文件已经更新成功
		} else {
			fmt.Printf("🔄 已重新加载WireGuard配置: %s\n", wgInterface.Name)
		}
	} else {
		fmt.Printf("💡 接口 %s 未运行，仅更新配置文件（启动时将自动应用）\n", wgInterface.Name)
	}

	return nil
}
