package wireguard

import (
	"fmt"
	"strings"
)

// WireGuardConfig WireGuard配置常量
type WireGuardConfig struct {
	// 默认网络配置
	DefaultInternalNetwork   string // 默认内网段
	DefaultGatewayIP         string // 默认网关IP
	DefaultInterfaceName     string // 默认接口名称
	DefaultExternalInterface string // 默认外部网络接口

	// DNS配置
	DefaultDNS string // 默认DNS服务器

	// 网络配置模板
	InternalNetworkTemplates map[string]string // 预定义的内网段模板
}

// DefaultWireGuardConfig 默认配置
var DefaultWireGuardConfig = &WireGuardConfig{
	// 默认网络配置
	DefaultInternalNetwork:   "192.168.1.0/24",
	DefaultGatewayIP:         "192.168.1.1",
	DefaultInterfaceName:     "wg0",
	DefaultExternalInterface: "eth0",

	// DNS配置
	DefaultDNS: "8.8.8.8,8.8.4.4",

	// 预定义网络段模板
	InternalNetworkTemplates: map[string]string{
		"home":       "192.168.1.0/24",   // 家庭网络
		"office":     "192.168.2.0/24",   // 办公网络
		"datacenter": "10.0.0.0/24",      // 数据中心
		"cloud":      "172.16.0.0/24",    // 云网络
		"iot":        "192.168.100.0/24", // 物联网设备
	},
}

// ConfigLoader 配置加载器接口
type ConfigLoader interface {
	GetSystemConfig(key string) (string, error)
}

// LoadConfigFromSystem 从系统配置加载WireGuard配置
func LoadConfigFromSystem(loader ConfigLoader) {
	// 尝试从系统配置加载各项配置
	if internalNetwork, err := loader.GetSystemConfig("wireguard.default_internal_network"); err == nil && internalNetwork != "" {
		DefaultWireGuardConfig.DefaultInternalNetwork = internalNetwork
	}

	if gatewayIP, err := loader.GetSystemConfig("wireguard.default_gateway_ip"); err == nil && gatewayIP != "" {
		DefaultWireGuardConfig.DefaultGatewayIP = gatewayIP
	}

	if interfaceName, err := loader.GetSystemConfig("wireguard.default_interface_name"); err == nil && interfaceName != "" {
		DefaultWireGuardConfig.DefaultInterfaceName = interfaceName
	}

	if externalInterface, err := loader.GetSystemConfig("wireguard.default_external_interface"); err == nil && externalInterface != "" {
		DefaultWireGuardConfig.DefaultExternalInterface = externalInterface
	}

	if dns, err := loader.GetSystemConfig("wireguard.default_dns"); err == nil && dns != "" {
		DefaultWireGuardConfig.DefaultDNS = dns
	}
}

// GetDefaultInternalNetwork 获取默认内网段
func GetDefaultInternalNetwork() string {
	return DefaultWireGuardConfig.DefaultInternalNetwork
}

// GetDefaultGatewayIP 获取默认网关IP
func GetDefaultGatewayIP() string {
	return DefaultWireGuardConfig.DefaultGatewayIP
}

// GetDefaultInterfaceName 获取默认接口名
func GetDefaultInterfaceName() string {
	return DefaultWireGuardConfig.DefaultInterfaceName
}

// GetDefaultExternalInterface 获取默认外部网络接口
func GetDefaultExternalInterface() string {
	return DefaultWireGuardConfig.DefaultExternalInterface
}

// GetDefaultDNS 获取默认DNS
func GetDefaultDNS() string {
	return DefaultWireGuardConfig.DefaultDNS
}

// DeduceGatewayFromNetwork 从网络段推导网关IP
// 例如：192.168.2.0/24 -> 192.168.2.1
func DeduceGatewayFromNetwork(network string) string {
	if network == "" {
		return GetDefaultGatewayIP()
	}

	// 检查是否为/24网段
	if !strings.Contains(network, "/24") {
		return GetDefaultGatewayIP()
	}

	// 分割网络段
	networkBase := strings.Split(network, "/")[0]
	parts := strings.Split(networkBase, ".")

	if len(parts) != 4 {
		return GetDefaultGatewayIP()
	}

	// 通常网关是网段的第一个IP (.1)
	return fmt.Sprintf("%s.%s.%s.1", parts[0], parts[1], parts[2])
}

// IsDefaultInternalNetwork 检查是否为默认内网段
func IsDefaultInternalNetwork(network string) bool {
	return network == GetDefaultInternalNetwork()
}

// GetNetworkTemplate 获取网络模板
func GetNetworkTemplate(templateName string) string {
	if network, exists := DefaultWireGuardConfig.InternalNetworkTemplates[templateName]; exists {
		return network
	}
	return GetDefaultInternalNetwork()
}

// ValidateNetworkSegment 验证网络段格式
func ValidateNetworkSegment(network string) bool {
	// 简单的网络段格式验证
	// 支持格式：192.168.1.0/24, 10.0.0.0/8, 172.16.0.0/16 等
	parts := strings.Split(network, "/")
	if len(parts) != 2 {
		return false
	}

	// 验证IP部分
	ipParts := strings.Split(parts[0], ".")
	if len(ipParts) != 4 {
		return false
	}

	// 这里可以添加更详细的IP和CIDR验证
	return true
}

// UpdateConfig 更新配置值
func UpdateConfig(key, value string) {
	switch key {
	case "default_internal_network":
		DefaultWireGuardConfig.DefaultInternalNetwork = value
	case "default_gateway_ip":
		DefaultWireGuardConfig.DefaultGatewayIP = value
	case "default_interface_name":
		DefaultWireGuardConfig.DefaultInterfaceName = value
	case "default_external_interface":
		DefaultWireGuardConfig.DefaultExternalInterface = value
	case "default_dns":
		DefaultWireGuardConfig.DefaultDNS = value
	}
}

// GetAllConfigs 获取所有配置值（用于管理界面）
func GetAllConfigs() map[string]string {
	return map[string]string{
		"default_internal_network":   GetDefaultInternalNetwork(),
		"default_gateway_ip":         GetDefaultGatewayIP(),
		"default_interface_name":     GetDefaultInterfaceName(),
		"default_external_interface": GetDefaultExternalInterface(),
		"default_dns":                GetDefaultDNS(),
	}
}

// DeduceNetworkFromServerIP 从服务器IP推导网络段
// 例如：10.0.8.1 -> 10.0.8.0/24
func DeduceNetworkFromServerIP(serverIP string) string {
	if serverIP == "" {
		return "10.0.0.0/24" // 默认网络段
	}

	// 分割IP地址
	parts := strings.Split(serverIP, ".")
	if len(parts) != 4 {
		return "10.0.0.0/24" // 默认网络段
	}

	// 构建网络段（将最后一部分设为0）
	return fmt.Sprintf("%s.%s.%s.0/24", parts[0], parts[1], parts[2])
}
