package services

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"eitec-vpn/internal/shared/config"
)

// WireGuardShowService WireGuard状态检测服务
type WireGuardShowService struct{}

// NewWireGuardShowService 创建WireGuard状态检测服务
func NewWireGuardShowService() *WireGuardShowService {
	return &WireGuardShowService{}
}

// =====================================================
// 实时状态数据结构
// =====================================================

// WireGuardSystemInfo WireGuard系统信息
type WireGuardSystemInfo struct {
	IsInstalled    bool                          `json:"is_installed"`    // 是否安装
	Version        string                        `json:"version"`         // 版本信息
	ServiceRunning bool                          `json:"service_running"` // 服务是否运行
	Interfaces     map[string]*InterfaceShowInfo `json:"interfaces"`      // 接口状态
}

// InterfaceShowInfo 接口show信息
type InterfaceShowInfo struct {
	Name          string                   `json:"name"`           // 接口名称
	PublicKey     string                   `json:"public_key"`     // 公钥
	ListenPort    int                      `json:"listen_port"`    // 监听端口
	IsActive      bool                     `json:"is_active"`      // 是否激活
	PeerCount     int                      `json:"peer_count"`     // peer总数
	ActivePeers   int                      `json:"active_peers"`   // 活跃peer数
	TotalTraffic  TrafficData              `json:"total_traffic"`  // 总流量
	LastHandshake *time.Time               `json:"last_handshake"` // 最后握手时间
	Peers         map[string]*PeerShowInfo `json:"peers"`          // peer列表
}

// PeerShowInfo peer的show信息
type PeerShowInfo struct {
	PublicKey       string      `json:"public_key"`       // 公钥
	Endpoint        string      `json:"endpoint"`         // 端点
	AllowedIPs      []string    `json:"allowed_ips"`      // 允许的IP
	LatestHandshake *time.Time  `json:"latest_handshake"` // 最后握手时间
	TrafficStats    TrafficData `json:"traffic_stats"`    // 流量统计
	PersistentKA    int         `json:"persistent_ka"`    // 保活间隔
	IsOnline        bool        `json:"is_online"`        // 是否在线
	LastSeenAgo     string      `json:"last_seen_ago"`    // 最后在线时间描述
}

// TrafficData 流量数据
type TrafficData struct {
	RxBytes uint64 `json:"rx_bytes"` // 接收字节数
	TxBytes uint64 `json:"tx_bytes"` // 发送字节数
	RxMB    string `json:"rx_mb"`    // 接收MB（格式化）
	TxMB    string `json:"tx_mb"`    // 发送MB（格式化）
	Total   string `json:"total"`    // 总流量（格式化）
}

// =====================================================
// 核心检测方法
// =====================================================

// CheckWireGuardInstalled 检查WireGuard是否安装
func (wss *WireGuardShowService) CheckWireGuardInstalled() bool {
	cmd := exec.Command("wg", "version")
	return cmd.Run() == nil
}

// GetWireGuardVersion 获取WireGuard版本
func (wss *WireGuardShowService) GetWireGuardVersion() string {
	if !wss.CheckWireGuardInstalled() {
		return "not_installed"
	}

	cmd := exec.Command("wg", "version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// 提取版本信息
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}

	return "unknown"
}

// CheckServiceRunning 检查WireGuard服务是否运行
func (wss *WireGuardShowService) CheckServiceRunning() bool {
	if !wss.CheckWireGuardInstalled() {
		return false
	}

	// 尝试执行wg show，如果有任何接口运行，服务就是运行的
	cmd := exec.Command("wg", "show")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// 如果有输出，说明至少有一个接口在运行
	return strings.TrimSpace(string(output)) != ""
}

// GetSystemInfo 获取WireGuard系统完整信息
func (wss *WireGuardShowService) GetSystemInfo() *WireGuardSystemInfo {
	info := &WireGuardSystemInfo{
		IsInstalled:    wss.CheckWireGuardInstalled(),
		ServiceRunning: false,
		Interfaces:     make(map[string]*InterfaceShowInfo),
	}

	if info.IsInstalled {
		info.Version = wss.GetWireGuardVersion()
		info.ServiceRunning = wss.CheckServiceRunning()

		if info.ServiceRunning {
			interfaces, err := wss.GetAllInterfacesInfo()
			if err == nil {
				info.Interfaces = interfaces
			}
		}
	}

	return info
}

// =====================================================
// 接口状态检测
// =====================================================

// GetInterfaceInfo 获取单个接口的实时状态
func (wss *WireGuardShowService) GetInterfaceInfo(interfaceName string) (*InterfaceShowInfo, error) {
	if !wss.CheckWireGuardInstalled() {
		return &InterfaceShowInfo{
			Name:     interfaceName,
			IsActive: false,
		}, nil
	}

	cmd := exec.Command("wg", "show", interfaceName, "dump")
	output, err := cmd.Output()
	if err != nil {
		// 接口不存在或未激活
		return &InterfaceShowInfo{
			Name:     interfaceName,
			IsActive: false,
		}, nil
	}

	return wss.parseInterfaceDump(interfaceName, string(output))
}

// GetAllInterfacesInfo 获取所有接口的实时状态
func (wss *WireGuardShowService) GetAllInterfacesInfo() (map[string]*InterfaceShowInfo, error) {
	if !wss.CheckWireGuardInstalled() {
		return make(map[string]*InterfaceShowInfo), nil
	}

	cmd := exec.Command("wg", "show", "all", "dump")
	output, err := cmd.Output()
	if err != nil {
		return make(map[string]*InterfaceShowInfo), nil
	}

	return wss.parseAllInterfacesDump(string(output))
}

// GetPeerInfo 获取特定peer的实时状态
func (wss *WireGuardShowService) GetPeerInfo(interfaceName, publicKey string) (*PeerShowInfo, error) {
	interfaceInfo, err := wss.GetInterfaceInfo(interfaceName)
	if err != nil {
		return nil, err
	}

	if peer, exists := interfaceInfo.Peers[publicKey]; exists {
		return peer, nil
	}

	return nil, fmt.Errorf("peer not found")
}

// =====================================================
// WireGuard dump格式解析
// =====================================================

// parseInterfaceDump 解析单个接口的dump输出
func (wss *WireGuardShowService) parseInterfaceDump(interfaceName, output string) (*InterfaceShowInfo, error) {
	info := &InterfaceShowInfo{
		Name:     interfaceName,
		IsActive: true,
		Peers:    make(map[string]*PeerShowInfo),
	}

	var totalRx, totalTx uint64
	var lastHandshake *time.Time

	lines := strings.Split(strings.TrimSpace(output), "\n")
	isFirstLine := true

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}

		// 第一行是接口信息：private_key \t public_key \t listen_port \t off
		if isFirstLine && len(fields) >= 3 {
			info.PublicKey = fields[1] // 第二个字段是public_key
			if port, err := strconv.Atoi(fields[2]); err == nil {
				info.ListenPort = port
			}
			isFirstLine = false
			continue
		}

		// 其他行是peer信息：peer_public_key \t preshared_key \t endpoint \t allowed_ips \t latest_handshake \t rx_bytes \t tx_bytes \t persistent_keepalive
		if len(fields) >= 6 {
			peer := &PeerShowInfo{
				PublicKey:  fields[0],                     // 第一个字段就是peer的public_key
				AllowedIPs: strings.Split(fields[3], ","), // 调整索引
			}

			// 解析端点
			if len(fields) > 2 && fields[2] != "(none)" {
				peer.Endpoint = fields[2]
			}

			// 解析握手时间
			if len(fields) > 4 && fields[4] != "0" && fields[4] != "" {
				if timestamp, err := strconv.ParseInt(fields[4], 10, 64); err == nil {
					handshakeTime := time.Unix(timestamp, 0)
					peer.LatestHandshake = &handshakeTime
					peer.LastSeenAgo = wss.formatTimeAgo(handshakeTime)

					// 使用统一的超时常量判断在线状态
					timeSinceHandshake := time.Since(handshakeTime)
					if timeSinceHandshake <= config.WireGuardOnlineTimeout {
						peer.IsOnline = true
						info.ActivePeers++
					}

					// 更新接口最后握手时间
					if lastHandshake == nil || handshakeTime.After(*lastHandshake) {
						lastHandshake = &handshakeTime
					}
				}
			} else {
				// 如果没有握手记录但有流量，也可能是在线的
				// 检查是否有流量传输
				var hasTraffic bool
				if len(fields) > 5 {
					if rx, err := strconv.ParseUint(fields[5], 10, 64); err == nil && rx > 0 {
						hasTraffic = true
					}
				}
				if len(fields) > 6 {
					if tx, err := strconv.ParseUint(fields[6], 10, 64); err == nil && tx > 0 {
						hasTraffic = true
					}
				}

				if hasTraffic {
					peer.IsOnline = true
					info.ActivePeers++
				}
			}

			// 解析流量数据
			var rxBytes, txBytes uint64
			if len(fields) > 5 {
				if rx, err := strconv.ParseUint(fields[5], 10, 64); err == nil {
					rxBytes = rx
					totalRx += rx
				}
			}
			if len(fields) > 6 {
				if tx, err := strconv.ParseUint(fields[6], 10, 64); err == nil {
					txBytes = tx
					totalTx += tx
				}
			}
			peer.TrafficStats = wss.formatTrafficData(rxBytes, txBytes)

			// 解析保活间隔
			if len(fields) > 7 && fields[7] != "off" {
				if ka, err := strconv.Atoi(fields[7]); err == nil {
					peer.PersistentKA = ka
				}
			}

			info.Peers[peer.PublicKey] = peer
			info.PeerCount++
		}
	}

	// 设置接口总流量和最后握手时间
	info.TotalTraffic = wss.formatTrafficData(totalRx, totalTx)
	info.LastHandshake = lastHandshake

	return info, nil
}

// parseAllInterfacesDump 解析所有接口的dump输出
func (wss *WireGuardShowService) parseAllInterfacesDump(output string) (map[string]*InterfaceShowInfo, error) {
	interfaces := make(map[string]*InterfaceShowInfo)

	lines := strings.Split(strings.TrimSpace(output), "\n")

	// 按接口分组
	interfaceLines := make(map[string][]string)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) >= 2 {
			interfaceName := fields[0]
			if interfaceLines[interfaceName] == nil {
				interfaceLines[interfaceName] = make([]string, 0)
			}
			interfaceLines[interfaceName] = append(interfaceLines[interfaceName], line)
		}
	}

	// 解析每个接口
	for interfaceName, lines := range interfaceLines {
		output := strings.Join(lines, "\n")
		if info, err := wss.parseInterfaceDump(interfaceName, output); err == nil {
			interfaces[interfaceName] = info
		}
	}

	return interfaces, nil
}

// =====================================================
// 工具函数
// =====================================================

// formatTrafficData 格式化流量数据
func (wss *WireGuardShowService) formatTrafficData(rxBytes, txBytes uint64) TrafficData {
	return TrafficData{
		RxBytes: rxBytes,
		TxBytes: txBytes,
		RxMB:    wss.formatBytes(rxBytes),
		TxMB:    wss.formatBytes(txBytes),
		Total:   wss.formatBytes(rxBytes + txBytes),
	}
}

// formatBytes 格式化字节数
func (wss *WireGuardShowService) formatBytes(bytes uint64) string {
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := float64(bytes) / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	sizes := []string{"B", "KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), sizes[exp+1])
}

// formatTimeAgo 格式化时间差
func (wss *WireGuardShowService) formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return fmt.Sprintf("%d秒前", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%d分钟前", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%d小时前", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%d天前", int(duration.Hours()/24))
	}
}

// =====================================================
// 便捷检查方法
// =====================================================

// IsInterfaceActive 检查接口是否激活
func (wss *WireGuardShowService) IsInterfaceActive(interfaceName string) bool {
	info, err := wss.GetInterfaceInfo(interfaceName)
	if err != nil {
		return false
	}
	return info.IsActive
}

// GetInterfacePeerCount 获取接口peer数量
func (wss *WireGuardShowService) GetInterfacePeerCount(interfaceName string) (total, active int) {
	info, err := wss.GetInterfaceInfo(interfaceName)
	if err != nil {
		return 0, 0
	}
	return info.PeerCount, info.ActivePeers
}

// IsPeerOnline 检查peer是否在线
func (wss *WireGuardShowService) IsPeerOnline(interfaceName, publicKey string) bool {
	peer, err := wss.GetPeerInfo(interfaceName, publicKey)
	if err != nil {
		return false
	}
	return peer.IsOnline
}

// CheckConfigExists 检查配置文件是否存在
func (wss *WireGuardShowService) CheckConfigExists(interfaceName string) bool {
	configPath := fmt.Sprintf("/etc/wireguard/%s.conf", interfaceName)
	cmd := exec.Command("test", "-f", configPath)
	return cmd.Run() == nil
}
