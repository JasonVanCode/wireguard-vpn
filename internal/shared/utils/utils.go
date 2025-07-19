package utils

import (
	"crypto/rand"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 密码加密
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes), nil
}

// ValidateIP 验证IP地址格式
func ValidateIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidIP 验证IP地址格式 (别名)
func IsValidIP(ip string) bool {
	return ValidateIP(ip)
}

// ValidateCIDR 验证CIDR格式
func ValidateCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

// IsValidCIDR 验证CIDR格式 (别名)
func IsValidCIDR(cidr string) bool {
	return ValidateCIDR(cidr)
}

// ValidatePort 验证端口号
func ValidatePort(port int) bool {
	return port > 0 && port <= 65535
}

// IsValidPort 验证端口号 (别名)
func IsValidPort(port int) bool {
	return ValidatePort(port)
}

// IsValidEndpoint 验证端点格式 (IP:Port 或 Domain:Port)
func IsValidEndpoint(endpoint string) bool {
	parts := strings.Split(endpoint, ":")
	if len(parts) != 2 {
		return false
	}

	host := parts[0]
	portStr := parts[1]

	// 验证端口
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return false
	}

	// 验证主机 (IP或域名)
	if net.ParseIP(host) != nil {
		return true // 有效IP
	}

	// 简单域名验证
	return len(host) > 0 && !strings.Contains(host, " ")
}

// IsValidDNSList 验证DNS服务器列表格式
func IsValidDNSList(dns string) bool {
	if dns == "" {
		return true
	}

	servers := strings.Split(dns, ",")
	for _, server := range servers {
		server = strings.TrimSpace(server)
		if net.ParseIP(server) == nil {
			return false
		}
	}
	return true
}

// CompareIPs 比较两个IP地址
func CompareIPs(ip1, ip2 string) int {
	a := net.ParseIP(ip1)
	b := net.ParseIP(ip2)
	if a == nil || b == nil {
		return 0
	}

	a4 := a.To4()
	b4 := b.To4()
	if a4 == nil || b4 == nil {
		return 0
	}

	for i := 0; i < 4; i++ {
		if a4[i] < b4[i] {
			return -1
		}
		if a4[i] > b4[i] {
			return 1
		}
	}
	return 0
}

// ValidatePortString 验证端口号字符串
func ValidatePortString(portStr string) bool {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return false
	}
	return ValidatePort(port)
}

// ValidateWireGuardKey 验证WireGuard密钥格式
func ValidateWireGuardKey(key string) bool {
	// WireGuard密钥是44字符的base64编码
	if len(key) != 44 {
		return false
	}
	// 基本正则验证base64格式
	matched, _ := regexp.MatchString(`^[A-Za-z0-9+/]{43}=$`, key)
	return matched
}

// ParsePagination 解析分页参数
func ParsePagination(r *http.Request) (page, size int) {
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err = strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		size = 10
	}

	return page, size
}

// TimeAgo 时间差描述
func TimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return fmt.Sprintf("%d秒前", int(diff.Seconds()))
	} else if diff < time.Hour {
		return fmt.Sprintf("%d分钟前", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%d小时前", int(diff.Hours()))
	} else if diff < 30*24*time.Hour {
		return fmt.Sprintf("%d天前", int(diff.Hours()/24))
	} else {
		return t.Format("2006-01-02")
	}
}

// FormatBytes 格式化字节数
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// IsValidEmail 验证邮箱格式
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	return emailRegex.MatchString(strings.ToLower(email))
}

// Sanitize 清理字符串，移除特殊字符
func Sanitize(input string) string {
	// 移除可能危险的字符
	reg := regexp.MustCompile(`[<>'"&\x00-\x1f\x7f-\x9f]`)
	return reg.ReplaceAllString(input, "")
}

// TrimSpaces 清理多余空格
func TrimSpaces(input string) string {
	// 清理首尾空格并将多个空格合并为一个
	spaceRegex := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(spaceRegex.ReplaceAllString(input, " "))
}

// Contains 检查字符串切片是否包含指定元素
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ParseDuration 解析时间长度字符串
func ParseDuration(s string) (time.Duration, error) {
	// 支持格式: "1h", "30m", "45s", "1h30m", "2d" 等
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

// GetClientIP 获取客户端真实IP
func GetClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 检查X-Real-IP头
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// 使用RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// MergeMaps 合并两个map
func MergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range map1 {
		result[k] = v
	}

	for k, v := range map2 {
		result[k] = v
	}

	return result
}

// FindWebPath 查找web目录路径
// 这个函数会在多个可能的路径中查找web资源目录
// 适用于从不同工作目录运行程序的场景
func FindWebPath(subPath string) string {
	// 可能的路径列表
	possiblePaths := []string{
		filepath.Join(".", "web", subPath),     // 在项目根目录执行
		filepath.Join("..", "web", subPath),    // 在bin目录执行
		filepath.Join("../..", "web", subPath), // 在更深的目录执行
		filepath.Join("web", subPath),          // 当前目录下的web
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 如果都找不到，返回默认路径
	return filepath.Join("web", subPath)
}

// GetProjectRoot 获取项目根目录
// 基于程序执行文件的位置推断项目根目录
func GetProjectRoot() (string, error) {
	// 获取程序执行文件的路径
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取程序执行路径失败: %v", err)
	}

	// 获取程序目录的父目录（项目根目录）
	// 假设程序在 bin/ 目录下，项目根目录是 bin/ 的父目录
	projectRoot := filepath.Dir(filepath.Dir(execPath))

	return projectRoot, nil
}

// GetAbsolutePath 获取相对路径的绝对路径
func GetAbsolutePath(relativePath string) (string, error) {
	root, err := GetProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, relativePath), nil
}

// EnsureDir 确保目录存在，如果不存在则创建
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// EnsureDirForFile 确保文件所在目录存在
func EnsureDirForFile(filePath string) error {
	dir := filepath.Dir(filePath)
	return EnsureDir(dir)
}
