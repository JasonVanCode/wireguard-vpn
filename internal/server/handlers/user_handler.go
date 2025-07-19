package handlers

import (
	"strconv"

	"eitec-vpn/internal/shared/auth"
	"eitec-vpn/internal/shared/response"
	"eitec-vpn/internal/shared/utils"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户管理处理器
type UserHandler struct {
	userService *auth.UserService
}

// NewUserHandler 创建用户管理处理器
func NewUserHandler(userService *auth.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	IsActive *bool  `json:"is_active"`
}

// GetUsers 获取用户列表
func (uh *UserHandler) GetUsers(c *gin.Context) {
	// 获取分页参数
	page, pageSize := utils.ParsePagination(c.Request)

	users, total, err := uh.userService.GetUsers(page, pageSize, make(map[string]interface{}))
	if err != nil {
		response.InternalError(c, "获取用户列表失败")
		return
	}

	userResponse := map[string]interface{}{
		"users": users,
		"total": total,
		"page":  page,
		"limit": pageSize,
	}

	response.Success(c, userResponse)
}

// CreateUser 创建用户
func (uh *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 创建用户
	user, err := uh.userService.CreateUser(req.Username, req.Password)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 移除密码字段
	user.Password = ""

	response.SuccessWithMessage(c, "用户创建成功", user)
}

// GetUser 获取用户信息
func (uh *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "用户ID格式错误")
		return
	}

	user, err := uh.userService.GetUserByID(uint(id))
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	response.Success(c, user)
}

// UpdateUser 更新用户信息
func (uh *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "用户ID格式错误")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := uh.userService.UpdateUser(uint(id), updates); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "用户更新成功", nil)
}

// DeleteUser 删除用户
func (uh *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "用户ID格式错误")
		return
	}

	if err := uh.userService.DeleteUser(uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "用户删除成功", nil)
}

// UpdateUserStatus 更新用户状态
func (uh *UserHandler) UpdateUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "用户ID格式错误")
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	updates := map[string]interface{}{
		"is_active": req.IsActive,
	}

	if err := uh.userService.UpdateUser(uint(id), updates); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	status := "启用"
	if !req.IsActive {
		status = "禁用"
	}

	response.SuccessWithMessage(c, "用户"+status+"成功", nil)
}

// ResetPassword 重置用户密码
func (uh *UserHandler) ResetPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "用户ID格式错误")
		return
	}

	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 重置密码 (直接更新密码字段)
	updates := map[string]interface{}{
		"password": req.NewPassword,
	}

	if err := uh.userService.UpdateUser(uint(id), updates); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "密码重置成功", nil)
}
