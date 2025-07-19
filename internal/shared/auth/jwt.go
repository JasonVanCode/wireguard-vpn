package auth

import (
	"errors"
	"fmt"
	"time"

	"eitec-vpn/internal/server/models"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   uint      `json:"user_id"`
	Username string    `json:"username"`
	Type     TokenType `json:"type"`
	jwt.RegisteredClaims
}

// TokenType Token类型
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// TokenPair 令牌对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// JWTService JWT服务
type JWTService struct {
	accessSecret  string
	refreshSecret string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTService 创建JWT服务
func NewJWTService(accessSecret, refreshSecret string, accessExpiry, refreshExpiry time.Duration) *JWTService {
	return &JWTService{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateTokenPair 生成令牌对
func (j *JWTService) GenerateTokenPair(user *models.User) (*TokenPair, error) {
	now := time.Now()

	// 生成访问令牌
	accessClaims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Type:     AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "eitec-vpn",
			Subject:   fmt.Sprintf("%d", user.ID),
			Audience:  []string{"eitec-vpn-server"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(j.accessSecret))
	if err != nil {
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	// 生成刷新令牌
	refreshClaims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Type:     RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "eitec-vpn",
			Subject:   fmt.Sprintf("%d", user.ID),
			Audience:  []string{"eitec-vpn-server"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(j.refreshSecret))
	if err != nil {
		return nil, fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(j.accessExpiry.Seconds()),
	}, nil
}

// ValidateAccessToken 验证访问令牌
func (j *JWTService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	return j.validateToken(tokenString, j.accessSecret, AccessToken)
}

// ValidateRefreshToken 验证刷新令牌
func (j *JWTService) ValidateRefreshToken(tokenString string) (*JWTClaims, error) {
	return j.validateToken(tokenString, j.refreshSecret, RefreshToken)
}

// validateToken 验证令牌
func (j *JWTService) validateToken(tokenString, secret string, expectedType TokenType) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析令牌失败: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("无效的令牌")
	}

	// 验证令牌类型
	if claims.Type != expectedType {
		return nil, fmt.Errorf("令牌类型不匹配，期望: %s, 实际: %s", expectedType, claims.Type)
	}

	// 验证是否过期
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("令牌已过期")
	}

	return claims, nil
}

// RefreshTokenPair 刷新令牌对
func (j *JWTService) RefreshTokenPair(refreshTokenString string, user *models.User) (*TokenPair, error) {
	// 验证刷新令牌
	claims, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("刷新令牌无效: %w", err)
	}

	// 验证用户ID是否匹配
	if claims.UserID != user.ID {
		return nil, errors.New("用户ID不匹配")
	}

	// 生成新的令牌对
	return j.GenerateTokenPair(user)
}

// ExtractTokenFromHeader 从HTTP头部提取令牌
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("授权头部为空")
	}

	// 期望格式: "Bearer <token>"
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("无效的授权头部格式")
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", errors.New("令牌为空")
	}

	return token, nil
}

// IsTokenExpired 检查令牌是否即将过期 (30分钟内)
func IsTokenExpired(claims *JWTClaims) bool {
	if claims.ExpiresAt == nil {
		return false
	}

	// 如果令牌在30分钟内过期，认为需要刷新
	return claims.ExpiresAt.Time.Sub(time.Now()) < 30*time.Minute
}

// GetUserFromClaims 从声明中获取用户信息
func GetUserFromClaims(claims *JWTClaims) *models.User {
	return &models.User{
		ID:       claims.UserID,
		Username: claims.Username,
	}
}

// CreateAPIKey 生成API密钥 (长期有效的特殊令牌)
func (j *JWTService) CreateAPIKey(user *models.User, description string, expiryDays int) (string, error) {
	now := time.Now()
	expiry := now.AddDate(0, 0, expiryDays)

	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Type:     "api_key",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "eitec-vpn",
			Subject:   fmt.Sprintf("api_%d", user.ID),
			Audience:  []string{"eitec-vpn-api"},
			ExpiresAt: jwt.NewNumericDate(expiry),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.accessSecret))
}

// ValidateAPIKey 验证API密钥
func (j *JWTService) ValidateAPIKey(tokenString string) (*JWTClaims, error) {
	return j.validateToken(tokenString, j.accessSecret, "api_key")
}
