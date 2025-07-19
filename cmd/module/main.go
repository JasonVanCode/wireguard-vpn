package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"eitec-vpn/internal/module/database"
	"eitec-vpn/internal/module/routes"
	"eitec-vpn/internal/module/services"
	"eitec-vpn/internal/shared/config"
	"eitec-vpn/internal/shared/utils"

	"github.com/gin-gonic/gin"
)

var (
	configFile  = flag.String("config", "configs/module.yaml", "配置文件路径")
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
	AppName = "EITEC VPN Module"
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

	// 加载基础配置
	cfg, err := config.LoadModuleConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化模块端专用数据库
	// 使用utils方法获取数据库绝对路径
	dbPath, err := utils.GetAbsolutePath("data/module.db")
	if err != nil {
		log.Fatalf("获取数据库路径失败: %v", err)
	}

	// 确保data目录存在
	dataDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}

	log.Printf("数据库路径: %s", dbPath)

	if *initDB {
		// 强制初始化模式
		if err := database.InitModuleDatabaseWithOptions(dbPath, true); err != nil {
			log.Fatalf("初始化模块数据库失败: %v", err)
		}
		log.Println("模块数据库强制初始化完成")
		return
	} else {
		// 正常启动模式
		if err := database.InitModuleDatabase(dbPath); err != nil {
			log.Fatalf("初始化模块数据库失败: %v", err)
		}
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建模块管理器
	moduleManager, err := services.NewModuleManager(cfg)
	if err != nil {
		log.Fatalf("创建模块管理器失败: %v", err)
	}

	// 获取服务实例并创建路由
	moduleService, statusService := moduleManager.GetServices()
	db := database.GetDB()
	router := routes.SetupModuleRoutes(moduleService, statusService, cfg, db)

	// 创建HTTP服务器并设置到管理器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: router,
	}
	moduleManager.SetServer(server)

	// 启动模块管理器
	if err := moduleManager.Start(); err != nil {
		log.Fatalf("启动模块管理器失败: %v", err)
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭模块...")

	// 优雅关闭
	if err := moduleManager.Stop(); err != nil {
		log.Printf("关闭模块管理器失败: %v", err)
	}

	log.Println("模块已关闭")
}
