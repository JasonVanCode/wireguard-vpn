package handlers

import (
	"strconv"
	"time"

	"eitec-vpn/internal/server/models"
	"eitec-vpn/internal/server/services"
	"eitec-vpn/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type InterfaceHandler struct {
	interfaceService *services.WireGuardInterfaceService
}

func NewInterfaceHandler() *InterfaceHandler {
	return &InterfaceHandler{
		interfaceService: services.NewWireGuardInterfaceService(),
	}
}

// GetInterfaces 获取所有WireGuard接口
func (h *InterfaceHandler) GetInterfaces(c *gin.Context) {
	interfaces, err := h.interfaceService.GetInterfaces()
	if err != nil {
		response.InternalError(c, "获取接口列表失败")
		return
	}

	response.Success(c, interfaces)
}

// GetInterface 获取单个WireGuard接口详情
func (h *InterfaceHandler) GetInterface(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的接口ID")
		return
	}

	wgInterface, err := h.interfaceService.GetInterface(uint(id))
	if err != nil {
		response.NotFound(c, "接口不存在")
		return
	}

	response.Success(c, wgInterface)
}

// GetInterfaceStats 获取接口统计信息
func (h *InterfaceHandler) GetInterfaceStats(c *gin.Context) {
	interfaces, err := h.interfaceService.GetInterfaces()
	if err != nil {
		response.InternalError(c, "获取接口统计失败")
		return
	}

	stats := map[string]interface{}{
		"total_interfaces":  len(interfaces),
		"active_interfaces": 0,
		"total_capacity":    0,
		"used_capacity":     0,
		"interfaces":        interfaces,
	}

	// 计算统计信息
	for _, iface := range interfaces {
		if iface.Status == models.InterfaceStatusUp {
			stats["active_interfaces"] = stats["active_interfaces"].(int) + 1
		}
		stats["total_capacity"] = stats["total_capacity"].(int) + iface.MaxPeers
		stats["used_capacity"] = stats["used_capacity"].(int) + iface.TotalPeers
	}

	response.Success(c, stats)
}

// StartInterface 启动接口
func (h *InterfaceHandler) StartInterface(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的接口ID")
		return
	}

	err = h.interfaceService.StartInterface(uint(id))
	if err != nil {
		response.InternalError(c, "启动接口失败")
		return
	}

	response.Success(c, gin.H{"message": "接口启动成功"})
}

// StopInterface 停止接口
func (h *InterfaceHandler) StopInterface(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的接口ID")
		return
	}

	err = h.interfaceService.StopInterface(uint(id))
	if err != nil {
		response.InternalError(c, "停止接口失败")
		return
	}

	response.Success(c, gin.H{"message": "接口停止成功"})
}

// CreateInterface 创建新的WireGuard接口
func (h *InterfaceHandler) CreateInterface(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Network     string `json:"network" binding:"required"`
		ListenPort  int    `json:"listen_port" binding:"required"`
		DNS         string `json:"dns"`
		MaxPeers    int    `json:"max_peers"`
		MTU         int    `json:"mtu"`
		PostUp      string `json:"post_up"`
		PostDown    string `json:"post_down"`
		AutoStart   bool   `json:"auto_start"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	// 创建接口数据
	interfaceTemplate := &models.InterfaceTemplate{
		Name:        req.Name,
		Description: req.Description,
		Network:     req.Network,
		ListenPort:  req.ListenPort,
		DNS:         req.DNS,
		MaxPeers:    req.MaxPeers,
		PostUp:      req.PostUp,
		PostDown:    req.PostDown,
	}

	// 设置默认值
	if interfaceTemplate.DNS == "" {
		interfaceTemplate.DNS = "8.8.8.8,8.8.4.4"
	}
	if interfaceTemplate.MaxPeers == 0 {
		interfaceTemplate.MaxPeers = 50
	}

	// 创建接口
	createdInterface, err := h.interfaceService.CreateInterface(interfaceTemplate)
	if err != nil {
		response.InternalError(c, "创建接口失败: "+err.Error())
		return
	}

	// 如果设置了自动启动，则启动接口
	if req.AutoStart {
		go func() {
			time.Sleep(1 * time.Second) // 稍等一下让接口创建完成
			if err := h.interfaceService.StartInterface(createdInterface.ID); err != nil {
				// 记录错误但不影响创建流程
				// TODO: 添加日志记录
			}
		}()
	}

	response.Success(c, createdInterface)
}

// DeleteInterface 删除WireGuard接口
func (h *InterfaceHandler) DeleteInterface(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的接口ID")
		return
	}

	err = h.interfaceService.DeleteInterface(uint(id))
	if err != nil {
		response.InternalError(c, "删除接口失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "接口删除成功"})
}

// GetInterfaceConfig 获取接口的配置文件内容
func (h *InterfaceHandler) GetInterfaceConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的接口ID")
		return
	}

	wgInterface, err := h.interfaceService.GetInterface(uint(id))
	if err != nil {
		response.NotFound(c, "接口不存在")
		return
	}

	// 生成实际的配置文件内容（包含所有模块和用户的Peer信息）
	configContent := h.interfaceService.GenerateInterfaceConfig(wgInterface)

	response.Success(c, gin.H{
		"config":    configContent,
		"interface": wgInterface,
	})
}
