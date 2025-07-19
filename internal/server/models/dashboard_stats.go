package models

// DashboardStats 仪表盘统计数据
type DashboardStats struct {
	TotalModules   int `json:"total_modules"`
	OnlineModules  int `json:"online_modules"`
	OfflineModules int `json:"offline_modules"`
	WarningModules int `json:"warning_modules"`
	TotalTraffic   struct {
		TxBytes uint64 `json:"tx_bytes"`
		RxBytes uint64 `json:"rx_bytes"`
	} `json:"total_traffic"`
}
