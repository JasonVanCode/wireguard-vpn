package services

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"eitec-vpn/internal/shared/config"
)

// ModuleManager 模块管理器
type ModuleManager struct {
	config        *config.ModuleConfig
	server        *http.Server
	moduleService *ModuleService
	statusService *StatusService
	serverClient  *ServerClient
	wgManager     *WireGuardManager
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewModuleManager 创建模块管理器
func NewModuleManager(cfg *config.ModuleConfig) (*ModuleManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建各种服务
	moduleService := NewModuleService(cfg)
	statusService := NewStatusService(cfg)
	serverClient := NewServerClient(cfg)
	wgManager := NewWireGuardManager(cfg)

	manager := &ModuleManager{
		config:        cfg,
		moduleService: moduleService,
		statusService: statusService,
		serverClient:  serverClient,
		wgManager:     wgManager,
		ctx:           ctx,
		cancel:        cancel,
	}

	return manager, nil
}

// SetServer 设置HTTP服务器（由main.go调用）
func (mm *ModuleManager) SetServer(server *http.Server) {
	mm.server = server
}

// GetServices 获取服务实例（用于创建路由）
func (mm *ModuleManager) GetServices() (*ModuleService, *StatusService) {
	return mm.moduleService, mm.statusService
}

// Start 启动模块管理器
func (mm *ModuleManager) Start() error {
	log.Println("启动模块管理器...")

	// 1. 检查模块配置
	if err := mm.initModuleConfig(); err != nil {
		log.Printf("模块配置检查警告: %v", err)
	}

	// 2. 启动WireGuard（如果已配置）
	if mm.isConfigured() {
		if err := mm.wgManager.Start(); err != nil {
			log.Printf("启动WireGuard失败: %v", err)
		} else {
			log.Println("WireGuard启动成功")
		}
	} else {
		log.Println("模块未配置，请通过Web界面完成配置")
	}

	// 3. 启动HTTP服务器（如果已设置）
	if mm.server != nil {
		go func() {
			log.Printf("模块Web界面启动在端口 %s", mm.server.Addr)
			if err := mm.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTP服务器启动失败: %v", err)
			}
		}()
	}

	log.Println("模块管理器启动完成")
	return nil
}

// Stop 停止模块管理器
func (mm *ModuleManager) Stop() error {
	log.Println("停止模块管理器...")

	// 1. 停止HTTP服务器
	if mm.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := mm.server.Shutdown(ctx); err != nil {
			log.Printf("停止HTTP服务器失败: %v", err)
		}
	}

	// 2. 停止WireGuard（注释掉，避免系统退出时关闭VPN）
	// if err := mm.wgManager.Stop(); err != nil {
	// 	log.Printf("停止WireGuard失败: %v", err)
	// } else {
	// 	log.Println("WireGuard停止成功")
	// }

	log.Println("模块管理器已停止")
	return nil
}

// initModuleConfig 初始化模块配置
func (mm *ModuleManager) initModuleConfig() error {
	// 检查配置是否完整
	if mm.config.Module.ID == 0 || mm.config.Module.PrivateKey == "" || mm.config.Module.ServerEndpoint == "" {
		log.Println("模块配置不完整，请检查配置文件")
		return nil
	}

	log.Println("模块配置检查完成")
	return nil
}

// isConfigured 检查模块是否已配置
func (mm *ModuleManager) isConfigured() bool {
	// 检查必要的配置项
	if mm.config.Module.ID == 0 || mm.config.Module.PrivateKey == "" || mm.config.Module.ServerEndpoint == "" {
		return false
	}

	// 检查WireGuard配置文件是否存在
	configPath := DefaultWireGuardConfigPath
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false
	}

	return true
}
