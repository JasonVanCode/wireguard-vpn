package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"eitec-vpn/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// 为路由文件提供的简化中间件别名
func LoggerMiddleware() gin.HandlerFunc {
	return RequestLogger()
}

func RecoveryMiddleware() gin.HandlerFunc {
	return Recovery()
}

func SecurityMiddleware() gin.HandlerFunc {
	return SecurityHeaders()
}

func TimeoutMiddleware() gin.HandlerFunc {
	return Timeout(30 * time.Second)
}

// RequireStringRole 需要特定角色的中间件 (临时简化实现)
func RequireStringRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置默认用户信息，模拟管理员权限
		c.Set("user_id", uint(1))
		c.Set("username", "admin")
		c.Set("user_role", "admin")
		c.Next()
	}
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSConfig 默认CORS配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
}

// CORS 跨域资源共享中间件
func CORS(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查Origin是否允许
		if len(config.AllowOrigins) > 0 {
			allowed := false
			for _, allowedOrigin := range config.AllowOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}

		// 设置允许的方法
		if len(config.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		}

		// 设置允许的头部
		if len(config.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		}

		// 设置暴露的头部
		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		// 设置是否允许凭证
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 设置预检请求缓存时间
		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", string(rune(int(config.MaxAge.Seconds()))))
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityHeaders 安全头中间件
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止XSS攻击
		c.Header("X-XSS-Protection", "1; mode=block")

		// 防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")

		// 强制HTTPS (在生产环境中启用)
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// 内容安全策略
		c.Header("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:")

		// 引用策略
		c.Header("Referrer-Policy", "no-referrer")

		c.Next()
	}
}

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("Panic recovered: %v", recovered)
		response.InternalError(c, "服务器内部错误")
	})
}

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// 生成简单的请求ID
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		c.Next()
	}
}

// Timeout 超时中间件
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})
		go func() {
			defer close(finished)
			c.Next()
		}()

		select {
		case <-finished:
			// 请求正常完成
		case <-ctx.Done():
			// 请求超时
			response.Error(c, http.StatusRequestTimeout, "请求超时")
			c.Abort()
		}
	}
}

// client 限流客户端
type client struct {
	requests []time.Time
	mutex    sync.Mutex
}

// RateLimit 限流中间件
func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	clients := make(map[string]*client)
	var mutex sync.RWMutex

	// 定期清理过期的客户端记录
	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()

		for range ticker.C {
			mutex.Lock()
			now := time.Now()
			for ip, c := range clients {
				c.mutex.Lock()
				var validRequests []time.Time
				for _, req := range c.requests {
					if now.Sub(req) < window {
						validRequests = append(validRequests, req)
					}
				}
				c.requests = validRequests
				c.mutex.Unlock()

				// 如果客户端没有最近的请求，删除记录
				if len(c.requests) == 0 {
					delete(clients, ip)
				}
			}
			mutex.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mutex.RLock()
		clientRecord, exists := clients[ip]
		mutex.RUnlock()

		if !exists {
			mutex.Lock()
			clientRecord = &client{requests: []time.Time{}}
			clients[ip] = clientRecord
			mutex.Unlock()
		}

		clientRecord.mutex.Lock()
		defer clientRecord.mutex.Unlock()

		// 清理过期请求
		var validRequests []time.Time
		for _, req := range clientRecord.requests {
			if now.Sub(req) < window {
				validRequests = append(validRequests, req)
			}
		}
		clientRecord.requests = validRequests

		// 检查是否超过限制
		if len(clientRecord.requests) >= maxRequests {
			response.Error(c, http.StatusTooManyRequests, "请求过于频繁")
			c.Abort()
			return
		}

		// 记录当前请求
		clientRecord.requests = append(clientRecord.requests, now)

		c.Next()
	}
}

// IPWhitelist IP白名单中间件
func IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// 检查IP是否在白名单中
		allowed := false
		for _, ip := range allowedIPs {
			if ip == clientIP {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Forbidden(c, "IP地址不在白名单中")
			c.Abort()
			return
		}

		c.Next()
	}
}

// Maintenance 维护模式中间件
func Maintenance(enabled bool, message string) gin.HandlerFunc {
	if message == "" {
		message = "系统正在维护中，请稍后再试"
	}

	return func(c *gin.Context) {
		if enabled {
			response.Error(c, http.StatusServiceUnavailable, message)
			c.Abort()
			return
		}

		c.Next()
	}
}

// HealthCheck 健康检查中间件
func HealthCheck(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == path {
			response.Success(c, gin.H{
				"status": "healthy",
				"time":   time.Now().Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
