package handlers

import (
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"eitec-vpn/internal/server/database"
	"eitec-vpn/internal/server/models"
	"eitec-vpn/internal/server/services"
	"eitec-vpn/internal/shared/config"
	"eitec-vpn/internal/shared/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DashboardHandler 仪表盘处理器
type DashboardHandler struct {
	dashboardService *services.DashboardService
	db               *gorm.DB
}

// NewDashboardHandler 创建仪表盘处理器
func NewDashboardHandler(dashboardService *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
		db:               database.DB,
	}
}

// GetDashboardStats 获取仪表盘统计数据（用于统计卡片）
func (dh *DashboardHandler) GetDashboardStats(c *gin.Context) {
	// 使用实时 WireGuard 状态获取模块统计
	moduleStats, err := dh.getModuleStatsFromWireGuard()
	if err != nil {
		// 如果获取实时状态失败，降级到数据库统计
		moduleStats, err = dh.getModuleStatsFromDB()
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "获取模块统计数据失败: "+err.Error())
			return
		}
	}

	// 获取系统资源使用率
	systemResources, err := dh.dashboardService.GetSystemHealth()
	if err != nil {
		// 如果获取系统资源失败，使用默认值
		stats := map[string]interface{}{
			"module_stats": map[string]interface{}{
				"online": moduleStats.OnlineModules,
				"total":  moduleStats.TotalModules,
			},
			"user_stats": map[string]interface{}{
				"online": moduleStats.OnlineUsers,
				"total":  moduleStats.TotalUsers,
			},
			"system_resources": map[string]interface{}{
				"cpu_usage":    0.0,
				"memory_usage": 0.0,
				"disk_usage":   0.0,
			},
			"service_status": map[string]interface{}{
				"wireguard_status": "unknown",
				"database_status":  "unknown",
				"api_status":       "unknown",
			},
		}
		response.Success(c, stats)
		return
	}

	// 从系统资源中提取CPU、内存、磁盘使用率
	var cpuUsage, memoryUsage, diskUsage float64
	var wireguardStatus, databaseStatus, apiStatus string

	if systemResources != nil {
		// 从SystemHealthInfo结构体中提取数据
		if systemResources.SystemRes != nil {
			cpuUsage = systemResources.SystemRes.CPUUsage
			memoryUsage = systemResources.SystemRes.MemoryUsage
			diskUsage = systemResources.SystemRes.DiskUsage
		}

		// 从健康检查中提取服务状态
		if systemResources.Checks != nil {
			if wgCheck, exists := systemResources.Checks["wireguard"]; exists {
				wireguardStatus = wgCheck.Status
			}
			if dbCheck, exists := systemResources.Checks["database"]; exists {
				databaseStatus = dbCheck.Status
			}
			if apiCheck, exists := systemResources.Checks["api"]; exists {
				apiStatus = apiCheck.Status
			}
		}
	}

	// 构建完整的统计数据
	stats := map[string]interface{}{
		"module_stats": map[string]interface{}{
			"online": moduleStats.OnlineModules,
			"total":  moduleStats.TotalModules,
		},
		"user_stats": map[string]interface{}{
			"online": moduleStats.OnlineUsers,
			"total":  moduleStats.TotalUsers,
		},
		"system_resources": map[string]interface{}{
			"cpu_usage":    cpuUsage,
			"memory_usage": memoryUsage,
			"disk_usage":   diskUsage,
		},
		"service_status": map[string]interface{}{
			"wireguard_status": wireguardStatus,
			"database_status":  databaseStatus,
			"api_status":       apiStatus,
		},
	}

	response.Success(c, stats)
}

// ModuleStats 模块统计信息
type ModuleStats struct {
	OnlineModules int64 `json:"online_modules"`
	TotalModules  int64 `json:"total_modules"`
	OnlineUsers   int64 `json:"online_users"`
	TotalUsers    int64 `json:"total_users"`
}

// getModuleStatsFromDB 从数据库获取模块统计信息
func (dh *DashboardHandler) getModuleStatsFromDB() (*ModuleStats, error) {
	stats := &ModuleStats{}

	// 统计总模块数
	if err := dh.db.Model(&models.Module{}).Count(&stats.TotalModules).Error; err != nil {
		return nil, err
	}

	// 使用统一的超时常量统计在线模块数
	onlineTime := time.Now().Add(-config.WireGuardOnlineTimeout)
	if err := dh.db.Model(&models.Module{}).
		Where("latest_handshake > ?", onlineTime).
		Count(&stats.OnlineModules).Error; err != nil {
		return nil, err
	}

	// 用户统计（这里可以根据实际需求调整）
	stats.TotalUsers = 0
	stats.OnlineUsers = 0

	return stats, nil
}

// getModuleStatsFromWireGuard 从实时 WireGuard 状态获取模块统计信息
func (dh *DashboardHandler) getModuleStatsFromWireGuard() (*ModuleStats, error) {
	stats := &ModuleStats{}

	// 创建WireGuard接口服务
	interfaceService := services.NewWireGuardInterfaceService()

	// 获取有状态的接口列表
	interfaces, err := interfaceService.GetInterfacesWithStatus()
	if err != nil {
		return nil, err
	}

	// 统计模块数据
	for _, iface := range interfaces {
		if len(iface.Modules) > 0 {
			for _, module := range iface.Modules {
				stats.TotalModules++

				// 使用实时在线状态
				if module.IsOnline {
					stats.OnlineModules++
				}

				// 统计用户数据
				stats.TotalUsers += int64(len(module.Users))
				for _, user := range module.Users {
					if user.IsActive {
						stats.OnlineUsers++
					}
				}
			}
		}
	}

	return stats, nil
}

// GetSystemHealth 获取系统健康状态（用于统计卡片）
func (dh *DashboardHandler) GetSystemHealth(c *gin.Context) {
	health, err := dh.dashboardService.GetSystemHealth()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取系统健康状态失败: "+err.Error())
		return
	}

	response.Success(c, health)
}

// GetNetworkInterfaces 获取系统网络接口列表（用于创建接口时的网络接口选择）
func (dh *DashboardHandler) GetNetworkInterfaces(c *gin.Context) {
	interfaces, err := getSystemNetworkInterfaces()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取网络接口失败: "+err.Error())
		return
	}

	response.Success(c, interfaces)
}

// GetWireGuardInterfacesWithStatus 获取WireGuard接口列表（包含实时状态）
func (dh *DashboardHandler) GetWireGuardInterfacesWithStatus(c *gin.Context) {
	// 创建WireGuard接口服务
	interfaceService := services.NewWireGuardInterfaceService()

	// 获取有状态的接口列表
	interfaces, err := interfaceService.GetInterfacesWithStatus()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取WireGuard接口状态失败: "+err.Error())
		return
	}

	response.Success(c, interfaces)
}

// 删除了额外的WireGuard状态API，保持简单，使用现有的两个API即可

// NetworkInterface 网络接口信息
type NetworkInterface struct {
	Name      string `json:"name"`       // 接口名称
	IP        string `json:"ip"`         // IP地址
	IsDefault bool   `json:"is_default"` // 是否为默认路由接口
	IsUp      bool   `json:"is_up"`      // 是否启动
}

// getSystemNetworkInterfaces 获取系统网络接口
func getSystemNetworkInterfaces() ([]NetworkInterface, error) {
	var result []NetworkInterface

	// 获取默认路由接口
	defaultInterface := getDefaultRouteInterface()

	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		// 跳过回环接口和虚拟接口
		if iface.Flags&net.FlagLoopback != 0 ||
			strings.HasPrefix(iface.Name, "lo") ||
			strings.HasPrefix(iface.Name, "docker") ||
			strings.HasPrefix(iface.Name, "br-") ||
			strings.HasPrefix(iface.Name, "veth") ||
			strings.HasPrefix(iface.Name, "wg") {
			continue
		}

		// 获取接口的IP地址
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var ip string
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ip = ipnet.IP.String()
					break
				}
			}
		}

		// 只添加有IP地址的接口
		if ip != "" {
			result = append(result, NetworkInterface{
				Name:      iface.Name,
				IP:        ip,
				IsDefault: iface.Name == defaultInterface,
				IsUp:      iface.Flags&net.FlagUp != 0,
			})
		}
	}

	return result, nil
}

// getDefaultRouteInterface 获取默认路由的网络接口
func getDefaultRouteInterface() string {
	// 在Linux/macOS上使用ip route或route命令
	cmd := exec.Command("sh", "-c", "ip route show default 2>/dev/null | awk '{print $5}' | head -1 || route get default 2>/dev/null | grep interface | awk '{print $2}'")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}
