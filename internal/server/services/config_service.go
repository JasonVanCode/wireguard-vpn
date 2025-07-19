package services

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"eitec-vpn/internal/server/database"
	"eitec-vpn/internal/server/models"
	"eitec-vpn/internal/shared/utils"
	"eitec-vpn/internal/shared/wireguard"

	"gorm.io/gorm"
)

// ConfigService 配置管理服务
type ConfigService struct {
	db *gorm.DB
}

// NewConfigService 创建配置管理服务
func NewConfigService() *ConfigService {
	return &ConfigService{
		db: database.DB,
	}
}

// SystemConfigInfo 系统配置信息
type SystemConfigInfo struct {
	Server    ServerConfigInfo    `json:"server"`
	WireGuard WireGuardConfigInfo `json:"wireguard"`
	Security  SecurityConfigInfo  `json:"security"`
	Network   NetworkConfigInfo   `json:"network"`
}

// ServerConfigInfo 服务器配置信息
type ServerConfigInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
	WebPort     int    `json:"web_port"`
	EnableHTTPS bool   `json:"enable_https"`
	DomainName  string `json:"domain_name"`
}

// WireGuardConfigInfo WireGuard配置信息
type WireGuardConfigInfo struct {
	Interface  string `json:"interface"`
	ListenPort int    `json:"listen_port"`
	Network    string `json:"network"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key,omitempty"` // 仅在管理员查看时显示
	DNS        string `json:"dns"`
	MTU        int    `json:"mtu"`
	PostUp     string `json:"post_up"`
	PostDown   string `json:"post_down"`
	SaveConfig bool   `json:"save_config"`
}

// SecurityConfigInfo 安全配置信息
type SecurityConfigInfo struct {
	JWTSecret         string `json:"jwt_secret,omitempty"`
	SessionTimeout    int    `json:"session_timeout"`
	MaxLoginAttempts  int    `json:"max_login_attempts"`
	LockoutDuration   int    `json:"lockout_duration"`
	EnableAPIAccess   bool   `json:"enable_api_access"`
	AllowedIPRanges   string `json:"allowed_ip_ranges"`
	EnableRateLimit   bool   `json:"enable_rate_limit"`
	RateLimitRequests int    `json:"rate_limit_requests"`
	RateLimitWindow   int    `json:"rate_limit_window"`
}

// NetworkConfigInfo 网络配置信息
type NetworkConfigInfo struct {
	VPNNetwork     string `json:"vpn_network"`
	IPPoolStart    string `json:"ip_pool_start"`
	IPPoolEnd      string `json:"ip_pool_end"`
	DefaultGateway string `json:"default_gateway"`
	DNSServers     string `json:"dns_servers"`
	EnableNAT      bool   `json:"enable_nat"`
	EnableIPv6     bool   `json:"enable_ipv6"`
}

// GetSystemConfig 获取系统配置
func (cs *ConfigService) GetSystemConfig(includeSecrets bool) (*SystemConfigInfo, error) {
	config := &SystemConfigInfo{}

	// 获取服务器配置
	serverConfig, err := cs.getServerConfig()
	if err != nil {
		return nil, fmt.Errorf("获取服务器配置失败: %w", err)
	}
	config.Server = *serverConfig

	// 获取WireGuard配置
	wgConfig, err := cs.getWireGuardConfig(includeSecrets)
	if err != nil {
		return nil, fmt.Errorf("获取WireGuard配置失败: %w", err)
	}
	config.WireGuard = *wgConfig

	// 获取安全配置
	securityConfig, err := cs.getSecurityConfig(includeSecrets)
	if err != nil {
		return nil, fmt.Errorf("获取安全配置失败: %w", err)
	}
	config.Security = *securityConfig

	// 获取网络配置
	networkConfig, err := cs.getNetworkConfig()
	if err != nil {
		return nil, fmt.Errorf("获取网络配置失败: %w", err)
	}
	config.Network = *networkConfig

	return config, nil
}

// getServerConfig 获取服务器配置
func (cs *ConfigService) getServerConfig() (*ServerConfigInfo, error) {
	name, _ := database.GetSystemConfig("server.name")
	if name == "" {
		name = "EITEC VPN Server"
	}

	description, _ := database.GetSystemConfig("server.description")
	if description == "" {
		description = "WireGuard VPN集中管理平台"
	}

	endpoint, _ := database.GetSystemConfig("server.endpoint")
	webPortStr, _ := database.GetSystemConfig("server.web_port")
	webPort, _ := strconv.Atoi(webPortStr)
	if webPort == 0 {
		webPort = 8080
	}

	enableHTTPSStr, _ := database.GetSystemConfig("server.enable_https")
	enableHTTPS, _ := strconv.ParseBool(enableHTTPSStr)

	domainName, _ := database.GetSystemConfig("server.domain_name")

	return &ServerConfigInfo{
		Name:        name,
		Description: description,
		Endpoint:    endpoint,
		WebPort:     webPort,
		EnableHTTPS: enableHTTPS,
		DomainName:  domainName,
	}, nil
}

// getWireGuardConfig 获取WireGuard配置
func (cs *ConfigService) getWireGuardConfig(includeSecrets bool) (*WireGuardConfigInfo, error) {
	interfaceName, _ := database.GetSystemConfig("wg.interface")
	if interfaceName == "" {
		interfaceName = "wg0"
	}

	listenPortStr, _ := database.GetSystemConfig("wg.listen_port")
	listenPort, _ := strconv.Atoi(listenPortStr)
	if listenPort == 0 {
		listenPort = 51820
	}

	network, _ := database.GetSystemConfig("wg.network")
	if network == "" {
		network = "10.10.0.0/24"
	}

	publicKey, _ := database.GetSystemConfig("server.public_key")
	privateKey := ""
	if includeSecrets {
		privateKey, _ = database.GetSystemConfig("server.private_key")
	}

	dns, _ := database.GetSystemConfig("wg.dns")
	if dns == "" {
		dns = "8.8.8.8,8.8.4.4"
	}

	mtuStr, _ := database.GetSystemConfig("wg.mtu")
	mtu, _ := strconv.Atoi(mtuStr)
	if mtu == 0 {
		mtu = 1420
	}

	postUp, _ := database.GetSystemConfig("wg.post_up")
	postDown, _ := database.GetSystemConfig("wg.post_down")

	saveConfigStr, _ := database.GetSystemConfig("wg.save_config")
	saveConfig, _ := strconv.ParseBool(saveConfigStr)

	return &WireGuardConfigInfo{
		Interface:  interfaceName,
		ListenPort: listenPort,
		Network:    network,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		DNS:        dns,
		MTU:        mtu,
		PostUp:     postUp,
		PostDown:   postDown,
		SaveConfig: saveConfig,
	}, nil
}

// getSecurityConfig 获取安全配置
func (cs *ConfigService) getSecurityConfig(includeSecrets bool) (*SecurityConfigInfo, error) {
	jwtSecret := ""
	if includeSecrets {
		jwtSecret, _ = database.GetSystemConfig("jwt.secret")
	}

	sessionTimeoutStr, _ := database.GetSystemConfig("security.session_timeout")
	sessionTimeout, _ := strconv.Atoi(sessionTimeoutStr)
	if sessionTimeout == 0 {
		sessionTimeout = 3600 // 1小时
	}

	maxLoginAttemptsStr, _ := database.GetSystemConfig("security.max_login_attempts")
	maxLoginAttempts, _ := strconv.Atoi(maxLoginAttemptsStr)
	if maxLoginAttempts == 0 {
		maxLoginAttempts = 5
	}

	lockoutDurationStr, _ := database.GetSystemConfig("security.lockout_duration")
	lockoutDuration, _ := strconv.Atoi(lockoutDurationStr)
	if lockoutDuration == 0 {
		lockoutDuration = 900 // 15分钟
	}

	enableAPIAccessStr, _ := database.GetSystemConfig("security.enable_api_access")
	enableAPIAccess, _ := strconv.ParseBool(enableAPIAccessStr)

	allowedIPRanges, _ := database.GetSystemConfig("security.allowed_ip_ranges")

	enableRateLimitStr, _ := database.GetSystemConfig("security.enable_rate_limit")
	enableRateLimit, _ := strconv.ParseBool(enableRateLimitStr)

	rateLimitRequestsStr, _ := database.GetSystemConfig("security.rate_limit_requests")
	rateLimitRequests, _ := strconv.Atoi(rateLimitRequestsStr)
	if rateLimitRequests == 0 {
		rateLimitRequests = 100
	}

	rateLimitWindowStr, _ := database.GetSystemConfig("security.rate_limit_window")
	rateLimitWindow, _ := strconv.Atoi(rateLimitWindowStr)
	if rateLimitWindow == 0 {
		rateLimitWindow = 3600 // 1小时
	}

	return &SecurityConfigInfo{
		JWTSecret:         jwtSecret,
		SessionTimeout:    sessionTimeout,
		MaxLoginAttempts:  maxLoginAttempts,
		LockoutDuration:   lockoutDuration,
		EnableAPIAccess:   enableAPIAccess,
		AllowedIPRanges:   allowedIPRanges,
		EnableRateLimit:   enableRateLimit,
		RateLimitRequests: rateLimitRequests,
		RateLimitWindow:   rateLimitWindow,
	}, nil
}

// getNetworkConfig 获取网络配置
func (cs *ConfigService) getNetworkConfig() (*NetworkConfigInfo, error) {
	vpnNetwork, _ := database.GetSystemConfig("wg.network")
	if vpnNetwork == "" {
		vpnNetwork = "10.10.0.0/24"
	}

	ipPoolStart, _ := database.GetSystemConfig("network.ip_pool_start")
	if ipPoolStart == "" {
		ipPoolStart = "10.10.0.2"
	}

	ipPoolEnd, _ := database.GetSystemConfig("network.ip_pool_end")
	if ipPoolEnd == "" {
		ipPoolEnd = "10.10.0.254"
	}

	defaultGateway, _ := database.GetSystemConfig("network.default_gateway")
	if defaultGateway == "" {
		defaultGateway = "10.10.0.1"
	}

	dnsServers, _ := database.GetSystemConfig("wg.dns")
	if dnsServers == "" {
		dnsServers = "8.8.8.8,8.8.4.4"
	}

	enableNATStr, _ := database.GetSystemConfig("network.enable_nat")
	enableNAT, _ := strconv.ParseBool(enableNATStr)

	enableIPv6Str, _ := database.GetSystemConfig("network.enable_ipv6")
	enableIPv6, _ := strconv.ParseBool(enableIPv6Str)

	return &NetworkConfigInfo{
		VPNNetwork:     vpnNetwork,
		IPPoolStart:    ipPoolStart,
		IPPoolEnd:      ipPoolEnd,
		DefaultGateway: defaultGateway,
		DNSServers:     dnsServers,
		EnableNAT:      enableNAT,
		EnableIPv6:     enableIPv6,
	}, nil
}

// UpdateSystemConfig 更新系统配置
func (cs *ConfigService) UpdateSystemConfig(updates map[string]interface{}) error {
	// 验证配置项
	if err := cs.validateConfigUpdates(updates); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 更新配置
	for key, value := range updates {
		valueStr := fmt.Sprintf("%v", value)
		if err := database.SetSystemConfig(key, valueStr); err != nil {
			return fmt.Errorf("更新配置 %s 失败: %w", key, err)
		}
	}

	return nil
}

// validateConfigUpdates 验证配置更新
func (cs *ConfigService) validateConfigUpdates(updates map[string]interface{}) error {
	for key, value := range updates {
		switch key {
		case "server.endpoint":
			endpoint := fmt.Sprintf("%v", value)
			if endpoint == "" {
				return errors.New("服务器端点不能为空")
			}
			// 验证端点格式 (IP:Port 或 Domain:Port)
			if !utils.IsValidEndpoint(endpoint) {
				return errors.New("服务器端点格式无效")
			}

		case "server.web_port":
			port, ok := value.(int)
			if !ok {
				if portStr, ok := value.(string); ok {
					var err error
					port, err = strconv.Atoi(portStr)
					if err != nil {
						return errors.New("Web端口必须是数字")
					}
				} else {
					return errors.New("Web端口必须是数字")
				}
			}
			if !utils.IsValidPort(port) {
				return errors.New("Web端口无效")
			}

		case "wg.listen_port":
			port, ok := value.(int)
			if !ok {
				if portStr, ok := value.(string); ok {
					var err error
					port, err = strconv.Atoi(portStr)
					if err != nil {
						return errors.New("WireGuard监听端口必须是数字")
					}
				} else {
					return errors.New("WireGuard监听端口必须是数字")
				}
			}
			if !utils.IsValidPort(port) {
				return errors.New("WireGuard监听端口无效")
			}

		case "wg.network":
			network := fmt.Sprintf("%v", value)
			if !utils.IsValidCIDR(network) {
				return errors.New("VPN网络CIDR格式无效")
			}

		case "wg.dns":
			dns := fmt.Sprintf("%v", value)
			if dns != "" && !utils.IsValidDNSList(dns) {
				return errors.New("DNS服务器列表格式无效")
			}

		case "network.ip_pool_start", "network.ip_pool_end", "network.default_gateway":
			ip := fmt.Sprintf("%v", value)
			if !utils.IsValidIP(ip) {
				return fmt.Errorf("IP地址 %s 无效", ip)
			}

		case "security.session_timeout", "security.max_login_attempts", "security.lockout_duration":
			timeout, ok := value.(int)
			if !ok {
				if timeoutStr, ok := value.(string); ok {
					var err error
					timeout, err = strconv.Atoi(timeoutStr)
					if err != nil {
						return fmt.Errorf("配置项 %s 必须是数字", key)
					}
				} else {
					return fmt.Errorf("配置项 %s 必须是数字", key)
				}
			}
			if timeout <= 0 {
				return fmt.Errorf("配置项 %s 必须大于0", key)
			}

		case "wg.mtu":
			mtu, ok := value.(int)
			if !ok {
				if mtuStr, ok := value.(string); ok {
					var err error
					mtu, err = strconv.Atoi(mtuStr)
					if err != nil {
						return errors.New("MTU必须是数字")
					}
				} else {
					return errors.New("MTU必须是数字")
				}
			}
			if mtu < 1280 || mtu > 1500 {
				return errors.New("MTU必须在1280-1500之间")
			}
		}
	}

	return nil
}

// InitializeWireGuardServer 初始化WireGuard服务器
func (cs *ConfigService) InitializeWireGuardServer() error {
	// 检查是否已经初始化
	publicKey, err := database.GetSystemConfig("server.public_key")
	if err == nil && publicKey != "" {
		return errors.New("WireGuard服务器已经初始化")
	}

	// 生成服务器密钥对
	keyPair, err := wireguard.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("生成服务器密钥对失败: %w", err)
	}

	// 保存密钥到系统配置
	if err := database.SetSystemConfig("server.public_key", keyPair.PublicKey); err != nil {
		return fmt.Errorf("保存服务器公钥失败: %w", err)
	}

	if err := database.SetSystemConfig("server.private_key", keyPair.PrivateKey); err != nil {
		return fmt.Errorf("保存服务器私钥失败: %w", err)
	}

	// 设置默认配置
	defaults := map[string]string{
		"wg.interface":            "wg0",
		"wg.listen_port":          "51820",
		"wg.network":              "10.10.0.0/24",
		"wg.dns":                  "8.8.8.8,8.8.4.4",
		"wg.mtu":                  "1420",
		"wg.save_config":          "true",
		"network.ip_pool_start":   "10.10.0.2",
		"network.ip_pool_end":     "10.10.0.254",
		"network.default_gateway": "10.10.0.1",
		"network.enable_nat":      "true",
	}

	for key, value := range defaults {
		if _, err := database.GetSystemConfig(key); err != nil {
			database.SetSystemConfig(key, value)
		}
	}

	return nil
}

// GenerateServerConfig 生成WireGuard服务器配置文件
func (cs *ConfigService) GenerateServerConfig() (string, error) {
	// 获取WireGuard配置
	wgConfig, err := cs.getWireGuardConfig(true)
	if err != nil {
		return "", fmt.Errorf("获取WireGuard配置失败: %w", err)
	}

	if wgConfig.PrivateKey == "" {
		return "", errors.New("服务器私钥未配置")
	}

	// 获取所有模块
	var modules []models.Module
	if err := cs.db.Find(&modules).Error; err != nil {
		return "", fmt.Errorf("查询模块列表失败: %w", err)
	}

	// 获取接口名称
	interfaceName := wgConfig.Interface
	if interfaceName == "" {
		interfaceName = "wg0"
	}

	// 生成服务器配置
	config := wireguard.GenerateServerConfig(wgConfig.PrivateKey, "10.10.0.1", wgConfig.ListenPort, modules, interfaceName)

	return config, nil
}

// ApplyWireGuardConfig 应用WireGuard配置
func (cs *ConfigService) ApplyWireGuardConfig() error {
	// 生成配置文件
	config, err := cs.GenerateServerConfig()
	if err != nil {
		return fmt.Errorf("生成配置文件失败: %w", err)
	}

	// 获取接口名称
	interfaceName, _ := database.GetSystemConfig("wg.interface")
	if interfaceName == "" {
		interfaceName = "wg0"
	}

	// 写入配置文件
	configPath := fmt.Sprintf("/etc/wireguard/%s.conf", interfaceName)
	if err := wireguard.WriteConfigFile(configPath, config); err != nil {
		return fmt.Errorf("写入WireGuard配置文件失败: %w", err)
	}

	// 应用配置
	if err := wireguard.ApplyConfig(interfaceName, configPath); err != nil {
		return fmt.Errorf("应用WireGuard配置失败: %w", err)
	}

	return nil
}

// ValidateNetworkSettings 验证网络设置
func (cs *ConfigService) ValidateNetworkSettings(network, ipStart, ipEnd string) error {
	// 验证网络CIDR
	_, ipNet, err := net.ParseCIDR(network)
	if err != nil {
		return fmt.Errorf("网络CIDR格式无效: %w", err)
	}

	// 验证起始IP
	startIP := net.ParseIP(ipStart)
	if startIP == nil {
		return errors.New("起始IP地址无效")
	}

	// 验证结束IP
	endIP := net.ParseIP(ipEnd)
	if endIP == nil {
		return errors.New("结束IP地址无效")
	}

	// 检查IP是否在网络范围内
	if !ipNet.Contains(startIP) {
		return errors.New("起始IP地址不在VPN网络范围内")
	}

	if !ipNet.Contains(endIP) {
		return errors.New("结束IP地址不在VPN网络范围内")
	}

	// 检查起始IP是否小于结束IP
	if utils.CompareIPs(startIP.String(), endIP.String()) >= 0 {
		return errors.New("起始IP地址必须小于结束IP地址")
	}

	return nil
}

// ResetToDefaults 重置为默认配置
func (cs *ConfigService) ResetToDefaults() error {
	defaults := map[string]string{
		"server.name":                  "EITEC VPN Server",
		"server.description":           "WireGuard VPN集中管理平台",
		"server.web_port":              "8080",
		"server.enable_https":          "false",
		"wg.interface":                 "wg0",
		"wg.listen_port":               "51820",
		"wg.network":                   "10.10.0.0/24",
		"wg.dns":                       "8.8.8.8,8.8.4.4",
		"wg.mtu":                       "1420",
		"wg.save_config":               "true",
		"network.ip_pool_start":        "10.10.0.2",
		"network.ip_pool_end":          "10.10.0.254",
		"network.default_gateway":      "10.10.0.1",
		"network.enable_nat":           "true",
		"network.enable_ipv6":          "false",
		"security.session_timeout":     "3600",
		"security.max_login_attempts":  "5",
		"security.lockout_duration":    "900",
		"security.enable_api_access":   "true",
		"security.enable_rate_limit":   "true",
		"security.rate_limit_requests": "100",
		"security.rate_limit_window":   "3600",
	}

	for key, value := range defaults {
		if err := database.SetSystemConfig(key, value); err != nil {
			return fmt.Errorf("重置配置 %s 失败: %w", key, err)
		}
	}

	return nil
}

// ExportConfig 导出配置
func (cs *ConfigService) ExportConfig() (map[string]string, error) {
	var configs []models.SystemConfig
	if err := cs.db.Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("查询系统配置失败: %w", err)
	}

	configMap := make(map[string]string)
	for _, config := range configs {
		// 排除敏感信息
		if config.Key != "server.private_key" && config.Key != "jwt.secret" {
			configMap[config.Key] = config.Value
		}
	}

	return configMap, nil
}

// ImportConfig 导入配置
func (cs *ConfigService) ImportConfig(configMap map[string]string) error {
	// 验证所有配置项
	if err := cs.validateConfigUpdates(convertToInterface(configMap)); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 导入配置
	for key, value := range configMap {
		// 跳过敏感信息
		if key == "server.private_key" || key == "jwt.secret" {
			continue
		}

		if err := database.SetSystemConfig(key, value); err != nil {
			return fmt.Errorf("导入配置 %s 失败: %w", key, err)
		}
	}

	return nil
}

// convertToInterface 将string map转换为interface{} map
func convertToInterface(stringMap map[string]string) map[string]interface{} {
	interfaceMap := make(map[string]interface{})
	for k, v := range stringMap {
		interfaceMap[k] = v
	}
	return interfaceMap
}
