package models

import (
	"time"

	"gorm.io/gorm"
)

// LocalModule 本地模块配置和状态
type LocalModule struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	ServerID     uint           `json:"server_id" gorm:"index"` // 对应服务端的模块ID
	Name         string         `json:"name" gorm:"not null;size:100"`
	Location     string         `json:"location" gorm:"size:200"`
	PublicKey    string         `json:"public_key" gorm:"not null;size:44"`
	PrivateKey   string         `json:"private_key" gorm:"not null;size:44"`
	IPAddress    string         `json:"ip_address" gorm:"size:15"`
	Status       ModuleStatus   `json:"status" gorm:"default:0"`
	LastSeen     *time.Time     `json:"last_seen"`
	LastSync     *time.Time     `json:"last_sync"`     // 最后同步时间
	ConfigHash   string         `json:"config_hash"`   // 配置哈希值，用于检测变更
	IsConfigured bool           `json:"is_configured"` // 是否已完成配置
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 服务器连接配置
	ServerURL      string `json:"server_url" gorm:"size:255"`
	ServerEndpoint string `json:"server_endpoint" gorm:"size:255"`
	APIKey         string `json:"api_key" gorm:"size:255"`

	// WireGuard 配置
	WireGuardInterface string `json:"wireguard_interface" gorm:"default:'wg0'"`
	AllowedIPs         string `json:"allowed_ips" gorm:"default:'0.0.0.0/0'"`
	PersistentKA       int    `json:"persistent_keepalive" gorm:"default:25"`
	DNS                string `json:"dns" gorm:"size:255"`
}

// ModuleStatus 模块状态枚举
type ModuleStatus int

const (
	ModuleStatusOffline      ModuleStatus = iota // 离线
	ModuleStatusOnline                           // 在线
	ModuleStatusConnecting                       // 连接中
	ModuleStatusError                            // 错误
	ModuleStatusUnconfigured                     // 未配置
	ModuleStatusSyncing                          // 同步中
)

func (s ModuleStatus) String() string {
	switch s {
	case ModuleStatusOnline:
		return "在线"
	case ModuleStatusOffline:
		return "离线"
	case ModuleStatusConnecting:
		return "连接中"
	case ModuleStatusError:
		return "错误"
	case ModuleStatusUnconfigured:
		return "未配置"
	case ModuleStatusSyncing:
		return "同步中"
	default:
		return "未知"
	}
}
