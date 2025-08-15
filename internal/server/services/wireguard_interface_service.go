package services

import (
	"errors"
	"fmt"
	"net"
	"os"

	"strings"
	"time"

	"eitec-vpn/internal/server/database"
	"eitec-vpn/internal/server/models"
	"eitec-vpn/internal/shared/config"
	"eitec-vpn/internal/shared/wireguard"

	"os/exec"

	"gorm.io/gorm"
)

// WireGuardInterfaceService WireGuard接口管理服务
type WireGuardInterfaceService struct {
	db *gorm.DB
}

// NewWireGuardInterfaceService 创建WireGuard接口管理服务
func NewWireGuardInterfaceService() *WireGuardInterfaceService {
	return &WireGuardInterfaceService{
		db: database.DB,
	}
}

// CreateInterface 创建WireGuard接口
func (wis *WireGuardInterfaceService) CreateInterface(template *models.InterfaceTemplate) (*models.WireGuardInterface, error) {
	// 检查接口名是否已存在
	var existingInterface models.WireGuardInterface
	if err := wis.db.Where("name = ?", template.Name).First(&existingInterface).Error; err == nil {
		return nil, errors.New("接口名称已存在")
	}

	// 验证网络段
	if err := wis.validateNetwork(template.Network); err != nil {
		return nil, fmt.Errorf("网络段验证失败: %w", err)
	}

	// 检查端口是否已被使用
	if err := wis.checkPortAvailable(template.ListenPort); err != nil {
		return nil, fmt.Errorf("端口验证失败: %w", err)
	}

	// 生成服务器密钥对
	keyPair, err := wireguard.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("生成密钥对失败: %w", err)
	}

	// 计算服务器IP（通常是网络段的第一个IP）
	serverIP, err := wis.calculateServerIP(template.Network)
	if err != nil {
		return nil, fmt.Errorf("计算服务器IP失败: %w", err)
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
		SaveConfig:  true,
	}

	if err := wis.db.Create(wgInterface).Error; err != nil {
		return nil, fmt.Errorf("创建接口失败: %w", err)
	}

	// 为接口创建IP池
	if err := wis.createIPPoolForInterface(wgInterface); err != nil {
		wis.db.Delete(wgInterface)
		return nil, fmt.Errorf("创建IP池失败: %w", err)
	}

	return wgInterface, nil
}

// GetInterfaces 获取WireGuard接口列表
func (wis *WireGuardInterfaceService) GetInterfaces() ([]models.WireGuardInterface, error) {
	var interfaces []models.WireGuardInterface
	if err := wis.db.Preload("Modules").Find(&interfaces).Error; err != nil {
		return nil, fmt.Errorf("查询接口列表失败: %w", err)
	}
	return interfaces, nil
}

// GetInterfacesWithStatus 获取WireGuard接口列表（包含实时状态）
func (wis *WireGuardInterfaceService) GetInterfacesWithStatus() ([]InterfaceWithRealTimeStatus, error) {
	// 一次性预加载所有关联数据，避免N+1查询问题
	var interfaces []models.WireGuardInterface
	if err := wis.db.Preload("Modules.UserVPNs").Find(&interfaces).Error; err != nil {
		return nil, fmt.Errorf("查询接口列表失败: %w", err)
	}

	// 创建WireGuard状态检测服务
	showService := NewWireGuardShowService()

	var result []InterfaceWithRealTimeStatus
	for _, iface := range interfaces {
		interfaceStatus := InterfaceWithRealTimeStatus{
			WireGuardInterface: iface,
			IsActive:           false,
			PeerCount:          0,
			ActivePeers:        0,
			TotalTraffic:       ShowTrafficData{},
			ConfigExists:       showService.CheckConfigExists(iface.Name),
			ServiceStatus:      "inactive",
			Modules:            []ModuleWithStatus{},
		}

		// 获取实时状态
		if showInfo, err := showService.GetInterfaceInfo(iface.Name); err == nil && showInfo.IsActive {
			interfaceStatus.IsActive = showInfo.IsActive
			interfaceStatus.PeerCount = showInfo.PeerCount
			interfaceStatus.ActivePeers = showInfo.ActivePeers
			interfaceStatus.TotalTraffic = ShowTrafficData{
				RxBytes: showInfo.TotalTraffic.RxBytes,
				TxBytes: showInfo.TotalTraffic.TxBytes,
				RxMB:    "[WG] " + showInfo.TotalTraffic.RxMB,
				TxMB:    "[WG] " + showInfo.TotalTraffic.TxMB,
				Total:   "[WG] " + showInfo.TotalTraffic.Total,
			}
			interfaceStatus.LastHandshake = showInfo.LastHandshake
			interfaceStatus.ServiceStatus = "[WG] active"
		}

		// 处理关联模块的实时状态（数据已通过Preload加载）
		if len(iface.Modules) > 0 {
			for _, module := range iface.Modules {
				// 直接使用预加载的用户数据，避免循环查询数据库
				users := wis.convertUserVPNsToModuleUserInfo(module.UserVPNs, interfaceStatus.IsActive)

				moduleStatus := ModuleWithStatus{
					Module:            module,
					IsOnline:          false,
					LastSeen:          nil,
					LatestHandshake:   nil,
					TrafficStats:      ShowTrafficData{},
					CurrentEndpoint:   "",
					ConnectionQuality: "unknown",
					PingLatency:       -1,
					UserCount:         len(users),
					Users:             users,
				}

				// 如果有实时状态，更新模块信息和用户状态
				if interfaceStatus.IsActive {
					if showInfo, err := showService.GetInterfaceInfo(iface.Name); err == nil {
						// 更新模块状态
						if peer, exists := showInfo.Peers[module.PublicKey]; exists {
							moduleStatus.IsOnline = peer.IsOnline
							moduleStatus.LatestHandshake = peer.LatestHandshake
							moduleStatus.TrafficStats = ShowTrafficData{
								RxBytes: peer.TrafficStats.RxBytes,
								TxBytes: peer.TrafficStats.TxBytes,
								RxMB:    "[WG] " + peer.TrafficStats.RxMB,
								TxMB:    "[WG] " + peer.TrafficStats.TxMB,
								Total:   "[WG] " + peer.TrafficStats.Total,
							}
							moduleStatus.CurrentEndpoint = "[WG] " + peer.Endpoint
							moduleStatus.LastSeen = peer.LatestHandshake
						}

						// 根据wg show输出更新用户状态（数据库状态作废）
						for i := range moduleStatus.Users {
							userPublicKey := ""
							// 根据用户ID查找对应的UserVPN记录获取公钥
							for _, userVPN := range module.UserVPNs {
								if userVPN.ID == moduleStatus.Users[i].ID {
									userPublicKey = userVPN.PublicKey
									break
								}
							}

							// 只根据wg show输出判断用户是否在线
							if userPublicKey != "" {
								if userPeer, exists := showInfo.Peers[userPublicKey]; exists && userPeer.IsOnline {
									moduleStatus.Users[i].IsActive = true
								} else {
									moduleStatus.Users[i].IsActive = false
								}
							} else {
								moduleStatus.Users[i].IsActive = false
							}
						}
					} else {
						// 无法获取wg show信息，所有用户设为离线
						for i := range moduleStatus.Users {
							moduleStatus.Users[i].IsActive = false
						}
					}
				} else {
					// 接口未激活，所有用户设为离线
					for i := range moduleStatus.Users {
						moduleStatus.Users[i].IsActive = false
					}
				}

				interfaceStatus.Modules = append(interfaceStatus.Modules, moduleStatus)
			}
		}

		result = append(result, interfaceStatus)
	}

	return result, nil
}

// GetInterface 获取单个WireGuard接口
func (wis *WireGuardInterfaceService) GetInterface(id uint) (*models.WireGuardInterface, error) {
	var wgInterface models.WireGuardInterface
	if err := wis.db.Preload("Modules").First(&wgInterface, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("接口不存在")
		}
		return nil, fmt.Errorf("查询接口失败: %w", err)
	}
	return &wgInterface, nil
}

// GetInterfaceByName 根据名称获取接口
func (wis *WireGuardInterfaceService) GetInterfaceByName(name string) (*models.WireGuardInterface, error) {
	var wgInterface models.WireGuardInterface
	if err := wis.db.Where("name = ?", name).First(&wgInterface).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("接口不存在")
		}
		return nil, fmt.Errorf("查询接口失败: %w", err)
	}
	return &wgInterface, nil
}

// StartInterface 启动WireGuard接口
func (wis *WireGuardInterfaceService) StartInterface(id uint) error {
	wgInterface, err := wis.GetInterface(id)
	if err != nil {
		return err
	}

	// 检查接口是否已经在运行
	if wireguard.IsInterfaceUp(wgInterface.Name) {
		return fmt.Errorf("接口 %s 已经在运行", wgInterface.Name)
	}

	// 更新状态为启动中
	wis.db.Model(wgInterface).Update("status", models.InterfaceStatusStarting)

	// 检查配置文件是否存在
	configPath := fmt.Sprintf("/etc/wireguard/%s.conf", wgInterface.Name)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，生成配置文件
		configContent := wis.GenerateInterfaceConfig(wgInterface)
		if err := wireguard.WriteConfigFile(configPath, configContent); err != nil {
			wis.db.Model(wgInterface).Update("status", models.InterfaceStatusError)
			return fmt.Errorf("写入配置文件失败: %w", err)
		}
	}

	// 启动接口
	if err := wis.startWireGuardInterface(wgInterface.Name); err != nil {
		wis.db.Model(wgInterface).Update("status", models.InterfaceStatusError)
		return fmt.Errorf("启动接口失败: %w", err)
	}

	// 更新状态为运行中
	wis.db.Model(wgInterface).Update("status", models.InterfaceStatusUp)

	return nil
}

// startWireGuardInterface 启动WireGuard接口（内部方法）
func (wis *WireGuardInterfaceService) startWireGuardInterface(interfaceName string) error {
	// 使用wg-quick up命令启动接口
	cmd := exec.Command("wg-quick", "up", interfaceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("启动WireGuard接口失败: %s, 输出: %s", err.Error(), string(output))
	}
	return nil
}

// StopInterface 停止WireGuard接口
func (wis *WireGuardInterfaceService) StopInterface(id uint) error {
	wgInterface, err := wis.GetInterface(id)
	if err != nil {
		return err
	}

	// 检查接口是否在运行
	if !wireguard.IsInterfaceUp(wgInterface.Name) {
		// 接口已经停止，直接更新状态
		wis.db.Model(wgInterface).Update("status", models.InterfaceStatusDown)
		return nil
	}

	// 更新状态为停止中
	wis.db.Model(wgInterface).Update("status", models.InterfaceStatusStopping)

	// 停止接口
	if err := wis.stopWireGuardInterface(wgInterface.Name); err != nil {
		wis.db.Model(wgInterface).Update("status", models.InterfaceStatusError)
		return fmt.Errorf("停止接口失败: %w", err)
	}

	// 更新状态为已停止
	wis.db.Model(wgInterface).Update("status", models.InterfaceStatusDown)

	return nil
}

// stopWireGuardInterface 停止WireGuard接口（内部方法）
func (wis *WireGuardInterfaceService) stopWireGuardInterface(interfaceName string) error {
	// 使用wg-quick down命令停止接口
	cmd := exec.Command("wg-quick", "down", interfaceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 如果接口不存在，不算错误
		if strings.Contains(string(output), "does not exist") || strings.Contains(err.Error(), "does not exist") {
			return nil
		}
		return fmt.Errorf("停止WireGuard接口失败: %s, 输出: %s", err.Error(), string(output))
	}
	return nil
}

// DeleteInterface 删除WireGuard接口
func (wis *WireGuardInterfaceService) DeleteInterface(id uint) error {
	wgInterface, err := wis.GetInterface(id)
	if err != nil {
		return err
	}

	// 检查是否有关联的模块
	var moduleCount int64
	if err := wis.db.Model(&models.Module{}).Where("interface_id = ?", id).Count(&moduleCount).Error; err != nil {
		return fmt.Errorf("检查关联模块失败: %w", err)
	}

	if moduleCount > 0 {
		return fmt.Errorf("接口下还有 %d 个模块，无法删除", moduleCount)
	}

	// 停止接口
	if wgInterface.Status == models.InterfaceStatusUp {
		wis.StopInterface(id)
	}

	// 删除配置文件
	configPath := fmt.Sprintf("/etc/wireguard/%s.conf", wgInterface.Name)
	if _, err := os.Stat(configPath); err == nil {
		os.Remove(configPath)
	}

	// 删除IP池
	wis.db.Unscoped().Where("network = ?", wgInterface.Network).Delete(&models.IPPool{})

	// 硬删除接口记录
	if err := wis.db.Unscoped().Delete(wgInterface).Error; err != nil {
		return fmt.Errorf("删除接口失败: %w", err)
	}

	return nil
}

// UpdateInterfaceConfig 更新接口配置文件
func (wis *WireGuardInterfaceService) UpdateInterfaceConfig(id uint) error {
	wgInterface, err := wis.GetInterface(id)
	if err != nil {
		return err
	}

	// 生成最新的配置文件内容
	configContent := wis.GenerateInterfaceConfig(wgInterface)
	configPath := fmt.Sprintf("/etc/wireguard/%s.conf", wgInterface.Name)

	// 写入配置文件
	if err := wireguard.WriteConfigFile(configPath, configContent); err != nil {
		return fmt.Errorf("更新配置文件失败: %w", err)
	}

	// 如果接口正在运行，重新加载配置
	if wireguard.IsInterfaceUp(wgInterface.Name) {
		if err := wis.reloadInterface(wgInterface.Name); err != nil {
			return fmt.Errorf("重新加载接口配置失败: %w", err)
		}
	}

	return nil
}

// reloadInterface 重新加载接口配置
func (wis *WireGuardInterfaceService) reloadInterface(interfaceName string) error {
	// 停止接口
	if err := wis.stopWireGuardInterface(interfaceName); err != nil {
		return fmt.Errorf("停止接口失败: %w", err)
	}

	// 重新启动接口
	if err := wis.startWireGuardInterface(interfaceName); err != nil {
		return fmt.Errorf("重新启动接口失败: %w", err)
	}

	return nil
}

// validateNetwork 验证网络段
func (wis *WireGuardInterfaceService) validateNetwork(network string) error {
	_, ipNet, err := net.ParseCIDR(network)
	if err != nil {
		return fmt.Errorf("无效的网络段格式: %w", err)
	}

	// 检查网络段是否已被使用
	var existingInterface models.WireGuardInterface
	if err := wis.db.Where("network = ?", network).First(&existingInterface).Error; err == nil {
		return errors.New("网络段已被使用")
	}

	// 检查是否为私有网络
	if !ipNet.IP.IsPrivate() {
		return errors.New("必须使用私有网络段")
	}

	return nil
}

// checkPortAvailable 检查端口是否可用
func (wis *WireGuardInterfaceService) checkPortAvailable(port int) error {
	if port < 1024 || port > 65535 {
		return errors.New("端口必须在1024-65535范围内")
	}

	var existingInterface models.WireGuardInterface
	if err := wis.db.Where("listen_port = ?", port).First(&existingInterface).Error; err == nil {
		return errors.New("端口已被使用")
	}

	return nil
}

// calculateServerIP 计算服务器IP地址
func (wis *WireGuardInterfaceService) calculateServerIP(network string) (string, error) {
	_, ipNet, err := net.ParseCIDR(network)
	if err != nil {
		return "", err
	}

	// 使用网络段的第一个可用IP作为服务器IP
	ip := ipNet.IP.To4()
	if ip == nil {
		return "", errors.New("不支持IPv6")
	}

	// 第一个IP通常是网络地址，第二个IP作为服务器IP
	ip[3] = ip[3] + 1
	return ip.String(), nil
}

// createIPPoolForInterface 为接口创建IP池
func (wis *WireGuardInterfaceService) createIPPoolForInterface(wgInterface *models.WireGuardInterface) error {
	_, ipNet, err := net.ParseCIDR(wgInterface.Network)
	if err != nil {
		return err
	}

	// 生成IP池
	ip := ipNet.IP.To4()
	if ip == nil {
		return errors.New("不支持IPv6")
	}

	// 跳过网络地址和服务器IP，从第三个IP开始
	startIP := make(net.IP, 4)
	copy(startIP, ip)
	startIP[3] = startIP[3] + 2

	// 计算可用IP数量
	ones, bits := ipNet.Mask.Size()
	availableIPs := 1<<(bits-ones) - 3 // 减去网络地址、广播地址和服务器IP

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
		if err := wis.db.CreateInBatches(ipPools, 100).Error; err != nil {
			return fmt.Errorf("批量创建IP池失败: %w", err)
		}
	}

	return nil
}

// generateInterfaceConfig 生成接口配置
func (wis *WireGuardInterfaceService) GenerateInterfaceConfig(wgInterface *models.WireGuardInterface) string {
	var config strings.Builder

	// Interface部分
	config.WriteString("[Interface]\n")
	config.WriteString(fmt.Sprintf("PrivateKey = %s\n", wgInterface.PrivateKey))
	config.WriteString(fmt.Sprintf("Address = %s/24\n", wgInterface.ServerIP))
	config.WriteString(fmt.Sprintf("ListenPort = %d\n", wgInterface.ListenPort))

	if wgInterface.DNS != "" {
		config.WriteString(fmt.Sprintf("DNS = %s\n", wgInterface.DNS))
	}

	if wgInterface.MTU > 0 {
		config.WriteString(fmt.Sprintf("MTU = %d\n", wgInterface.MTU))
	}

	if wgInterface.SaveConfig {
		config.WriteString("SaveConfig = true\n")
	}

	// 基础PostUp/PostDown规则 - 参考用户成功配置
	// 确定网络接口名称
	networkInterface := "eth0" // 默认值
	if wgInterface.NetworkInterface != "" {
		networkInterface = wgInterface.NetworkInterface
	}

	if wgInterface.PostUp != "" {
		config.WriteString(fmt.Sprintf("PostUp = %s\n", wgInterface.PostUp))
	} else {
		// 使用成功验证的规则格式：简洁且使用%i占位符和动态网络接口
		config.WriteString(fmt.Sprintf("PostUp = iptables -A FORWARD -i %%i -j ACCEPT; iptables -A FORWARD -o %%i -j ACCEPT; iptables -t nat -A POSTROUTING -o %s -j MASQUERADE\n", networkInterface))
	}

	if wgInterface.PostDown != "" {
		config.WriteString(fmt.Sprintf("PostDown = %s\n", wgInterface.PostDown))
	} else {
		// 对应的清理规则
		config.WriteString(fmt.Sprintf("PostDown = iptables -D FORWARD -i %%i -j ACCEPT; iptables -D FORWARD -o %%i -j ACCEPT; iptables -t nat -D POSTROUTING -o %s -j MASQUERADE\n", networkInterface))
	}

	// 获取所有模块信息（用于生成Peer配置）
	var modules []models.Module
	wis.db.Where("interface_id = ?", wgInterface.ID).Find(&modules)

	// 注意：不再自动生成硬编码的iptables规则
	// 用户反馈：这些规则不够灵活，应该由用户自定义或使用默认规则

	if wgInterface.PreUp != "" {
		config.WriteString(fmt.Sprintf("PreUp = %s\n", wgInterface.PreUp))
	}

	if wgInterface.PreDown != "" {
		config.WriteString(fmt.Sprintf("PreDown = %s\n", wgInterface.PreDown))
	}

	// Peer部分 - 添加所有关联的模块
	for _, module := range modules {
		config.WriteString("\n[Peer]\n")
		config.WriteString(fmt.Sprintf("# %s - %s\n", module.Name, module.Location))
		config.WriteString(fmt.Sprintf("PublicKey = %s\n", module.PublicKey))

		// 添加预共享密钥（如果有）
		if module.PresharedKey != "" {
			config.WriteString(fmt.Sprintf("PresharedKey = %s\n", module.PresharedKey))
		}

		// AllowedIPs格式：模块VPN_IP/32, 内网网段
		config.WriteString(fmt.Sprintf("AllowedIPs = %s/32", module.IPAddress))
		if module.AllowedIPs != "" && module.AllowedIPs != "192.168.1.0/24" {
			config.WriteString(fmt.Sprintf(", %s", module.AllowedIPs))
		}
		config.WriteString("\n")

		// 添加Endpoint（如果有配置）
		if module.Endpoint != "" {
			config.WriteString(fmt.Sprintf("Endpoint = %s\n", module.Endpoint))
		}

		if module.PersistentKA > 0 {
			config.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", module.PersistentKA))
		}
	}

	// Peer部分 - 添加所有关联的用户VPN
	var userVPNs []models.UserVPN
	wis.db.Joins("JOIN modules ON user_vpns.module_id = modules.id").
		Where("modules.interface_id = ? AND user_vpns.is_active = ?", wgInterface.ID, true).
		Find(&userVPNs)

	for _, userVPN := range userVPNs {
		config.WriteString("\n[Peer]\n")
		config.WriteString(fmt.Sprintf("# User: %s\n", userVPN.Username))
		config.WriteString(fmt.Sprintf("PublicKey = %s\n", userVPN.PublicKey))

		// 参考用户成功配置：AllowedIPs = 10.10.0.3/32
		// 只包含用户的VPN IP，不包含网段
		config.WriteString(fmt.Sprintf("AllowedIPs = %s/32\n", userVPN.IPAddress))

		// 添加预共享密钥（如果有）
		if userVPN.PresharedKey != "" {
			config.WriteString(fmt.Sprintf("PresharedKey = %s\n", userVPN.PresharedKey))
		}

		// 添加Endpoint（如果需要）
		// 注意：用户客户端配置中的Endpoint是服务器端点，服务端配置中不需要

		if userVPN.PersistentKA > 0 {
			config.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", userVPN.PersistentKA))
		}
	}

	return config.String()
}

// GetAvailableIPForInterface 为指定接口获取可用IP
func (wis *WireGuardInterfaceService) GetAvailableIPForInterface(interfaceID uint) (string, error) {
	wgInterface, err := wis.GetInterface(interfaceID)
	if err != nil {
		return "", err
	}

	var ipPool models.IPPool
	if err := wis.db.Where("network = ? AND is_used = ?", wgInterface.Network, false).First(&ipPool).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("接口 %s 没有可用的IP地址", wgInterface.Name)
		}
		return "", fmt.Errorf("查询可用IP失败: %w", err)
	}

	return ipPool.IPAddress, nil
}

// GetInterfaceTrafficStats 获取接口流量统计
func (wis *WireGuardInterfaceService) GetInterfaceTrafficStats(interfaceName string) (*models.TrafficStats, error) {
	// 获取WireGuard状态信息
	peers, err := wireguard.GetWireGuardStatus(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("获取WireGuard状态失败: %w", err)
	}

	var totalRx, totalTx uint64
	for _, peer := range peers {
		totalRx += peer.TransferRxBytes
		totalTx += peer.TransferTxBytes
	}

	return &models.TrafficStats{
		InterfaceName: interfaceName,
		TotalRx:       totalRx,
		TotalTx:       totalTx,
		TotalBytes:    totalRx + totalTx,
		PeerCount:     len(peers),
		Timestamp:     time.Now(),
	}, nil
}

// GetAllInterfacesTrafficStats 获取所有接口的流量统计
func (wis *WireGuardInterfaceService) GetAllInterfacesTrafficStats() ([]models.TrafficStats, error) {
	interfaces, err := wis.GetInterfaces()
	if err != nil {
		return nil, err
	}

	var allStats []models.TrafficStats
	for _, iface := range interfaces {
		// 只统计运行中的接口
		if iface.Status == models.InterfaceStatusUp {
			stats, err := wis.GetInterfaceTrafficStats(iface.Name)
			if err != nil {
				// 如果获取单个接口统计失败，记录错误但继续处理其他接口
				fmt.Printf("获取接口 %s 流量统计失败: %v\n", iface.Name, err)
				continue
			}
			allStats = append(allStats, *stats)
		}
	}

	return allStats, nil
}

// =====================================================
// WireGuard Show 命令解析功能
// =====================================================

// WireGuardShowInfo WireGuard show命令输出的完整信息
type WireGuardShowInfo struct {
	InterfaceName string              `json:"interface_name"`
	PublicKey     string              `json:"public_key"`
	ListenPort    int                 `json:"listen_port"`
	Peers         []WireGuardPeerInfo `json:"peers"`
	TotalPeers    int                 `json:"total_peers"`
	OnlinePeers   int                 `json:"online_peers"`
}

// WireGuardPeerInfo WireGuard对等端信息
type WireGuardPeerInfo struct {
	PublicKey           string    `json:"public_key"`
	Endpoint            string    `json:"endpoint"`
	AllowedIPs          []string  `json:"allowed_ips"`
	LatestHandshake     time.Time `json:"latest_handshake"`
	TransferRxBytes     uint64    `json:"transfer_rx_bytes"`
	TransferTxBytes     uint64    `json:"transfer_tx_bytes"`
	TransferRxFormatted string    `json:"transfer_rx_formatted"`
	TransferTxFormatted string    `json:"transfer_tx_formatted"`
	PersistentKeepalive int       `json:"persistent_keepalive"`
	IsOnline            bool      `json:"is_online"`
	LastSeenAgo         string    `json:"last_seen_ago"`
}

// ParseWireGuardShow 解析wg show命令的输出
func (wis *WireGuardInterfaceService) ParseWireGuardShow(interfaceName string) (*WireGuardShowInfo, error) {
	// 执行wg show命令
	cmd := exec.Command("wg", "show", interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行wg show命令失败: %w", err)
	}

	// 解析输出
	return wis.parseWireGuardShowOutput(string(output))
}

// parseWireGuardShowOutput 解析wg show命令的文本输出
func (wis *WireGuardInterfaceService) parseWireGuardShowOutput(output string) (*WireGuardShowInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("wg show输出为空")
	}

	info := &WireGuardShowInfo{
		Peers: make([]WireGuardPeerInfo, 0),
	}

	var currentPeer *WireGuardPeerInfo
	var currentSection string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 判断当前行属于哪个部分
		if strings.HasPrefix(line, "interface:") {
			currentSection = "interface"
			info.InterfaceName = strings.TrimSpace(strings.TrimPrefix(line, "interface:"))
		} else if strings.HasPrefix(line, "peer:") {
			currentSection = "peer"
			// 如果有上一个peer，先保存
			if currentPeer != nil {
				info.Peers = append(info.Peers, *currentPeer)
			}
			// 开始新的peer
			currentPeer = &WireGuardPeerInfo{
				PublicKey: strings.TrimSpace(strings.TrimPrefix(line, "peer:")),
			}
		} else if currentSection == "interface" {
			// 解析接口信息
			if strings.HasPrefix(line, "public key:") {
				info.PublicKey = strings.TrimSpace(strings.TrimPrefix(line, "public key:"))
			} else if strings.HasPrefix(line, "listening port:") {
				portStr := strings.TrimSpace(strings.TrimPrefix(line, "listening port:"))
				if port, err := parsePort(portStr); err == nil {
					info.ListenPort = port
				}
			}
		} else if currentSection == "peer" && currentPeer != nil {
			// 解析peer信息
			if strings.HasPrefix(line, "endpoint:") {
				currentPeer.Endpoint = strings.TrimSpace(strings.TrimPrefix(line, "endpoint:"))
			} else if strings.HasPrefix(line, "allowed ips:") {
				ipsStr := strings.TrimSpace(strings.TrimPrefix(line, "allowed ips:"))
				currentPeer.AllowedIPs = parseAllowedIPs(ipsStr)
			} else if strings.HasPrefix(line, "latest handshake:") {
				handshakeStr := strings.TrimSpace(strings.TrimPrefix(line, "latest handshake:"))
				currentPeer.LatestHandshake, currentPeer.LastSeenAgo = parseLatestHandshake(handshakeStr)
			} else if strings.HasPrefix(line, "transfer:") {
				transferStr := strings.TrimSpace(strings.TrimPrefix(line, "transfer:"))
				currentPeer.TransferRxBytes, currentPeer.TransferTxBytes, currentPeer.TransferRxFormatted, currentPeer.TransferTxFormatted = parseTransfer(transferStr)
			} else if strings.HasPrefix(line, "persistent keepalive:") {
				keepaliveStr := strings.TrimSpace(strings.TrimPrefix(line, "persistent keepalive:"))
				currentPeer.PersistentKeepalive = parsePersistentKeepalive(keepaliveStr)
			}
		}
	}

	// 保存最后一个peer
	if currentPeer != nil {
		info.Peers = append(info.Peers, *currentPeer)
	}

	// 计算统计信息
	info.TotalPeers = len(info.Peers)
	info.OnlinePeers = 0
	now := time.Now()

	for i := range info.Peers {
		// 使用统一的超时常量判断peer是否在线
		if info.Peers[i].LatestHandshake.After(now.Add(-config.WireGuardOnlineTimeout)) {
			info.Peers[i].IsOnline = true
			info.OnlinePeers++
		} else {
			info.Peers[i].IsOnline = false
		}
	}

	return info, nil
}

// parsePort 解析端口号
func parsePort(portStr string) (int, error) {
	var port int
	_, err := fmt.Sscanf(portStr, "%d", &port)
	return port, err
}

// parseAllowedIPs 解析允许的IP地址列表
func parseAllowedIPs(ipsStr string) []string {
	ips := strings.Split(ipsStr, ",")
	result := make([]string, 0, len(ips))
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			result = append(result, ip)
		}
	}
	return result
}

// parseLatestHandshake 解析最新握手时间
func parseLatestHandshake(handshakeStr string) (time.Time, string) {
	now := time.Now()

	// 处理相对时间格式
	if strings.Contains(handshakeStr, "ago") {
		// 移除"ago"并解析时间
		timeStr := strings.TrimSpace(strings.TrimSuffix(handshakeStr, "ago"))

		// 解析各种时间格式
		if strings.Contains(timeStr, "second") {
			var seconds int
			fmt.Sscanf(timeStr, "%d second", &seconds)
			if strings.Contains(timeStr, "seconds") {
				fmt.Sscanf(timeStr, "%d seconds", &seconds)
			}
			return now.Add(-time.Duration(seconds) * time.Second), handshakeStr
		} else if strings.Contains(timeStr, "minute") {
			var minutes int
			fmt.Sscanf(timeStr, "%d minute", &minutes)
			if strings.Contains(timeStr, "minutes") {
				fmt.Sscanf(timeStr, "%d minutes", &minutes)
			}
			return now.Add(-time.Duration(minutes) * time.Minute), handshakeStr
		} else if strings.Contains(timeStr, "hour") {
			var hours int
			fmt.Sscanf(timeStr, "%d hour", &hours)
			if strings.Contains(timeStr, "hours") {
				fmt.Sscanf(timeStr, "%d hours", &hours)
			}
			return now.Add(-time.Duration(hours) * time.Hour), handshakeStr
		} else if strings.Contains(timeStr, "day") {
			var days int
			fmt.Sscanf(timeStr, "%d day", &days)
			if strings.Contains(timeStr, "days") {
				fmt.Sscanf(timeStr, "%d days", &days)
			}
			return now.Add(-time.Duration(days) * 24 * time.Hour), handshakeStr
		}
	}

	// 如果无法解析，返回当前时间
	return now, handshakeStr
}

// parseTransfer 解析传输数据
func parseTransfer(transferStr string) (rxBytes, txBytes uint64, rxFormatted, txFormatted string) {
	// 示例: "60.07 KiB received, 851.23 KiB sent"
	parts := strings.Split(transferStr, ",")
	if len(parts) != 2 {
		return 0, 0, "", ""
	}

	// 解析接收数据
	rxPart := strings.TrimSpace(parts[0])
	rxFormatted = rxPart
	if strings.Contains(rxPart, "received") {
		rxBytes = parseDataSize(rxPart)
	}

	// 解析发送数据
	txPart := strings.TrimSpace(parts[1])
	txFormatted = txPart
	if strings.Contains(txPart, "sent") {
		txBytes = parseDataSize(txPart)
	}

	return rxBytes, txBytes, rxFormatted, txFormatted
}

// parseDataSize 解析数据大小（KiB, MiB, GiB等）
func parseDataSize(sizeStr string) uint64 {
	var size float64
	var unit string

	// 提取数字和单位
	fmt.Sscanf(sizeStr, "%f %s", &size, &unit)

	// 根据单位转换为字节
	switch strings.ToLower(unit) {
	case "b", "byte", "bytes":
		return uint64(size)
	case "kib":
		return uint64(size * 1024)
	case "mib":
		return uint64(size * 1024 * 1024)
	case "gib":
		return uint64(size * 1024 * 1024 * 1024)
	case "kb":
		return uint64(size * 1000)
	case "mb":
		return uint64(size * 1000 * 1000)
	case "gb":
		return uint64(size * 1000 * 1000 * 1000)
	default:
		return uint64(size)
	}
}

// parsePersistentKeepalive 解析保活间隔
func parsePersistentKeepalive(keepaliveStr string) int {
	// 示例: "every 25 seconds"
	if strings.Contains(keepaliveStr, "every") {
		var seconds int
		fmt.Sscanf(keepaliveStr, "every %d seconds", &seconds)
		return seconds
	}
	return 0
}

// GetWireGuardShowInfo 获取WireGuard接口的show信息（便捷方法）
func (wis *WireGuardInterfaceService) GetWireGuardShowInfo(interfaceName string) (*WireGuardShowInfo, error) {
	return wis.ParseWireGuardShow(interfaceName)
}

// =====================================================
// 实时状态检测相关结构体和方法
// =====================================================

// InterfaceWithRealTimeStatus 带实时状态的接口信息
type InterfaceWithRealTimeStatus struct {
	models.WireGuardInterface

	// 实时状态信息（从wg show获取）
	IsActive      bool            `json:"is_active"`      // 接口是否激活
	PeerCount     int             `json:"peer_count"`     // 当前连接的peer数量
	ActivePeers   int             `json:"active_peers"`   // 活跃的peer数量
	TotalTraffic  ShowTrafficData `json:"total_traffic"`  // 总流量统计
	LastHandshake *time.Time      `json:"last_handshake"` // 最近握手时间

	// 系统状态
	ConfigExists  bool   `json:"config_exists"`  // 配置文件是否存在
	ServiceStatus string `json:"service_status"` // 系统服务状态

	// 模块信息
	Modules []ModuleWithStatus `json:"modules,omitempty"`
}

// ModuleWithStatus 带实时状态的模块信息
type ModuleWithStatus struct {
	models.Module

	// 实时状态信息（从wg show获取）
	IsOnline        bool            `json:"is_online"`        // 是否在线
	LastSeen        *time.Time      `json:"last_seen"`        // 最后见到时间
	LatestHandshake *time.Time      `json:"latest_handshake"` // 最新握手时间
	TrafficStats    ShowTrafficData `json:"traffic_stats"`    // 流量统计
	CurrentEndpoint string          `json:"current_endpoint"` // 当前连接端点

	// 连接质量
	ConnectionQuality string `json:"connection_quality"` // 连接质量评估
	PingLatency       int    `json:"ping_latency"`       // ping延迟(ms)

	// 用户信息（直接返回用户列表而不是数量）
	UserCount int              `json:"user_count"` // 关联的用户数量（保持兼容性）
	Users     []ModuleUserInfo `json:"users"`      // 用户详细信息列表
}

// ModuleUserInfo 模块用户信息
type ModuleUserInfo struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email,omitempty"`
	IsActive  bool   `json:"is_active"`
	IPAddress string `json:"ip_address"`
}

// ShowTrafficData 流量数据（使用show service的格式）
type ShowTrafficData struct {
	RxBytes uint64 `json:"rx_bytes"` // 接收字节数
	TxBytes uint64 `json:"tx_bytes"` // 发送字节数
	RxMB    string `json:"rx_mb"`    // 接收MB（格式化）
	TxMB    string `json:"tx_mb"`    // 发送MB（格式化）
	Total   string `json:"total"`    // 总流量（格式化）
}

// 旧的重复代码已移动到 wireguard_show_service.go

// getUserCountForModule 获取模块的用户数量
func (wis *WireGuardInterfaceService) getUserCountForModule(moduleID uint) int {
	var count int64
	wis.db.Model(&models.UserVPN{}).Where("module_id = ?", moduleID).Count(&count)
	return int(count)
}

// convertUserVPNsToModuleUserInfo 将预加载的UserVPN数据转换为ModuleUserInfo
// 用户状态统一由wg show输出决定，数据库状态作废
func (wis *WireGuardInterfaceService) convertUserVPNsToModuleUserInfo(userVPNs []models.UserVPN, interfaceIsActive bool) []ModuleUserInfo {
	if len(userVPNs) == 0 {
		return []ModuleUserInfo{}
	}

	users := make([]ModuleUserInfo, 0, len(userVPNs))
	for _, userVPN := range userVPNs {
		users = append(users, ModuleUserInfo{
			ID:        userVPN.ID,
			Username:  userVPN.Username,
			Email:     userVPN.Email,
			IsActive:  false, // 默认离线，后续根据wg show输出更新
			IPAddress: userVPN.IPAddress,
		})
	}

	return users
}

// getModuleUsers 获取模块的用户列表（保留为向后兼容）
func (wis *WireGuardInterfaceService) getModuleUsers(moduleID uint, interfaceIsActive bool) []ModuleUserInfo {
	var userVPNs []models.UserVPN
	if err := wis.db.Where("module_id = ?", moduleID).Find(&userVPNs).Error; err != nil {
		return []ModuleUserInfo{}
	}

	return wis.convertUserVPNsToModuleUserInfo(userVPNs, interfaceIsActive)
}
