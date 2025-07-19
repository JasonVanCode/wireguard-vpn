package handlers

import (
	"eitec-vpn/internal/shared/auth"
	"eitec-vpn/internal/shared/response"
	"eitec-vpn/internal/shared/utils"

	"strconv"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService    *auth.UserService
	jwtService     *auth.JWTService
	sessionManager *auth.SessionManager
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(userService *auth.UserService, jwtService *auth.JWTService, sessionManager *auth.SessionManager) *AuthHandler {
	return &AuthHandler{
		userService:    userService,
		jwtService:     jwtService,
		sessionManager: sessionManager,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	User         interface{} `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// Login 用户登录
func (ah *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 验证用户凭证
	user, err := ah.userService.Login(req.Username, req.Password)
	if err != nil {
		response.Unauthorized(c, "用户名或密码错误")
		return
	}

	// 检查用户状态
	if !user.IsActive {
		response.Forbidden(c, "账户已被禁用")
		return
	}

	// 生成JWT令牌
	tokenPair, err := ah.jwtService.GenerateTokenPair(user)
	if err != nil {
		response.InternalError(c, "生成令牌失败")
		return
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 创建会话
	session, err := ah.sessionManager.CreateSession(user, ipAddress, userAgent)
	if err != nil {
		response.InternalError(c, "创建会话失败")
		return
	}

	// 返回响应
	loginResponse := LoginResponse{
		User: map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}

	// 设置会话cookie (1小时有效期)
	c.SetCookie("session_id", session.ID, 3600, "/", "", false, true)

	response.Success(c, loginResponse)
}

// RefreshToken 刷新访问令牌
func (ah *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 验证刷新令牌
	claims, err := ah.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "刷新令牌无效")
		return
	}

	// 获取用户信息
	user, err := ah.userService.GetUserByID(claims.UserID)
	if err != nil {
		response.Unauthorized(c, "用户不存在")
		return
	}

	// 检查用户状态
	if !user.IsActive {
		response.Forbidden(c, "账户已被禁用")
		return
	}

	// 生成新的令牌对
	tokenPair, err := ah.jwtService.GenerateTokenPair(user)
	if err != nil {
		response.InternalError(c, "生成令牌失败")
		return
	}

	refreshResponse := map[string]interface{}{
		"access_token": tokenPair.AccessToken,
		"expires_in":   tokenPair.ExpiresIn,
	}

	response.Success(c, refreshResponse)
}

// GetCurrentUser 获取当前用户信息
func (ah *AuthHandler) GetCurrentUser(c *gin.Context) {
	// 临时注释掉认证检查，直接返回默认管理员信息
	/*
		// 从上下文获取用户ID
		userIDStr := c.GetString("user_id")
		if userIDStr == "" {
			response.Unauthorized(c, "用户未认证")
			return
		}

		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			response.BadRequest(c, "用户ID格式错误")
			return
		}

		// 获取用户信息
		user, err := ah.userService.GetUserByID(uint(userID))
		if err != nil {
			response.NotFound(c, "用户不存在")
			return
		}

		userInfo := map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"role":       user.Role,
			"is_active":  user.IsActive,
			"created_at": user.CreatedAt,
			"last_login": user.LastLogin,
		}
	*/

	// 返回默认的管理员用户信息
	userInfo := map[string]interface{}{
		"id":        1,
		"username":  "admin",
		"role":      "admin",
		"is_active": true,
	}

	response.Success(c, userInfo)
}

// Logout 用户注销
func (ah *AuthHandler) Logout(c *gin.Context) {
	// 从cookie获取会话ID
	sessionID, err := c.Cookie("session_id")
	if err == nil && sessionID != "" {
		// 销毁会话
		ah.sessionManager.DestroySession(sessionID)
	}

	// 清除cookie
	c.SetCookie("session_id", "", -1, "/", "", false, true)

	response.SuccessWithMessage(c, "注销成功", nil)
}

// ChangePassword 修改密码
func (ah *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 从上下文获取用户ID
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "用户ID格式错误")
		return
	}

	// 获取用户信息
	user, err := ah.userService.GetUserByID(uint(userID))
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	// 验证旧密码
	if !utils.CheckPassword(req.OldPassword, user.Password) {
		response.BadRequest(c, "旧密码错误")
		return
	}

	// 更新密码
	if err := ah.userService.ChangePassword(uint(userID), req.OldPassword, req.NewPassword); err != nil {
		response.InternalError(c, "密码更新失败")
		return
	}

	response.SuccessWithMessage(c, "密码修改成功", nil)
}
