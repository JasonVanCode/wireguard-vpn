package services

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// CronScheduler 定时任务调度器
type CronScheduler struct {
	cron        *cron.Cron
	syncService *WireGuardSyncService
}

// NewCronScheduler 创建定时任务调度器
func NewCronScheduler() *CronScheduler {
	// 创建cron实例，支持秒级精度
	c := cron.New(cron.WithSeconds())

	return &CronScheduler{
		cron:        c,
		syncService: NewWireGuardSyncService(),
	}
}

// Start 启动定时任务调度器
func (cs *CronScheduler) Start() error {
	fmt.Println("🚀 [定时调度器] 正在启动定时任务...")

	// 1. 每10秒同步WireGuard状态 - 高频率，确保实时性
	_, err := cs.cron.AddFunc("*/10 * * * * *", func() {
		fmt.Printf("🔄 [定时同步] 开始同步WireGuard状态 - %s\n", time.Now().Format("15:04:05"))
		if err := cs.syncService.SyncAllInterfaces(); err != nil {
			fmt.Printf("❌ [定时同步] WireGuard同步失败: %v\n", err)
		} else {
			fmt.Printf("✅ [定时同步] WireGuard同步完成 - %s\n", time.Now().Format("15:04:05"))
		}
	})
	if err != nil {
		return fmt.Errorf("添加WireGuard同步任务失败: %w", err)
	}

	// 2. 每30秒标记离线peer - 中频率，清理过时连接
	_, err = cs.cron.AddFunc("*/30 * * * * *", func() {
		fmt.Printf("🧹 [定时清理] 开始标记离线peer - %s\n", time.Now().Format("15:04:05"))
		if err := cs.syncService.MarkOfflinePeers(); err != nil {
			fmt.Printf("❌ [定时清理] 标记离线peer失败: %v\n", err)
		} else {
			fmt.Printf("✅ [定时清理] 离线peer标记完成 - %s\n", time.Now().Format("15:04:05"))
		}
	})
	if err != nil {
		return fmt.Errorf("添加离线标记任务失败: %w", err)
	}

	// 3. 每5分钟进行完整的状态同步 - 低频率，确保数据一致性
	_, err = cs.cron.AddFunc("0 */5 * * * *", func() {
		fmt.Printf("🔧 [完整同步] 开始完整状态同步 - %s\n", time.Now().Format("15:04:05"))
		if err := cs.fullStatusSync(); err != nil {
			fmt.Printf("❌ [完整同步] 完整同步失败: %v\n", err)
		} else {
			fmt.Printf("✅ [完整同步] 完整同步完成 - %s\n", time.Now().Format("15:04:05"))
		}
	})
	if err != nil {
		return fmt.Errorf("添加完整同步任务失败: %w", err)
	}

	// 4. 每小时清理统计数据 - 数据维护
	_, err = cs.cron.AddFunc("0 0 * * * *", func() {
		fmt.Printf("📊 [数据维护] 开始数据维护任务 - %s\n", time.Now().Format("15:04:05"))
		if err := cs.dataMaintenanceTask(); err != nil {
			fmt.Printf("❌ [数据维护] 数据维护失败: %v\n", err)
		} else {
			fmt.Printf("✅ [数据维护] 数据维护完成 - %s\n", time.Now().Format("15:04:05"))
		}
	})
	if err != nil {
		return fmt.Errorf("添加数据维护任务失败: %w", err)
	}

	// 启动调度器
	cs.cron.Start()
	fmt.Println("✅ [定时调度器] 定时任务调度器已启动")
	fmt.Println("📋 [定时调度器] 任务列表:")
	fmt.Println("   • WireGuard状态同步: 每10秒")
	fmt.Println("   • 离线peer标记: 每30秒")
	fmt.Println("   • 完整状态同步: 每5分钟")
	fmt.Println("   • 数据维护: 每小时")

	return nil
}

// Stop 停止定时任务调度器
func (cs *CronScheduler) Stop() {
	fmt.Println("🛑 [定时调度器] 正在停止定时任务...")
	cs.cron.Stop()
	fmt.Println("✅ [定时调度器] 定时任务调度器已停止")
}

// GetRunningJobs 获取正在运行的任务列表
func (cs *CronScheduler) GetRunningJobs() []cron.Entry {
	return cs.cron.Entries()
}

// fullStatusSync 完整状态同步
func (cs *CronScheduler) fullStatusSync() error {
	// 1. 同步WireGuard状态
	if err := cs.syncService.SyncAllInterfaces(); err != nil {
		return fmt.Errorf("同步接口状态失败: %w", err)
	}

	// 2. 标记离线peer
	if err := cs.syncService.MarkOfflinePeers(); err != nil {
		return fmt.Errorf("标记离线peer失败: %w", err)
	}

	// 3. 更新接口状态统计
	if err := cs.updateInterfaceStats(); err != nil {
		return fmt.Errorf("更新接口统计失败: %w", err)
	}

	return nil
}

// updateInterfaceStats 更新接口统计信息
func (cs *CronScheduler) updateInterfaceStats() error {
	fmt.Printf("📊 [统计更新] 更新接口统计信息\n")

	// 这里可以添加更多的统计逻辑
	// 比如计算平均连接时间、流量趋势等

	return nil
}

// dataMaintenanceTask 数据维护任务
func (cs *CronScheduler) dataMaintenanceTask() error {
	fmt.Printf("🧹 [数据维护] 执行数据清理和优化\n")

	// 可以添加以下维护任务：
	// 1. 清理过期的连接记录
	// 2. 压缩历史流量数据
	// 3. 更新统计缓存
	// 4. 数据库优化

	return nil
}

// AddCustomJob 添加自定义定时任务
func (cs *CronScheduler) AddCustomJob(spec string, jobFunc func()) (cron.EntryID, error) {
	return cs.cron.AddFunc(spec, jobFunc)
}

// RemoveJob 移除定时任务
func (cs *CronScheduler) RemoveJob(id cron.EntryID) {
	cs.cron.Remove(id)
}

// GetStats 获取调度器统计信息
func (cs *CronScheduler) GetStats() map[string]interface{} {
	entries := cs.cron.Entries()

	stats := map[string]interface{}{
		"total_jobs":   len(entries),
		"is_running":   len(entries) > 0,
		"last_updated": time.Now(),
	}

	// 添加各个任务的下次执行时间
	var jobStats []map[string]interface{}
	for i, entry := range entries {
		jobStats = append(jobStats, map[string]interface{}{
			"job_id":   i + 1,
			"next_run": entry.Next,
			"prev_run": entry.Prev,
		})
	}
	stats["jobs"] = jobStats

	return stats
}
