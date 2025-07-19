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

	// 删除IP池
	wis.db.Where("network = ?", wgInterface.Network).Delete(&models.IPPool{})

	// 删除接口记录
	if err := wis.db.Delete(wgInterface).Error; err != nil {
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

	// 基础PostUp/PostDown规则
	if wgInterface.PostUp != "" {
		config.WriteString(fmt.Sprintf("PostUp = %s\n", wgInterface.PostUp))
	} else {
		// 默认规则：NAT + FORWARD + INPUT
		config.WriteString(fmt.Sprintf("PostUp = iptables -t nat -A POSTROUTING -s %s -o eth0 -j MASQUERADE; iptables -A INPUT -p udp -m udp --dport %d -j ACCEPT; iptables -A FORWARD -i %s -j ACCEPT; iptables -A FORWARD -o %s -j ACCEPT;\n",
			wgInterface.Network, wgInterface.ListenPort, wgInterface.Name, wgInterface.Name))
	}

	if wgInterface.PostDown != "" {
		config.WriteString(fmt.Sprintf("PostDown = %s\n", wgInterface.PostDown))
	} else {
		// 默认清理规则
		config.WriteString(fmt.Sprintf("PostDown = iptables -t nat -D POSTROUTING -s %s -o eth0 -j MASQUERADE; iptables -D INPUT -p udp -m udp --dport %d -j ACCEPT; iptables -D FORWARD -i %s -j ACCEPT; iptables -D FORWARD -o %s -j ACCEPT;\n",
			wgInterface.Network, wgInterface.ListenPort, wgInterface.Name, wgInterface.Name))
	}

	// 获取所有模块并生成内网穿透规则
	var modules []models.Module
	wis.db.Where("interface_id = ?", wgInterface.ID).Find(&modules)

	// 收集所有需要内网穿透的网段
	internalNetworks := make(map[string]bool)
	for _, module := range modules {
		if module.AllowedIPs != "" && module.AllowedIPs != "192.168.1.0/24" && module.AllowedIPs != wgInterface.Network {
			internalNetworks[module.AllowedIPs] = true
		}
	}

	// 为每个内网段添加FORWARD规则
	for network := range internalNetworks {
		config.WriteString(fmt.Sprintf("PostUp = iptables -I FORWARD -s %s -i %s -d %s -j ACCEPT\n", wgInterface.Network, wgInterface.Name, network))
		config.WriteString(fmt.Sprintf("PostUp = iptables -I FORWARD -s %s -i %s -d %s -j ACCEPT\n", network, wgInterface.Name, wgInterface.Network))
		config.WriteString(fmt.Sprintf("PostDown = iptables -D FORWARD -s %s -i %s -d %s -j ACCEPT\n", wgInterface.Network, wgInterface.Name, network))
		config.WriteString(fmt.Sprintf("PostDown = iptables -D FORWARD -s %s -i %s -d %s -j ACCEPT\n", network, wgInterface.Name, wgInterface.Network))
	}

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
		config.WriteString(fmt.Sprintf("AllowedIPs = %s/32", module.IPAddress))

		// 如果模块配置了内网访问，添加到AllowedIPs
		if module.AllowedIPs != "" && module.AllowedIPs != "192.168.1.0/24" {
			config.WriteString(fmt.Sprintf(",%s", module.AllowedIPs))
		}
		config.WriteString("\n")

		// 添加预共享密钥（如果有）
		if module.PresharedKey != "" {
			config.WriteString(fmt.Sprintf("PresharedKey = %s\n", module.PresharedKey))
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
		config.WriteString(fmt.Sprintf("AllowedIPs = %s/32", userVPN.IPAddress))

		// 如果用户配置了特定的网段访问，添加到AllowedIPs
		if userVPN.AllowedIPs != "" && userVPN.AllowedIPs != "0.0.0.0/0" {
			config.WriteString(fmt.Sprintf(",%s", userVPN.AllowedIPs))
		}
		config.WriteString("\n")

		// 添加预共享密钥（如果有）
		if userVPN.PresharedKey != "" {
			config.WriteString(fmt.Sprintf("PresharedKey = %s\n", userVPN.PresharedKey))
		}

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
