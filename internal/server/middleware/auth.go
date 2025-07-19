package middleware

import (
	"eitec-vpn/internal/shared/auth"
	"eitec-vpn/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "缺少认证信息")
			c.Abort()
			return
		}

		// 提取token
		token, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			response.Unauthorized(c, "无效的认证格式")
			c.Abort()
			return
		}

		// 验证token
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			response.Unauthorized(c, "无效的认证令牌")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("claims", claims)

		c.Next()
	}
}

// SessionAuthMiddleware 会话认证中间件
func SessionAuthMiddleware(sessionManager *auth.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Cookie获取会话ID
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		// 验证会话
		session, err := sessionManager.GetSession(sessionID)
		if err != nil {
			response.Unauthorized(c, "会话已过期")
			c.Abort()
			return
		}

		// 更新会话活跃时间
		sessionManager.UpdateSession(sessionID)

		// 设置用户信息到上下文
		c.Set("user_id", session.UserID)
		c.Set("username", session.Username)
		c.Set("session_id", sessionID)

		c.Next()
	}
}

// APIKeyAuthMiddleware API密钥认证中间件
func APIKeyAuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取API Key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			response.Unauthorized(c, "缺少API密钥")
			c.Abort()
			return
		}

		// 验证API Key
		claims, err := jwtService.ValidateAPIKey(apiKey)
		if err != nil {
			response.Unauthorized(c, "无效的API密钥")
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("claims", claims)

		c.Next()
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
