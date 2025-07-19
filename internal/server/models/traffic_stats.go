package models

import "time"

// TrafficStats 流量统计模型
type TrafficStats struct {
	InterfaceName string    `json:"interface_name"`
	TotalRx       uint64    `json:"total_rx"`    // 总接收字节数
	TotalTx       uint64    `json:"total_tx"`    // 总发送字节数
	TotalBytes    uint64    `json:"total_bytes"` // 总流量字节数
	PeerCount     int       `json:"peer_count"`  // Peer数量
	Timestamp     time.Time `json:"timestamp"`   // 统计时间
}

// RealTimeTrafficData 实时流量数据
type RealTimeTrafficData struct {
	Timestamp  time.Time `json:"timestamp"`
	UploadMB   float64   `json:"upload_mb"`   // 上传速度 MB/s
	DownloadMB float64   `json:"download_mb"` // 下载速度 MB/s
}

// TrafficHistoryData 流量历史数据
type TrafficHistoryData struct {
	Interface string                `json:"interface"`
	TimeRange string                `json:"time_range"` // 1h, 6h, 24h
	Data      []RealTimeTrafficData `json:"data"`
}
