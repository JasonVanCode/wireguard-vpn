package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"eitec-vpn/internal/module/database"
	"eitec-vpn/internal/module/models"
	"eitec-vpn/internal/shared/config"

	"gorm.io/gorm"
)

// ModuleManager 模块管理器
type ModuleManager struct {
	config          *config.ModuleConfig
	db              *gorm.DB
	server          *http.Server
	moduleService   *ModuleService
	statusService   *StatusService
	serverClient    *ServerClient
	wgManager       *WireGuardManager
	backgroundTasks []BackgroundTask
	ctx             context.Context
	cancel          context.CancelFunc
}

// BackgroundTask 后台任务接口
type BackgroundTask interface {
	Start(ctx context.Context) error
	Stop() error
	Name() string
}

// NewModuleManager 创建模块管理器
func NewModuleManager(cfg *config.ModuleConfig) (*ModuleManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 获取模块端数据库连接
	db := database.DB
	if db == nil {
		return nil, fmt.Errorf("模块端数据库未初始化")
	}

	// 创建各种服务
	moduleService := NewModuleService(cfg)
	statusService := NewStatusService(cfg)
	serverClient := NewServerClient(cfg)
	wgManager := NewWireGuardManager(cfg)

	manager := &ModuleManager{
		config:        cfg,
		db:            db,
		moduleService: moduleService,
		statusService: statusService,
		serverClient:  serverClient,
		wgManager:     wgManager,
		ctx:           ctx,
		cancel:        cancel,
	}

	// 初始化后台任务
	manager.initBackgroundTasks()

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

	// 1. 初始化模块配置到数据库
	if err := mm.initModuleConfig(); err != nil {
		return fmt.Errorf("初始化模块配置失败: %w", err)
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

	// 4. 启动后台任务
	for _, task := range mm.backgroundTasks {
		go func(t BackgroundTask) {
			if err := t.Start(mm.ctx); err != nil {
				log.Printf("启动后台任务 %s 失败: %v", t.Name(), err)
			} else {
				log.Printf("后台任务 %s 启动成功", t.Name())
			}
		}(task)
	}

	log.Println("模块管理器启动完成")
	return nil
}

// Stop 停止模块管理器
func (mm *ModuleManager) Stop() error {
	log.Println("停止模块管理器...")

	// 1. 取消后台任务
	mm.cancel()
	for _, task := range mm.backgroundTasks {
		if err := task.Stop(); err != nil {
			log.Printf("停止后台任务 %s 失败: %v", task.Name(), err)
		}
	}

	// 2. 停止HTTP服务器
	if mm.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := mm.server.Shutdown(ctx); err != nil {
			log.Printf("停止HTTP服务器失败: %v", err)
		}
	}

	// 3. 停止WireGuard
	if err := mm.wgManager.Stop(); err != nil {
		log.Printf("停止WireGuard失败: %v", err)
	} else {
		log.Println("WireGuard停止成功")
	}

	log.Println("模块管理器已停止")
	return nil
}

// initModuleConfig 初始化模块配置到数据库
func (mm *ModuleManager) initModuleConfig() error {
	// 检查数据库中是否已有模块配置
	var moduleCount int64
	if err := mm.db.Model(&models.LocalModule{}).Count(&moduleCount).Error; err != nil {
		return fmt.Errorf("检查模块配置失败: %w", err)
	}

	// 如果数据库中没有配置且YAML中有配置，则初始化
	if moduleCount == 0 && mm.config.Module.ID != 0 {
		module := &models.LocalModule{
			ServerID:           mm.config.Module.ID,
			Name:               mm.config.Module.Name,
			Location:           mm.config.Module.Location,
			PrivateKey:         mm.config.Module.PrivateKey,
			ServerURL:          mm.config.Server.URL,
			ServerEndpoint:     mm.config.Module.ServerEndpoint,
			APIKey:             mm.config.Module.APIKey,
			WireGuardInterface: mm.config.WireGuard.Interface,
			Status:             models.ModuleStatusUnconfigured,
			IsConfigured:       false,
		}

		// 从私钥生成公钥
		if publicKey, err := mm.wgManager.GetPublicKeyFromPrivate(mm.config.Module.PrivateKey); err == nil {
			module.PublicKey = publicKey
		}

		if err := mm.db.Create(module).Error; err != nil {
			return fmt.Errorf("创建模块配置失败: %w", err)
		}
		log.Println("模块配置已初始化到数据库")
	}

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

// initBackgroundTasks 初始化后台任务
func (mm *ModuleManager) initBackgroundTasks() {
	// 心跳任务
	heartbeatTask := &HeartbeatTask{
		client:   mm.serverClient,
		interval: time.Duration(mm.config.Server.HeartbeatInterval) * time.Second,
	}
	mm.backgroundTasks = append(mm.backgroundTasks, heartbeatTask)

	// 流量上报任务
	trafficTask := &TrafficReportTask{
		name:          "流量上报任务",
		client:        mm.serverClient,
		statusService: mm.statusService,
	}
	mm.backgroundTasks = append(mm.backgroundTasks, trafficTask)

	// 配置同步任务
	syncTask := &ConfigSyncTask{
		client:    mm.serverClient,
		wgManager: mm.wgManager,
		interval:  time.Duration(mm.config.Server.SyncInterval) * time.Second,
	}
	mm.backgroundTasks = append(mm.backgroundTasks, syncTask)

	// 数据清理任务（每天执行一次）
	cleanupTask := &DataCleanupTask{
		interval: 24 * time.Hour,
	}
	mm.backgroundTasks = append(mm.backgroundTasks, cleanupTask)
}

// HeartbeatTask 心跳任务
type HeartbeatTask struct {
	client   *ServerClient
	interval time.Duration
}

func (h *HeartbeatTask) Name() string { return "心跳任务" }

func (h *HeartbeatTask) Start(ctx context.Context) error {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := h.client.SendHeartbeat(); err != nil {
				log.Printf("发送心跳失败: %v", err)
			}
		}
	}
}

func (h *HeartbeatTask) Stop() error { return nil }

// TrafficReportTask 流量上报任务
type TrafficReportTask struct {
	name          string
	client        *ServerClient
	statusService *StatusService
}

func (t *TrafficReportTask) Name() string {
	return t.name
}

func (t *TrafficReportTask) Start(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// 获取流量统计
			stats, err := t.statusService.GetTrafficStats()
			if err != nil {
				log.Printf("获取流量统计失败: %v", err)
				continue
			}

			// 记录流量统计
			if stats != nil {
				log.Printf("流量统计: 上传 %d bytes, 下载 %d bytes", stats.TxBytes, stats.RxBytes)
				if err := t.client.ReportTraffic(*stats); err != nil {
					log.Printf("上报流量统计失败: %v", err)
				}
			}

			// 等待下一次执行
			time.Sleep(30 * time.Second)
		}
	}
}

func (t *TrafficReportTask) Stop() error { return nil }

// ConfigSyncTask 配置同步任务
type ConfigSyncTask struct {
	client    *ServerClient
	wgManager *WireGuardManager
	interval  time.Duration
}

func (c *ConfigSyncTask) Name() string { return "配置同步任务" }

func (c *ConfigSyncTask) Start(ctx context.Context) error {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			configContent, err := c.client.GetConfiguration()
			if err != nil {
				log.Printf("获取配置失败: %v", err)
				continue
			}

			if err := c.wgManager.UpdateConfig(configContent); err != nil {
				log.Printf("更新配置失败: %v", err)
			} else {
				log.Println("配置同步成功")
			}
		}
	}
}

func (c *ConfigSyncTask) Stop() error { return nil }

// DataCleanupTask 数据清理任务
type DataCleanupTask struct {
	interval time.Duration
}

func (d *DataCleanupTask) Name() string { return "数据清理任务" }

func (d *DataCleanupTask) Start(ctx context.Context) error {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// 清理旧数据
			log.Println("开始清理旧数据...")
			// 这里可以添加具体的清理逻辑
			log.Println("数据清理完成")
		}
	}
}

func (d *DataCleanupTask) Stop() error { return nil }
