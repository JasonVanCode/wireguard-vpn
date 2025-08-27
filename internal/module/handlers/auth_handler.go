package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"log"

	"eitec-vpn/internal/module/models"
	"eitec-vpn/internal/shared/response"
	"eitec-vpn/internal/shared/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	db *gorm.DB
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// Login 处理登录请求
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 查找用户
	var user models.LocalUser
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Unauthorized(c, "用户名或密码错误")
			return
		}
		log.Printf("查询用户失败: %v", err)
		response.InternalError(c, "系统错误")
		return
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.Password) {
		response.Unauthorized(c, "用户名或密码错误")
		return
	}

	// 生成会话token
	token, err := generateToken()
	if err != nil {
		log.Printf("生成token失败: %v", err)
		response.InternalError(c, "系统错误")
		return
	}

	// 设置cookie - 使用module_token保持一致性
	c.SetCookie("module_token", token, 3600*24, "/", "", false, true)

	// 记录登录日志
	log.Printf("用户 %s 登录成功", user.Username)

	response.Success(c, gin.H{
		"message": "登录成功",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

// Logout 处理登出请求
func (h *AuthHandler) Logout(c *gin.Context) {
	// 清除cookie - 使用module_token保持一致性
	c.SetCookie("module_token", "", -1, "/", "", false, true)

	// 记录登出日志
	if username, exists := c.Get("username"); exists {
		log.Printf("用户 %s 登出", username)
	}

	response.Success(c, gin.H{
		"message": "登出成功",
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	// 查找用户
	var user models.LocalUser
	if err := h.db.First(&user, userID).Error; err != nil {
		log.Printf("查询用户失败: %v", err)
		response.InternalError(c, "系统错误")
		return
	}

	// 验证旧密码
	if !utils.CheckPassword(req.OldPassword, user.Password) {
		response.BadRequest(c, "旧密码错误")
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("密码加密失败: %v", err)
		response.InternalError(c, "系统错误")
		return
	}

	// 更新密码
	if err := h.db.Model(&user).Update("password", hashedPassword).Error; err != nil {
		log.Printf("更新密码失败: %v", err)
		response.InternalError(c, "系统错误")
		return
	}

	// 记录修改密码日志
	log.Printf("用户 %s 修改密码成功", user.Username)

	response.Success(c, gin.H{
		"message": "密码修改成功",
	})
}

// Verify 验证token有效性
func (h *AuthHandler) Verify(c *gin.Context) {
	// 从认证中间件获取用户信息
	username, exists := c.Get("username")
	if !exists {
		response.Unauthorized(c, "token无效")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "token无效")
		return
	}

	response.Success(c, gin.H{
		"message": "token有效",
		"user": gin.H{
			"id":       userID,
			"username": username,
		},
	})
}

// generateToken 生成随机token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
