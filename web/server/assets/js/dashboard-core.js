// =====================================================
// EITEC VPN - 核心仪表板功能
// =====================================================
//
// 📋 功能概述：
// - 仪表板的核心显示和数据加载逻辑
// - ECharts图表初始化和数据更新
// - 系统状态监控和统计数据展示
//
// 🔗 依赖关系：
// - 依赖：shared-utils.js (工具函数)
// - 依赖：echarts (图表库)
// - 依赖：module-management.js (updateModulesTable函数)
// - 最后加载，协调所有模块的数据显示
//
// 📦 主要功能：
// - loadAllData() - 加载所有仪表板数据
// - initCharts() - 初始化图表（已删除ECharts）
// - updateStatsCards() - 更新统计卡片
// - updateSystemHealth() - 更新系统健康状态
//
// 📏 文件大小：15.9KB (原文件的 15.2%)
// =====================================================



// 页面加载时初始化
document.addEventListener('DOMContentLoaded', function() {
    initCharts();
    
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

// 初始化图表（已删除ECharts相关功能）
function initCharts() {
    // 图表功能已删除，保留函数避免错误
    console.log('图表功能已删除');
}



// 加载所有数据
async function loadAllData() {
    try {
        const token = localStorage.getItem('access_token');
        if (!token) {
            window.location.href = '/login';
            return;
        }

        // 只调用一个API获取所有数据
        const statsResponse = await fetch('/api/v1/dashboard/stats', { 
            headers: { 'Authorization': `Bearer ${token}` } 
        });

        if (statsResponse.ok) {
            const stats = await statsResponse.json();
            console.log('Dashboard stats received:', stats);
            
            // 更新统计卡片
            updateStatsCards(stats);
            
            // 更新系统健康状态
            updateSystemHealth(stats);
            
            // 更新头部状态
            updateHeaderStatus(stats);
            
            // 统计接口已经包含所有必要数据，无需额外调用
            console.log('Dashboard stats loaded successfully');
        } else {
            console.log('Stats response failed, dashboard data unavailable');
        }

        // 刷新接口-模块网格
        if (typeof refreshInterfaceModuleGrid === 'function') {
            refreshInterfaceModuleGrid();
        }

    } catch (error) {
        console.error('加载仪表板数据失败:', error);
        // 显示错误提示
        if (typeof showToast === 'function') {
            showToast('加载数据失败，请检查网络连接', 'error');
        } else {
            console.error('showToast function not available');
        }
    }
}

// 直接加载模块数据的函数（已优化，使用带状态的接口API）
async function loadModulesDirectly() {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/system/wireguard-interfaces', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.ok) {
            const result = await response.json();
            const interfaces = result.data || result;
            console.log('Direct interfaces fetch (with status):', interfaces);
            
            // 从接口数据中提取模块信息
            const allModules = [];
            interfaces.forEach(iface => {
                if (iface.modules && Array.isArray(iface.modules)) {
                    allModules.push(...iface.modules);
                }
            });
            console.log('Extracted modules:', allModules);
        } else {
            console.log('Direct interfaces fetch failed');
        }
    } catch (error) {
        console.error('直接加载接口数据失败:', error);
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

// 更新状态图表
function updateStatusChart(statusData) {
    if (!statusChart) return;
    
    const chartData = statusData.filter(item => item.status !== 'total').map(item => ({
        value: item.count,
        name: getChineseStatusName(item.status),
        itemStyle: { color: item.color }
    }));
    
    const option = statusChart.getOption();
    option.series[0].data = chartData;
    statusChart.setOption(option);
}

// 获取中文状态名称
function getChineseStatusName(status) {
    switch(status) {
        case 'online': return '在线';
        case 'offline': return '离线';
        case 'warning': return '警告';
        case 'unconfigured': return '未配置';
        default: return status;
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

// 更新系统信息显示
function updateSystemInfo(interfaceStats) {
    const data = interfaceStats.data || interfaceStats;
    
    // 更新接口统计信息
    const totalInterfacesEl = document.getElementById('totalInterfaces');
    const activeInterfacesEl = document.getElementById('activeInterfaces');
    
    if (totalInterfacesEl) {
        totalInterfacesEl.textContent = data.total_interfaces || 0;
    }
    if (activeInterfacesEl) {
        activeInterfacesEl.textContent = data.active_interfaces || 0;
    }
    
    // 格式化容量显示
    const usedCapacity = data.used_capacity || 0;
    const totalCapacity = data.total_capacity || 0;
    const capacityText = totalCapacity > 0 ? `${usedCapacity}/${totalCapacity}` : '--';
    
    const interfaceCapacityEl = document.getElementById('interfaceCapacity');
    if (interfaceCapacityEl) {
        interfaceCapacityEl.textContent = capacityText;
    }
    
    // 更新网络配置显示
    if (data.interfaces && data.interfaces.length > 0) {
        // 如果有多个接口，显示所有接口的端口信息
        const ports = data.interfaces.map(iface => iface.listen_port).join(', ');
        const networks = data.interfaces.map(iface => iface.network).join(', ');
        const dnsServers = [...new Set(data.interfaces.map(iface => iface.dns).filter(dns => dns))].join(', ');
        
        const networkConfigEl = document.getElementById('networkConfig');
        const portConfigEl = document.getElementById('portConfig');
        const dnsConfigEl = document.getElementById('dnsConfig');
        
        if (networkConfigEl) networkConfigEl.textContent = networks || '无接口';
        if (portConfigEl) portConfigEl.textContent = ports || '无端口';
        if (dnsConfigEl) dnsConfigEl.textContent = dnsServers || '8.8.8.8';
    } else {
        // 如果没有接口数据，显示提示信息
        const networkConfigEl = document.getElementById('networkConfig');
        const portConfigEl = document.getElementById('portConfig');
        const dnsConfigEl = document.getElementById('dnsConfig');
        
        if (networkConfigEl) networkConfigEl.textContent = '暂无接口';
        if (portConfigEl) portConfigEl.textContent = '暂无端口';
        if (dnsConfigEl) dnsConfigEl.textContent = '8.8.8.8';
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
function logout() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    window.location.href = '/login';
}

// 全局导出核心仪表板函数
window.loadAllData = loadAllData;
window.refreshAllData = refreshAllData;
window.logout = logout; 