package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"eitec-vpn/internal/shared/config"
)

// ServerClient 服务器客户端
type ServerClient struct {
	config     *config.ModuleConfig
	httpClient *http.Client
}

// NewServerClient 创建服务器客户端
func NewServerClient(cfg *config.ModuleConfig) *ServerClient {
	return &ServerClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// sendRequest 发送HTTP请求的通用方法
func (sc *ServerClient) sendRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", sc.config.Server.URL, endpoint)

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求数据失败: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("X-API-Key", sc.config.Module.APIKey)
	req.Header.Set("Content-Type", "application/json")

	return sc.httpClient.Do(req)
}

// SendHeartbeat 发送心跳
func (sc *ServerClient) SendHeartbeat() error {
	endpoint := fmt.Sprintf("/api/v1/modules/%d/heartbeat", sc.config.Module.ID)

	resp, err := sc.sendRequest("POST", endpoint, nil)
	if err != nil {
		return fmt.Errorf("发送心跳请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("服务器返回错误 %d: %s", resp.StatusCode, body)
	}

	return nil
}

// ReportTraffic 上报流量统计
func (sc *ServerClient) ReportTraffic(stats TrafficStats) error {
	endpoint := fmt.Sprintf("/api/v1/modules/%d/traffic", sc.config.Module.ID)

	resp, err := sc.sendRequest("POST", endpoint, stats)
	if err != nil {
		return fmt.Errorf("上报流量统计请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("服务器返回错误 %d: %s", resp.StatusCode, body)
	}

	return nil
}

// GetConfiguration 获取最新配置
func (sc *ServerClient) GetConfiguration() ([]byte, error) {
	endpoint := fmt.Sprintf("/api/v1/modules/%d/config", sc.config.Module.ID)

	resp, err := sc.sendRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("获取配置请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("服务器返回错误 %d: %s", resp.StatusCode, body)
	}

	return io.ReadAll(resp.Body)
}
