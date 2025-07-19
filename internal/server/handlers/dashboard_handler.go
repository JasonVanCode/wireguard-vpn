package handlers

import (
	"net/http"
	"strconv"
	"time"

	"eitec-vpn/internal/server/services"
	"eitec-vpn/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// DashboardHandler 仪表盘处理器
type DashboardHandler struct {
	dashboardService *services.DashboardService
}

// NewDashboardHandler 创建仪表盘处理器
func NewDashboardHandler(dashboardService *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// GetDashboardStats 获取仪表盘统计数据
func (dh *DashboardHandler) GetDashboardStats(c *gin.Context) {
	stats, err := dh.dashboardService.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取仪表盘数据失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    stats,
	})
}

// GetSystemHealth 获取系统健康状态
func (dh *DashboardHandler) GetSystemHealth(c *gin.Context) {
	health, err := dh.dashboardService.GetSystemHealth()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取系统健康状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    health,
	})
}

// GetModuleRanking 获取模块流量排行
func (dh *DashboardHandler) GetModuleRanking(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	ranking, err := dh.dashboardService.GetModuleRanking(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取模块排行失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    ranking,
	})
}

// GetTrafficData 获取流量数据
func (h *DashboardHandler) GetTrafficData(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "1h")

	// 获取WireGuard接口服务
	interfaceService := services.NewWireGuardInterfaceService()

	// 获取所有接口的流量统计
	allStats, err := interfaceService.GetAllInterfacesTrafficStats()
	if err != nil {
		response.InternalError(c, "获取流量统计失败: "+err.Error())
		return
	}

	// 生成实时流量数据（这里可以实现更复杂的历史数据逻辑）
	timeLabels := generateTimeLabels(timeRange)
	uploadData := make([]float64, len(timeLabels))
	downloadData := make([]float64, len(timeLabels))

	// 如果有流量数据，模拟流量变化（实际应用中可以存储历史数据）
	if len(allStats) > 0 {
		var totalTx, totalRx uint64
		for _, stat := range allStats {
			totalTx += stat.TotalTx
			totalRx += stat.TotalRx
		}

		// 基于总流量生成模拟的时间序列数据
		for i := range timeLabels {
			// 简单的模拟逻辑：基于当前总流量生成变化
			uploadData[i] = float64(totalTx) / (1024 * 1024) / float64(len(timeLabels)) * (0.8 + 0.4*float64(i%3))
			downloadData[i] = float64(totalRx) / (1024 * 1024) / float64(len(timeLabels)) * (0.9 + 0.2*float64(i%2))
		}
	}

	result := map[string]interface{}{
		"time_labels":   timeLabels,
		"upload_data":   uploadData,
		"download_data": downloadData,
		"total_stats":   allStats,
	}

	response.Success(c, result)
}

// generateTimeLabels 生成时间标签
func generateTimeLabels(timeRange string) []string {
	var labels []string
	var count int
	var interval time.Duration

	switch timeRange {
	case "1h":
		count = 12
		interval = 5 * time.Minute
	case "6h":
		count = 12
		interval = 30 * time.Minute
	case "24h":
		count = 12
		interval = 2 * time.Hour
	default:
		count = 12
		interval = 5 * time.Minute
	}

	now := time.Now()
	for i := count - 1; i >= 0; i-- {
		t := now.Add(-time.Duration(i) * interval)
		labels = append(labels, t.Format("15:04"))
	}

	return labels
}
