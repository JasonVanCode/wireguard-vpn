package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"eitec-vpn/internal/server/database"
	"eitec-vpn/internal/server/routes"
	"eitec-vpn/internal/server/services"
	"eitec-vpn/internal/shared/auth"
	"eitec-vpn/internal/shared/config"
	"eitec-vpn/internal/shared/utils"

	"github.com/gin-gonic/gin"
)

var (
	configFile  = flag.String("config", "configs/server.yaml", "配置文件路径")
	versionFlag = flag.Bool("version", false, "显示版本信息")
	help        = flag.Bool("help", false, "显示帮助信息")
	initDB      = flag.Bool("init", false, "初始化数据库和默认数据")
)

// 这些变量可以在构建时通过-ldflags设置
var (
	version   string = "2.0.0"
	buildTime string = "2024-01-01"
)

const (
	AppName = "EITEC VPN Server"
)

func init() {
	// 解析命令行参数
	flag.Parse()

	// 显示版本信息
	if *versionFlag {
		log.Printf("%s v%s (built at %s)", AppName, version, buildTime)
		os.Exit(0)
	}

	// 显示帮助信息
	if *help {
		flag.Usage()
		os.Exit(0)
	}
}

func main() {
	log.Printf("启动 %s v%s", AppName, version)

	// 加载配置
	cfg, err := config.LoadServerConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置全局配置
	config.SetGlobalServerConfig(cfg)

	// 设置Gin模式
	gin.SetMode(cfg.App.Mode)

	// 处理数据库路径 - 转换为绝对路径
	dbPath, err := utils.GetAbsolutePath(cfg.Database.Path)
	if err != nil {
		log.Fatalf("获取数据库路径失败: %v", err)
	}

	// 确保数据库目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("创建数据库目录失败: %v", err)
	}

	log.Printf("数据库路径: %s", dbPath)

	// 初始化数据库
	if *initDB {
		// 强制初始化模式
		if err := database.InitDatabaseWithOptions(dbPath, true); err != nil {
			log.Fatalf("初始化数据库失败: %v", err)
		}
		log.Println("数据库强制初始化完成")
		return // 初始化完成后退出
	} else {
		// 正常启动模式
		if err := database.InitDatabase(dbPath); err != nil {
			log.Fatalf("初始化数据库失败: %v", err)
		}
	}
	log.Println("数据库初始化成功")

	// 创建服务层
	moduleService := services.NewModuleService()
	dashboardService := services.NewDashboardService(moduleService)
	configService := services.NewConfigService()

	// 创建认证服务
	jwtService := auth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.RefreshSecret, cfg.Auth.AccessExpiry, cfg.Auth.RefreshExpiry)
	userService := auth.NewUserService()
	sessionManager := auth.NewSessionManager(cfg.Auth.SessionTimeout)

	// 启动基于Cron的WireGuard状态同步调度器
	cronScheduler := services.NewCronScheduler()
	if err := cronScheduler.Start(); err != nil {
		log.Fatalf("启动定时任务调度器失败: %v", err)
	}

	// 设置路由
	var router *gin.Engine
	if cfg.App.Mode == "api" {
		// 仅API模式
		router = routes.SetupAPIRoutes(moduleService, dashboardService, configService, userService, jwtService, sessionManager, cronScheduler)
		log.Println("运行模式: API Only")
	} else {
		// 完整模式 (API + Web界面)
		router = routes.SetupRoutes(moduleService, dashboardService, configService, userService, jwtService, sessionManager, cronScheduler)
		log.Println("运行模式: Full Stack")
	}

	// 创建HTTP服务器
	server := &http.Server{
		Addr:           cfg.App.Listen,
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.App.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.App.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(cfg.App.IdleTimeout) * time.Second,
		MaxHeaderBytes: cfg.App.MaxHeaderBytes << 20, // MB to bytes
	}

	// 启动HTTP服务器
	go func() {
		log.Printf("HTTP服务器启动在 %s", cfg.App.Listen)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP服务器启动失败: %v", err)
		}
	}()

	// 如果启用HTTPS
	if cfg.App.TLS.Enabled {
		httpsServer := &http.Server{
			Addr:           cfg.App.TLS.Listen,
			Handler:        router,
			ReadTimeout:    time.Duration(cfg.App.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(cfg.App.WriteTimeout) * time.Second,
			IdleTimeout:    time.Duration(cfg.App.IdleTimeout) * time.Second,
			MaxHeaderBytes: cfg.App.MaxHeaderBytes << 20,
		}

		go func() {
			log.Printf("HTTPS服务器启动在 %s", cfg.App.TLS.Listen)
			if err := httpsServer.ListenAndServeTLS(cfg.App.TLS.CertFile, cfg.App.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTPS服务器启动失败: %v", err)
			}
		}()
	}

	// 启动后台任务
	startBackgroundTasks(moduleService, cfg)

	// 在服务关闭时停止调度器
	defer cronScheduler.Stop()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	// 优雅关闭
	gracefulShutdown(server)
}

// startBackgroundTasks 启动后台任务
func startBackgroundTasks(moduleService *services.ModuleService, cfg *config.ServerConfig) {
	// 启动模块状态同步任务
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.WireGuard.SyncInterval) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := moduleService.SyncModuleStatus(); err != nil {
				log.Printf("同步模块状态失败: %v", err)
			}
		}
	}()

	// 启动会话清理任务
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
		defer ticker.Stop()

		for range ticker.C {
			// TODO: 实现会话清理逻辑
			log.Println("执行会话清理任务")
		}
	}()

	// 启动日志轮转任务
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // 每天执行一次
		defer ticker.Stop()

		for range ticker.C {
			// TODO: 实现日志轮转逻辑
			log.Println("执行日志轮转任务")
		}
	}()

	log.Println("后台任务已启动")
}

// gracefulShutdown 优雅关闭服务器
func gracefulShutdown(server *http.Server) {
	// 关闭数据库连接
	if err := database.Close(); err != nil {
		log.Printf("关闭数据库连接失败: %v", err)
	} else {
		log.Println("数据库连接已关闭")
	}

	log.Println("服务器已关闭")
}
