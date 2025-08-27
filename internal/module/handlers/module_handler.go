package handlers

import (
	"eitec-vpn/internal/module/services"
	"eitec-vpn/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// ModuleHandler 模块处理器 - 简化版本，只保留必要的基础功能
type ModuleHandler struct {
	moduleService *services.ModuleService
	statusService *services.StatusService
}

// NewModuleHandler 创建模块处理器
func NewModuleHandler(moduleService *services.ModuleService, statusService *services.StatusService) *ModuleHandler {
	return &ModuleHandler{
		moduleService: moduleService,
		statusService: statusService,
	}
}

// GetStatus 获取模块基础状态
func (h *ModuleHandler) GetStatus(c *gin.Context) {
	wgStatus, err := h.statusService.GetWireGuardStatus()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 获取模块配置信息
	moduleInfo := h.moduleService.GetModuleInfo()

	response.Success(c, gin.H{
		"wireguard": wgStatus,
		"module":    moduleInfo,
	})
}

// GetStats 获取基础流量统计
func (h *ModuleHandler) GetStats(c *gin.Context) {
	trafficStats, err := h.statusService.GetTrafficStats()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, trafficStats)
}

// ConfigureModule 配置模块 - 用于初始配置
func (h *ModuleHandler) ConfigureModule(c *gin.Context) {
	var req services.SetupInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if err := h.moduleService.ApplySetup(&req); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Module configured successfully", nil)
}

// StartWireGuard 启动WireGuard
func (h *ModuleHandler) StartWireGuard(c *gin.Context) {
	if err := h.moduleService.StartWireGuard(); err != nil {
		response.InternalError(c, "启动WireGuard失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "WireGuard已启动", nil)
}

// StopWireGuard 停止WireGuard
func (h *ModuleHandler) StopWireGuard(c *gin.Context) {
	if err := h.moduleService.StopWireGuard(); err != nil {
		response.InternalError(c, "停止WireGuard失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "WireGuard已停止", nil)
}

// RestartWireGuard 重启WireGuard
func (h *ModuleHandler) RestartWireGuard(c *gin.Context) {
	if err := h.moduleService.RestartWireGuard(); err != nil {
		response.InternalError(c, "重启WireGuard失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "WireGuard已重启", nil)
}
