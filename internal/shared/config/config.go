package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// WireGuard 超时时间常量
const (
	// WireGuardOnlineTimeout WireGuard连接在线判断超时时间（2分钟）
	WireGuardOnlineTimeout = 2 * time.Minute

	// WireGuardOfflineTimeout WireGuard连接离线标记超时时间（2分钟）
	WireGuardOfflineTimeout = 2 * time.Minute
)

// 全局配置变量
var (
	globalServerConfig *ServerConfig
	configMutex        sync.RWMutex
)

// SetGlobalServerConfig 设置全局服务器配置
func SetGlobalServerConfig(config *ServerConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalServerConfig = config
}

// GetGlobalServerConfig 获取全局服务器配置
func GetGlobalServerConfig() *ServerConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalServerConfig
}

// ServerConfig 服务器端配置
type ServerConfig struct {
	App struct {
		Name           string        `yaml:"name"`
		Port           int           `yaml:"port"`
		Mode           string        `yaml:"mode"`
		Secret         string        `yaml:"secret"`
		Listen         string        `yaml:"listen"`
		ReadTimeout    time.Duration `yaml:"read_timeout"`
		WriteTimeout   time.Duration `yaml:"write_timeout"`
		IdleTimeout    time.Duration `yaml:"idle_timeout"`
		MaxHeaderBytes int           `yaml:"max_header_bytes"`
		ServerIP       string        `yaml:"server_ip"`
		TLS            struct {
			Enabled  bool   `yaml:"enabled"`
			Listen   string `yaml:"listen"`
			CertFile string `yaml:"cert_file"`
			KeyFile  string `yaml:"key_file"`
		} `yaml:"tls"`
	} `yaml:"app"`

	WireGuard struct {
		Interface    string `yaml:"interface"`
		Port         int    `yaml:"port"`
		Network      string `yaml:"network"`
		DNS          string `yaml:"dns"`
		SyncInterval int    `yaml:"sync_interval"`
	} `yaml:"wireguard"`

	Database struct {
		Type string `yaml:"type"`
		Path string `yaml:"path"`
	} `yaml:"database"`

	Auth struct {
		AdminUsername  string        `yaml:"admin_username"`
		AdminPassword  string        `yaml:"admin_password"`
		JWTSecret      string        `yaml:"jwt_secret"`
		RefreshSecret  string        `yaml:"refresh_secret"`
		AccessExpiry   time.Duration `yaml:"access_expiry"`
		RefreshExpiry  time.Duration `yaml:"refresh_expiry"`
		SessionTimeout time.Duration `yaml:"session_timeout"`
	} `yaml:"auth"`
}

// ModuleConfig 模块端配置
type ModuleConfig struct {
	App struct {
		Name   string `yaml:"name"`
		Port   int    `yaml:"port"`
		Secret string `yaml:"secret"`
	} `yaml:"app"`

	Module struct {
		ID             uint   `yaml:"id"`
		Name           string `yaml:"name"`
		Location       string `yaml:"location"`
		PrivateKey     string `yaml:"private_key"`
		ServerEndpoint string `yaml:"server_endpoint"`
		APIKey         string `yaml:"api_key"`
	} `yaml:"module"`

	Server struct {
		URL               string `yaml:"url"`
		HeartbeatInterval int    `yaml:"heartbeat_interval"`
		ReportInterval    int    `yaml:"report_interval"`
		SyncInterval      int    `yaml:"sync_interval"`
	} `yaml:"server"`

	WireGuard struct {
		Interface string `yaml:"interface"`
	} `yaml:"wireguard"`
}

// findConfigFile 智能查找配置文件
func findConfigFile(filename string) (string, error) {
	// 候选路径列表
	candidates := []string{
		filename,                                    // 当前目录
		filepath.Join("configs", filename),          // 当前目录的 configs 子目录
		filepath.Join("..", filename),               // 上级目录
		filepath.Join("..", "configs", filename),    // 上级目录的 configs 子目录
		filepath.Join("../..", "configs", filename), // 上上级目录的 configs 子目录
	}

	// 依次检查每个候选路径
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			abs, err := filepath.Abs(candidate)
			if err != nil {
				return candidate, nil // 返回相对路径
			}
			return abs, nil
		}
	}

	return "", fmt.Errorf("配置文件 %s 未找到，已搜索路径: %v", filename, candidates)
}

// LoadServerConfig 加载服务器配置
func LoadServerConfig(configPath string) (*ServerConfig, error) {
	config := &ServerConfig{}

	// 设置默认值
	config.App.Name = "EiTec VPN Server"
	config.App.Port = 8080
	config.App.Mode = "release"
	config.App.Listen = ":8080"
	config.App.ReadTimeout = 15 * time.Second
	config.App.WriteTimeout = 15 * time.Second
	config.App.IdleTimeout = 60 * time.Second
	config.App.MaxHeaderBytes = 1
	config.WireGuard.Interface = "wg0"
	config.WireGuard.Port = 51820
	config.WireGuard.Network = "10.10.0.0/24"
	config.WireGuard.DNS = "8.8.8.8,8.8.4.4"
	config.WireGuard.SyncInterval = 300
	config.Database.Type = "sqlite"
	config.Database.Path = "data/eitec-vpn.db"
	config.Auth.AdminUsername = "admin"
	config.Auth.AdminPassword = "admin123"
	config.Auth.JWTSecret = "your-jwt-secret-key"
	config.Auth.RefreshSecret = "your-refresh-secret-key"
	config.Auth.AccessExpiry = 1 * time.Hour
	config.Auth.RefreshExpiry = 24 * time.Hour
	config.Auth.SessionTimeout = 24 * time.Hour

	if configPath != "" {
		// 智能查找配置文件
		actualPath, err := findConfigFile(configPath)
		if err != nil {
			return nil, err
		}

		data, err := os.ReadFile(actualPath)
		if err != nil {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("解析配置文件失败: %w", err)
		}
	}

	// 验证必需配置
	if config.App.Secret == "" {
		return nil, fmt.Errorf("app.secret 不能为空")
	}

	return config, nil
}

// LoadModuleConfig 加载模块配置
func LoadModuleConfig(configPath string) (*ModuleConfig, error) {
	config := &ModuleConfig{}

	// 设置默认值
	config.App.Name = "EiTec VPN Module"
	config.App.Port = 8080
	config.Module.Name = "默认模块"
	config.Module.Location = "未设置"
	config.Server.HeartbeatInterval = 30
	config.Server.ReportInterval = 60
	config.Server.SyncInterval = 300
	config.WireGuard.Interface = "wg0"

	if configPath != "" {
		// 智能查找配置文件
		actualPath, err := findConfigFile(configPath)
		if err != nil {
			return nil, err
		}

		data, err := os.ReadFile(actualPath)
		if err != nil {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("解析配置文件失败: %w", err)
		}
	}

	// 验证必需配置
	if config.App.Secret == "" {
		return nil, fmt.Errorf("app.secret 不能为空")
	}

	return config, nil
}

// SaveServerConfig 保存服务器配置
func SaveServerConfig(config *ServerConfig, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	return nil
}

// SaveModuleConfig 保存模块配置
func SaveModuleConfig(config *ModuleConfig, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	return nil
}
