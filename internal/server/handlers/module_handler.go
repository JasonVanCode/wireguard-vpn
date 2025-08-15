package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"eitec-vpn/internal/server/models"
	"eitec-vpn/internal/server/services"
	"eitec-vpn/internal/shared/utils"

	"github.com/gin-gonic/gin"
)

// ModuleHandler 模块管理处理器
type ModuleHandler struct {
	moduleService *services.ModuleService
}

// NewModuleHandler 创建模块处理器
func NewModuleHandler(moduleService *services.ModuleService) *ModuleHandler {
	return &ModuleHandler{
		moduleService: moduleService,
	}
}

// CreateModuleRequest 创建模块请求
type CreateModuleRequest struct {
	Name                string `json:"name" binding:"required,min=1,max=50"`
	Location            string `json:"location" binding:"required,min=1,max=100"`
	Description         string `json:"description,omitempty"`
	InterfaceID         uint   `json:"interface_id" binding:"required"`
	LocalIP             string `json:"local_ip,omitempty"`
	AllowedIPs          string `json:"allowed_ips,omitempty"`
	PersistentKeepalive int    `json:"persistent_keepalive,omitempty"`
	DNS                 string `json:"dns,omitempty"`
	AutoGenerateKeys    bool   `json:"auto_generate_keys,omitempty"`
	AutoAssignIP        bool   `json:"auto_assign_ip,omitempty"`
	ConfigTemplate      string `json:"config_template,omitempty"`
}

// UpdateModuleRequest 更新模块请求
type UpdateModuleRequest struct {
	Name         *string `json:"name,omitempty"`
	Location     *string `json:"location,omitempty"`
	AllowedIPs   *string `json:"allowed_ips,omitempty"`
	PersistentKA *int    `json:"persistent_keepalive,omitempty"`
}

// CreateModule 创建模块
func (mh *ModuleHandler) CreateModule(c *gin.Context) {
	var req CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	// 清理输入
	req.Name = utils.TrimSpaces(utils.Sanitize(req.Name))
	req.Location = utils.TrimSpaces(utils.Sanitize(req.Location))
	req.Description = utils.TrimSpaces(utils.Sanitize(req.Description))
	req.LocalIP = utils.TrimSpaces(utils.Sanitize(req.LocalIP))

	// 设置默认值
	if req.AllowedIPs == "" {
		req.AllowedIPs = "192.168.50.0/24" // 使用配置文档中的默认网段
	}
	if req.PersistentKeepalive == 0 {
		req.PersistentKeepalive = 25
	}
	if req.DNS == "" {
		req.DNS = "8.8.8.8,8.8.4.4"
	}
	if req.ConfigTemplate == "" {
		req.ConfigTemplate = "default"
	}
	// 默认启用自动生成密钥和分配IP
	if !req.AutoGenerateKeys {
		req.AutoGenerateKeys = true
	}
	if !req.AutoAssignIP {
		req.AutoAssignIP = true
	}

	// 验证配置参数
	if req.PersistentKeepalive < 0 || req.PersistentKeepalive > 300 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "持久化保活时间必须在0-300秒之间",
		})
		return
	}

	// 构建模块创建请求
	moduleData := &models.ModuleCreateRequest{
		Name:                req.Name,
		Location:            req.Location,
		Description:         req.Description,
		InterfaceID:         req.InterfaceID,
		AllowedIPs:          req.AllowedIPs,
		LocalIP:             req.LocalIP,
		PersistentKeepalive: req.PersistentKeepalive,
		DNS:                 req.DNS,
		AutoGenerateKeys:    req.AutoGenerateKeys,
		AutoAssignIP:        req.AutoAssignIP,
		ConfigTemplate:      req.ConfigTemplate,
	}

	module, err := mh.moduleService.CreateModule(moduleData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "模块创建成功，WireGuard配置已自动更新",
		"data":    module,
		"note":    "如果接口正在运行，配置文件已自动重新生成并应用",
	})
}

// GetModule 获取模块详情
func (mh *ModuleHandler) GetModule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "模块ID无效",
		})
		return
	}

	module, err := mh.moduleService.GetModule(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    module,
	})
}

// GetModules 获取模块列表
func (mh *ModuleHandler) GetModules(c *gin.Context) {
	// 解析分页参数
	page := 1
	size := 10
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			size = s
		}
	}

	// 解析过滤条件
	filters := make(map[string]interface{})
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}
	if location := c.Query("location"); location != "" {
		filters["location"] = location
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if ipAddress := c.Query("ip_address"); ipAddress != "" {
		filters["ip_address"] = ipAddress
	}

	modules, total, err := mh.moduleService.GetModules(page, size, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询模块列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    modules,
		"total":   total,
		"page":    page,
		"size":    size,
	})
}

// UpdateModule 更新模块
func (mh *ModuleHandler) UpdateModule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "模块ID无效",
		})
		return
	}

	var req UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Name != nil {
		name := utils.TrimSpaces(utils.Sanitize(*req.Name))
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "模块名称不能为空",
			})
			return
		}
		updates["name"] = name
	}
	if req.Location != nil {
		location := utils.TrimSpaces(utils.Sanitize(*req.Location))
		if location == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "模块位置不能为空",
			})
			return
		}
		updates["location"] = location
	}
	if req.AllowedIPs != nil {
		updates["allowed_ips"] = *req.AllowedIPs
	}
	if req.PersistentKA != nil {
		if *req.PersistentKA < 0 || *req.PersistentKA > 300 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "持久化保活时间必须在0-300之间",
			})
			return
		}
		updates["persistent_ka"] = *req.PersistentKA
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "没有需要更新的字段",
		})
		return
	}

	if err := mh.moduleService.UpdateModule(uint(id), updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "模块更新成功",
	})
}

// DeleteModule 删除模块
func (mh *ModuleHandler) DeleteModule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "模块ID无效",
		})
		return
	}

	if err := mh.moduleService.DeleteModule(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "模块删除成功",
	})
}

// GenerateModuleConfig 生成模块配置
func (mh *ModuleHandler) GenerateModuleConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "模块ID无效",
		})
		return
	}

	// 获取模块信息用于文件名
	module, err := mh.moduleService.GetModule(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "模块不存在: " + err.Error(),
		})
		return
	}

	config, err := mh.moduleService.GenerateModuleConfig(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成配置失败: " + err.Error(),
		})
		return
	}

	// 使用英文文件名，避免中文编码问题
	filename := fmt.Sprintf("module_%d.conf", module.ID)

	// 设置下载头部
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.String(http.StatusOK, config)
}

// GeneratePeerConfig 生成运维端配置
func (mh *ModuleHandler) GeneratePeerConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "模块ID无效",
		})
		return
	}

	config, err := mh.moduleService.GeneratePeerConfig(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": map[string]string{
			"config": config,
		},
	})
}

// RegenerateKeys 重新生成模块密钥
func (mh *ModuleHandler) RegenerateKeys(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "模块ID无效",
		})
		return
	}

	module, err := mh.moduleService.RegenerateModuleKeys(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "重新生成密钥失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密钥重新生成成功",
		"data":    module,
	})
}

// GetModuleStats 获取模块统计信息
func (mh *ModuleHandler) GetModuleStats(c *gin.Context) {
	stats, err := mh.moduleService.GetModuleStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取统计信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    stats,
	})
}

// SyncModuleStatus 同步模块状态
func (mh *ModuleHandler) SyncModuleStatus(c *gin.Context) {
	if err := mh.moduleService.SyncModuleStatus(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "同步模块状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "模块状态同步成功",
	})
}

// UpdateModuleStatusRequest 更新模块状态请求
type UpdateModuleStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateModuleStatus 更新模块状态
func (mh *ModuleHandler) UpdateModuleStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "模块ID无效",
		})
		return
	}

	var req UpdateModuleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	// 验证状态值
	var status models.ModuleStatus
	switch req.Status {
	case "online":
		status = models.ModuleStatusOnline
	case "offline":
		status = models.ModuleStatusOffline
	case "warning":
		status = models.ModuleStatusWarning
	case "unconfigured":
		status = models.ModuleStatusUnconfigured
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的状态值",
		})
		return
	}

	if err := mh.moduleService.UpdateModuleStatus(uint(id), status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "模块状态更新成功",
	})
}

// BatchDeleteModules 批量删除模块
func (mh *ModuleHandler) BatchDeleteModules(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	var errors []string
	successCount := 0

	for _, id := range req.IDs {
		if err := mh.moduleService.DeleteModule(id); err != nil {
			errors = append(errors, fmt.Sprintf("删除模块 %d 失败: %s", id, err.Error()))
		} else {
			successCount++
		}
	}

	result := gin.H{
		"code":          200,
		"message":       fmt.Sprintf("批量删除完成，成功: %d, 失败: %d", successCount, len(errors)),
		"success_count": successCount,
		"error_count":   len(errors),
	}

	if len(errors) > 0 {
		result["errors"] = errors
	}

	c.JSON(http.StatusOK, result)
}

// GetRecentlyActiveModules 获取最近活跃的模块
func (mh *ModuleHandler) GetRecentlyActiveModules(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	modules, err := mh.moduleService.GetRecentlyActiveModules(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取最近活跃模块失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    modules,
	})
}

// UpdateModuleHeartbeat 更新模块心跳
func (mh *ModuleHandler) UpdateModuleHeartbeat(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "模块ID无效",
		})
		return
	}

	// 更新模块最后握手时间
	if err := mh.moduleService.UpdateModuleHandshake(uint(id), time.Now()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新心跳失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "心跳更新成功",
	})
}

// UpdateModuleTraffic 更新模块流量
func (mh *ModuleHandler) UpdateModuleTraffic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "模块ID无效",
		})
		return
	}

	var req struct {
		RxBytes uint64 `json:"rx_bytes"`
		TxBytes uint64 `json:"tx_bytes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	// 更新模块流量统计
	if err := mh.moduleService.UpdateModuleTraffic(uint(id), req.RxBytes, req.TxBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新流量失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "流量更新成功",
	})
}
