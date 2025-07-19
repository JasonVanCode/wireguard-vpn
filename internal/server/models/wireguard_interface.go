package models

import (
	"time"

	"gorm.io/gorm"
)

// WireGuardInterface WireGuard接口配置
type WireGuardInterface struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	Name        string          `json:"name" gorm:"not null;size:20;uniqueIndex"` // wg0, wg1, wg2等
	Description string          `json:"description" gorm:"size:200"`              // 接口描述
	Network     string          `json:"network" gorm:"not null;size:20"`          // 网络段，如10.10.0.0/24
	ServerIP    string          `json:"server_ip" gorm:"not null;size:15"`        // 服务器IP，如10.10.0.1
	ListenPort  int             `json:"listen_port" gorm:"not null"`              // 监听端口
	PublicKey   string          `json:"public_key" gorm:"not null;size:44"`       // 服务器公钥
	PrivateKey  string          `json:"private_key" gorm:"not null;size:44"`      // 服务器私钥
	Status      InterfaceStatus `json:"status" gorm:"default:0"`                  // 接口状态
	MaxPeers    int             `json:"max_peers" gorm:"default:100"`             // 最大连接数
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `json:"deleted_at" gorm:"index"`

	// 配置选项
	DNS        string `json:"dns" gorm:"default:'8.8.8.8,8.8.4.4'"` // DNS服务器
	MTU        int    `json:"mtu" gorm:"default:1420"`              // MTU大小
	PostUp     string `json:"post_up" gorm:"size:500"`              // 启动后执行的命令
	PostDown   string `json:"post_down" gorm:"size:500"`            // 停止后执行的命令
	PreUp      string `json:"pre_up" gorm:"size:500"`               // 启动前执行的命令
	PreDown    string `json:"pre_down" gorm:"size:500"`             // 停止前执行的命令
	SaveConfig bool   `json:"save_config" gorm:"default:true"`      // 是否保存配置

	// 统计信息
	TotalPeers    int       `json:"total_peers" gorm:"default:0"`   // 总连接数
	ActivePeers   int       `json:"active_peers" gorm:"default:0"`  // 活跃连接数
	TotalTraffic  uint64    `json:"total_traffic" gorm:"default:0"` // 总流量
	LastHeartbeat time.Time `json:"last_heartbeat"`                 // 最后心跳时间

	// 关联
	Modules []Module `json:"modules,omitempty" gorm:"foreignKey:InterfaceID"`
}

// InterfaceStatus 接口状态枚举
type InterfaceStatus int

const (
	InterfaceStatusDown     InterfaceStatus = iota // 停止
	InterfaceStatusUp                              // 运行中
	InterfaceStatusError                           // 错误
	InterfaceStatusStarting                        // 启动中
	InterfaceStatusStopping                        // 停止中
)

func (s InterfaceStatus) String() string {
	switch s {
	case InterfaceStatusUp:
		return "运行中"
	case InterfaceStatusDown:
		return "已停止"
	case InterfaceStatusError:
		return "错误"
	case InterfaceStatusStarting:
		return "启动中"
	case InterfaceStatusStopping:
		return "停止中"
	default:
		return "未知"
	}
}

// InterfaceTemplate 接口模板
type InterfaceTemplate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Network     string `json:"network"`
	ListenPort  int    `json:"listen_port"`
	MaxPeers    int    `json:"max_peers"`
	DNS         string `json:"dns"`
	PostUp      string `json:"post_up"`
	PostDown    string `json:"post_down"`
}

// GetDefaultTemplates 获取默认接口模板
func GetDefaultTemplates() []InterfaceTemplate {
	return []InterfaceTemplate{
		{
			Name:        "wg0",
			Description: "主接口 - 生产环境",
			Network:     "10.10.0.0/24",
			ListenPort:  51820,
			MaxPeers:    100,
			DNS:         "8.8.8.8,8.8.4.4",
			PostUp:      "iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE",
			PostDown:    "iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE",
		},
		{
			Name:        "wg1",
			Description: "北京节点专用",
			Network:     "10.11.0.0/24",
			ListenPort:  51821,
			MaxPeers:    50,
			DNS:         "8.8.8.8,8.8.4.4",
			PostUp:      "iptables -A FORWARD -i wg1 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE",
			PostDown:    "iptables -D FORWARD -i wg1 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE",
		},
		{
			Name:        "wg2",
			Description: "上海节点专用",
			Network:     "10.12.0.0/24",
			ListenPort:  51822,
			MaxPeers:    50,
			DNS:         "8.8.8.8,8.8.4.4",
			PostUp:      "iptables -A FORWARD -i wg2 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE",
			PostDown:    "iptables -D FORWARD -i wg2 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE",
		},
		{
			Name:        "wg99",
			Description: "测试环境专用",
			Network:     "10.99.0.0/24",
			ListenPort:  51899,
			MaxPeers:    10,
			DNS:         "8.8.8.8,8.8.4.4",
			PostUp:      "iptables -A FORWARD -i wg99 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE",
			PostDown:    "iptables -D FORWARD -i wg99 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE",
		},
	}
}
