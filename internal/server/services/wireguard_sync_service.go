package services

import (
	"fmt"
	"time"

	"eitec-vpn/internal/server/database"
	"eitec-vpn/internal/server/models"
	"eitec-vpn/internal/shared/wireguard"

	"gorm.io/gorm"
)

// WireGuardSyncService WireGuard状态同步服务
type WireGuardSyncService struct {
	db *gorm.DB
}

// NewWireGuardSyncService 创建WireGuard状态同步服务
func NewWireGuardSyncService() *WireGuardSyncService {
	return &WireGuardSyncService{
		db: database.DB,
	}
}

// SyncAllInterfaces 同步所有接口的状态和流量数据
func (wss *WireGuardSyncService) SyncAllInterfaces() error {
	// 获取所有接口
	var interfaces []models.WireGuardInterface
	if err := wss.db.Find(&interfaces).Error; err != nil {
		return fmt.Errorf("查询接口列表失败: %w", err)
	}

	fmt.Printf("🔄 [WireGuard同步] 开始同步 %d 个接口的状态\n", len(interfaces))

	for _, iface := range interfaces {
		if err := wss.SyncInterface(iface.Name, iface.ID); err != nil {
			fmt.Printf("❌ [WireGuard同步] 同步接口 %s 失败: %v\n", iface.Name, err)
			continue
		}
	}

	fmt.Printf("✅ [WireGuard同步] 所有接口同步完成\n")
	return nil
}

// SyncInterface 同步单个接口的状态
func (wss *WireGuardSyncService) SyncInterface(interfaceName string, interfaceID uint) error {
	fmt.Printf("🔍 [WireGuard同步] 开始同步接口: %s\n", interfaceName)

	// 获取WireGuard实际状态
	peers, err := wireguard.GetWireGuardStatus(interfaceName)
	if err != nil {
		return fmt.Errorf("获取WireGuard状态失败: %w", err)
	}

	fmt.Printf("📊 [WireGuard同步] 接口 %s 发现 %d 个活跃peer\n", interfaceName, len(peers))

	// 更新接口连接数和状态
	var connectedPeers int
	var totalRx, totalTx uint64

	for publicKey, peer := range peers {
		connectedPeers++
		totalRx += peer.TransferRxBytes
		totalTx += peer.TransferTxBytes

		fmt.Printf("🔗 [WireGuard同步] Peer: %s..., RX: %d bytes, TX: %d bytes\n",
			publicKey[:20], peer.TransferRxBytes, peer.TransferTxBytes)

		// 更新模块状态
		if err := wss.updateModuleStatus(publicKey, peer); err != nil {
			fmt.Printf("⚠️  [WireGuard同步] 更新模块状态失败: %v\n", err)
		}

		// 更新用户VPN状态
		if err := wss.updateUserVPNStatus(publicKey, peer); err != nil {
			fmt.Printf("⚠️  [WireGuard同步] 更新用户VPN状态失败: %v\n", err)
		}
	}

	// 更新接口统计信息
	updateData := map[string]interface{}{
		"total_peers":    connectedPeers,
		"active_peers":   connectedPeers,
		"total_traffic":  totalRx + totalTx,
		"last_heartbeat": time.Now(),
		"status":         models.InterfaceStatusUp,
	}

	if err := wss.db.Model(&models.WireGuardInterface{}).Where("id = ?", interfaceID).Updates(updateData).Error; err != nil {
		return fmt.Errorf("更新接口统计失败: %w", err)
	}

	fmt.Printf("✅ [WireGuard同步] 接口 %s 同步完成，连接数: %d，总流量: %d bytes\n",
		interfaceName, connectedPeers, totalRx+totalTx)

	return nil
}

// updateModuleStatus 更新模块状态
func (wss *WireGuardSyncService) updateModuleStatus(publicKey string, peer wireguard.WireGuardPeer) error {
	var module models.Module
	if err := wss.db.Where("public_key = ?", publicKey).First(&module).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 这个公钥可能是用户VPN，不是模块
			return nil
		}
		return err
	}

	// 判断状态
	status := models.ModuleStatusOnline
	if time.Since(peer.LatestHandshake) > 5*time.Minute {
		status = models.ModuleStatusWarning
	}

	// 更新模块数据
	updateData := map[string]interface{}{
		"status":           status,
		"latest_handshake": peer.LatestHandshake,
		"total_rx_bytes":   peer.TransferRxBytes,
		"total_tx_bytes":   peer.TransferTxBytes,
		"last_seen":        time.Now(),
	}

	if err := wss.db.Model(&module).Updates(updateData).Error; err != nil {
		return err
	}

	fmt.Printf("📦 [模块同步] 模块 %s 状态更新: %s, 流量 RX:%d TX:%d\n",
		module.Name, status.String(), peer.TransferRxBytes, peer.TransferTxBytes)

	return nil
}

// updateUserVPNStatus 更新用户VPN状态
func (wss *WireGuardSyncService) updateUserVPNStatus(publicKey string, peer wireguard.WireGuardPeer) error {
	var userVPN models.UserVPN
	if err := wss.db.Where("public_key = ?", publicKey).First(&userVPN).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 这个公钥可能是模块，不是用户VPN
			return nil
		}
		return err
	}

	// 判断状态
	status := models.UserVPNStatusOnline
	if time.Since(peer.LatestHandshake) > 3*time.Minute {
		status = models.UserVPNStatusOffline
	}

	// 更新用户VPN数据
	updateData := map[string]interface{}{
		"status":         status,
		"total_rx_bytes": peer.TransferRxBytes,
		"total_tx_bytes": peer.TransferTxBytes,
		"last_seen":      time.Now(),
	}

	if err := wss.db.Model(&userVPN).Updates(updateData).Error; err != nil {
		return err
	}

	fmt.Printf("👤 [用户VPN同步] 用户 %s 状态更新: %s, 流量 RX:%d TX:%d\n",
		userVPN.Username, status.String(), peer.TransferRxBytes, peer.TransferTxBytes)

	return nil
}

// MarkOfflinePeers 标记离线的peer
func (wss *WireGuardSyncService) MarkOfflinePeers() error {
	// 标记长时间未更新的模块为离线
	if err := wss.db.Model(&models.Module{}).
		Where("last_seen < ? AND status != ?", time.Now().Add(-10*time.Minute), models.ModuleStatusOffline).
		Update("status", models.ModuleStatusOffline).Error; err != nil {
		return fmt.Errorf("标记离线模块失败: %w", err)
	}

	// 标记长时间未更新的用户VPN为离线
	if err := wss.db.Model(&models.UserVPN{}).
		Where("last_seen < ? AND status != ?", time.Now().Add(-5*time.Minute), models.UserVPNStatusOffline).
		Update("status", models.UserVPNStatusOffline).Error; err != nil {
		return fmt.Errorf("标记离线用户VPN失败: %w", err)
	}

	return nil
}

// InterfaceStats 接口统计信息
type InterfaceStats struct {
	InterfaceName string    `json:"interface_name"`
	TotalPeers    int       `json:"total_peers"`
	OnlinePeers   int       `json:"online_peers"`
	TotalRxBytes  uint64    `json:"total_rx_bytes"`
	TotalTxBytes  uint64    `json:"total_tx_bytes"`
	LastUpdated   time.Time `json:"last_updated"`
}

// GetInterfaceRealTimeStats 获取接口实时统计
func (wss *WireGuardSyncService) GetInterfaceRealTimeStats(interfaceName string) (*InterfaceStats, error) {
	peers, err := wireguard.GetWireGuardStatus(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("获取WireGuard状态失败: %w", err)
	}

	var totalRx, totalTx uint64
	var onlinePeers, totalPeers int

	for _, peer := range peers {
		totalPeers++
		totalRx += peer.TransferRxBytes
		totalTx += peer.TransferTxBytes

		// 3分钟内有握手认为在线
		if time.Since(peer.LatestHandshake) <= 3*time.Minute {
			onlinePeers++
		}
	}

	return &InterfaceStats{
		InterfaceName: interfaceName,
		TotalPeers:    totalPeers,
		OnlinePeers:   onlinePeers,
		TotalRxBytes:  totalRx,
		TotalTxBytes:  totalTx,
		LastUpdated:   time.Now(),
	}, nil
}
