package services

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"eitec-vpn/internal/module/database"
	"eitec-vpn/internal/module/models"
	"eitec-vpn/internal/shared/config"
	"eitec-vpn/internal/shared/utils"
)

// 配置路径常量
const (
	DefaultWireGuardConfigPath = "/etc/wireguard/wg0.conf"
	DefaultModuleConfigDir     = "/etc/eitec-vpn"
	DefaultModuleInfoPath      = "/etc/eitec-vpn/module.info"
)

// ModuleService 模块服务
type ModuleService struct {
	config     *config.ModuleConfig
	configDir  string
	wgConfPath string
}

// NewModuleService 创建新的模块服务
func NewModuleService(cfg *config.ModuleConfig) *ModuleService {
	// 从环境变量获取配置目录，如果没有则使用默认值
	configDir := os.Getenv("EITEC_CONFIG_DIR")
	if configDir == "" {
		configDir = DefaultModuleConfigDir
	}

	wgConfPath := os.Getenv("WIREGUARD_CONFIG_PATH")
	if wgConfPath == "" {
		wgConfPath = DefaultWireGuardConfigPath
	}

	return &ModuleService{
		config:     cfg,
		configDir:  configDir,
		wgConfPath: wgConfPath,
	}
}

// ModuleInfo 模块信息
type ModuleInfo struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Status   string `json:"status"`
}

// SetupInfo 设置信息
type SetupInfo struct {
	ModuleID   uint   `json:"module_id"`
	APIKey     string `json:"api_key"`
	ServerURL  string `json:"server_url"`
	ConfigData string `json:"config_data"`
}

// GetModuleInfo 获取模块信息
func (ms *ModuleService) GetModuleInfo() *ModuleInfo {
	var moduleInfo *ModuleInfo

	// 优先从数据库获取
	db := database.DB
	if db != nil {
		var module models.LocalModule
		if err := db.First(&module).Error; err == nil {
			status := "未配置"
			if module.IsConfigured {
				if ms.IsWireGuardRunning() {
					status = "运行中"
				} else {
					status = "已停止"
				}
			}

			moduleInfo = &ModuleInfo{
				ID:       module.ServerID,
				Name:     module.Name,
				Location: module.Location,
				Status:   status,
			}
			return moduleInfo
		}
	}

	// 兼容性：从配置文件获取
	status := "未配置"
	if ms.IsConfigured() {
		if ms.IsWireGuardRunning() {
			status = "运行中"
		} else {
			status = "已停止"
		}
	}

	return &ModuleInfo{
		ID:       ms.getModuleID(),
		Name:     ms.config.Module.Name,
		Location: ms.config.Module.Location,
		Status:   status,
	}
}

// IsConfigured 检查是否已配置
func (ms *ModuleService) IsConfigured() bool {
	// 首先检查数据库
	db := database.DB
	if db != nil {
		var module models.LocalModule
		if err := db.Where("is_configured = ?", 1).First(&module).Error; err == nil {
			return true
		}
	}

	// 兼容性检查：检查配置文件是否存在
	configPath := DefaultWireGuardConfigPath
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false
	}

	// 检查必要配置项
	if ms.getModuleID() == 0 {
		return false
	}

	return true
}

// IsWireGuardRunning 检查WireGuard是否运行
func (ms *ModuleService) IsWireGuardRunning() bool {
	cmd := exec.Command("wg", "show", "wg0")
	err := cmd.Run()
	return err == nil
}

// ApplySetup 应用设置
func (ms *ModuleService) ApplySetup(setup *SetupInfo) error {
	// 写入WireGuard配置文件
	configPath := DefaultWireGuardConfigPath
	if err := os.WriteFile(configPath, []byte(setup.ConfigData), 0600); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	// 保存到数据库
	if err := ms.saveToDatabase(setup); err != nil {
		return fmt.Errorf("保存配置到数据库失败: %v", err)
	}

	// 更新模块配置文件（保持兼容性）
	ms.updateModuleConfig(setup)

	// 保存配置文件
	if err := ms.saveConfig(); err != nil {
		return fmt.Errorf("保存配置失败: %v", err)
	}

	return nil
}

// StartWireGuard 启动WireGuard
func (ms *ModuleService) StartWireGuard() error {
	if !ms.IsConfigured() {
		return fmt.Errorf("模块未配置")
	}

	// 停止现有实例
	exec.Command("wg-quick", "down", "wg0").Run()

	// 启动WireGuard
	cmd := exec.Command("wg-quick", "up", "wg0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("启动WireGuard失败: %v, 输出: %s", err, output)
	}

	return nil
}

// StopWireGuard 停止WireGuard
func (ms *ModuleService) StopWireGuard() error {
	cmd := exec.Command("wg-quick", "down", "wg0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("停止WireGuard失败: %v, 输出: %s", err, output)
	}

	return nil
}

// RestartWireGuard 重启WireGuard
func (ms *ModuleService) RestartWireGuard() error {
	if err := ms.StopWireGuard(); err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	return ms.StartWireGuard()
}

// GetWireGuardConfig 获取WireGuard配置
func (ms *ModuleService) GetWireGuardConfig() (string, error) {
	configPath := DefaultWireGuardConfigPath
	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("读取配置文件失败: %v", err)
	}

	return string(content), nil
}

// UpdateWireGuardConfig 更新WireGuard配置
func (ms *ModuleService) UpdateWireGuardConfig(config string) error {
	configPath := DefaultWireGuardConfigPath

	// 备份当前配置
	backupPath := configPath + ".backup"
	if _, err := os.Stat(configPath); err == nil {
		if err := exec.Command("cp", configPath, backupPath).Run(); err != nil {
			return fmt.Errorf("备份配置失败: %v", err)
		}
	}

	// 写入新配置
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	// 如果WireGuard正在运行，重启它
	if ms.IsWireGuardRunning() {
		if err := ms.RestartWireGuard(); err != nil {
			// 恢复备份
			exec.Command("mv", backupPath, configPath).Run()
			return fmt.Errorf("重启WireGuard失败: %v", err)
		}
	}

	// 删除备份
	os.Remove(backupPath)

	return nil
}

// ResetConfiguration 重置配置
func (ms *ModuleService) ResetConfiguration() error {
	// 停止WireGuard
	ms.StopWireGuard()

	// 删除配置文件
	configPath := DefaultWireGuardConfigPath
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除配置文件失败: %v", err)
	}

	// 重置模块配置
	ms.resetModuleConfig()

	// 保存配置
	return ms.saveConfig()
}

// getModuleID 获取模块ID
func (ms *ModuleService) getModuleID() uint {
	// 从配置文件或环境变量获取模块ID
	if id := os.Getenv("MODULE_ID"); id != "" {
		if idNum, err := strconv.ParseUint(id, 10, 32); err == nil {
			return uint(idNum)
		}
	}

	// 从配置信息文件获取
	infoPath := filepath.Join(ms.configDir, "module.info")
	if content, err := os.ReadFile(infoPath); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "MODULE_ID=") {
				idStr := strings.TrimPrefix(line, "MODULE_ID=")
				if idNum, err := strconv.ParseUint(idStr, 10, 32); err == nil {
					return uint(idNum)
				}
			}
		}
	}

	return 0
}

// updateModuleConfig 更新模块配置
func (ms *ModuleService) updateModuleConfig(setup *SetupInfo) {
	// 创建配置目录
	if err := utils.EnsureDir(ms.configDir); err != nil {
		log.Printf("创建配置目录失败: %v", err)
		return
	}

	// 写入模块信息
	infoContent := fmt.Sprintf("MODULE_ID=%d\nAPI_KEY=%s\nSERVER_URL=%s\n",
		setup.ModuleID, setup.APIKey, setup.ServerURL)

	infoPath := filepath.Join(ms.configDir, "module.info")
	if err := os.WriteFile(infoPath, []byte(infoContent), 0600); err != nil {
		log.Printf("写入模块信息失败: %v", err)
	}
}

// resetModuleConfig 重置模块配置
func (ms *ModuleService) resetModuleConfig() {
	// 删除模块信息文件
	infoPath := filepath.Join(ms.configDir, "module.info")
	if err := os.Remove(infoPath); err != nil && !os.IsNotExist(err) {
		log.Printf("删除模块信息文件失败: %v", err)
	}
}

// saveConfig 保存配置
func (ms *ModuleService) saveConfig() error {
	// 这里可以实现配置保存逻辑
	// 当前使用文件系统保存
	return nil
}

// saveToDatabase 保存配置到数据库
func (ms *ModuleService) saveToDatabase(setup *SetupInfo) error {
	db := database.DB
	if db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	// 计算配置哈希
	configHash := fmt.Sprintf("%x", md5.Sum([]byte(setup.ConfigData)))

	// 查找或创建模块记录
	var module models.LocalModule
	result := db.Where("server_id = ?", setup.ModuleID).First(&module)

	now := time.Now()

	if result.Error != nil {
		// 创建新记录
		module = models.LocalModule{
			ServerID:     setup.ModuleID,
			Name:         fmt.Sprintf("Module-%d", setup.ModuleID),
			ServerURL:    setup.ServerURL,
			APIKey:       setup.APIKey,
			ConfigHash:   configHash,
			IsConfigured: true,
			LastSync:     &now,
			Status:       models.ModuleStatusOnline,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := db.Create(&module).Error; err != nil {
			return fmt.Errorf("创建模块记录失败: %w", err)
		}
	} else {
		// 更新现有记录
		updates := map[string]interface{}{
			"server_url":    setup.ServerURL,
			"api_key":       setup.APIKey,
			"config_hash":   configHash,
			"is_configured": true,
			"last_sync":     &now,
			"status":        models.ModuleStatusOnline,
			"updated_at":    now,
		}

		if err := db.Model(&module).Updates(updates).Error; err != nil {
			return fmt.Errorf("更新模块记录失败: %w", err)
		}
	}

	return nil
}

// GetSystemInfo 获取系统信息
func (ms *ModuleService) GetSystemInfo() map[string]interface{} {
	info := make(map[string]interface{})

	// 获取系统时间
	info["system_time"] = time.Now().Format("2006-01-02 15:04:05")

	// 获取运行时间
	if content, err := os.ReadFile("/proc/uptime"); err == nil {
		uptimeStr := strings.Fields(string(content))[0]
		if uptime, err := strconv.ParseFloat(uptimeStr, 64); err == nil {
			info["uptime"] = int(uptime)
		}
	}

	// 获取内存信息
	if content, err := os.ReadFile("/proc/meminfo"); err == nil {
		lines := strings.Split(string(content), "\n")
		memInfo := make(map[string]string)
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				key := strings.TrimSuffix(fields[0], ":")
				memInfo[key] = fields[1]
			}
		}
		info["memory"] = memInfo
	}

	// 获取负载
	if content, err := os.ReadFile("/proc/loadavg"); err == nil {
		info["load_average"] = strings.TrimSpace(string(content))
	}

	// 获取WireGuard状态
	info["wireguard_running"] = ms.IsWireGuardRunning()
	info["wireguard_configured"] = ms.IsConfigured()

	return info
}

// ValidateConfig 验证配置
func (ms *ModuleService) ValidateConfig(config string) error {
	// 检查配置格式
	if !strings.Contains(config, "[Interface]") {
		return fmt.Errorf("配置缺少 [Interface] 部分")
	}

	if !strings.Contains(config, "[Peer]") {
		return fmt.Errorf("配置缺少 [Peer] 部分")
	}

	// 检查必要字段
	requiredFields := []string{"PrivateKey", "Address", "PublicKey", "Endpoint"}
	for _, field := range requiredFields {
		if !strings.Contains(config, field) {
			return fmt.Errorf("配置缺少必要字段: %s", field)
		}
	}

	return nil
}
