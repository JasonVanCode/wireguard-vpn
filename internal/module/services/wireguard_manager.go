package services

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"eitec-vpn/internal/shared/config"
)

// WireGuardManager WireGuard管理器
type WireGuardManager struct {
	config     *config.ModuleConfig
	configPath string
}

// NewWireGuardManager 创建WireGuard管理器
func NewWireGuardManager(cfg *config.ModuleConfig) *WireGuardManager {
	configPath := os.Getenv("WIREGUARD_CONFIG_PATH")
	if configPath == "" {
		configPath = DefaultWireGuardConfigPath
	}

	return &WireGuardManager{
		config:     cfg,
		configPath: configPath,
	}
}

// Start 启动WireGuard
func (wm *WireGuardManager) Start() error {
	log.Println("启动WireGuard...")

	// 停止现有的WireGuard实例
	exec.Command("wg-quick", "down", "wg0").Run()

	// 启动WireGuard
	cmd := exec.Command("wg-quick", "up", "wg0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("启动WireGuard失败: %v, 输出: %s", err, output)
	}

	log.Println("WireGuard启动成功")
	return nil
}

// Stop 停止WireGuard
func (wm *WireGuardManager) Stop() error {
	log.Println("停止WireGuard...")
	if err := exec.Command("wg-quick", "down", "wg0").Run(); err != nil {
		return fmt.Errorf("停止WireGuard失败: %v", err)
	}
	return nil
}

// Restart 重启WireGuard
func (wm *WireGuardManager) Restart() error {
	// 停止WireGuard
	if err := wm.Stop(); err != nil {
		log.Printf("停止WireGuard警告: %v", err)
	}

	// 启动WireGuard
	return wm.Start()
}

// UpdateConfig 更新WireGuard配置
func (wm *WireGuardManager) UpdateConfig(configContent []byte) error {
	// 检查配置是否发生变化
	currentConfig, err := os.ReadFile(wm.configPath)
	if err == nil && string(currentConfig) == string(configContent) {
		// 配置没有变化
		return nil
	}

	// 写入新配置
	if err := os.WriteFile(wm.configPath, configContent, 0600); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	// 重启WireGuard
	log.Println("检测到配置变化，重启WireGuard...")
	if err := wm.Restart(); err != nil {
		return fmt.Errorf("重启WireGuard失败: %w", err)
	}

	log.Println("配置同步成功")
	return nil
}

// GetPublicKeyFromPrivate 从私钥生成公钥
func (wm *WireGuardManager) GetPublicKeyFromPrivate(privateKey string) (string, error) {
	// 临时文件存储私钥
	tmpFile, err := os.CreateTemp("", "wg-private-*")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入私钥
	if _, err := tmpFile.WriteString(privateKey); err != nil {
		return "", fmt.Errorf("写入私钥失败: %w", err)
	}
	tmpFile.Close()

	// 使用wg工具生成公钥
	cmd := exec.Command("wg", "pubkey")
	stdinFile, err := os.Open(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("打开临时文件失败: %w", err)
	}
	defer stdinFile.Close()
	cmd.Stdin = stdinFile

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("生成公钥失败: %w", err)
	}

	// 去除换行符
	publicKey := string(output)
	if len(publicKey) > 0 && publicKey[len(publicKey)-1] == '\n' {
		publicKey = publicKey[:len(publicKey)-1]
	}

	return publicKey, nil
}

// IsRunning 检查WireGuard是否运行
func (wm *WireGuardManager) IsRunning() bool {
	cmd := exec.Command("wg", "show", "wg0")
	err := cmd.Run()
	return err == nil
}

// GetStatus 获取WireGuard状态
func (wm *WireGuardManager) GetStatus() (map[string]interface{}, error) {
	cmd := exec.Command("wg", "show", "wg0", "dump")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取WireGuard状态失败: %w", err)
	}

	// 简化处理，返回原始输出
	return map[string]interface{}{
		"raw_output": string(output),
		"running":    true,
	}, nil
}
