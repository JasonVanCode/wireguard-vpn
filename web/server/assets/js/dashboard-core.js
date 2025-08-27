// =====================================================
// EITEC VPN - 核心仪表板功能
// =====================================================
//
// 📋 功能概述：
// - 仪表板的核心显示和数据加载逻辑
// - 系统状态监控和统计数据展示
// - 接口-模块网格数据刷新
//
// 🔗 依赖关系：
// - 依赖：shared-utils.js (工具函数)
// - 依赖：api.js (统一API管理)
// - 依赖：interface-management.js (接口-模块网格)
// - 最后加载，协调所有模块的数据显示
//
// 📦 主要功能：
// - loadAllData() - 加载所有仪表板数据
// - updateStatsCards() - 更新统计卡片
// - updateSystemHealth() - 更新系统健康状态
// - updateHeaderStatus() - 更新头部状态
// - updateTime() - 更新时间显示
//
// 📏 文件大小：约8KB (优化后)
// =====================================================



// 页面加载时初始化
document.addEventListener('DOMContentLoaded', function() {
    // 初始化接口-模块网格
    if (typeof renderInterfaceModuleGrid === 'function') {
        renderInterfaceModuleGrid();
    }
    
    loadAllData();
    updateTime();
    
    // 每30秒刷新一次数据
    setInterval(loadAllData, 30000);
    // 每秒更新时间
    setInterval(updateTime, 1000);
});

// 更新时间显示
function updateTime() {
    const now = new Date();
    const timeString = now.toLocaleTimeString('zh-CN', { hour12: false });
    document.getElementById('currentTime').textContent = timeString;
}





// 加载所有数据
async function loadAllData() {
    try {
        // 检查认证状态
        if (!api.auth.isAuthenticated()) {
            window.location.href = '/login';
            return;
        }

        apiHelper.showLoading('加载仪表板数据...');

        // 使用新的API管理系统获取仪表板统计数据
        const stats = await api.system.getDashboardStats();
        console.log('Dashboard stats received:', stats);
        
        // 更新统计卡片
        updateStatsCards(stats);
        
        // 更新系统健康状态
        updateSystemHealth(stats);
        
        // 更新头部状态
        updateHeaderStatus(stats);
        
        console.log('Dashboard stats loaded successfully');
        
        // 刷新接口-模块网格
        if (typeof refreshInterfaceModuleGrid === 'function') {
            refreshInterfaceModuleGrid();
        }

        apiHelper.hideLoading();

    } catch (error) {
        console.error('加载仪表板数据失败:', error);
        apiHelper.hideLoading();
        
        // 显示错误提示
        if (typeof showToast === 'function') {
            showToast('加载数据失败，请检查网络连接', 'error');
        } else {
            apiHelper.handleError(error, '加载仪表板数据失败');
        }
    }
}



// 更新头部服务状态
function updateHeaderStatus(stats) {
    const data = stats.data || stats;
    const serviceStatus = data.service_status || {};
    
    // 更新WireGuard状态
    const wgStatus = document.getElementById('headerWgStatus');
    if (wgStatus) {
        const wgDot = wgStatus.querySelector('.status-dot');
        if (wgDot) {
            if (serviceStatus.wireguard_status === 'ok' || serviceStatus.wireguard_status === 'running') {
                wgDot.className = 'status-dot status-running';
            } else {
                wgDot.className = 'status-dot status-error';
            }
        }
    }
    
    // 更新数据库状态
    const dbStatus = document.getElementById('headerDbStatus');
    if (dbStatus) {
        const dbDot = dbStatus.querySelector('.status-dot');
        if (dbDot) {
            if (serviceStatus.database_status === 'ok' || serviceStatus.database_status === 'connected') {
                dbDot.className = 'status-dot status-normal';
            } else {
                dbDot.className = 'status-dot status-error';
            }
        }
    }
    
    // 更新API状态
    const apiStatus = document.getElementById('headerApiStatus');
    if (apiStatus) {
        const apiDot = apiStatus.querySelector('.status-dot');
        if (apiDot) {
            if (serviceStatus.api_status === 'ok' || serviceStatus.api_status === 'healthy') {
                apiDot.className = 'status-dot status-normal';
            } else {
                apiDot.className = 'status-dot status-error';
            }
        }
    }
}

// 更新统计卡片
function updateStatsCards(stats) {
    const data = stats.data || stats; // 处理两种可能的数据结构
    
    // 模块统计数据
    const moduleStats = data.module_stats || {};
    const moduleStatusEl = document.getElementById('moduleStatus');
    
    if (moduleStatusEl) {
        const online = moduleStats.online || 0;
        const total = moduleStats.total || 0;
        moduleStatusEl.textContent = `${online}/${total}`;
    }
    
    // 系统资源数据
    const systemResources = data.system_resources || {};
    const cpuUsageEl = document.getElementById('cpuUsage');
    const memoryUsageEl = document.getElementById('memoryUsage');
    const diskUsageEl = document.getElementById('diskUsage');
    
    if (cpuUsageEl) {
        const cpuUsage = systemResources.cpu_usage || 0;
        cpuUsageEl.textContent = cpuUsage.toFixed(1) + '%';
    }
    
    if (memoryUsageEl) {
        const memoryUsage = systemResources.memory_usage || 0;
        memoryUsageEl.textContent = memoryUsage.toFixed(1) + '%';
    }
    
    if (diskUsageEl) {
        const diskUsage = systemResources.disk_usage || 0;
        diskUsageEl.textContent = diskUsage.toFixed(1) + '%';
    }
}



// 更新系统健康状态
function updateSystemHealth(stats) {
    const data = stats.data || stats; // 处理两种可能的数据结构
    const systemRes = data.system_resources;
    
    if (systemRes) {
        const cpuUsage = systemRes.cpu_usage || 0;
        const memoryUsage = systemRes.memory_usage || 0;
        const diskUsage = systemRes.disk_usage || 0;

        // 更新CPU使用率（如果元素存在）
        const cpuUsageEl = document.getElementById('cpuUsage');
        if (cpuUsageEl) cpuUsageEl.textContent = cpuUsage.toFixed(1) + '%';

        // 更新内存使用率（如果元素存在）
        const memoryUsageEl = document.getElementById('memoryUsage');
        if (memoryUsageEl) memoryUsageEl.textContent = memoryUsage.toFixed(1) + '%';

        // 更新磁盘使用率（如果元素存在）
        const diskUsageEl = document.getElementById('diskUsage');
        if (diskUsageEl) diskUsageEl.textContent = diskUsage.toFixed(1) + '%';
    }
}





// 刷新所有数据
async function refreshAllData() {
    await loadAllData();
    
    // 额外刷新接口-模块网格
    if (typeof refreshInterfaceModuleGrid === 'function') {
        refreshInterfaceModuleGrid();
    }
}

// 退出登录
async function logout() {
    try {
        // 调用后端登出API
        await api.auth.logout();
    } catch (error) {
        console.warn('登出API调用失败，继续本地清理:', error);
    } finally {
        // 清理本地认证信息
        api.auth.clearAuth();
        // 跳转到登录页
        window.location.href = '/login';
    }
}

// 全局导出核心仪表板函数
window.loadAllData = loadAllData;
window.refreshAllData = refreshAllData;
window.logout = logout; 