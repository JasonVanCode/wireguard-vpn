package services

import (
	"context"
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

	// 检查并处理resolvconf依赖问题
	if err := ms.checkAndFixDNSSupport(); err != nil {
		log.Printf("DNS支持检查警告: %v", err)
	}

	// 停止现有实例（带超时）
	log.Println("检查并停止现有WireGuard实例...")
	if err := ms.stopWireGuardWithTimeout(); err != nil {
		log.Printf("停止现有WireGuard实例时出错: %v", err)
		// 不返回错误，继续尝试启动
	}

	// 启动WireGuard（带超时和重试机制）
	log.Println("开始启动WireGuard...")
	return ms.startWireGuardWithTimeout()
}

// checkAndFixDNSSupport 检查并修复DNS支持
func (ms *ModuleService) checkAndFixDNSSupport() error {
	// 检查resolvconf是否可用
	if !ms.isResolvconfAvailable() {
		log.Println("检测到系统缺少resolvconf，尝试使用替代方案")
		return ms.createDNSFriendlyConfig()
	}
	return nil
}

// isResolvconfAvailable 检查resolvconf是否可用
func (ms *ModuleService) isResolvconfAvailable() bool {
	cmd := exec.Command("which", "resolvconf")
	return cmd.Run() == nil
}

// createDNSFriendlyConfig 创建兼容的配置文件
func (ms *ModuleService) createDNSFriendlyConfig() error {
	configPath := DefaultWireGuardConfigPath

	// 读取当前配置
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	configStr := string(content)

	// 如果配置包含DNS设置，创建一个不包含DNS的备用配置
	if strings.Contains(configStr, "DNS") {
		// 移除DNS行，避免触发resolvconf
		lines := strings.Split(configStr, "\n")
		var newLines []string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "DNS") {
				newLines = append(newLines, line)
			}
		}

		newConfig := strings.Join(newLines, "\n")

		// 创建备用配置文件
		backupPath := configPath + ".nodns"
		if err := os.WriteFile(backupPath, []byte(newConfig), 0600); err != nil {
			return fmt.Errorf("创建备用配置失败: %v", err)
		}

		log.Printf("已创建无DNS配置文件: %s", backupPath)
	}

	return nil
}

// stopWireGuardWithTimeout 带超时的停止WireGuard
func (ms *ModuleService) stopWireGuardWithTimeout() error {
	// 首先检查接口是否存在
	if !ms.wireGuardInterfaceExists() {
		log.Println("WireGuard接口不存在，跳过停止操作")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "wg-quick", "down", "wg0")
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("停止WireGuard超时")
	}

	if err != nil {
		// 检查是否是"接口不存在"的错误
		outputStr := string(output)
		if strings.Contains(outputStr, "is not a WireGuard interface") ||
			strings.Contains(outputStr, "does not exist") {
			log.Println("WireGuard接口已经不存在，无需停止")
			return nil
		}

		// 其他错误，尝试强制清理
		ms.forceCleanupWireGuard()
		return fmt.Errorf("停止WireGuard失败: %v, 输出: %s", err, output)
	}

	return nil
}

// wireGuardInterfaceExists 检查WireGuard接口是否存在
func (ms *ModuleService) wireGuardInterfaceExists() bool {
	// 方法1: 使用 wg show 命令检查
	cmd := exec.Command("wg", "show", "wg0")
	if err := cmd.Run(); err == nil {
		return true
	}

	// 方法2: 使用 ip link show 命令检查
	cmd = exec.Command("ip", "link", "show", "wg0")
	if err := cmd.Run(); err == nil {
		return true
	}

	return false
}

// startWireGuardWithTimeout 带超时的启动WireGuard
func (ms *ModuleService) startWireGuardWithTimeout() error {
	configPath := DefaultWireGuardConfigPath

	// 首先尝试使用wg-quick启动
	if err := ms.attemptStartWithConfig(configPath); err != nil {
		log.Printf("使用wg-quick启动失败: %v", err)
		log.Println("尝试使用手动方式启动WireGuard")

		// 尝试手动启动（不依赖wg-quick）
		if err := ms.manualStartWireGuard(configPath); err != nil {
			log.Printf("手动启动也失败: %v", err)

			// 如果失败且存在备用配置，尝试使用备用配置
			backupPath := configPath + ".nodns"
			if _, statErr := os.Stat(backupPath); statErr == nil {
				log.Println("尝试使用无DNS配置手动启动")

				// 临时替换配置文件
				if err := ms.useBackupConfig(configPath, backupPath); err != nil {
					return fmt.Errorf("切换到备用配置失败: %v", err)
				}

				// 尝试手动启动
				if err := ms.manualStartWireGuard(configPath); err != nil {
					return fmt.Errorf("使用备用配置手动启动也失败: %v", err)
				}

				// 启动成功后手动设置DNS
				go ms.setDNSManually()

				log.Println("WireGuard已使用备用配置手动启动，DNS将手动设置")
				return nil
			}

			return err
		}

		log.Println("WireGuard手动启动成功")
		return nil
	}

	return nil
}

// attemptStartWithConfig 尝试使用指定配置启动
func (ms *ModuleService) attemptStartWithConfig(configPath string) error {
	log.Printf("尝试启动WireGuard，配置文件: %s", configPath)

	// 先检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 减少超时时间，快速发现问题
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("执行命令: wg-quick up wg0")
	cmd := exec.CommandContext(ctx, "wg-quick", "up", "wg0")

	// 启动命令并监控进度
	done := make(chan error, 1)
	go func() {
		output, err := cmd.CombinedOutput()
		if err != nil {
			done <- fmt.Errorf("命令执行失败: %v, 输出: %s", err, output)
		} else {
			log.Printf("wg-quick执行成功，输出: %s", output)
			done <- nil
		}
	}()

	// 等待命令完成或超时
	select {
	case err := <-done:
		if err != nil {
			log.Printf("WireGuard启动失败: %v", err)
			return err
		}
		log.Println("wg-quick命令执行完成")

	case <-ctx.Done():
		log.Println("WireGuard启动超时，尝试终止进程")
		// 尝试终止进程
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		// 清理可能的残留接口
		ms.forceCleanupWireGuard()
		return fmt.Errorf("启动WireGuard超时(15秒)")
	}

	// 验证接口是否真的启动了
	log.Println("验证WireGuard接口状态")
	if !ms.IsWireGuardRunning() {
		return fmt.Errorf("WireGuard启动命令成功但接口未运行")
	}

	log.Println("WireGuard启动成功并验证完成")
	return nil
}

// manualStartWireGuard 手动启动WireGuard（不依赖wg-quick）
func (ms *ModuleService) manualStartWireGuard(configPath string) error {
	log.Println("开始手动启动WireGuard")

	// 解析配置文件
	config, err := ms.parseWireGuardConfig(configPath)
	if err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 1. 创建网络接口
	log.Println("步骤1: 创建WireGuard接口")
	if err := ms.runCommandWithTimeout("ip", []string{"link", "add", "wg0", "type", "wireguard"}, 5); err != nil {
		return fmt.Errorf("创建接口失败: %v", err)
	}

	// 2. 设置私钥和端口
	log.Println("步骤2: 配置私钥")
	if err := ms.setWireGuardPrivateKey(config.PrivateKey); err != nil {
		ms.runCommandWithTimeout("ip", []string{"link", "delete", "wg0"}, 3)
		return fmt.Errorf("设置私钥失败: %v", err)
	}

	// 3. 配置IP地址
	log.Println("步骤3: 设置IP地址")
	if err := ms.runCommandWithTimeout("ip", []string{"-4", "address", "add", config.Address, "dev", "wg0"}, 5); err != nil {
		ms.runCommandWithTimeout("ip", []string{"link", "delete", "wg0"}, 3)
		return fmt.Errorf("设置IP地址失败: %v", err)
	}

	// 4. 设置MTU
	if config.MTU != "" {
		log.Printf("步骤4: 设置MTU为%s", config.MTU)
		if err := ms.runCommandWithTimeout("ip", []string{"link", "set", "mtu", config.MTU, "dev", "wg0"}, 5); err != nil {
			log.Printf("设置MTU失败: %v", err) // 不是致命错误
		}
	}

	// 5. 启动接口
	log.Println("步骤5: 启动网络接口")
	if err := ms.runCommandWithTimeout("ip", []string{"link", "set", "up", "dev", "wg0"}, 5); err != nil {
		ms.runCommandWithTimeout("ip", []string{"link", "delete", "wg0"}, 3)
		return fmt.Errorf("启动接口失败: %v", err)
	}

	// 6. 配置Peer（如果有）
	if config.PeerPublicKey != "" {
		log.Println("步骤6: 配置Peer")
		if err := ms.configurePeer(config); err != nil {
			ms.runCommandWithTimeout("ip", []string{"link", "delete", "wg0"}, 3)
			return fmt.Errorf("配置Peer失败: %v", err)
		}
	}

	// 7. 验证接口状态
	log.Println("步骤7: 验证接口状态")
	if !ms.IsWireGuardRunning() {
		ms.runCommandWithTimeout("ip", []string{"link", "delete", "wg0"}, 3)
		return fmt.Errorf("接口创建成功但未运行")
	}

	log.Println("WireGuard手动启动成功")
	return nil
}

// 简化的配置结构
type WireGuardConfig struct {
	PrivateKey     string
	Address        string
	MTU            string
	PeerPublicKey  string
	PeerEndpoint   string
	PeerAllowedIPs string
}

// parseWireGuardConfig 解析WireGuard配置文件
func (ms *ModuleService) parseWireGuardConfig(configPath string) (*WireGuardConfig, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := &WireGuardConfig{}
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "PrivateKey") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				config.PrivateKey = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Address") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				config.Address = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "MTU") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				config.MTU = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "PublicKey") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				config.PeerPublicKey = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Endpoint") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				config.PeerEndpoint = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "AllowedIPs") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				config.PeerAllowedIPs = strings.TrimSpace(parts[1])
			}
		}
	}

	return config, nil
}

// runCommandWithTimeout 运行带超时的命令
func (ms *ModuleService) runCommandWithTimeout(command string, args []string, timeoutSec int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("命令超时: %s %v", command, args)
	}

	if err != nil {
		return fmt.Errorf("命令失败: %s %v, 错误: %v, 输出: %s", command, args, err, output)
	}

	log.Printf("命令成功: %s %v", command, args)
	return nil
}

// setWireGuardPrivateKey 设置WireGuard私钥
func (ms *ModuleService) setWireGuardPrivateKey(privateKey string) error {
	if privateKey == "" {
		return fmt.Errorf("私钥为空")
	}

	// 使用echo和管道传递私钥给wg命令
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("echo '%s' | wg set wg0 private-key /dev/stdin", privateKey))
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("设置私钥失败: %v, 输出: %s", err, output)
	}

	return nil
}

// configurePeer 配置Peer
func (ms *ModuleService) configurePeer(config *WireGuardConfig) error {
	if config.PeerPublicKey == "" {
		return nil // 没有Peer配置
	}

	args := []string{"set", "wg0", "peer", config.PeerPublicKey}

	if config.PeerEndpoint != "" {
		args = append(args, "endpoint", config.PeerEndpoint)
	}

	if config.PeerAllowedIPs != "" {
		args = append(args, "allowed-ips", config.PeerAllowedIPs)
	}

	return ms.runCommandWithTimeout("wg", args, 5)
}

// useBackupConfig 使用备用配置
func (ms *ModuleService) useBackupConfig(originalPath, backupPath string) error {
	// 备份原始配置
	originalBackup := originalPath + ".original"
	if err := exec.Command("cp", originalPath, originalBackup).Run(); err != nil {
		return fmt.Errorf("备份原始配置失败: %v", err)
	}

	// 使用备用配置
	if err := exec.Command("cp", backupPath, originalPath).Run(); err != nil {
		return fmt.Errorf("复制备用配置失败: %v", err)
	}

	return nil
}

// setDNSManually 手动设置DNS
func (ms *ModuleService) setDNSManually() {
	time.Sleep(2 * time.Second) // 等待接口完全启动

	// 尝试多种DNS设置方法
	ms.trySetDNSMethods()
}

// trySetDNSMethods 尝试各种DNS设置方法
func (ms *ModuleService) trySetDNSMethods() {
	dnsServers := []string{"8.8.8.8", "8.8.4.4"}

	// 方法1: 尝试使用systemd-resolved
	if ms.trySystemdResolved(dnsServers) {
		log.Println("DNS已通过systemd-resolved设置")
		return
	}

	// 方法2: 尝试直接修改/etc/resolv.conf (临时)
	if ms.tryDirectResolvConf(dnsServers) {
		log.Println("DNS已通过直接修改resolv.conf设置")
		return
	}

	// 方法3: 使用ip命令设置路由
	if ms.tryIPRoute(dnsServers) {
		log.Println("DNS路由已设置")
		return
	}

	log.Println("警告: 无法自动设置DNS，可能需要手动配置")
}

// trySystemdResolved 尝试使用systemd-resolved
func (ms *ModuleService) trySystemdResolved(dnsServers []string) bool {
	for _, dns := range dnsServers {
		cmd := exec.Command("systemd-resolve", "--set-dns", dns, "--interface", "wg0")
		if err := cmd.Run(); err == nil {
			return true
		}

		// 尝试新版本命令
		cmd = exec.Command("resolvectl", "dns", "wg0", dns)
		if err := cmd.Run(); err == nil {
			return true
		}
	}
	return false
}

// tryDirectResolvConf 尝试直接修改resolv.conf
func (ms *ModuleService) tryDirectResolvConf(dnsServers []string) bool {
	// 这种方法风险较高，仅作为最后手段
	return false
}

// tryIPRoute 尝试设置IP路由
func (ms *ModuleService) tryIPRoute(dnsServers []string) bool {
	// 为DNS服务器添加路由
	for _, dns := range dnsServers {
		cmd := exec.Command("ip", "route", "add", dns, "dev", "wg0")
		cmd.Run() // 忽略错误，因为路由可能已存在
	}
	return true
}

// forceCleanupWireGuard 强制清理WireGuard接口
func (ms *ModuleService) forceCleanupWireGuard() {
	// 尝试删除接口
	exec.Command("ip", "link", "delete", "wg0").Run()

	// 清理可能的iptables规则
	exec.Command("iptables", "-D", "FORWARD", "-i", "wg0", "-j", "ACCEPT").Run()
	exec.Command("iptables", "-D", "FORWARD", "-o", "wg0", "-j", "ACCEPT").Run()
	exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-o", "eth0", "-j", "MASQUERADE").Run()
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
