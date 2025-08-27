package middleware

import (
	"eitec-vpn/internal/module/handlers"
	"eitec-vpn/internal/module/models"
	"eitec-vpn/internal/shared/response"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(authHandler *handlers.AuthHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查Authorization header中的Bearer token
		authHeader := c.GetHeader("Authorization")
		var token string

		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// 如果没有Authorization header，尝试从cookie获取
			if cookieToken, err := c.Cookie("module_token"); err == nil {
				token = cookieToken
			}
		}

		if token == "" {
			response.Unauthorized(c, "未认证")
			c.Abort()
			return
		}

		// 验证token格式（检查token长度）
		if len(token) < 20 {
			response.Unauthorized(c, "无效token")
			c.Abort()
			return
		}

		// 验证token有效性 - 这里简化处理，实际应该从数据库验证
		// 暂时使用默认管理员信息，后续可以扩展为数据库验证
		c.Set("username", "admin")
		c.Set("user_id", uint(1))
		c.Set("role", models.LocalUserRoleAdmin)

		c.Next()
	}
}
