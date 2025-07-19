package routes

import (
	"eitec-vpn/internal/module/handlers"
	"eitec-vpn/internal/module/middleware"
	"eitec-vpn/internal/module/services"
	"eitec-vpn/internal/shared/config"
	"eitec-vpn/internal/shared/utils"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupModuleRoutes 设置模块路由 - 参考server端设计，使用handlers模式
func SetupModuleRoutes(moduleService *services.ModuleService, statusService *services.StatusService, cfg *config.ModuleConfig, db *gorm.DB) *gin.Engine {
	router := gin.New()

	// 基础中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 智能查找静态文件和模板路径
	staticPath := utils.FindWebPath("module/static")
	templatesPath := utils.FindWebPath("module/templates")

	// 调试输出
	fmt.Printf("Module static path: %s\n", staticPath)
	fmt.Printf("Module templates path: %s\n", templatesPath)

	// 设置静态文件
	router.Static("/static", staticPath)

	// 查找所有模板文件
	pattern1 := filepath.Join(templatesPath, "*.html")
	pattern2 := filepath.Join(templatesPath, "*", "*.html")

	files1, _ := filepath.Glob(pattern1)
	files2, _ := filepath.Glob(pattern2)

	allTemplateFiles := append(files1, files2...)
	fmt.Printf("Found module template files: %v\n", allTemplateFiles)

	if len(allTemplateFiles) > 0 {
		router.LoadHTMLFiles(allTemplateFiles...)
	} else {
		fmt.Printf("No template files found, trying simple pattern\n")
		router.LoadHTMLGlob(pattern1)
	}

	// 创建处理器
	authHandler := handlers.NewAuthHandler(db)
	moduleHandler := handlers.NewModuleHandler(moduleService, statusService)

	// 健康检查 (无需认证)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"module": cfg.Module.Name,
		})
	})

	// 页面路由
	router.GET("/", func(c *gin.Context) {
		// 主页面显示index.html，集成了配置功能，由前端JavaScript处理认证和配置状态检查
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "模块管理",
		})
	})

	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "模块登录",
		})
	})

	// 兼容旧的配置路由，重定向到主页面
	router.GET("/config", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	// 404处理
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"code":    404,
			"message": "页面未找到",
		})
	})

	// API路由组
	api := router.Group("/api/v1")
	{
		// 公共API路由 (无需认证)
		public := api.Group("")
		{
			public.POST("/auth/login", authHandler.Login)
			// 添加配置状态检查接口（用于前端路由判断）
			public.GET("/config/status", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"configured": moduleService.IsConfigured(),
				})
			})
			// 初始配置接口（无需认证，用于首次配置）
			public.POST("/configure", moduleHandler.ConfigureModule)
		}

		// 认证API路由 (需要认证)
		auth := api.Group("")
		auth.Use(middleware.AuthMiddleware(authHandler))
		{
			// 认证相关
			auth.GET("/auth/verify", authHandler.Verify)
			auth.POST("/auth/logout", authHandler.Logout)
			auth.POST("/auth/change-password", authHandler.ChangePassword)

			// 状态相关
			auth.GET("/status", moduleHandler.GetStatus)
			auth.POST("/status", moduleHandler.UpdateStatus)

			// 配置相关
			auth.GET("/config", moduleHandler.GetConfig)
			auth.POST("/config", moduleHandler.UpdateConfig)

			// 流量统计
			auth.GET("/stats", moduleHandler.GetStats)

			// WireGuard控制
			auth.POST("/wireguard/start", moduleHandler.StartWireGuard)
			auth.POST("/wireguard/stop", moduleHandler.StopWireGuard)
			auth.POST("/wireguard/restart", moduleHandler.RestartWireGuard)
		}
	}

	return router
}
