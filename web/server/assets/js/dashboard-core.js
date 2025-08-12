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
// - initCharts() - 初始化ECharts图表
// - updateTrafficChart() - 更新流量趋势图
// - updateStatsCards() - 更新统计卡片
// - updateSystemHealth() - 更新系统健康状态
// - switchTimeRange() - 切换时间范围
//
// 📏 文件大小：15.9KB (原文件的 15.2%)
// =====================================================

let trafficChart, statusChart;
let currentTimeRange = '1h';

// 页面加载时初始化
document.addEventListener('DOMContentLoaded', function() {
    initCharts();
    
    // 初始化模块表格为加载状态
    updateModulesTable(null);
    
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

// 初始化图表
function initCharts() {
    // 流量趋势图
    const trafficDiv = document.getElementById('trafficChart');
    trafficChart = echarts.init(trafficDiv);
    
    const trafficOption = {
        backgroundColor: 'transparent',
        tooltip: {
            trigger: 'axis',
            backgroundColor: '#1e293b',
            borderColor: '#334155',
            textStyle: { color: '#f8fafc' }
        },
        legend: {
            data: ['上传', '下载'],
            textStyle: { color: '#cbd5e1' },
            bottom: 5
        },
        grid: {
            left: '12%',
            right: '4%',
            bottom: '12%',
            top: '8%',
            containLabel: true
        },
        xAxis: {
            type: 'category',
            data: [],
            axisLabel: { color: '#cbd5e1' },
            axisLine: { lineStyle: { color: '#334155' } }
        },
        yAxis: {
            type: 'value',
            name: 'MB/s',
            nameTextStyle: { 
                color: '#cbd5e1',
                padding: [0, 0, 0, 15]
            },
            nameGap: 25,
            axisLabel: { color: '#cbd5e1' },
            axisLine: { lineStyle: { color: '#334155' } },
            splitLine: { lineStyle: { color: '#334155' } }
        },
        series: [
            {
                name: '上传',
                type: 'line',
                data: [],
                smooth: true,
                itemStyle: { color: '#10b981' },
                areaStyle: { opacity: 0.3, color: '#10b981' }
            },
            {
                name: '下载',
                type: 'line',
                data: [],
                smooth: true,
                itemStyle: { color: '#3b82f6' },
                areaStyle: { opacity: 0.3, color: '#3b82f6' }
            }
        ]
    };
    
    trafficChart.setOption(trafficOption);

    // 状态分布图
    const statusDiv = document.getElementById('statusChart');
    statusChart = echarts.init(statusDiv);
    
    const statusOption = {
        backgroundColor: 'transparent',
        tooltip: {
            trigger: 'item',
            backgroundColor: '#1e293b',
            borderColor: '#334155',
            textStyle: { color: '#f8fafc' }
        },
        legend: {
            orient: 'vertical',
            right: '5%',
            top: 'center',
            textStyle: { color: '#cbd5e1', fontSize: 12 }
        },
        series: [
            {
                name: '模块状态',
                type: 'pie',
                radius: ['45%', '90%'],
                center: ['42%', '55%'],
                data: [
                    { value: 12, name: '在线', itemStyle: { color: '#10b981' } },
                    { value: 3, name: '离线', itemStyle: { color: '#6b7280' } },
                    { value: 1, name: '故障', itemStyle: { color: '#ef4444' } }
                ],
                label: { 
                    color: '#cbd5e1',
                    fontSize: 11,
                    formatter: '{b}: {c}'
                },
                labelLine: {
                    lineStyle: { color: '#cbd5e1' }
                },
                emphasis: {
                    itemStyle: {
                        shadowBlur: 10,
                        shadowOffsetX: 0,
                        shadowColor: 'rgba(0, 0, 0, 0.5)'
                    }
                }
            }
        ]
    };
    
    statusChart.setOption(statusOption);
    
    // 窗口大小变化时重新调整图表大小
    window.addEventListener('resize', function() {
        trafficChart.resize();
        statusChart.resize();
    });
}

// 生成时间标签
function generateTimeLabels() {
    const labels = [];
    const now = new Date();
    for (let i = 11; i >= 0; i--) {
        const time = new Date(now.getTime() - i * 5 * 60 * 1000);
        labels.push(time.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }));
    }
    return labels;
}

// 获取真实流量数据
async function getTrafficData(timeRange = '1h') {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch(`/api/v1/dashboard/traffic?range=${timeRange}`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.ok) {
            const result = await response.json();
            return result.data || result;
        }
    } catch (error) {
        console.error('获取流量数据失败:', error);
    }
    
    // 如果获取失败，返回空数据
    return {
        time_labels: generateTimeLabels(),
        upload_data: Array.from({ length: 12 }, () => 0),
        download_data: Array.from({ length: 12 }, () => 0),
        total_stats: []
    };
}

// 更新流量图表
async function updateTrafficChart(timeRange) {
    if (!trafficChart) return;
    
    const trafficData = await getTrafficData(timeRange);
    
    const option = trafficChart.getOption();
    option.xAxis[0].data = trafficData.time_labels;
    option.series[0].data = trafficData.upload_data;
    option.series[1].data = trafficData.download_data;
    trafficChart.setOption(option);
    
    // 更新今日流量显示
    if (trafficData.total_stats && trafficData.total_stats.length > 0) {
        let totalBytes = 0;
        trafficData.total_stats.forEach(stat => {
            totalBytes += stat.total_bytes || 0;
        });
        document.getElementById('todayTraffic').textContent = formatBytes(totalBytes);
    }
}

// 加载所有数据
async function loadAllData() {
    try {
        const token = localStorage.getItem('access_token');
        if (!token) {
            window.location.href = '/login';
            return;
        }

        // 并行加载所有数据
        const [statsResponse, healthResponse, interfaceResponse] = await Promise.all([
            fetch('/api/v1/dashboard/stats', { headers: { 'Authorization': `Bearer ${token}` } }),
            fetch('/api/v1/dashboard/health', { headers: { 'Authorization': `Bearer ${token}` } }),
            fetch('/api/v1/interfaces/stats', { headers: { 'Authorization': `Bearer ${token}` } })
        ]);

        if (statsResponse.ok) {
            const stats = await statsResponse.json();
            console.log('Dashboard stats received:', stats);
            updateStatsCards(stats);
            
            // 从统计数据中获取模块列表
            if (stats.data && stats.data.module_list) {
                console.log('Module list from stats.data:', stats.data.module_list);
                updateModulesTable(stats.data.module_list);
            } else if (stats.module_list) {
                console.log('Module list from stats:', stats.module_list);
                updateModulesTable(stats.module_list);
            } else {
                console.log('No module list found in stats response, fetching modules directly...');
                // 如果统计接口没有模块列表，直接调用模块接口
                await loadModulesDirectly();
            }
        } else {
            console.log('Stats response failed, loading modules directly...');
            // 如果统计接口失败，直接加载模块数据
            await loadModulesDirectly();
        }

        if (healthResponse.ok) {
            const health = await healthResponse.json();
            updateSystemHealth(health);
            updateHeaderStatus(health); // 更新头部状态
        }

        if (interfaceResponse.ok) {
            const interfaceStats = await interfaceResponse.json();
            updateSystemInfo(interfaceStats);
        }

        // 加载流量数据并更新图表
        await updateTrafficChart(currentTimeRange);

    } catch (error) {
        console.error('加载数据失败:', error);
        // 即使出错也要确保显示空状态而不是一直loading
        updateModulesTable([]);
    }
}

// 直接加载模块数据的函数
async function loadModulesDirectly() {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/modules', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.ok) {
            const result = await response.json();
            const modules = result.data || result;
            console.log('Direct modules fetch:', modules);
            updateModulesTable(modules);
        } else {
            console.log('Direct modules fetch failed');
            updateModulesTable([]);
        }
    } catch (error) {
        console.error('直接加载模块失败:', error);
        updateModulesTable([]);
    }
}

// 更新头部服务状态
function updateHeaderStatus(health) {
    const data = health.data || health;
    
    // 更新WireGuard状态
    const wgStatus = document.getElementById('headerWgStatus');
    const wgDot = wgStatus.querySelector('.status-dot');
    if (data.wireguard_status === 'running') {
        wgDot.className = 'status-dot status-running';
    } else {
        wgDot.className = 'status-dot status-error';
    }
    
    // 更新数据库状态
    const dbStatus = document.getElementById('headerDbStatus');
    const dbDot = dbStatus.querySelector('.status-dot');
    if (data.database_status === 'connected') {
        dbDot.className = 'status-dot status-normal';
    } else {
        dbDot.className = 'status-dot status-error';
    }
    
    // 更新API状态
    const apiStatus = document.getElementById('headerApiStatus');
    const apiDot = apiStatus.querySelector('.status-dot');
    if (data.api_status === 'healthy') {
        apiDot.className = 'status-dot status-normal';
    } else {
        apiDot.className = 'status-dot status-error';
    }
}

// 更新统计卡片
function updateStatsCards(stats) {
    const data = stats.data || stats; // 处理两种可能的数据结构
    
    // 模块统计数据
    const moduleStats = data.module_stats || {};
    document.getElementById('onlineModules').textContent = moduleStats.online || 0;
    document.getElementById('totalModules').textContent = moduleStats.total || 0;
    
    // 流量统计数据
    const trafficStats = data.traffic_stats || {};
    document.getElementById('todayTraffic').textContent = formatBytes(trafficStats.today_total || 0);
    
    // 系统统计数据 (这里使用占位数据，实际可能需要从health接口获取)
    document.getElementById('systemLoad').textContent = '0.45';
    
    // 更新状态图表
    if (data.status_chart) {
        updateStatusChart(data.status_chart);
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
function updateSystemHealth(health) {
    const data = health.data || health; // 处理两种可能的数据结构
    const systemRes = data.system_resources;
    
    if (systemRes) {
        const cpuUsage = systemRes.cpu_usage || 0;
        const memoryUsage = systemRes.memory_usage || 0;
        const diskUsage = systemRes.disk_usage || 0;

        document.getElementById('cpuUsage').textContent = cpuUsage.toFixed(1) + '%';
        document.getElementById('cpuProgress').style.width = cpuUsage + '%';

        document.getElementById('memoryUsage').textContent = memoryUsage.toFixed(1) + '%';
        document.getElementById('memoryProgress').style.width = memoryUsage + '%';

        document.getElementById('diskUsage').textContent = diskUsage.toFixed(1) + '%';
        document.getElementById('diskProgress').style.width = diskUsage + '%';
    } else {
        // 如果没有系统资源数据，显示占位符
        document.getElementById('cpuUsage').textContent = '--';
        document.getElementById('cpuProgress').style.width = '0%';
        document.getElementById('memoryUsage').textContent = '--';
        document.getElementById('memoryProgress').style.width = '0%';
        document.getElementById('diskUsage').textContent = '--';
        document.getElementById('diskProgress').style.width = '0%';
    }
}

// 更新系统信息显示
function updateSystemInfo(interfaceStats) {
    const data = interfaceStats.data || interfaceStats;
    
    // 更新接口统计信息
    document.getElementById('totalInterfaces').textContent = data.total_interfaces || 0;
    document.getElementById('activeInterfaces').textContent = data.active_interfaces || 0;
    
    // 格式化容量显示
    const usedCapacity = data.used_capacity || 0;
    const totalCapacity = data.total_capacity || 0;
    const capacityText = totalCapacity > 0 ? `${usedCapacity}/${totalCapacity}` : '--';
    document.getElementById('interfaceCapacity').textContent = capacityText;
    
    // 更新网络配置显示
    if (data.interfaces && data.interfaces.length > 0) {
        // 如果有多个接口，显示所有接口的端口信息
        const ports = data.interfaces.map(iface => iface.listen_port).join(', ');
        const networks = data.interfaces.map(iface => iface.network).join(', ');
        const dnsServers = [...new Set(data.interfaces.map(iface => iface.dns).filter(dns => dns))].join(', ');
        
        document.getElementById('networkConfig').textContent = networks || '无接口';
        document.getElementById('portConfig').textContent = ports || '无端口';
        document.getElementById('dnsConfig').textContent = dnsServers || '8.8.8.8';
    } else {
        // 如果没有接口数据，显示提示信息
        document.getElementById('networkConfig').textContent = '暂无接口';
        document.getElementById('portConfig').textContent = '暂无端口';
        document.getElementById('dnsConfig').textContent = '8.8.8.8';
    }
}

// 切换时间范围
async function switchTimeRange(range) {
    currentTimeRange = range;
    document.querySelectorAll('.chart-control-btn').forEach(btn => btn.classList.remove('active'));
    event.target.classList.add('active');
    
    // 更新图表数据
    await updateTrafficChart(range);
}

// 刷新所有数据
async function refreshAllData() {
    await loadAllData();
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
window.switchTimeRange = switchTimeRange;
window.logout = logout; 