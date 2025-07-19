package handlers

import (
	"net/http"

	"eitec-vpn/internal/server/database"
	"eitec-vpn/internal/server/services"

	"github.com/gin-gonic/gin"
)

// ConfigHandler 配置管理处理器
type ConfigHandler struct {
	configService *services.ConfigService
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(configService *services.ConfigService) *ConfigHandler {
	return &ConfigHandler{
		configService: configService,
	}
}

// GetSystemConfig 获取系统配置
func (ch *ConfigHandler) GetSystemConfig(c *gin.Context) {
	// 简化：移除敏感信息检查，统一返回所有配置
	config, err := ch.configService.GetSystemConfig(true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取系统配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    config,
	})
}

// UpdateSystemConfig 更新系统配置
func (ch *ConfigHandler) UpdateSystemConfig(c *gin.Context) {
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	if err := ch.configService.UpdateSystemConfig(updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "系统配置更新成功",
	})
}

// ResetToDefaults 重置为默认配置
func (ch *ConfigHandler) ResetToDefaults(c *gin.Context) {
	if err := ch.configService.ResetToDefaults(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "重置配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "配置重置成功",
	})
}

// ExportConfig 导出配置
func (ch *ConfigHandler) ExportConfig(c *gin.Context) {
	config, err := ch.configService.ExportConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "导出配置失败: " + err.Error(),
		})
		return
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=eitec-vpn-config.json")
	c.JSON(http.StatusOK, config)
}

// ImportConfig 导入配置
func (ch *ConfigHandler) ImportConfig(c *gin.Context) {
	var configMap map[string]string
	if err := c.ShouldBindJSON(&configMap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	if err := ch.configService.ImportConfig(configMap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "配置导入成功",
	})
}

// InitializeWireGuard 初始化WireGuard配置
func (ch *ConfigHandler) InitializeWireGuard(c *gin.Context) {
	if err := ch.configService.InitializeWireGuardServer(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "WireGuard服务器初始化成功",
	})
}

// ApplyWireGuardConfig 应用WireGuard配置
func (ch *ConfigHandler) ApplyWireGuardConfig(c *gin.Context) {
	if err := ch.configService.ApplyWireGuardConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "应用WireGuard配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "WireGuard配置应用成功",
	})
}

// GenerateServerConfig 生成服务器配置文件
func (ch *ConfigHandler) GenerateServerConfig(c *gin.Context) {
	config, err := ch.configService.GenerateServerConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成服务器配置失败: " + err.Error(),
		})
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=wg0.conf")
	c.String(http.StatusOK, config)
}

// ValidateNetworkSettings 验证网络设置
func (ch *ConfigHandler) ValidateNetworkSettings(c *gin.Context) {
	var req struct {
		Network string `json:"network" binding:"required"`
		IPStart string `json:"ip_start" binding:"required"`
		IPEnd   string `json:"ip_end" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	if err := ch.configService.ValidateNetworkSettings(req.Network, req.IPStart, req.IPEnd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "网络设置验证通过",
	})
}

// GetConfigStatus 获取配置状态
func (ch *ConfigHandler) GetConfigStatus(c *gin.Context) {
	// 检查WireGuard是否已初始化
	serverPublicKey, err := database.GetSystemConfig("server.public_key")
	isWireGuardConfigured := err == nil && serverPublicKey != ""

	// 检查服务器端点是否已配置
	serverEndpoint, err := database.GetSystemConfig("server.endpoint")
	isEndpointConfigured := err == nil && serverEndpoint != ""

	// 获取基本配置信息
	config, _ := ch.configService.GetSystemConfig(false)

	status := gin.H{
		"wireguard_configured": isWireGuardConfigured,
		"endpoint_configured":  isEndpointConfigured,
		"fully_configured":     isWireGuardConfigured && isEndpointConfigured,
		"config":               config,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    status,
	})
}
