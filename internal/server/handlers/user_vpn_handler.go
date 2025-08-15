package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"eitec-vpn/internal/server/models"
	"eitec-vpn/internal/server/services"

	"github.com/gin-gonic/gin"
)

// UserVPNHandler 用户VPN处理器
type UserVPNHandler struct {
	userVPNService *services.UserVPNService
}

// NewUserVPNHandler 创建用户VPN处理器
func NewUserVPNHandler() *UserVPNHandler {
	return &UserVPNHandler{
		userVPNService: services.NewUserVPNService(),
	}
}

// CreateUserVPN 创建用户VPN
func (h *UserVPNHandler) CreateUserVPN(c *gin.Context) {
	var config models.UserVPNConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数", "message": err.Error()})
		return
	}

	userVPN, err := h.userVPNService.CreateUserVPN(&config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户VPN失败", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "用户VPN创建成功",
		"data":    userVPN,
	})
}

// GetUserVPNsByModule 获取模块的用户VPN列表
func (h *UserVPNHandler) GetUserVPNsByModule(c *gin.Context) {
	moduleIDStr := c.Param("id") // 从 moduleId 改为 id
	moduleID, err := strconv.ParseUint(moduleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模块ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	userVPNs, total, err := h.userVPNService.GetUserVPNsByModule(uint(moduleID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户VPN列表失败", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": userVPNs,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetUserVPN 获取单个用户VPN信息
func (h *UserVPNHandler) GetUserVPN(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户VPN ID"})
		return
	}

	userVPN, err := h.userVPNService.GetUserVPN(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户VPN不存在", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": userVPN})
}

// UpdateUserVPN 更新用户VPN信息
func (h *UserVPNHandler) UpdateUserVPN(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户VPN ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数", "message": err.Error()})
		return
	}

	if err := h.userVPNService.UpdateUserVPN(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户VPN失败", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户VPN更新成功"})
}

// DeleteUserVPN 删除用户VPN
func (h *UserVPNHandler) DeleteUserVPN(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户VPN ID"})
		return
	}

	if err := h.userVPNService.DeleteUserVPN(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户VPN失败", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户VPN删除成功"})
}

// GenerateUserVPNConfig 生成用户VPN配置文件
func (h *UserVPNHandler) GenerateUserVPNConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户VPN ID"})
		return
	}

	config, err := h.userVPNService.GenerateUserVPNConfig(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成配置文件失败", "message": err.Error()})
		return
	}

	// 获取用户信息用于文件名
	userVPN, err := h.userVPNService.GetUserVPN(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败", "message": err.Error()})
		return
	}

	// 使用英文文件名，避免中文编码问题
	filename := fmt.Sprintf("user_%d_vpn_config.conf", userVPN.ID)

	// 设置响应头
	c.Header("Content-Type", "text/plain")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.String(http.StatusOK, config)
}

// GetUserVPNStats 获取模块的用户VPN统计
func (h *UserVPNHandler) GetUserVPNStats(c *gin.Context) {
	moduleIDStr := c.Param("id") // 从 moduleId 改为 id
	moduleID, err := strconv.ParseUint(moduleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模块ID"})
		return
	}

	stats, err := h.userVPNService.GetUserVPNStats(uint(moduleID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}
