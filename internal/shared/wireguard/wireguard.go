package wireguard

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"eitec-vpn/internal/server/models"

	"golang.org/x/crypto/curve25519"
)

// GenerateKeyPair 生成WireGuard密钥对
func GenerateKeyPair() (*models.WireGuardKey, error) {
	// 生成32字节随机私钥
	privateKey := make([]byte, 32)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, fmt.Errorf("生成私钥失败: %w", err)
	}

	// 计算对应的公钥
	var publicKey [32]byte
	curve25519.ScalarBaseMult(&publicKey, (*[32]byte)(privateKey))

	// 编码为base64
	privateKeyB64 := base64.StdEncoding.EncodeToString(privateKey)
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKey[:])

	return &models.WireGuardKey{
		PrivateKey: privateKeyB64,
		PublicKey:  publicKeyB64,
	}, nil
}

// ValidateKey 验证WireGuard密钥格式
func ValidateKey(key string) bool {
	if len(key) != 44 {
		return false
	}

	// 检查是否为有效的base64编码
	data, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return false
	}

	// WireGuard密钥应该是32字节
	return len(data) == 32
}

// GenerateModuleConfig 生成模块配置文件内容
func GenerateModuleConfig(module *models.Module, wgInterface *models.WireGuardInterface, serverEndpoint, dns string) string {
	return GenerateModuleConfigWithLocalIP(module, wgInterface, serverEndpoint, dns, "")
}

// GenerateModuleConfigWithLocalIP 生成模块配置文件内容（支持指定local_ip）
func GenerateModuleConfigWithLocalIP(module *models.Module, wgInterface *models.WireGuardInterface, serverEndpoint, dns, moduleLocalIP string) string {
	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/32`, module.PrivateKey, module.IPAddress)

	// 使用DNS配置或默认值
	dnsToUse := dns
	if dnsToUse == "" {
		dnsToUse = GetDefaultDNS()
	}
	if dnsToUse != "" {
		config += fmt.Sprintf("\nDNS = %s", dnsToUse)
	}

	// 确定模块的内网IP（模块在内网的IP地址）
	var finalLocalIP string

	// 按优先级选择local_ip：
	// 1. 函数参数传入的moduleLocalIP（从系统配置获取）
	// 2. module.LocalIP字段
	// 3. 从内网段推导
	// 4. 默认网关IP
	if moduleLocalIP != "" {
		finalLocalIP = moduleLocalIP
	} else if module.LocalIP != "" {
		finalLocalIP = module.LocalIP
	} else {
		// 如果LocalIP为空，从内网段推导
		if module.AllowedIPs != "" && !IsDefaultInternalNetwork(module.AllowedIPs) {
			// 从自定义内网段推导网关IP
			finalLocalIP = DeduceGatewayFromNetwork(module.AllowedIPs)
		} else {
			// 使用默认网关IP
			finalLocalIP = GetDefaultGatewayIP()
		}
	}

	// 如果推导失败，使用默认IP
	if finalLocalIP == "" {
		finalLocalIP = GetDefaultGatewayIP()
	}

	// 获取接口名称用于PostUp脚本
	interfaceName := wgInterface.Name
	if interfaceName == "" {
		interfaceName = GetDefaultInterfaceName()
	}

	// 添加NAT转发规则，实现内网穿透
	config += fmt.Sprintf(`
# NAT转发规则 - 实现内网穿透功能
# 参考用户成功配置，添加SNAT和FORWARD规则
PostUp = iptables -t nat -A POSTROUTING -s %s -j SNAT --to-source %s; iptables -A FORWARD -i %s -j ACCEPT; iptables -A FORWARD -o %s -j ACCEPT
PostDown = iptables -t nat -D POSTROUTING -s %s -j SNAT --to-source %s; iptables -D FORWARD -i %s -j ACCEPT; iptables -D FORWARD -o %s -j ACCEPT`,
		wgInterface.Network, finalLocalIP, interfaceName, interfaceName,
		wgInterface.Network, finalLocalIP, interfaceName, interfaceName)

	// 根据用户成功配置的模式设置AllowedIPs
	// 参考：AllowedIPs = 10.0.8.0/24 (整个VPN网段，实现VPN内部互通)
	allowedIPs := wgInterface.Network // 使用接口的整个网络段

	// 如果模块配置了额外的内网访问权限，添加到AllowedIPs中
	if module.AllowedIPs != "" {
		if !IsDefaultInternalNetwork(module.AllowedIPs) {
			// 自定义内网段，添加到VPN网段后面
			allowedIPs += fmt.Sprintf(", %s", module.AllowedIPs)
		} else {
			// 默认内网段，也添加进去
			allowedIPs += fmt.Sprintf(", %s", GetDefaultInternalNetwork())
		}
	}

	config += fmt.Sprintf(`

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = %s`, wgInterface.PublicKey, serverEndpoint, allowedIPs)

	// 如果有预共享密钥，添加到配置中
	if module.PresharedKey != "" {
		config += fmt.Sprintf("\nPresharedKey = %s", module.PresharedKey)
	}

	config += fmt.Sprintf("\nPersistentKeepalive = %d", module.PersistentKA)

	return config
}

// GeneratePeerConfig 生成运维端Peer配置
func GeneratePeerConfig(module *models.Module) string {
	allowedIPs := module.AllowedIPs
	if allowedIPs == "" {
		allowedIPs = GetDefaultInternalNetwork()
	}

	config := fmt.Sprintf(`[Peer]
# %s - %s
PublicKey = %s
AllowedIPs = %s/32, %s`,
		module.Name,
		module.Location,
		module.PublicKey,
		module.IPAddress,
		allowedIPs)

	return config
}

// GenerateServerConfig 生成服务器端WireGuard配置
func GenerateServerConfig(serverPrivateKey, serverIP string, port int, modules []models.Module, interfaceName string) string {
	// 使用传入的接口名，如果为空则使用默认值
	if interfaceName == "" {
		interfaceName = GetDefaultInterfaceName()
	}

	// 使用配置的外部网络接口名称
	externalInterface := GetDefaultExternalInterface()

	// 计算VPN网络段（从服务器IP推导）
	vpnNetwork := DeduceNetworkFromServerIP(serverIP)

	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/24
ListenPort = %d
SaveConfig = true

# PostUp和PostDown脚本用于防火墙规则和内网穿透
PostUp = iptables -t nat -A POSTROUTING -s %s -o %s -j MASQUERADE; iptables -A INPUT -p udp -m udp --dport %d -j ACCEPT; iptables -A FORWARD -i %s -j ACCEPT; iptables -A FORWARD -o %s -j ACCEPT;
PostDown = iptables -t nat -D POSTROUTING -s %s -o %s -j MASQUERADE; iptables -D INPUT -p udp -m udp --dport %d -j ACCEPT; iptables -D FORWARD -i %s -j ACCEPT; iptables -D FORWARD -o %s -j ACCEPT;

`, serverPrivateKey, serverIP, port, vpnNetwork, externalInterface, port, interfaceName, interfaceName, vpnNetwork, externalInterface, port, interfaceName, interfaceName)

	// 收集所有需要内网穿透的网段
	internalNetworks := make(map[string]bool)
	for _, module := range modules {
		if module.AllowedIPs != "" && !IsDefaultInternalNetwork(module.AllowedIPs) && module.AllowedIPs != vpnNetwork {
			internalNetworks[module.AllowedIPs] = true
		}
	}

	// 为每个内网段添加FORWARD规则
	for network := range internalNetworks {
		config += fmt.Sprintf("PostUp = iptables -I FORWARD -s %s -i %s -d %s -j ACCEPT\n", vpnNetwork, interfaceName, network)
		config += fmt.Sprintf("PostUp = iptables -I FORWARD -s %s -i %s -d %s -j ACCEPT\n", network, interfaceName, vpnNetwork)
		config += fmt.Sprintf("PostDown = iptables -D FORWARD -s %s -i %s -d %s -j ACCEPT\n", vpnNetwork, interfaceName, network)
		config += fmt.Sprintf("PostDown = iptables -D FORWARD -s %s -i %s -d %s -j ACCEPT\n", network, interfaceName, vpnNetwork)
	}

	config += "\n"

	// 添加所有模块的Peer配置
	for _, module := range modules {
		config += fmt.Sprintf(`
[Peer]
# %s - %s
PublicKey = %s
AllowedIPs = %s/32`,
			module.Name,
			module.Location,
			module.PublicKey,
			module.IPAddress)

		// 如果模块配置了内网访问，添加到AllowedIPs
		if module.AllowedIPs != "" && !IsDefaultInternalNetwork(module.AllowedIPs) {
			config += fmt.Sprintf(",%s", module.AllowedIPs)
		}

		// 添加预共享密钥（如果有）
		if module.PresharedKey != "" {
			config += fmt.Sprintf("\nPresharedKey = %s", module.PresharedKey)
		}

		config += "\n"
	}

	return config
}

// WriteConfigFile 写入配置文件
func WriteConfigFile(configPath, content string) error {
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	return nil
}

// RestartWireGuard 重启WireGuard接口
func RestartWireGuard(interfaceName string) error {
	// 停止接口
	if err := exec.Command("wg-quick", "down", interfaceName).Run(); err != nil {
		// 如果接口不存在，忽略错误
		if !strings.Contains(err.Error(), "does not exist") {
			return fmt.Errorf("停止WireGuard接口失败: %w", err)
		}
	}

	// 启动接口
	if err := exec.Command("wg-quick", "up", interfaceName).Run(); err != nil {
		return fmt.Errorf("启动WireGuard接口失败: %w", err)
	}

	return nil
}

// GetWireGuardStatus 获取WireGuard状态信息
func GetWireGuardStatus(interfaceName string) (map[string]WireGuardPeer, error) {
	cmd := exec.Command("wg", "show", interfaceName, "dump")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取WireGuard状态失败: %w", err)
	}

	return parseWireGuardStatus(string(output))
}

// WireGuardPeer Peer状态信息
type WireGuardPeer struct {
	PublicKey       string    `json:"public_key"`
	Endpoint        string    `json:"endpoint"`
	AllowedIPs      []string  `json:"allowed_ips"`
	LatestHandshake time.Time `json:"latest_handshake"`
	TransferRxBytes uint64    `json:"transfer_rx_bytes"`
	TransferTxBytes uint64    `json:"transfer_tx_bytes"`
	PersistentKA    int       `json:"persistent_keepalive"`
}

// parseWireGuardStatus 解析wg show dump输出
func parseWireGuardStatus(output string) (map[string]WireGuardPeer, error) {
	peers := make(map[string]WireGuardPeer)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 6 {
			continue
		}

		// 跳过接口行
		if fields[0] != "" && len(fields) == 6 {
			continue
		}

		publicKey := fields[0]
		if publicKey == "" {
			continue
		}

		peer := WireGuardPeer{
			PublicKey: publicKey,
			Endpoint:  fields[2],
		}

		// 解析AllowedIPs
		if fields[3] != "" {
			peer.AllowedIPs = strings.Split(fields[3], ",")
		}

		// 解析最新握手时间
		if fields[4] != "0" {
			if timestamp, err := parseTimestamp(fields[4]); err == nil {
				peer.LatestHandshake = timestamp
			}
		}

		// 解析传输字节数
		if rxBytes, err := parseBytes(fields[5]); err == nil {
			peer.TransferRxBytes = rxBytes
		}
		if len(fields) > 6 {
			if txBytes, err := parseBytes(fields[6]); err == nil {
				peer.TransferTxBytes = txBytes
			}
		}

		// 解析PersistentKeepalive
		if len(fields) > 7 && fields[7] != "0" {
			if ka, err := parseBytes(fields[7]); err == nil {
				peer.PersistentKA = int(ka)
			}
		}

		peers[publicKey] = peer
	}

	return peers, nil
}

// parseTimestamp 解析时间戳
func parseTimestamp(s string) (time.Time, error) {
	timestamp, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		// 尝试Unix时间戳
		if unix, err2 := time.Parse("1136239445", s); err2 == nil {
			return unix, nil
		}
		return time.Time{}, err
	}
	return timestamp, nil
}

// parseBytes 解析字节数
func parseBytes(s string) (uint64, error) {
	var bytes uint64
	_, err := fmt.Sscanf(s, "%d", &bytes)
	return bytes, err
}

// IsWireGuardInstalled 检查WireGuard是否已安装
func IsWireGuardInstalled() bool {
	_, err := exec.LookPath("wg")
	return err == nil
}

// IsInterfaceUp 检查接口是否已启动
func IsInterfaceUp(interfaceName string) bool {
	cmd := exec.Command("wg", "show", interfaceName)
	return cmd.Run() == nil
}

// ValidateConfig 验证WireGuard配置文件
func ValidateConfig(configContent string) error {
	// 检查必需的部分
	if !strings.Contains(configContent, "[Interface]") {
		return fmt.Errorf("配置缺少[Interface]部分")
	}

	// 检查私钥
	privateKeyRegex := regexp.MustCompile(`PrivateKey\s*=\s*([A-Za-z0-9+/]{43}=)`)
	if !privateKeyRegex.MatchString(configContent) {
		return fmt.Errorf("配置缺少有效的PrivateKey")
	}

	// 检查地址
	addressRegex := regexp.MustCompile(`Address\s*=\s*(\d+\.\d+\.\d+\.\d+/\d+)`)
	if !addressRegex.MatchString(configContent) {
		return fmt.Errorf("配置缺少有效的Address")
	}

	return nil
}

// ApplyConfig 应用WireGuard配置
func ApplyConfig(interfaceName, configPath string) error {
	// 验证配置文件
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := ValidateConfig(string(content)); err != nil {
		return fmt.Errorf("配置文件验证失败: %w", err)
	}

	// 重启WireGuard接口
	return RestartWireGuard(interfaceName)
}

// GeneratePresharedKey 生成WireGuard预共享密钥
func GeneratePresharedKey() (string, error) {
	// 生成32字节随机预共享密钥
	presharedKey := make([]byte, 32)
	if _, err := rand.Read(presharedKey); err != nil {
		return "", fmt.Errorf("生成预共享密钥失败: %w", err)
	}

	// 编码为base64
	presharedKeyB64 := base64.StdEncoding.EncodeToString(presharedKey)
	return presharedKeyB64, nil
}
