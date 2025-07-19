package services

import (
	"bufio"
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

// GetWireGuardStatus 获取WireGuard状态
func (ss *StatusService) GetWireGuardStatus() (*WireGuardStatus, error) {
	status := &WireGuardStatus{
		Interface:   "wg0",
		Status:      "stopped",
		LastUpdated: time.Now(),
	}

	// 检查WireGuard是否运行
	cmd := exec.Command("wg", "show", "wg0")
	output, err := cmd.Output()
	if err != nil {
		return status, nil // WireGuard未运行
	}

	status.Status = "running"

	// 解析WireGuard输出
	lines := strings.Split(string(output), "\n")
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
			// 新的peer开始
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
				if t, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
					currentPeer.LatestHandshake = t
				}
			}
		case "transfer:":
			if currentPeer != nil && len(fields) >= 5 {
				// 格式: transfer: 1.23 MiB received, 4.56 KiB sent
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

	// 添加最后一个peer
	if currentPeer != nil {
		status.Peers = append(status.Peers, *currentPeer)
	}

	return status, nil
}

// GetTrafficStats 获取流量统计
func (ss *StatusService) GetTrafficStats() (*TrafficStats, error) {
	stats := &TrafficStats{
		LastUpdated: time.Now(),
	}

	// 从网络接口统计信息获取流量
	content, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		// 在非Linux系统或文件不存在时，返回默认值而不是错误
		return stats, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "wg0:") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				// 接收字节数 (第2个字段)
				if rx, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					stats.RxBytes = rx
				}
				// 接收包数 (第3个字段)
				if rxPkts, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
					stats.RxPackets = rxPkts
				}
				// 发送字节数 (第10个字段)
				if tx, err := strconv.ParseUint(fields[9], 10, 64); err == nil {
					stats.TxBytes = tx
				}
				// 发送包数 (第11个字段)
				if txPkts, err := strconv.ParseUint(fields[10], 10, 64); err == nil {
					stats.TxPackets = txPkts
				}
			}
			break
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
