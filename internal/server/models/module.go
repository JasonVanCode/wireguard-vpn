package models

import (
	"time"
)

// Module WireGuard模块信息
type Module struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"not null;size:100"`
	Location    string       `json:"location" gorm:"size:200"`
	Description string       `json:"description" gorm:"size:500"`
	InterfaceID uint         `json:"interface_id" gorm:"not null;index"` // 关联的WireGuard接口ID
	PublicKey   string       `json:"public_key" gorm:"not null;size:44;uniqueIndex"`
	PrivateKey  string       `json:"private_key" gorm:"not null;size:44"`
	IPAddress   string       `json:"ip_address" gorm:"not null;size:15;uniqueIndex"` // VPN网段中的IP地址
	LocalIP     string       `json:"local_ip" gorm:"size:15"`                        // 模块在内网的IP地址，用于NAT转发
	Status      ModuleStatus `json:"status" gorm:"default:0"`
	LastSeen    *time.Time   `json:"last_seen"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`

	// 统计信息
	TotalTxBytes uint64 `json:"total_tx_bytes" gorm:"default:0"`
	TotalRxBytes uint64 `json:"total_rx_bytes" gorm:"default:0"`

	// 握手信息
	LatestHandshake *time.Time `json:"latest_handshake"`

	// 配置信息
	AllowedIPs   string `json:"allowed_ips" gorm:"default:'192.168.1.0/24'"`
	PersistentKA int    `json:"persistent_keepalive" gorm:"default:25"`
	PresharedKey string `json:"preshared_key" gorm:"size:44"` // 预共享密钥，增强安全性

	// 关联
	Interface *WireGuardInterface `json:"interface,omitempty" gorm:"foreignKey:InterfaceID"`
}

// ModuleStatus 模块状态枚举
type ModuleStatus int

const (
	ModuleStatusOffline      ModuleStatus = iota // 离线
	ModuleStatusOnline                           // 在线
	ModuleStatusWarning                          // 警告
	ModuleStatusUnconfigured                     // 未配置
)

func (s ModuleStatus) String() string {
	switch s {
	case ModuleStatusOnline:
		return "在线"
	case ModuleStatusOffline:
		return "离线"
	case ModuleStatusWarning:
		return "警告"
	case ModuleStatusUnconfigured:
		return "未配置"
	default:
		return "未知"
	}
}
