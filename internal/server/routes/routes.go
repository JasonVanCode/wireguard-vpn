package routes

import (
	"eitec-vpn/internal/server/handlers"
	"eitec-vpn/internal/server/middleware"
	"eitec-vpn/internal/server/services"
	"eitec-vpn/internal/shared/auth"
	"eitec-vpn/internal/shared/utils"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// PageRoute 页面路由配置
type PageRoute struct {
	Path     string
	Template string
	Title    string
}

// RedirectRoute 重定向路由配置
type RedirectRoute struct {
	From string
	To   string
}

// SetupRoutes 设置服务端路由 - 优化版本，消除重复代码
func SetupRoutes(
	moduleService *services.ModuleService,
	dashboardService *services.DashboardService,
	configService *services.ConfigService,
	userService *auth.UserService,
	jwtService *auth.JWTService,
	sessionManager *auth.SessionManager,
) *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.SecurityMiddleware())
	r.Use(middleware.TimeoutMiddleware())

	// 创建处理器
	moduleHandler := handlers.NewModuleHandler(moduleService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	configHandler := handlers.NewConfigHandler(configService)
	authHandler := handlers.NewAuthHandler(userService, jwtService, sessionManager)
	userHandler := handlers.NewUserHandler(userService)
	interfaceHandler := handlers.NewInterfaceHandler()

	// 健康检查 (无需认证)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "EITEC VPN Server is running",
		})
	})

	// 设置静态文件和模板
	setupStaticAndTemplates(r)

	// 设置页面路由
	setupPageRoutes(r)

	// 设置API路由
	setupAPIRoutes(r, moduleHandler, dashboardHandler, configHandler, authHandler, userHandler, interfaceHandler)

	return r
}

// SetupAPIRoutes 设置API路由 (用于生产环境，仅API)
func SetupAPIRoutes(
	moduleService *services.ModuleService,
	dashboardService *services.DashboardService,
	configService *services.ConfigService,
	userService *auth.UserService,
	jwtService *auth.JWTService,
	sessionManager *auth.SessionManager,
) *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.SecurityMiddleware())
	r.Use(middleware.TimeoutMiddleware())

	// 创建处理器
	moduleHandler := handlers.NewModuleHandler(moduleService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	configHandler := handlers.NewConfigHandler(configService)
	authHandler := handlers.NewAuthHandler(userService, jwtService, sessionManager)
	userHandler := handlers.NewUserHandler(userService)
	interfaceHandler := handlers.NewInterfaceHandler()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "EITEC VPN Server API is running",
		})
	})

	// API版本信息
	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": "2.0.0",
			"name":    "EITEC VPN Server",
			"build":   "development",
		})
	})

	// 设置API路由
	setupAPIRoutes(r, moduleHandler, dashboardHandler, configHandler, authHandler, userHandler, interfaceHandler)

	return r
}

// setupStaticAndTemplates 设置静态文件和模板
func setupStaticAndTemplates(r *gin.Engine) {
	// 智能查找静态文件和模板路径
	staticPath := utils.FindWebPath("server/static")
	assetsPath := utils.FindWebPath("server/assets")
	templatesPath := utils.FindWebPath("server/templates")

	// 调试输出
	fmt.Printf("Static path: %s\n", staticPath)
	fmt.Printf("Assets path: %s\n", assetsPath)
	fmt.Printf("Templates path: %s\n", templatesPath)

	// 静态文件服务
	r.Static("/static", staticPath)
	r.Static("/assets", assetsPath)

	// 查找所有模板文件
	pattern1 := filepath.Join(templatesPath, "*.html")
	pattern2 := filepath.Join(templatesPath, "*", "*.html")

	files1, _ := filepath.Glob(pattern1)
	files2, _ := filepath.Glob(pattern2)

	allTemplateFiles := append(files1, files2...)
	fmt.Printf("Found template files: %v\n", allTemplateFiles)

	if len(allTemplateFiles) > 0 {
		r.LoadHTMLFiles(allTemplateFiles...)
	} else {
		fmt.Printf("No template files found, trying simple pattern\n")
		r.LoadHTMLGlob(pattern1)
	}
}

// setupPageRoutes 设置页面路由
func setupPageRoutes(r *gin.Engine) {
	// 页面路由配置
	pageRoutes := []PageRoute{
		{"/", "index.html", "EITEC VPN 管理平台"},
		{"/login", "login.html", "登录 - EITEC VPN"},
		{"/modules", "modules.html", "模块管理 - EITEC VPN"},
		{"/config", "config.html", "系统配置 - EITEC VPN"},
		{"/users", "users.html", "用户管理 - EITEC VPN"},
	}

	// 批量设置页面路由
	for _, route := range pageRoutes {
		path := route.Path
		template := route.Template
		title := route.Title
		r.GET(path, func(c *gin.Context) {
			c.HTML(http.StatusOK, template, gin.H{
				"title": title,
			})
		})
	}

	// 兼容性重定向路由
	redirectRoutes := []RedirectRoute{
		{"/index", "/"},
		{"/index.html", "/"},
	}

	for _, redirect := range redirectRoutes {
		from := redirect.From
		to := redirect.To
		r.GET(from, func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, to)
		})
	}

	// 404处理
	r.NoRoute(func(c *gin.Context) {
		c.HTML(404, "404.html", gin.H{
			"title": "页面未找到 - EITEC VPN",
		})
	})
}

// setupAPIRoutes 设置API路由
func setupAPIRoutes(
	r *gin.Engine,
	moduleHandler *handlers.ModuleHandler,
	dashboardHandler *handlers.DashboardHandler,
	configHandler *handlers.ConfigHandler,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	interfaceHandler *handlers.InterfaceHandler,
) {

	// API路由组
	api := r.Group("/api/v1")
	{
		// 公共路由 (无需认证)
		public := api.Group("")
		{
			public.POST("/auth/login", authHandler.Login)
			public.POST("/auth/refresh", authHandler.RefreshToken)
		}

		// 认证路由 (需要JWT认证) - 暂时注释掉认证要求
		auth := api.Group("")
		// auth.Use(middleware.JWTAuthMiddleware()) // 注释掉JWT认证中间件
		{
			// 认证相关
			setupAuthRoutes(auth, authHandler)

			// 仪表盘相关
			setupDashboardRoutes(auth, dashboardHandler)

			// 模块管理相关
			setupModuleRoutes(auth, moduleHandler)

			// 系统配置相关 (仅管理员) - 注释掉权限检查
			setupConfigRoutes(auth, configHandler)

			// 用户管理相关 (仅管理员) - 注释掉权限检查
			setupUserRoutes(auth, userHandler)

			// WireGuard接口管理相关
			setupInterfaceRoutes(auth, interfaceHandler)

		}
	}
}

// setupAuthRoutes 设置认证相关路由
func setupAuthRoutes(auth *gin.RouterGroup, authHandler *handlers.AuthHandler) {
	auth.GET("/auth/me", authHandler.GetCurrentUser)
	auth.POST("/auth/logout", authHandler.Logout)
	auth.POST("/auth/change-password", authHandler.ChangePassword)
}

// setupDashboardRoutes 设置仪表盘相关路由
func setupDashboardRoutes(auth *gin.RouterGroup, dashboardHandler *handlers.DashboardHandler) {
	auth.GET("/dashboard/stats", dashboardHandler.GetDashboardStats)
	auth.GET("/dashboard/health", dashboardHandler.GetSystemHealth)

	// 保持简单，使用现有的两个API：/dashboard/stats 和 /system/wireguard-interfaces

	// 系统相关
	auth.GET("/system/network-interfaces", dashboardHandler.GetNetworkInterfaces) // 获取网络接口列表

	// WireGuard接口状态 - 优化版本，返回实时状态
	auth.GET("/system/wireguard-interfaces", dashboardHandler.GetWireGuardInterfacesWithStatus) // 获取WireGuard接口实时状态
}

// setupModuleRoutes 设置模块管理相关路由
func setupModuleRoutes(auth *gin.RouterGroup, moduleHandler *handlers.ModuleHandler) {
	modules := auth.Group("/modules")
	{
		modules.GET("", moduleHandler.GetModules)
		modules.POST("" /* middleware.RequireRole("admin"), */, moduleHandler.CreateModule) // 注释掉角色权限检查
		modules.GET("/stats", moduleHandler.GetModuleStats)
		modules.POST("/sync" /* middleware.RequireRole("admin"), */, moduleHandler.SyncModuleStatus)      // 注释掉角色权限检查
		modules.DELETE("/batch" /* middleware.RequireRole("admin"), */, moduleHandler.BatchDeleteModules) // 注释掉角色权限检查

		// 单个模块操作
		module := modules.Group("/:id")
		{
			module.GET("", moduleHandler.GetModule)
			module.PUT("" /* middleware.RequireRole("admin"), */, moduleHandler.UpdateModule)                    // 注释掉角色权限检查
			module.DELETE("" /* middleware.RequireRole("admin"), */, moduleHandler.DeleteModule)                 // 注释掉角色权限检查
			module.PUT("/status" /* middleware.RequireRole("admin"), */, moduleHandler.UpdateModuleStatus)       // 注释掉角色权限检查
			module.POST("/regenerate-keys" /* middleware.RequireRole("admin"), */, moduleHandler.RegenerateKeys) // 注释掉角色权限检查
			module.GET("/config", moduleHandler.GenerateModuleConfig)
			module.GET("/peer-config", moduleHandler.GeneratePeerConfig)
		}
	}

	// 用户VPN管理路由
	userVPNHandler := handlers.NewUserVPNHandler()

	// 用户VPN CRUD操作
	auth.POST("/user-vpn", userVPNHandler.CreateUserVPN)
	auth.GET("/user-vpn/:id", userVPNHandler.GetUserVPN)
	auth.PUT("/user-vpn/:id", userVPNHandler.UpdateUserVPN)
	auth.DELETE("/user-vpn/:id", userVPNHandler.DeleteUserVPN)
	auth.GET("/user-vpn/:id/config", userVPNHandler.GenerateUserVPNConfig)

	// 模块相关的用户VPN操作 - 修复参数名冲突
	auth.GET("/modules/:id/users", userVPNHandler.GetUserVPNsByModule)
	auth.GET("/modules/:id/user-stats", userVPNHandler.GetUserVPNStats)
}

// setupConfigRoutes 设置系统配置相关路由
func setupConfigRoutes(auth *gin.RouterGroup, configHandler *handlers.ConfigHandler) {
	config := auth.Group("/config")
	// config.Use(middleware.RequireRole("admin")) // 注释掉管理员权限检查
	{
		config.GET("", configHandler.GetSystemConfig)
		config.GET("/status", configHandler.GetConfigStatus)
		config.PUT("", configHandler.UpdateSystemConfig)
		config.POST("/reset", configHandler.ResetToDefaults)
		config.GET("/export", configHandler.ExportConfig)
		config.POST("/import", configHandler.ImportConfig)
		config.POST("/wireguard/init", configHandler.InitializeWireGuard)
		config.POST("/wireguard/apply", configHandler.ApplyWireGuardConfig)
		config.GET("/wireguard/server-config", configHandler.GenerateServerConfig)
		config.POST("/wireguard/validate", configHandler.ValidateNetworkSettings)
	}
}

// setupUserRoutes 设置用户管理相关路由
func setupUserRoutes(auth *gin.RouterGroup, userHandler *handlers.UserHandler) {
	users := auth.Group("/users")
	// users.Use(middleware.RequireRole("admin")) // 注释掉管理员权限检查
	{
		users.GET("", userHandler.GetUsers)
		users.POST("", userHandler.CreateUser)
		users.GET("/:id", userHandler.GetUser)
		users.PUT("/:id", userHandler.UpdateUser)
		users.DELETE("/:id", userHandler.DeleteUser)
		users.PUT("/:id/status", userHandler.UpdateUserStatus)
		users.POST("/:id/reset-password", userHandler.ResetPassword)
	}
}

// setupInterfaceRoutes 设置WireGuard接口管理相关路由
func setupInterfaceRoutes(auth *gin.RouterGroup, interfaceHandler *handlers.InterfaceHandler) {
	interfaces := auth.Group("/interfaces")
	{
		interfaces.GET("", interfaceHandler.GetInterfaces)
		interfaces.POST("", interfaceHandler.CreateInterface) // 添加创建接口路由
		interfaces.GET("/:id", interfaceHandler.GetInterface)
		interfaces.GET("/:id/config", interfaceHandler.GetInterfaceConfig) // 添加获取接口配置路由
		interfaces.PUT("/:id/start", interfaceHandler.StartInterface)
		interfaces.PUT("/:id/stop", interfaceHandler.StopInterface)
		interfaces.DELETE("/:id", interfaceHandler.DeleteInterface) // 添加删除接口路由
		interfaces.GET("/stats", interfaceHandler.GetInterfaceStats)
	}
}
