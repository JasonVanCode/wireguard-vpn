package services

import (
	"fmt"
	"time"

	"eitec-vpn/internal/server/database"
	"eitec-vpn/internal/server/models"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"gorm.io/gorm"
)

// DashboardService 仪表盘服务
type DashboardService struct {
	db            *gorm.DB
	moduleService *ModuleService
}

// NewDashboardService 创建仪表盘服务
func NewDashboardService(moduleService *ModuleService) *DashboardService {
	return &DashboardService{
		db:            database.DB,
		moduleService: moduleService,
	}
}

// DashboardStats 仪表盘统计数据
type DashboardStats struct {
	ModuleStats    ModuleStatsInfo    `json:"module_stats"`
	TrafficStats   TrafficStatsInfo   `json:"traffic_stats"`
	SystemStats    SystemStatsInfo    `json:"system_stats"`
	RecentActivity []ActivityInfo     `json:"recent_activity"`
	TrafficChart   []TrafficDataPoint `json:"traffic_chart"`
	StatusChart    []StatusDataPoint  `json:"status_chart"`
	ModuleList     []ModuleListInfo   `json:"module_list"`
}

// ModuleStatsInfo 模块统计信息
type ModuleStatsInfo struct {
	Total        int64   `json:"total"`
	Online       int64   `json:"online"`
	Offline      int64   `json:"offline"`
	Warning      int64   `json:"warning"`
	Unconfigured int64   `json:"unconfigured"`
	OnlineRate   float64 `json:"online_rate"`
}

// TrafficStatsInfo 流量统计信息
type TrafficStatsInfo struct {
	TotalRx      uint64  `json:"total_rx"`
	TotalTx      uint64  `json:"total_tx"`
	Total        uint64  `json:"total"`
	TodayRx      uint64  `json:"today_rx"`
	TodayTx      uint64  `json:"today_tx"`
	TodayTotal   uint64  `json:"today_total"`
	AvgPerModule float64 `json:"avg_per_module"`
}

// SystemStatsInfo 系统统计信息
type SystemStatsInfo struct {
	Uptime           string    `json:"uptime"`
	StartTime        time.Time `json:"start_time"`
	UserCount        int64     `json:"user_count"`
	TotalConnections int64     `json:"total_connections"`
	ServerStatus     string    `json:"server_status"`
	IPPoolUsage      float64   `json:"ip_pool_usage"`
}

// ActivityInfo 活动信息
type ActivityInfo struct {
	ID        uint      `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	ModuleID  *uint     `json:"module_id,omitempty"`
	UserID    *uint     `json:"user_id,omitempty"`
}

// TrafficDataPoint 流量图表数据点
type TrafficDataPoint struct {
	Time time.Time `json:"time"`
	Rx   uint64    `json:"rx"`
	Tx   uint64    `json:"tx"`
}

// StatusDataPoint 状态图表数据点
type StatusDataPoint struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
	Color  string `json:"color"`
}

// ModuleListInfo 模块列表信息
type ModuleListInfo struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Location     string    `json:"location"`
	IPAddress    string    `json:"ip_address"`
	LastSeen     time.Time `json:"last_seen"`
	TotalRx      uint64    `json:"total_rx"`
	TotalTx      uint64    `json:"total_tx"`
	TotalTraffic uint64    `json:"total_traffic"`
	CreatedAt    time.Time `json:"created_at"`

	// 接口信息
	InterfaceID   uint   `json:"interface_id"`
	InterfaceName string `json:"interface_name"`
	Network       string `json:"network"`
}

// GetDashboardStats 获取仪表盘统计数据
func (ds *DashboardService) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	// 获取模块统计
	moduleStats, err := ds.getModuleStats()
	if err != nil {
		return nil, fmt.Errorf("获取模块统计失败: %w", err)
	}
	stats.ModuleStats = *moduleStats

	// 获取流量统计
	trafficStats, err := ds.getTrafficStats()
	if err != nil {
		return nil, fmt.Errorf("获取流量统计失败: %w", err)
	}
	stats.TrafficStats = *trafficStats

	// 获取系统统计
	systemStats, err := ds.getSystemStats()
	if err != nil {
		return nil, fmt.Errorf("获取系统统计失败: %w", err)
	}
	stats.SystemStats = *systemStats

	// 获取最近活动
	recentActivity, err := ds.getRecentActivity(10)
	if err != nil {
		return nil, fmt.Errorf("获取最近活动失败: %w", err)
	}
	stats.RecentActivity = recentActivity

	// 获取流量图表数据
	trafficChart, err := ds.getTrafficChart(24) // 最近24小时
	if err != nil {
		return nil, fmt.Errorf("获取流量图表数据失败: %w", err)
	}
	stats.TrafficChart = trafficChart

	// 获取状态图表数据
	statusChart, err := ds.getStatusChart()
	if err != nil {
		return nil, fmt.Errorf("获取状态图表数据失败: %w", err)
	}
	stats.StatusChart = statusChart

	// 获取模块列表
	moduleList, err := ds.getModuleList(20) // 获取最近20个模块
	if err != nil {
		return nil, fmt.Errorf("获取模块列表失败: %w", err)
	}
	stats.ModuleList = moduleList

	return stats, nil
}

// getModuleStats 获取模块统计信息
func (ds *DashboardService) getModuleStats() (*ModuleStatsInfo, error) {
	stats, err := ds.moduleService.GetModuleStats()
	if err != nil {
		return nil, err
	}

	onlineRate := 0.0
	if stats["total"] > 0 {
		onlineRate = float64(stats["online"]) / float64(stats["total"]) * 100
	}

	return &ModuleStatsInfo{
		Total:        stats["total"],
		Online:       stats["online"],
		Offline:      stats["offline"],
		Warning:      stats["warning"],
		Unconfigured: stats["unconfigured"],
		OnlineRate:   onlineRate,
	}, nil
}

// getTrafficStats 获取流量统计信息
func (ds *DashboardService) getTrafficStats() (*TrafficStatsInfo, error) {
	// 获取总流量
	totalTraffic, err := ds.moduleService.GetTotalTraffic()
	if err != nil {
		return nil, err
	}

	// 获取今日流量 (这里简化处理，实际应该有流量历史记录表)
	todayRx := totalTraffic["total_rx"] // 简化处理
	todayTx := totalTraffic["total_tx"]

	// 计算每个模块的平均流量
	var moduleCount int64
	if err := ds.db.Model(&models.Module{}).Count(&moduleCount).Error; err != nil {
		return nil, fmt.Errorf("统计模块数量失败: %w", err)
	}

	avgPerModule := 0.0
	if moduleCount > 0 {
		avgPerModule = float64(totalTraffic["total"]) / float64(moduleCount)
	}

	return &TrafficStatsInfo{
		TotalRx:      totalTraffic["total_rx"],
		TotalTx:      totalTraffic["total_tx"],
		Total:        totalTraffic["total"],
		TodayRx:      todayRx,
		TodayTx:      todayTx,
		TodayTotal:   todayRx + todayTx,
		AvgPerModule: avgPerModule,
	}, nil
}

// getSystemStats 获取系统统计信息
func (ds *DashboardService) getSystemStats() (*SystemStatsInfo, error) {
	// 获取系统启动时间
	startTimeStr, err := database.GetSystemConfig("system.start_time")
	if err != nil {
		startTimeStr = time.Now().Format(time.RFC3339)
		database.SetSystemConfig("system.start_time", startTimeStr)
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		startTime = time.Now()
	}

	// 计算运行时间
	uptime := time.Since(startTime).Round(time.Minute).String()

	// 获取用户数量
	var userCount int64
	if err := ds.db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return nil, fmt.Errorf("统计用户数量失败: %w", err)
	}

	// 获取总连接数 (模块数)
	var totalConnections int64
	if err := ds.db.Model(&models.Module{}).Count(&totalConnections).Error; err != nil {
		return nil, fmt.Errorf("统计连接数失败: %w", err)
	}

	// 计算IP池使用率
	ipPoolUsage, err := ds.getIPPoolUsage()
	if err != nil {
		return nil, fmt.Errorf("计算IP池使用率失败: %w", err)
	}

	return &SystemStatsInfo{
		Uptime:           uptime,
		StartTime:        startTime,
		UserCount:        userCount,
		TotalConnections: totalConnections,
		ServerStatus:     "running",
		IPPoolUsage:      ipPoolUsage,
	}, nil
}

// getIPPoolUsage 计算IP池使用率
func (ds *DashboardService) getIPPoolUsage() (float64, error) {
	var usedCount int64
	if err := ds.db.Model(&models.IPPool{}).Where("is_used = ?", true).Count(&usedCount).Error; err != nil {
		return 0, err
	}

	var totalCount int64
	if err := ds.db.Model(&models.IPPool{}).Count(&totalCount).Error; err != nil {
		return 0, err
	}

	if totalCount == 0 {
		return 0, nil
	}

	return float64(usedCount) / float64(totalCount) * 100, nil
}

// getRecentActivity 获取最近活动
func (ds *DashboardService) getRecentActivity(limit int) ([]ActivityInfo, error) {
	var activities []ActivityInfo

	// 简化：不再从数据库获取活动日志，可以从系统日志或其他方式获取
	// 这里返回空切片或模拟数据

	return activities, nil
}

// getTrafficChart 获取流量图表数据
func (ds *DashboardService) getTrafficChart(hours int) ([]TrafficDataPoint, error) {
	var dataPoints []TrafficDataPoint

	// 获取当前真实流量统计
	totalTraffic, err := ds.moduleService.GetTotalTraffic()
	if err != nil {
		return dataPoints, fmt.Errorf("获取总流量失败: %w", err)
	}

	// 使用当前流量数据作为最新点
	now := time.Now()
	currentDataPoint := TrafficDataPoint{
		Time: now,
		Rx:   totalTraffic["total_rx"],
		Tx:   totalTraffic["total_tx"],
	}

	// 如果没有历史数据表，至少返回当前流量点
	// 在实际应用中，应该创建流量历史记录表来存储时序数据
	for i := hours; i >= 0; i-- {
		pointTime := now.Add(time.Duration(-i) * time.Hour)

		if i == 0 {
			// 最新的点使用真实数据
			dataPoints = append(dataPoints, currentDataPoint)
		} else {
			// 历史点暂时使用零值，待实现流量历史记录功能
			dataPoints = append(dataPoints, TrafficDataPoint{
				Time: pointTime,
				Rx:   0,
				Tx:   0,
			})
		}
	}

	return dataPoints, nil
}

// getStatusChart 获取状态图表数据
func (ds *DashboardService) getStatusChart() ([]StatusDataPoint, error) {
	stats, err := ds.moduleService.GetModuleStats()
	if err != nil {
		return nil, err
	}

	statusColors := map[string]string{
		"online":       "#10b981", // green
		"offline":      "#ef4444", // red
		"warning":      "#f59e0b", // yellow
		"unconfigured": "#6b7280", // gray
	}

	var dataPoints []StatusDataPoint
	for status, count := range stats {
		if count > 0 {
			dataPoints = append(dataPoints, StatusDataPoint{
				Status: status,
				Count:  count,
				Color:  statusColors[status],
			})
		}
	}

	return dataPoints, nil
}

// GetModuleRanking 获取模块流量排行
func (ds *DashboardService) GetModuleRanking(limit int) ([]ModuleRankingInfo, error) {
	var modules []models.Module
	if err := ds.db.Order("total_rx_bytes + total_tx_bytes DESC").
		Limit(limit).
		Find(&modules).Error; err != nil {
		return nil, fmt.Errorf("查询模块排行失败: %w", err)
	}

	var ranking []ModuleRankingInfo
	for i, module := range modules {
		ranking = append(ranking, ModuleRankingInfo{
			Rank:       i + 1,
			ModuleID:   module.ID,
			ModuleName: module.Name,
			Location:   module.Location,
			TotalRx:    module.TotalRxBytes,
			TotalTx:    module.TotalTxBytes,
			Total:      module.TotalRxBytes + module.TotalTxBytes,
			Status:     module.Status.String(),
		})
	}

	return ranking, nil
}

// ModuleRankingInfo 模块排行信息
type ModuleRankingInfo struct {
	Rank       int    `json:"rank"`
	ModuleID   uint   `json:"module_id"`
	ModuleName string `json:"module_name"`
	Location   string `json:"location"`
	TotalRx    uint64 `json:"total_rx"`
	TotalTx    uint64 `json:"total_tx"`
	Total      uint64 `json:"total"`
	Status     string `json:"status"`
}

// GetSystemHealth 获取系统健康状态
func (ds *DashboardService) GetSystemHealth() (*SystemHealthInfo, error) {
	health := &SystemHealthInfo{
		Status: "healthy",
		Checks: make(map[string]HealthCheck),
	}

	// 检查数据库连接
	if err := ds.db.Exec("SELECT 1").Error; err != nil {
		health.Checks["database"] = HealthCheck{
			Status:  "error",
			Message: "数据库连接失败",
			Error:   err.Error(),
		}
		health.Status = "degraded"
	} else {
		health.Checks["database"] = HealthCheck{
			Status:  "ok",
			Message: "数据库连接正常",
		}
	}

	// 检查WireGuard状态
	wgStatus := ds.checkWireGuardStatus()
	health.Checks["wireguard"] = wgStatus
	if wgStatus.Status != "ok" && health.Status == "healthy" {
		health.Status = "degraded"
	}

	// 检查IP池状态
	ipPoolStatus := ds.checkIPPoolStatus()
	health.Checks["ip_pool"] = ipPoolStatus
	if ipPoolStatus.Status == "warning" && health.Status == "healthy" {
		health.Status = "warning"
	}

	// 获取系统资源信息
	systemRes, err := ds.getSystemResources()
	if err != nil {
		health.Checks["system_resources"] = HealthCheck{
			Status:  "error",
			Message: "获取系统资源失败",
			Error:   err.Error(),
		}
		if health.Status == "healthy" {
			health.Status = "degraded"
		}
	} else {
		health.SystemRes = systemRes
		health.Checks["system_resources"] = HealthCheck{
			Status:  "ok",
			Message: "系统资源正常",
		}
	}

	return health, nil
}

// SystemHealthInfo 系统健康信息
type SystemHealthInfo struct {
	Status    string                 `json:"status"`
	Checks    map[string]HealthCheck `json:"checks"`
	SystemRes *SystemResourceInfo    `json:"system_resources"`
}

// SystemResourceInfo 系统资源信息
type SystemResourceInfo struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	MemoryTotal uint64  `json:"memory_total"`
	MemoryUsed  uint64  `json:"memory_used"`
	DiskTotal   uint64  `json:"disk_total"`
	DiskUsed    uint64  `json:"disk_used"`
}

// HealthCheck 健康检查项
type HealthCheck struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// checkWireGuardStatus 检查WireGuard状态
func (ds *DashboardService) checkWireGuardStatus() HealthCheck {
	// 这里简化处理，实际应该检查WireGuard接口状态
	serverPublicKey, err := database.GetSystemConfig("server.public_key")
	if err != nil || serverPublicKey == "" {
		return HealthCheck{
			Status:  "error",
			Message: "WireGuard服务器未配置",
			Error:   "缺少服务器公钥配置",
		}
	}

	return HealthCheck{
		Status:  "ok",
		Message: "WireGuard服务正常",
	}
}

// checkIPPoolStatus 检查IP池状态
func (ds *DashboardService) checkIPPoolStatus() HealthCheck {
	usage, err := ds.getIPPoolUsage()
	if err != nil {
		return HealthCheck{
			Status:  "error",
			Message: "IP池状态检查失败",
			Error:   err.Error(),
		}
	}

	if usage > 90 {
		return HealthCheck{
			Status:  "warning",
			Message: fmt.Sprintf("IP池使用率过高: %.1f%%", usage),
		}
	}

	return HealthCheck{
		Status:  "ok",
		Message: fmt.Sprintf("IP池使用率正常: %.1f%%", usage),
	}
}

// getModuleList 获取模块列表
func (ds *DashboardService) getModuleList(limit int) ([]ModuleListInfo, error) {
	var modules []models.Module
	if err := ds.db.Preload("Interface").Order("last_seen DESC").Limit(limit).Find(&modules).Error; err != nil {
		return nil, fmt.Errorf("查询模块列表失败: %w", err)
	}

	var moduleList []ModuleListInfo
	for _, module := range modules {
		lastSeen := time.Time{}
		if module.LastSeen != nil {
			lastSeen = *module.LastSeen
		}

		// 获取接口信息
		interfaceName := ""
		network := ""
		if module.Interface != nil {
			interfaceName = module.Interface.Name
			network = module.Interface.Network
		}

		moduleList = append(moduleList, ModuleListInfo{
			ID:           module.ID,
			Name:         module.Name,
			Status:       module.Status.String(),
			Location:     module.Location,
			IPAddress:    module.IPAddress,
			LastSeen:     lastSeen,
			TotalRx:      module.TotalRxBytes,
			TotalTx:      module.TotalTxBytes,
			TotalTraffic: module.TotalRxBytes + module.TotalTxBytes,
			CreatedAt:    module.CreatedAt,

			// 接口信息
			InterfaceID:   module.InterfaceID,
			InterfaceName: interfaceName,
			Network:       network,
		})
	}

	return moduleList, nil
}

// getSystemResources 获取系统资源信息
func (ds *DashboardService) getSystemResources() (*SystemResourceInfo, error) {
	var sysRes SystemResourceInfo

	// 获取CPU使用率
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("获取CPU使用率失败: %w", err)
	}
	if len(cpuPercent) > 0 {
		sysRes.CPUUsage = cpuPercent[0]
	}

	// 获取内存信息
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("获取内存信息失败: %w", err)
	}
	sysRes.MemoryUsage = memInfo.UsedPercent
	sysRes.MemoryTotal = memInfo.Total
	sysRes.MemoryUsed = memInfo.Used

	// 获取磁盘信息 (根目录)
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("获取磁盘信息失败: %w", err)
	}
	sysRes.DiskUsage = diskInfo.UsedPercent
	sysRes.DiskTotal = diskInfo.Total
	sysRes.DiskUsed = diskInfo.Used

	return &sysRes, nil
}
