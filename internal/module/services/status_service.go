package services

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"eitec-vpn/internal/shared/config"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// StatusService 状态服务
type StatusService struct {
	config *config.ModuleConfig
}

// NewStatusService 创建状态服务
func NewStatusService(cfg *config.ModuleConfig) *StatusService {
	return &StatusService{
		config: cfg,
	}
}

// WireGuardStatus WireGuard状态
type WireGuardStatus struct {
	Interface   string       `json:"interface"`
	Status      string       `json:"status"`
	PublicKey   string       `json:"public_key"`
	ListenPort  int          `json:"listen_port"`
	Peers       []PeerStatus `json:"peers"`
	LastUpdated time.Time    `json:"last_updated"`
}

// PeerStatus 对等节点状态
type PeerStatus struct {
	PublicKey       string    `json:"public_key"`
	Endpoint        string    `json:"endpoint"`
	AllowedIPs      []string  `json:"allowed_ips"`
	LatestHandshake time.Time `json:"latest_handshake"`
	TransferRx      uint64    `json:"transfer_rx"`
	TransferTx      uint64    `json:"transfer_tx"`
	PersistentKA    int       `json:"persistent_keepalive"`
}

// TrafficStats 流量统计
type TrafficStats struct {
	RxBytes     uint64    `json:"rx_bytes"`
	TxBytes     uint64    `json:"tx_bytes"`
	RxPackets   uint64    `json:"rx_packets"`
	TxPackets   uint64    `json:"tx_packets"`
	LastUpdated time.Time `json:"last_updated"`
}

// SystemStatus 系统状态
type SystemStatus struct {
	Uptime      int64       `json:"uptime"`
	LoadAverage string      `json:"load_average"`
	Memory      MemoryInfo  `json:"memory"`
	Disk        DiskInfo    `json:"disk"`
	Network     NetworkInfo `json:"network"`
	LastUpdated time.Time   `json:"last_updated"`
}

// MemoryInfo 内存信息
type MemoryInfo struct {
	Total     uint64  `json:"total"`
	Free      uint64  `json:"free"`
	Available uint64  `json:"available"`
	Used      uint64  `json:"used"`
	Percent   float64 `json:"percent"`
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	Total   uint64  `json:"total"`
	Free    uint64  `json:"free"`
	Used    uint64  `json:"used"`
	Percent float64 `json:"percent"`
}

// NetworkInfo 网络信息
type NetworkInfo struct {
	Interfaces []NetworkInterface `json:"interfaces"`
}

// NetworkInterface 网络接口
type NetworkInterface struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	RxBytes uint64 `json:"rx_bytes"`
	TxBytes uint64 `json:"tx_bytes"`
}

// VPNStatus VPN状态信息
type VPNStatus struct {
	Status            string    `json:"status"`             // running/stopped
	ConnectionQuality string    `json:"connection_quality"` // excellent/good/fair/poor
	Latency           int       `json:"latency"`            // 延迟(ms)
	LastHandshake     time.Time `json:"last_handshake"`     // 最后握手时间
	Uptime            string    `json:"uptime"`             // 运行时间
}

// NetworkMetrics 网络性能指标
type NetworkMetrics struct {
	Latency    int     `json:"latency"`     // 延迟(ms)
	PacketLoss float64 `json:"packet_loss"` // 丢包率(%)
	Bandwidth  float64 `json:"bandwidth"`   // 带宽(Mbps)
	Quality    string  `json:"quality"`     // 网络质量
	Status     string  `json:"status"`      // 连接状态
}

// GetWireGuardStatus 获取WireGuard状态
func (ss *StatusService) GetWireGuardStatus() (*WireGuardStatus, error) {
	status := &WireGuardStatus{
		Interface:   "wg0",
		Status:      "stopped",
		LastUpdated: time.Now(),
	}

	// 首先检查网络接口是否存在
	if _, err := os.Stat("/sys/class/net/wg0"); err != nil {
		status.Status = "stopped"
		return status, nil
	}

	// 使用wg show wg0 dump命令获取更简洁的输出
	cmd := exec.Command("wg", "show", "wg0", "dump")
	output, err := cmd.Output()
	if err != nil {
		// 如果dump命令失败，尝试普通的wg show命令
		cmd = exec.Command("wg", "show", "wg0")
		output, err = cmd.Output()
		if err != nil {
			// 接口存在但命令失败，可能是配置问题
			status.Status = "configured"
			return status, nil
		}
		// 使用普通格式解析
		return ss.parseWireGuardOutput(string(output))
	}

	// 使用dump格式解析
	return ss.parseWireGuardDump(string(output))
}

// parseWireGuardDump 解析wg show wg0 dump输出
func (ss *StatusService) parseWireGuardDump(output string) (*WireGuardStatus, error) {
	status := &WireGuardStatus{
		Interface:   "wg0",
		Status:      "running",
		LastUpdated: time.Now(),
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 1 {
		return status, nil
	}

	// 第一行是interface信息
	// 格式: private_key public_key listen_port status
	interfaceFields := strings.Fields(lines[0])
	if len(interfaceFields) >= 4 {
		status.PublicKey = interfaceFields[1]
		if port, err := strconv.Atoi(interfaceFields[2]); err == nil {
			status.ListenPort = port
		}
		// interfaceFields[3] 是状态 (off/on)
		if len(interfaceFields) >= 4 {
			if interfaceFields[3] == "off" {
				status.Status = "stopped"
			} else {
				status.Status = "running"
			}
		}
	}

	// 后续行是peer信息
	// 格式: public_key preshared_key endpoint allowed_ips last_handshake rx_bytes tx_bytes persistent_keepalive
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 8 {
			peer := &PeerStatus{
				PublicKey: fields[0],
				Endpoint:  fields[2],
			}

			// 解析allowed ips
			if len(fields) >= 4 {
				peer.AllowedIPs = strings.Split(fields[3], ",")
			}

			// 解析last handshake (Unix timestamp)
			if len(fields) >= 5 {
				if timestamp, err := strconv.ParseInt(fields[4], 10, 64); err == nil {
					peer.LatestHandshake = time.Unix(timestamp, 0)
				}
			}

			// 解析流量统计
			if len(fields) >= 7 {
				if rx, err := strconv.ParseUint(fields[5], 10, 64); err == nil {
					peer.TransferRx = rx
				}
				if tx, err := strconv.ParseUint(fields[6], 10, 64); err == nil {
					peer.TransferTx = tx
				}
			}

			// 解析persistent keepalive
			if len(fields) >= 8 {
				if ka, err := strconv.Atoi(fields[7]); err == nil {
					peer.PersistentKA = ka
				}
			}

			status.Peers = append(status.Peers, *peer)
		}
	}

	return status, nil
}

// parseWireGuardOutput 解析普通的wg show wg0输出（备用方法）
func (ss *StatusService) parseWireGuardOutput(output string) (*WireGuardStatus, error) {
	status := &WireGuardStatus{
		Interface:   "wg0",
		Status:      "stopped", // 默认为stopped，需要根据实际输出判断
		LastUpdated: time.Now(),
	}

	lines := strings.Split(output, "\n")
	var currentPeer *PeerStatus

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "interface:":
			status.Interface = fields[1]
		case "public":
			if len(fields) >= 3 && fields[1] == "key:" {
				status.PublicKey = fields[2]
			}
		case "listening":
			if len(fields) >= 3 && fields[1] == "port:" {
				if port, err := strconv.Atoi(fields[2]); err == nil {
					status.ListenPort = port
				}
			}
		case "peer:":
			if currentPeer != nil {
				status.Peers = append(status.Peers, *currentPeer)
			}
			currentPeer = &PeerStatus{
				PublicKey: fields[1],
			}
		case "endpoint:":
			if currentPeer != nil {
				currentPeer.Endpoint = fields[1]
			}
		case "allowed":
			if currentPeer != nil && len(fields) >= 3 && fields[1] == "ips:" {
				ips := strings.Join(fields[2:], " ")
				currentPeer.AllowedIPs = strings.Split(ips, ",")
				for i, ip := range currentPeer.AllowedIPs {
					currentPeer.AllowedIPs[i] = strings.TrimSpace(ip)
				}
			}
		case "latest":
			if currentPeer != nil && len(fields) >= 3 && fields[1] == "handshake:" {
				timeStr := strings.Join(fields[2:], " ")
				if strings.Contains(timeStr, "ago") {
					currentPeer.LatestHandshake = time.Now().Add(-time.Minute)
				} else if t, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
					currentPeer.LatestHandshake = t
				}
			}
		case "transfer:":
			if currentPeer != nil && len(fields) >= 5 {
				if rx, err := parseTransferSize(fields[1], fields[2]); err == nil {
					currentPeer.TransferRx = rx
				}
				if len(fields) >= 6 {
					if tx, err := parseTransferSize(fields[4], fields[5]); err == nil {
						currentPeer.TransferTx = tx
					}
				}
			}
		case "persistent":
			if currentPeer != nil && len(fields) >= 4 && fields[1] == "keepalive:" {
				if ka, err := strconv.Atoi(fields[3]); err == nil {
					currentPeer.PersistentKA = ka
				}
			}
		}
	}

	if currentPeer != nil {
		status.Peers = append(status.Peers, *currentPeer)
	}

	// 根据是否有peer和handshake时间来判断状态
	if len(status.Peers) > 0 {
		// 检查最新的handshake时间
		latestHandshake := time.Time{}
		for _, peer := range status.Peers {
			if peer.LatestHandshake.After(latestHandshake) {
				latestHandshake = peer.LatestHandshake
			}
		}

		// 如果handshake时间在2分钟内，认为是running状态
		if time.Since(latestHandshake) < 2*time.Minute {
			status.Status = "running"
		} else {
			status.Status = "stopped"
		}
	} else {
		// 没有peer，可能是configured状态
		status.Status = "configured"
	}

	return status, nil
}

// GetTrafficStats 获取流量统计
func (ss *StatusService) GetTrafficStats() (*TrafficStats, error) {
	stats := &TrafficStats{
		LastUpdated: time.Now(),
	}

	// 直接执行 wg show wg0 dump 命令获取状态
	cmd := exec.Command("wg", "show", "wg0", "dump")
	output, err := cmd.Output()
	if err != nil {
		// 命令执行失败，说明接口不存在或未配置
		return stats, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) <= 1 {
		// 没有peer，返回默认值
		return stats, nil
	}

	// 找到最新的handshake时间和对应的流量数据
	latestHandshake := time.Time{}
	var rxBytes, txBytes uint64
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 7 {
			// 第5个字段是last_handshake (Unix timestamp)
			if timestamp, err := strconv.ParseInt(fields[4], 10, 64); err == nil {
				handshakeTime := time.Unix(timestamp, 0)
				if handshakeTime.After(latestHandshake) {
					latestHandshake = handshakeTime
					// 第6、7个字段是rx_bytes和tx_bytes
					if rx, err := strconv.ParseUint(fields[5], 10, 64); err == nil {
						rxBytes = rx
					}
					if tx, err := strconv.ParseUint(fields[6], 10, 64); err == nil {
						txBytes = tx
					}
				}
			}
		}
	}

	// 基于handshake时间判断是否有活跃连接
	// 注意：即使接口状态是"off"，如果有活跃的handshake，仍然认为连接是活跃的
	if !latestHandshake.IsZero() && time.Since(latestHandshake) < 2*time.Minute {
		stats.RxBytes = rxBytes
		stats.TxBytes = txBytes
		// 简单估算包数
		if stats.RxBytes > 0 {
			stats.RxPackets = stats.RxBytes / 1024
		}
		if stats.TxBytes > 0 {
			stats.TxPackets = stats.TxBytes / 1024
		}
	}

	return stats, nil
}

// GetSystemStatus 获取系统状态
func (ss *StatusService) GetSystemStatus() (*SystemStatus, error) {
	status := &SystemStatus{
		LastUpdated: time.Now(),
	}

	// 获取系统运行时间 - 使用gopsutil
	if hostStat, err := host.Info(); err == nil {
		status.Uptime = int64(hostStat.Uptime)
	}

	// 获取负载平均值 - 尝试使用gopsutil的load模块，如果失败则尝试传统方法
	if avgStat, err := cpu.Percent(time.Second, false); err == nil && len(avgStat) > 0 {
		status.LoadAverage = strconv.FormatFloat(avgStat[0], 'f', 2, 64) + "%"
	} else {
		// 回退到读取文件的方法（仅Linux）
		if content, err := os.ReadFile("/proc/loadavg"); err == nil {
			status.LoadAverage = strings.TrimSpace(string(content))
		}
	}

	// 获取内存信息 - 使用gopsutil
	if memInfo, err := ss.getMemoryInfoV2(); err == nil {
		status.Memory = *memInfo
	} else {
		// 提供默认内存信息
		status.Memory = MemoryInfo{
			Total:     0,
			Free:      0,
			Available: 0,
			Used:      0,
			Percent:   0,
		}
	}

	// 获取磁盘信息 - 使用gopsutil
	if diskInfo, err := ss.getDiskInfoV2(); err == nil {
		status.Disk = *diskInfo
	} else {
		// 提供默认磁盘信息
		status.Disk = DiskInfo{
			Total:   0,
			Free:    0,
			Used:    0,
			Percent: 0,
		}
	}

	// 获取网络信息 - 错误不影响整体返回
	if networkInfo, err := ss.getNetworkInfo(); err == nil {
		status.Network = *networkInfo
	} else {
		// 提供默认网络信息
		status.Network = NetworkInfo{
			Interfaces: []NetworkInterface{},
		}
	}

	return status, nil
}

// getMemoryInfo 获取内存信息
func (ss *StatusService) getMemoryInfo() (*MemoryInfo, error) {
	memInfo := &MemoryInfo{}

	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		// 在非Linux系统或文件不存在时，返回默认值
		return memInfo, nil
	}

	lines := strings.Split(string(content), "\n")
	memData := make(map[string]uint64)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			key := strings.TrimSuffix(fields[0], ":")
			if value, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				memData[key] = value * 1024 // 转换为字节
			}
		}
	}

	memInfo.Total = memData["MemTotal"]
	memInfo.Free = memData["MemFree"]
	memInfo.Available = memData["MemAvailable"]
	memInfo.Used = memInfo.Total - memInfo.Available

	if memInfo.Total > 0 {
		memInfo.Percent = float64(memInfo.Used) / float64(memInfo.Total) * 100
	}

	return memInfo, nil
}

// getDiskInfo 获取磁盘信息
func (ss *StatusService) getDiskInfo() (*DiskInfo, error) {
	diskInfo := &DiskInfo{}

	// 使用df命令获取磁盘使用情况
	cmd := exec.Command("df", "-B1", "/")
	output, err := cmd.Output()
	if err != nil {
		return diskInfo, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) >= 2 {
		fields := strings.Fields(lines[1])
		if len(fields) >= 4 {
			if total, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				diskInfo.Total = total
			}
			if used, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
				diskInfo.Used = used
			}
			if free, err := strconv.ParseUint(fields[3], 10, 64); err == nil {
				diskInfo.Free = free
			}
		}
	}

	if diskInfo.Total > 0 {
		diskInfo.Percent = float64(diskInfo.Used) / float64(diskInfo.Total) * 100
	}

	return diskInfo, nil
}

// getNetworkInfo 获取网络信息
func (ss *StatusService) getNetworkInfo() (*NetworkInfo, error) {
	networkInfo := &NetworkInfo{}

	// 读取网络接口统计信息
	content, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		// 在非Linux系统或文件不存在时，返回空列表而不是错误
		return networkInfo, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	// 跳过头两行
	scanner.Scan()
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 10 {
			name := strings.TrimSuffix(fields[0], ":")

			// 跳过loopback接口
			if name == "lo" {
				continue
			}

			iface := NetworkInterface{
				Name:   name,
				Status: "up", // 简化处理，实际可以读取/sys/class/net/{name}/operstate
			}

			if rx, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				iface.RxBytes = rx
			}
			if tx, err := strconv.ParseUint(fields[9], 10, 64); err == nil {
				iface.TxBytes = tx
			}

			networkInfo.Interfaces = append(networkInfo.Interfaces, iface)
		}
	}

	return networkInfo, nil
}

// parseTransferSize 解析传输大小
func parseTransferSize(sizeStr, unitStr string) (uint64, error) {
	size, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		return 0, err
	}

	unit := strings.ToLower(unitStr)
	switch unit {
	case "b", "bytes":
		return uint64(size), nil
	case "kib":
		return uint64(size * 1024), nil
	case "mib":
		return uint64(size * 1024 * 1024), nil
	case "gib":
		return uint64(size * 1024 * 1024 * 1024), nil
	default:
		return uint64(size), nil
	}
}

// IsHealthy 检查系统健康状态
func (ss *StatusService) IsHealthy() bool {
	// 检查WireGuard是否运行
	wgStatus, err := ss.GetWireGuardStatus()
	if err != nil || wgStatus.Status != "running" {
		return false
	}

	// 检查系统资源
	sysStatus, err := ss.GetSystemStatus()
	if err != nil {
		return false
	}

	// 检查内存使用率 (超过90%认为不健康)
	if sysStatus.Memory.Percent > 90 {
		return false
	}

	// 检查磁盘使用率 (超过95%认为不健康)
	if sysStatus.Disk.Percent > 95 {
		return false
	}

	return true
}

// GetConnectionQuality 获取连接质量
func (ss *StatusService) GetConnectionQuality() map[string]interface{} {
	quality := make(map[string]interface{})

	wgStatus, err := ss.GetWireGuardStatus()
	if err != nil || wgStatus.Status != "running" {
		quality["status"] = "disconnected"
		quality["latency"] = -1
		quality["packet_loss"] = -1
		return quality
	}

	quality["status"] = "connected"

	// 如果有peer，检查连接质量
	if len(wgStatus.Peers) > 0 {
		peer := wgStatus.Peers[0]

		// 检查最后握手时间
		timeSinceHandshake := time.Since(peer.LatestHandshake)
		if timeSinceHandshake > 2*time.Minute {
			quality["status"] = "unstable"
		}

		quality["last_handshake"] = peer.LatestHandshake
		quality["time_since_handshake"] = timeSinceHandshake.Seconds()
	}

	return quality
}

// GetDetailedStats 获取详细统计信息
func (ss *StatusService) GetDetailedStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// WireGuard状态
	wgStatus, err := ss.GetWireGuardStatus()
	if err == nil {
		stats["wireguard"] = wgStatus
	}

	// 流量统计
	trafficStats, err := ss.GetTrafficStats()
	if err == nil {
		stats["traffic"] = trafficStats
	}

	// 系统状态
	systemStatus, err := ss.GetSystemStatus()
	if err == nil {
		stats["system"] = systemStatus
	}

	// 连接质量
	stats["connection_quality"] = ss.GetConnectionQuality()

	// 健康状态
	stats["healthy"] = ss.IsHealthy()

	stats["last_updated"] = time.Now()

	return stats, nil
}

// getMemoryInfoV2 获取内存信息 - 使用gopsutil (跨平台)
func (ss *StatusService) getMemoryInfoV2() (*MemoryInfo, error) {
	memInfo := &MemoryInfo{}

	// 使用gopsutil获取内存信息
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return memInfo, err
	}

	memInfo.Total = vmStat.Total
	memInfo.Free = vmStat.Free
	memInfo.Available = vmStat.Available
	memInfo.Used = vmStat.Used
	memInfo.Percent = vmStat.UsedPercent

	return memInfo, nil
}

// getDiskInfoV2 获取磁盘信息 - 使用gopsutil (跨平台)
func (ss *StatusService) getDiskInfoV2() (*DiskInfo, error) {
	diskInfo := &DiskInfo{}

	// 获取根目录磁盘使用情况
	diskStat, err := disk.Usage("/")
	if err != nil {
		return diskInfo, err
	}

	diskInfo.Total = diskStat.Total
	diskInfo.Free = diskStat.Free
	diskInfo.Used = diskStat.Used
	diskInfo.Percent = diskStat.UsedPercent

	return diskInfo, nil
}

// GetVPNStatus 获取VPN状态信息
func (ss *StatusService) GetVPNStatus() (*VPNStatus, error) {
	status := &VPNStatus{
		Status:            "stopped", // 默认停止
		ConnectionQuality: "disconnected",
		Latency:           0,
		LastHandshake:     time.Now(),
		Uptime:            "0秒",
	}

	// 直接执行 wg show wg0 dump 命令获取状态
	cmd := exec.Command("wg", "show", "wg0", "dump")
	output, err := cmd.Output()
	if err != nil {
		// 命令执行失败，说明接口不存在或未配置
		status.Status = "stopped"
		status.ConnectionQuality = "disconnected"
		status.Uptime = "接口未配置"
		return status, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 1 {
		return status, nil
	}

	// 检查是否有peer连接
	if len(lines) <= 1 {
		// 没有peer，可能是configured状态
		status.Status = "configured"
		status.ConnectionQuality = "unknown"
		status.Uptime = "已配置"
		return status, nil
	}

	// 找到最新的handshake时间
	latestHandshake := time.Time{}
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 {
			// 第5个字段是last_handshake (Unix timestamp)
			if timestamp, err := strconv.ParseInt(fields[4], 10, 64); err == nil {
				handshakeTime := time.Unix(timestamp, 0)
				if handshakeTime.After(latestHandshake) {
					latestHandshake = handshakeTime
				}
			}
		}
	}

	// 如果没有找到handshake时间，说明没有活跃连接
	if latestHandshake.IsZero() {
		status.Status = "stopped"
		status.ConnectionQuality = "disconnected"
		status.Uptime = "无连接记录"
		return status, nil
	}

	status.LastHandshake = latestHandshake
	timeSinceHandshake := time.Since(latestHandshake)

	// 核心判断逻辑：如果handshake时间在2分钟内，说明在线
	// 注意：即使接口状态是"off"，如果有活跃的handshake，仍然认为连接是活跃的
	if timeSinceHandshake < 2*time.Minute {
		status.Status = "running"

		// 计算连接质量
		if timeSinceHandshake < time.Minute {
			status.ConnectionQuality = "excellent"
		} else {
			status.ConnectionQuality = "good"
		}

		// 计算运行时间（从最后握手时间开始）
		status.Uptime = formatDuration(timeSinceHandshake)
	} else {
		// 超过2分钟没有handshake，算离线
		status.Status = "stopped"
		status.ConnectionQuality = "disconnected"
		status.Uptime = formatDuration(timeSinceHandshake) + " (离线)"
	}

	return status, nil
}

// formatDuration 格式化时间间隔
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f秒", d.Seconds())
	} else if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%d分%d秒", minutes, seconds)
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%d小时%d分", hours, minutes)
	}
}

// GetNetworkMetrics 获取网络性能指标
func (ss *StatusService) GetNetworkMetrics() (*NetworkMetrics, error) {
	metrics := &NetworkMetrics{}

	// 直接执行 wg show wg0 dump 命令获取状态
	cmd := exec.Command("wg", "show", "wg0", "dump")
	output, err := cmd.Output()
	if err != nil {
		// 命令执行失败，说明接口不存在或未配置
		metrics.Status = "disconnected"
		metrics.Quality = "unknown"
		return metrics, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) <= 1 {
		// 没有peer，可能是configured状态
		metrics.Status = "configured"
		metrics.Quality = "unknown"
		return metrics, nil
	}

	// 找到最新的handshake时间和对应的endpoint
	latestHandshake := time.Time{}
	var endpoint string
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 {
			// 第5个字段是last_handshake (Unix timestamp)
			if timestamp, err := strconv.ParseInt(fields[4], 10, 64); err == nil {
				handshakeTime := time.Unix(timestamp, 0)
				if handshakeTime.After(latestHandshake) {
					latestHandshake = handshakeTime
					// 第3个字段是endpoint
					if len(fields) >= 3 {
						endpoint = fields[2]
					}
				}
			}
		}
	}

	// 基于handshake时间判断连接状态
	// 注意：即使接口状态是"off"，如果有活跃的handshake，仍然认为连接是活跃的
	timeSinceHandshake := time.Since(latestHandshake)
	if timeSinceHandshake < 2*time.Minute {
		// 2分钟内有handshake，说明连接是活跃的
		metrics.Status = "connected"

		// 测量延迟
		if endpoint != "" {
			latency, err := ss.measureLatency(endpoint)
			if err == nil {
				metrics.Latency = latency
			}
		}

		// 计算丢包率
		if endpoint != "" {
			packetLoss, err := ss.measurePacketLoss(endpoint)
			if err == nil {
				metrics.PacketLoss = packetLoss
			}
		}

		// 评估网络质量
		metrics.Quality = ss.evaluateNetworkQuality(metrics.Latency, metrics.PacketLoss)
	} else {
		// 超过2分钟没有handshake，算离线
		metrics.Status = "disconnected"
		metrics.Quality = "disconnected"
	}

	return metrics, nil
}

// measureLatency 测量延迟
func (ss *StatusService) measureLatency(endpoint string) (int, error) {
	// 解析endpoint获取IP地址
	ip := strings.Split(endpoint, ":")[0]

	// 使用ping命令测量延迟
	cmd := exec.Command("ping", "-c", "3", "-W", "1", ip)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// 解析ping输出获取平均延迟
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "avg") {
			// 提取延迟数值
			// 格式: round-trip min/avg/max/mdev = 0.123/0.456/0.789/0.123 ms
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				avgPart := strings.Split(parts[1], "/")[1]
				avgPart = strings.TrimSpace(avgPart)
				if latency, err := strconv.ParseFloat(avgPart, 64); err == nil {
					return int(latency), nil
				}
			}
		}
	}

	return 0, fmt.Errorf("无法解析ping结果")
}

// measurePacketLoss 测量丢包率
func (ss *StatusService) measurePacketLoss(endpoint string) (float64, error) {
	// 解析endpoint获取IP地址
	ip := strings.Split(endpoint, ":")[0]

	// 使用ping命令测量丢包率
	cmd := exec.Command("ping", "-c", "10", "-W", "1", ip)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// 解析ping输出获取丢包率
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "packets transmitted") {
			// 格式: 10 packets transmitted, 9 received, 10% packet loss
			parts := strings.Split(line, ",")
			if len(parts) >= 3 {
				lossPart := strings.TrimSpace(parts[2])
				if strings.Contains(lossPart, "%") {
					lossStr := strings.Split(lossPart, "%")[0]
					if loss, err := strconv.ParseFloat(lossStr, 64); err == nil {
						return loss, nil
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("无法解析ping结果")
}

// evaluateNetworkQuality 评估网络质量
func (ss *StatusService) evaluateNetworkQuality(latency int, packetLoss float64) string {
	if latency < 50 && packetLoss < 1 {
		return "excellent"
	} else if latency < 100 && packetLoss < 5 {
		return "good"
	} else if latency < 200 && packetLoss < 10 {
		return "fair"
	} else {
		return "poor"
	}
}
