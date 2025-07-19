package models

import (
	"time"
)

// UserVPN 用户VPN配置信息
type UserVPN struct {
	ID          uint          `json:"id" gorm:"primaryKey"`
	ModuleID    uint          `json:"module_id" gorm:"not null;index"`                // 关联的模块ID
	Username    string        `json:"username" gorm:"not null;size:100"`              // 用户名
	Email       string        `json:"email" gorm:"size:255"`                          // 用户邮箱
	Description string        `json:"description" gorm:"size:500"`                    // 用户描述
	PublicKey   string        `json:"public_key" gorm:"not null;size:44;uniqueIndex"` // 用户的WireGuard公钥
	PrivateKey  string        `json:"private_key" gorm:"not null;size:44"`            // 用户的WireGuard私钥
	IPAddress   string        `json:"ip_address" gorm:"not null;size:15;uniqueIndex"` // 分配给用户的IP地址
	Status      UserVPNStatus `json:"status" gorm:"default:0"`                        // 用户状态
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`

	// 统计信息
	TotalTxBytes uint64     `json:"total_tx_bytes" gorm:"default:0"` // 总发送字节数
	TotalRxBytes uint64     `json:"total_rx_bytes" gorm:"default:0"` // 总接收字节数
	LastSeen     *time.Time `json:"last_seen"`                       // 最后在线时间

	// 握手信息
	LatestHandshake *time.Time `json:"latest_handshake"`

	// 配置信息
	AllowedIPs   string     `json:"allowed_ips" gorm:"default:'0.0.0.0/0'"` // 用户可访问的网段
	PersistentKA int        `json:"persistent_keepalive" gorm:"default:25"` // 保活间隔
	PresharedKey string     `json:"preshared_key" gorm:"size:44"`           // 预共享密钥
	ExpiresAt    *time.Time `json:"expires_at"`                             // 配置过期时间

	// 权限控制
	IsActive   bool `json:"is_active" gorm:"default:true"` // 是否激活
	MaxDevices int  `json:"max_devices" gorm:"default:1"`  // 最大设备数

	// 关联
	Module *Module `json:"module,omitempty" gorm:"foreignKey:ModuleID"`
}

// UserVPNStatus 用户VPN状态枚举
type UserVPNStatus int

const (
	UserVPNStatusOffline   UserVPNStatus = iota // 离线
	UserVPNStatusOnline                         // 在线
	UserVPNStatusSuspended                      // 暂停
	UserVPNStatusExpired                        // 已过期
)

func (s UserVPNStatus) String() string {
	switch s {
	case UserVPNStatusOnline:
		return "在线"
	case UserVPNStatusOffline:
		return "离线"
	case UserVPNStatusSuspended:
		return "暂停"
	case UserVPNStatusExpired:
		return "已过期"
	default:
		return "未知"
	}
}

// UserVPNConfig 用户VPN配置请求结构
type UserVPNConfig struct {
	ModuleID    uint       `json:"module_id" binding:"required"`
	Username    string     `json:"username" binding:"required"`
	Email       string     `json:"email"`
	Description string     `json:"description"`
	AllowedIPs  string     `json:"allowed_ips"`
	ExpiresAt   *time.Time `json:"expires_at"`
	MaxDevices  int        `json:"max_devices"`
}
