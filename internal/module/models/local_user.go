package models

import (
	"time"

	"gorm.io/gorm"
)

// LocalUser 模块端本地用户
type LocalUser struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"not null;size:50;uniqueIndex"`
	Password  string         `json:"-" gorm:"not null;size:255"` // 哈希后的密码
	Role      LocalUserRole  `json:"role" gorm:"default:1"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	LastLogin *time.Time     `json:"last_login"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// LocalUserRole 本地用户角色
type LocalUserRole int

const (
	LocalUserRoleViewer LocalUserRole = iota // 只读用户
	LocalUserRoleAdmin                       // 管理员
)

func (r LocalUserRole) String() string {
	switch r {
	case LocalUserRoleAdmin:
		return "admin"
	case LocalUserRoleViewer:
		return "viewer"
	default:
		return "unknown"
	}
}

// TableName 指定表名
func (LocalUser) TableName() string {
	return "local_users"
}
