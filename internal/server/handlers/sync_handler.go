package handlers

import (
	"eitec-vpn/internal/server/services"
	"eitec-vpn/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// SyncHandler 同步处理器
type SyncHandler struct {
	syncService   *services.WireGuardSyncService
	cronScheduler *services.CronScheduler
}

// NewSyncHandler 创建同步处理器
func NewSyncHandler(cronScheduler *services.CronScheduler) *SyncHandler {
	return &SyncHandler{
		syncService:   services.NewWireGuardSyncService(),
		cronScheduler: cronScheduler,
	}
}

// GetInterfaceRealTimeStats 获取接口实时统计
func (h *SyncHandler) GetInterfaceRealTimeStats(c *gin.Context) {
	interfaceName := c.Param("name")
	if interfaceName == "" {
		response.BadRequest(c, "接口名称不能为空")
		return
	}

	stats, err := h.syncService.GetInterfaceRealTimeStats(interfaceName)
	if err != nil {
		response.InternalError(c, "获取接口实时统计失败: "+err.Error())
		return
	}

	response.Success(c, stats)
}

// GetCronSchedulerStats 获取定时任务调度器状态
func (h *SyncHandler) GetCronSchedulerStats(c *gin.Context) {
	if h.cronScheduler == nil {
		response.InternalError(c, "定时任务调度器未初始化")
		return
	}

	stats := h.cronScheduler.GetStats()
	response.Success(c, stats)
}

// GetRunningJobs 获取正在运行的定时任务
func (h *SyncHandler) GetRunningJobs(c *gin.Context) {
	if h.cronScheduler == nil {
		response.InternalError(c, "定时任务调度器未初始化")
		return
	}

	jobs := h.cronScheduler.GetRunningJobs()

	var jobList []map[string]interface{}
	for i, job := range jobs {
		jobList = append(jobList, map[string]interface{}{
			"id":       i + 1,
			"next_run": job.Next,
			"prev_run": job.Prev,
		})
	}

	response.Success(c, gin.H{
		"total_jobs": len(jobs),
		"jobs":       jobList,
	})
}
