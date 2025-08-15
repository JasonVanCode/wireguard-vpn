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
	Endpoint     string `json:"endpoint" gorm:"size:100"`     // 服务端端点（公网IP:端口）

	// 关联
	Interface *WireGuardInterface `json:"interface,omitempty" gorm:"foreignKey:InterfaceID"`
	UserVPNs  []UserVPN           `json:"user_vpns,omitempty" gorm:"foreignKey:ModuleID"`
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

// ModuleCreateRequest 模块创建请求结构
type ModuleCreateRequest struct {
	Name                string `json:"name" binding:"required"`         // 模块名称
	Location            string `json:"location" binding:"required"`     // 部署位置
	Description         string `json:"description"`                     // 描述
	InterfaceID         uint   `json:"interface_id" binding:"required"` // WireGuard接口ID
	AllowedIPs          string `json:"allowed_ips" binding:"required"`  // 允许访问的网段
	LocalIP             string `json:"local_ip"`                        // 模块内网IP地址
	PersistentKeepalive int    `json:"persistent_keepalive"`            // 保活间隔
	DNS                 string `json:"dns"`                             // DNS服务器
	AutoGenerateKeys    bool   `json:"auto_generate_keys"`              // 自动生成密钥对
	AutoAssignIP        bool   `json:"auto_assign_ip"`                  // 自动分配IP地址
	ConfigTemplate      string `json:"config_template"`                 // 配置模板
	PublicKey           string `json:"public_key,omitempty"`            // 自定义公钥（当AutoGenerateKeys为false时使用）
	PrivateKey          string `json:"private_key,omitempty"`           // 自定义私钥（当AutoGenerateKeys为false时使用）
	IPAddress           string `json:"ip_address,omitempty"`            // 自定义IP地址（当AutoAssignIP为false时使用）
}
