package handlers

import (
	"eitec-vpn/internal/module/services"
	"eitec-vpn/internal/shared/response"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// DashboardHandler 仪表板控制器
type DashboardHandler struct {
	statusService *services.StatusService
	moduleService *services.ModuleService
}

// NewDashboardHandler 创建仪表板控制器
func NewDashboardHandler(statusService *services.StatusService, moduleService *services.ModuleService) *DashboardHandler {
	return &DashboardHandler{
		statusService: statusService,
		moduleService: moduleService,
	}
}

// DashboardStats 仪表板统计数据
type DashboardStats struct {
	VPNStatus      *services.VPNStatus      `json:"vpn_status"`
	TrafficStats   *services.TrafficStats   `json:"traffic_stats"`
	NetworkMetrics *services.NetworkMetrics `json:"network_metrics"`
	SystemStatus   *services.SystemStatus   `json:"system_status"`
	LastUpdated    time.Time                `json:"last_updated"`
}

// GetDashboardStats 获取仪表板统计数据
func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	stats := &DashboardStats{
		LastUpdated: time.Now(),
	}

	// 并行获取各种状态信息
	var wg sync.WaitGroup
	var errChan = make(chan error, 4)

	wg.Add(4)

	// 获取VPN状态
	go func() {
		defer wg.Done()
		vpnStatus, err := h.statusService.GetVPNStatus()
		if err != nil {
			errChan <- err
			return
		}
		stats.VPNStatus = vpnStatus
	}()

	// 获取流量统计
	go func() {
		defer wg.Done()
		trafficStats, err := h.statusService.GetTrafficStats()
		if err != nil {
			errChan <- err
			return
		}
		stats.TrafficStats = trafficStats
	}()

	// 获取网络性能
	go func() {
		defer wg.Done()
		networkMetrics, err := h.statusService.GetNetworkMetrics()
		if err != nil {
			errChan <- err
			return
		}
		stats.NetworkMetrics = networkMetrics
	}()

	// 获取系统资源
	go func() {
		defer wg.Done()
		systemStatus, err := h.statusService.GetSystemStatus()
		if err != nil {
			errChan <- err
			return
		}
		stats.SystemStatus = systemStatus
	}()

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	if len(errChan) > 0 {
		response.InternalError(c, "获取统计数据失败")
		return
	}

	response.Success(c, stats)
}

// GetVPNStatus 获取VPN状态
func (h *DashboardHandler) GetVPNStatus(c *gin.Context) {
	vpnStatus, err := h.statusService.GetVPNStatus()
	if err != nil {
		response.InternalError(c, "获取VPN状态失败: "+err.Error())
		return
	}

	response.Success(c, vpnStatus)
}

// GetTrafficStats 获取流量统计
func (h *DashboardHandler) GetTrafficStats(c *gin.Context) {
	trafficStats, err := h.statusService.GetTrafficStats()
	if err != nil {
		response.InternalError(c, "获取流量统计失败: "+err.Error())
		return
	}

	response.Success(c, trafficStats)
}

// GetNetworkMetrics 获取网络性能指标
func (h *DashboardHandler) GetNetworkMetrics(c *gin.Context) {
	networkMetrics, err := h.statusService.GetNetworkMetrics()
	if err != nil {
		response.InternalError(c, "获取网络性能指标失败: "+err.Error())
		return
	}

	response.Success(c, networkMetrics)
}

// GetSystemStatus 获取系统状态
func (h *DashboardHandler) GetSystemStatus(c *gin.Context) {
	systemStatus, err := h.statusService.GetSystemStatus()
	if err != nil {
		response.InternalError(c, "获取系统状态失败: "+err.Error())
		return
	}

	response.Success(c, systemStatus)
}

// RefreshDashboard 刷新仪表板数据
func (h *DashboardHandler) RefreshDashboard(c *gin.Context) {
	// 强制刷新所有数据
	stats := &DashboardStats{
		LastUpdated: time.Now(),
	}

	var wg sync.WaitGroup
	var errChan = make(chan error, 4)

	wg.Add(4)

	// 并行刷新所有数据
	go func() {
		defer wg.Done()
		vpnStatus, err := h.statusService.GetVPNStatus()
		if err != nil {
			errChan <- err
			return
		}
		stats.VPNStatus = vpnStatus
	}()

	go func() {
		defer wg.Done()
		trafficStats, err := h.statusService.GetTrafficStats()
		if err != nil {
			errChan <- err
			return
		}
		stats.TrafficStats = trafficStats
	}()

	go func() {
		defer wg.Done()
		networkMetrics, err := h.statusService.GetNetworkMetrics()
		if err != nil {
			errChan <- err
			return
		}
		stats.NetworkMetrics = networkMetrics
	}()

	go func() {
		defer wg.Done()
		systemStatus, err := h.statusService.GetSystemStatus()
		if err != nil {
			errChan <- err
			return
		}
		stats.SystemStatus = systemStatus
	}()

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		response.InternalError(c, "刷新仪表板数据失败")
		return
	}

	response.SuccessWithMessage(c, "仪表板数据已刷新", stats)
}

// WireGuardControlRequest WireGuard控制请求
type WireGuardControlRequest struct {
	Action    string `json:"action" binding:"required"`    // start, stop, restart
	Interface string `json:"interface" binding:"required"` // wg0, wg1等
}

// ConfigUploadRequest 配置上传请求
type ConfigUploadRequest struct {
	Interface  string `json:"interface" binding:"required"`   // wg0, wg1等
	ConfigData string `json:"config_data" binding:"required"` // 配置文件内容
}

// GetWireGuardInterfaces 获取WireGuard接口列表
func (h *DashboardHandler) GetWireGuardInterfaces(c *gin.Context) {
	interfaces, err := h.moduleService.GetWireGuardInterfaces()
	if err != nil {
		response.InternalError(c, "获取接口列表失败")
		return
	}
	response.Success(c, interfaces)
}

// ControlWireGuard 控制WireGuard接口
func (h *DashboardHandler) ControlWireGuard(c *gin.Context) {
	var req WireGuardControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	var err error
	switch req.Action {
	case "start":
		err = h.moduleService.StartWireGuardInterface(req.Interface)
	case "stop":
		err = h.moduleService.StopWireGuardInterface(req.Interface)
	case "restart":
		err = h.moduleService.RestartWireGuardInterface(req.Interface)
	default:
		response.BadRequest(c, "不支持的操作")
		return
	}

	if err != nil {
		response.InternalError(c, fmt.Sprintf("操作失败: %v", err))
		return
	}

	response.Success(c, gin.H{"message": "操作成功"})
}

// UploadWireGuardConfig 上传WireGuard配置
func (h *DashboardHandler) UploadWireGuardConfig(c *gin.Context) {
	var req ConfigUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 验证配置格式
	if err := h.moduleService.ValidateConfig(req.ConfigData); err != nil {
		response.BadRequest(c, fmt.Sprintf("配置格式错误: %v", err))
		return
	}

	// 更新配置
	if err := h.moduleService.UpdateWireGuardConfigWithInterface(req.Interface, req.ConfigData); err != nil {
		response.InternalError(c, fmt.Sprintf("更新配置失败: %v", err))
		return
	}

	response.Success(c, gin.H{"message": "配置更新成功"})
}

// GetWireGuardConfigFile 读取指定接口的WireGuard配置文件
func (h *DashboardHandler) GetWireGuardConfigFile(c *gin.Context) {
	interfaceName := c.Param("interface")

	// 验证接口名称格式
	if interfaceName == "" || len(interfaceName) < 2 || interfaceName[:2] != "wg" {
		response.BadRequest(c, "无效的接口名称")
		return
	}

	// 构建配置文件路径
	configPath := "/etc/wireguard/" + interfaceName + ".conf"

	// 读取配置文件内容
	configContent, err := h.moduleService.ReadWireGuardConfigFile(configPath)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("读取配置文件失败: %v", err))
		return
	}

	response.Success(c, gin.H{
		"interface":      interfaceName,
		"config_path":    configPath,
		"config_content": configContent,
		"file_size":      len(configContent),
		"last_modified":  time.Now(), // 这里可以添加文件修改时间
	})
}
