package auth

import (
	"errors"
	"fmt"
	"time"

	"eitec-vpn/internal/server/database"
	"eitec-vpn/internal/server/models"
	"eitec-vpn/internal/shared/utils"

	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		db: database.DB,
	}
}

// Login 用户登录
func (us *UserService) Login(username, password string) (*models.User, error) {
	var user models.User
	if err := us.db.Where("username = ? AND is_active = ?", username, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 验证密码
	if !utils.CheckPassword(password, user.Password) {
		return nil, errors.New("用户名或密码错误")
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now
	us.db.Save(&user)

	return &user, nil
}

// GetUserByID 根据ID获取用户
func (us *UserService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := us.db.Where("id = ? AND is_active = ?", id, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (us *UserService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := us.db.Where("username = ? AND is_active = ?", username, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return &user, nil
}

// CreateUser 创建用户
func (us *UserService) CreateUser(username, password string) (*models.User, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	if err := us.db.Where("username = ?", username).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	user := &models.User{
		Username: username,
		Password: hashedPassword,
		IsActive: true,
	}

	if err := us.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

// UpdateUser 更新用户信息
func (us *UserService) UpdateUser(id uint, updates map[string]interface{}) error {
	// 如果要更新密码，需要加密
	if password, exists := updates["password"]; exists {
		if passwordStr, ok := password.(string); ok && passwordStr != "" {
			hashedPassword, err := utils.HashPassword(passwordStr)
			if err != nil {
				return fmt.Errorf("密码加密失败: %w", err)
			}
			updates["password"] = hashedPassword
		}
	}

	result := us.db.Model(&models.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("更新用户失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("用户不存在")
	}

	return nil
}

// DeleteUser 删除用户 (软删除)
func (us *UserService) DeleteUser(id uint) error {
	result := us.db.Delete(&models.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("删除用户失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("用户不存在")
	}

	return nil
}

// DeactivateUser 停用用户
func (us *UserService) DeactivateUser(id uint) error {
	return us.UpdateUser(id, map[string]interface{}{"is_active": false})
}

// ActivateUser 激活用户
func (us *UserService) ActivateUser(id uint) error {
	return us.UpdateUser(id, map[string]interface{}{"is_active": true})
}

// ChangePassword 修改密码
func (us *UserService) ChangePassword(id uint, oldPassword, newPassword string) error {
	user, err := us.GetUserByID(id)
	if err != nil {
		return err
	}

	// 验证旧密码
	if !utils.CheckPassword(oldPassword, user.Password) {
		return errors.New("原密码错误")
	}

	// 更新密码
	return us.UpdateUser(id, map[string]interface{}{"password": newPassword})
}

// GetUsers 获取用户列表 (分页)
func (us *UserService) GetUsers(page, pageSize int, filters map[string]interface{}) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := us.db.Model(&models.User{})

	// 应用过滤条件
	for key, value := range filters {
		switch key {
		case "username":
			query = query.Where("username LIKE ?", "%"+value.(string)+"%")
		case "role":
			query = query.Where("role = ?", value)
		case "is_active":
			query = query.Where("is_active = ?", value)
		}
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计用户数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}

	return users, total, nil
}

// GetUserStats 获取用户统计信息
func (us *UserService) GetUserStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// 总用户数
	var total int64
	if err := us.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("统计总用户数失败: %w", err)
	}
	stats["total"] = total

	// 活跃用户数
	var active int64
	if err := us.db.Model(&models.User{}).Where("is_active = ?", true).Count(&active).Error; err != nil {
		return nil, fmt.Errorf("统计活跃用户数失败: %w", err)
	}
	stats["active"] = active

	// 今日新增用户
	today := time.Now().Truncate(24 * time.Hour)
	var todayCount int64
	if err := us.db.Model(&models.User{}).Where("created_at >= ?", today).Count(&todayCount).Error; err != nil {
		return nil, fmt.Errorf("统计今日新增用户失败: %w", err)
	}
	stats["today"] = todayCount

	return stats, nil
}

// InitializeDefaultAdmin 初始化默认管理员
func (us *UserService) InitializeDefaultAdmin(username, password string) error {
	// 检查是否已有用户
	var count int64
	if err := us.db.Model(&models.User{}).Count(&count).Error; err != nil {
		return fmt.Errorf("检查用户数量失败: %w", err)
	}

	if count > 0 {
		return nil // 已有用户，跳过
	}

	// 创建默认管理员
	_, err := us.CreateUser(username, password)
	return err
}

// GetRecentLogins 获取最近登录记录
func (us *UserService) GetRecentLogins(limit int) ([]models.User, error) {
	var users []models.User
	if err := us.db.Where("last_login IS NOT NULL").
		Order("last_login DESC").
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("查询最近登录记录失败: %w", err)
	}

	return users, nil
}

// CreateAuthService 创建认证服务
func CreateAuthService(config AuthConfig) (*AuthServiceImpl, error) {
	jwtService := NewJWTService(
		config.AccessSecret,
		config.RefreshSecret,
		config.AccessExpiry,
		config.RefreshExpiry,
	)

	sessionManager := NewSessionManager(config.SessionTimeout)
	userService := NewUserService()

	return &AuthServiceImpl{
		jwtService:     jwtService,
		sessionManager: sessionManager,
		userService:    userService,
	}, nil
}

// AuthConfig 认证配置
type AuthConfig struct {
	AccessSecret   string
	RefreshSecret  string
	AccessExpiry   time.Duration
	RefreshExpiry  time.Duration
	SessionTimeout time.Duration
}

// AuthServiceImpl 认证服务实现
type AuthServiceImpl struct {
	jwtService     *JWTService
	sessionManager *SessionManager
	userService    *UserService
}

// ValidateAccessToken 实现AuthService接口
func (as *AuthServiceImpl) ValidateAccessToken(token string) (*JWTClaims, error) {
	return as.jwtService.ValidateAccessToken(token)
}

// ValidateAPIKey 实现AuthService接口
func (as *AuthServiceImpl) ValidateAPIKey(token string) (*JWTClaims, error) {
	return as.jwtService.ValidateAPIKey(token)
}

// Login 用户登录
func (as *AuthServiceImpl) Login(username, password, ipAddress, userAgent string) (*TokenPair, *Session, error) {
	// 验证用户
	user, err := as.userService.Login(username, password)
	if err != nil {
		return nil, nil, err
	}

	// 生成JWT令牌对
	tokenPair, err := as.jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	// 创建会话
	session, err := as.sessionManager.CreateSession(user, ipAddress, userAgent)
	if err != nil {
		return nil, nil, fmt.Errorf("创建会话失败: %w", err)
	}

	return tokenPair, session, nil
}

// Logout 用户退出
func (as *AuthServiceImpl) Logout(sessionID string) error {
	return as.sessionManager.DestroySession(sessionID)
}

// RefreshToken 刷新令牌
func (as *AuthServiceImpl) RefreshToken(refreshToken string) (*TokenPair, error) {
	// 验证刷新令牌
	claims, err := as.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 获取用户信息
	user, err := as.userService.GetUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	// 生成新的令牌对
	return as.jwtService.RefreshTokenPair(refreshToken, user)
}
