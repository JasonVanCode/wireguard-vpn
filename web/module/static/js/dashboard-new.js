// EITEC VPN 模块管理中心 - 仪表板JavaScript

// 全局变量
let dashboardData = {};
let refreshInterval = null;
let charts = {};

// 当前选择的接口
let currentInterface = 'wg0';

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    console.log('初始化仪表板...');
    try {
        initializeDashboard();
        startAutoRefresh();
    } catch (error) {
        console.error('ERROR: 初始化失败', error);
    }
});

// 初始化仪表板
async function initializeDashboard() {
    try {
        // 先加载WireGuard详情
        getWireGuardDetails();
        
        // 等待3秒后再加载仪表板数据，避免API请求冲突
        setTimeout(async () => {
            await refreshDashboardData();
            initializeCharts();
        }, 3000);
    } catch (error) {
        console.error('初始化仪表板失败:', error);
        showNotification('初始化失败', 'error');
    }
}

// 刷新仪表板数据
async function refreshDashboardData() {
    try {
        showLoadingState();
        
        // 使用封装的API
        const data = await DashboardAPI.getStats();
        
        // 如果API返回null（通常是认证问题），直接返回
        if (data === null) {
            hideLoadingState();
            return;
        }
        
        dashboardData = data;
        updateAllCards();
        hideLoadingState();
        console.log('仪表板数据更新成功');
    } catch (error) {
        console.error('刷新仪表板失败:', error);
        hideLoadingState();
        showNotification('刷新失败: ' + error.message, 'error');
    }
}

// 更新所有卡片
function updateAllCards() {
    console.log('开始更新所有卡片，数据:', dashboardData);
    
    if (dashboardData.vpn_status) {
        console.log('更新VPN状态:', dashboardData.vpn_status);
        updateVPNStatus(dashboardData.vpn_status);
    }
    
    if (dashboardData.traffic_stats) {
        console.log('更新流量统计:', dashboardData.traffic_stats);
        updateTrafficStats(dashboardData.traffic_stats);
    }
    
    if (dashboardData.network_metrics) {
        console.log('更新网络指标:', dashboardData.network_metrics);
        updateNetworkMetrics(dashboardData.network_metrics);
    }
    
    if (dashboardData.system_status) {
        console.log('更新系统状态:', dashboardData.system_status);
        updateSystemStatus(dashboardData.system_status);
    }
    
    // 更新连接状态信息
    updateConnectionInfo();
    
    // 更新最后更新时间
    updateLastUpdated();
}

// 更新VPN状态卡片
function updateVPNStatus(vpnStatus) {
    const statusElement = document.getElementById('vpnStatus');
    const detailElement = document.getElementById('vpnDetail');
    const qualityElement = document.getElementById('connectionQuality');
    const latencyElement = document.getElementById('latency');

    console.log('VPN状态数据:', vpnStatus); // 调试日志

    if (statusElement) {
        // 根据状态显示文本和样式
        let statusText, statusClass;
        
        switch (vpnStatus.status) {
            case 'running':
                statusText = '已连接';
                statusClass = 'main-value success';
                break;
            case 'stopped':
                statusText = '未连接';
                statusClass = 'main-value danger';
                break;
            case 'configured':
                statusText = '已配置';
                statusClass = 'main-value warning';
                break;
            default:
                statusText = '未知状态';
                statusClass = 'main-value muted';
        }
        
        statusElement.textContent = statusText;
        statusElement.className = statusClass;
    }
    
    if (detailElement) {
        const quality = vpnStatus.connection_quality || '未知';
        detailElement.textContent = `质量: ${quality}`;
    }
    
    if (qualityElement) {
        const quality = vpnStatus.connection_quality || '未知';
        qualityElement.textContent = quality;
        
        // 根据质量设置颜色
        if (quality === 'excellent') {
            qualityElement.className = 'value success';
        } else if (quality === 'good') {
            qualityElement.className = 'value info';
        } else if (quality === 'fair') {
            qualityElement.className = 'value warning';
        } else if (quality === 'poor') {
            qualityElement.className = 'value danger';
        } else if (quality === 'disconnected') {
            qualityElement.className = 'value danger';
        } else {
            qualityElement.className = 'value muted';
        }
    }
    
    if (latencyElement) {
        if (vpnStatus.latency > 0) {
            latencyElement.textContent = `${vpnStatus.latency}ms`;
            
            // 根据延迟设置颜色
            if (vpnStatus.latency < 50) {
                latencyElement.className = 'value success';
            } else if (vpnStatus.latency < 100) {
                latencyElement.className = 'value info';
            } else if (vpnStatus.latency < 200) {
                latencyElement.className = 'value warning';
            } else {
                latencyElement.className = 'value danger';
            }
        } else {
            latencyElement.textContent = '--';
            latencyElement.className = 'value muted';
        }
    }
    
    // 根据VPN状态更新按钮显示
    updateVPNButtons(vpnStatus.status);
    
    // 更新顶部导航栏状态
    updateTopStatus(vpnStatus.status);
}

// 更新流量统计卡片
function updateTrafficStats(trafficStats) {
    const uploadElement = document.getElementById('uploadTraffic');
    const downloadElement = document.getElementById('downloadTraffic');

    if (uploadElement) {
        if (trafficStats.tx_bytes > 0) {
            uploadElement.textContent = formatBytes(trafficStats.tx_bytes);
            uploadElement.className = 'value success';
        } else {
            uploadElement.textContent = '0 B';
            uploadElement.className = 'value muted';
        }
    }
    
    if (downloadElement) {
        if (trafficStats.rx_bytes > 0) {
            downloadElement.textContent = formatBytes(trafficStats.rx_bytes);
            downloadElement.className = 'value info';
        } else {
            downloadElement.textContent = '0 B';
            downloadElement.className = 'value muted';
        }
    }

    // 更新流量图表
    updateTrafficChart(trafficStats);
}

// 更新网络性能卡片
function updateNetworkMetrics(networkMetrics) {
    const latencyElement = document.getElementById('networkLatency');
    const packetLossElement = document.getElementById('packetLoss');
    const bandwidthElement = document.getElementById('bandwidth');
    const qualityElement = document.getElementById('networkQuality');
    const statusElement = document.getElementById('networkStatus');

    if (latencyElement) {
        if (networkMetrics.latency > 0) {
            latencyElement.textContent = `${networkMetrics.latency}ms`;
            
            // 根据延迟设置颜色
            if (networkMetrics.latency < 50) {
                latencyElement.className = 'metric-value success';
            } else if (networkMetrics.latency < 100) {
                latencyElement.className = 'metric-value info';
            } else if (networkMetrics.latency < 200) {
                latencyElement.className = 'metric-value warning';
            } else {
                latencyElement.className = 'metric-value danger';
            }
        } else {
            latencyElement.textContent = '--';
            latencyElement.className = 'metric-value muted';
        }
    }
    
    if (packetLossElement) {
        if (networkMetrics.packet_loss > 0) {
            packetLossElement.textContent = `${networkMetrics.packet_loss.toFixed(2)}%`;
            
            // 根据丢包率设置颜色
            if (networkMetrics.packet_loss < 1) {
                packetLossElement.className = 'metric-value success';
            } else if (networkMetrics.packet_loss < 5) {
                packetLossElement.className = 'metric-value info';
            } else if (networkMetrics.packet_loss < 10) {
                packetLossElement.className = 'metric-value warning';
            } else {
                packetLossElement.className = 'metric-value danger';
            }
        } else {
            packetLossElement.textContent = '0%';
            packetLossElement.className = 'metric-value success';
        }
    }
    
    if (bandwidthElement) {
        if (networkMetrics.bandwidth > 0) {
            bandwidthElement.textContent = `${networkMetrics.bandwidth.toFixed(2)}Mbps`;
            bandwidthElement.className = 'metric-value info';
        } else {
            bandwidthElement.textContent = '--';
            bandwidthElement.className = 'metric-value muted';
        }
    }
    
    if (qualityElement) {
        const quality = networkMetrics.quality || '未知';
        qualityElement.textContent = quality;
        
        // 根据质量设置颜色
        if (quality === 'excellent') {
            qualityElement.className = 'status success';
        } else if (quality === 'good') {
            qualityElement.className = 'status info';
        } else if (quality === 'fair') {
            qualityElement.className = 'status warning';
        } else if (quality === 'poor') {
            qualityElement.className = 'status danger';
        } else {
            qualityElement.className = 'status muted';
        }
    }
    
    if (statusElement) {
        const status = networkMetrics.status || '未知';
        statusElement.textContent = status;
        
        // 根据状态设置颜色
        if (status === 'connected') {
            statusElement.className = 'status success';
        } else if (status === 'disconnected') {
            statusElement.className = 'status danger';
        } else {
            statusElement.className = 'status muted';
        }
    }
}

// 更新系统资源卡片
function updateSystemStatus(systemStatus) {
    const cpuElement = document.getElementById('cpuUsage');
    const memoryElement = document.getElementById('memoryUsage');
    const diskElement = document.getElementById('diskUsage');

    if (cpuElement) {
        // 注意：这里应该显示CPU使用率，但后端返回的是内存使用率
        // 暂时使用内存使用率，后续需要后端提供真正的CPU使用率
        const cpuPercent = systemStatus.memory.percent;
        cpuElement.textContent = `${cpuPercent.toFixed(1)}%`;
        
        // 根据使用率设置颜色
        if (cpuPercent < 50) {
            cpuElement.className = 'resource-value success';
        } else if (cpuPercent < 80) {
            cpuElement.className = 'resource-value warning';
        } else {
            cpuElement.className = 'resource-value danger';
        }
    }
    
    if (memoryElement) {
        const memPercent = systemStatus.memory.percent;
        memoryElement.textContent = `${memPercent.toFixed(1)}%`;
        
        // 根据使用率设置颜色
        if (memPercent < 50) {
            memoryElement.className = 'resource-value success';
        } else if (memPercent < 80) {
            memoryElement.className = 'resource-value warning';
        } else {
            memoryElement.className = 'resource-value danger';
        }
    }
    
    if (diskElement) {
        const diskPercent = systemStatus.disk.percent;
        diskElement.textContent = `${diskPercent.toFixed(1)}%`;
        
        // 根据使用率设置颜色
        if (diskPercent < 50) {
            diskElement.className = 'resource-value success';
        } else if (diskPercent < 80) {
            diskElement.className = 'resource-value warning';
        } else {
            diskElement.className = 'resource-value danger';
        }
    }

    // 更新系统资源图表
    updateSystemCharts(systemStatus);
}

// 初始化图表
function initializeCharts() {
    // 流量图表
    const trafficCtx = document.getElementById('trafficChart');
    if (trafficCtx) {
        charts.traffic = new Chart(trafficCtx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: '上传',
                    data: [],
                    borderColor: '#10b981',
                    backgroundColor: 'rgba(16, 185, 129, 0.1)',
                    tension: 0.4
                }, {
                    label: '下载',
                    data: [],
                    borderColor: '#3b82f6',
                    backgroundColor: 'rgba(59, 130, 246, 0.1)',
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        display: false
                    },
                    x: {
                        display: false
                    }
                }
            }
        });
    }

    // 系统资源图表
    initializeSystemCharts();
}

// 初始化系统资源图表
function initializeSystemCharts() {
    const chartIds = ['cpuChart', 'memoryChart', 'diskChart'];
    
    chartIds.forEach(id => {
        const ctx = document.getElementById(id);
        if (ctx) {
            charts[id] = new Chart(ctx, {
                type: 'doughnut',
                data: {
                    datasets: [{
                        data: [0, 100],
                        backgroundColor: ['#3b82f6', '#1e293b'],
                        borderWidth: 0
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    cutout: '70%',
                    plugins: {
                        legend: {
                            display: false
                        }
                    }
                }
            });
        }
    });
}

// 更新流量图表
function updateTrafficChart(trafficStats) {
    if (!charts.traffic) return;

    const now = new Date();
    const timeLabel = now.toLocaleTimeString('zh-CN', { hour12: false });

    // 添加新数据点
    charts.traffic.data.labels.push(timeLabel);
    charts.traffic.data.datasets[0].data.push(trafficStats.tx_bytes / 1024 / 1024); // 转换为MB
    charts.traffic.data.datasets[1].data.push(trafficStats.rx_bytes / 1024 / 1024);

    // 保持最近20个数据点
    if (charts.traffic.data.labels.length > 20) {
        charts.traffic.data.labels.shift();
        charts.traffic.data.datasets[0].data.shift();
        charts.traffic.data.datasets[1].data.shift();
    }

    charts.traffic.update();
}

// 更新系统资源图表
function updateSystemCharts(systemStatus) {
    // CPU图表
    if (charts.cpuChart) {
        const cpuPercent = systemStatus.memory.percent;
        charts.cpuChart.data.datasets[0].data = [cpuPercent, 100 - cpuPercent];
        charts.cpuChart.update();
    }

    // 内存图表
    if (charts.memoryChart) {
        const memPercent = systemStatus.memory.percent;
        charts.memoryChart.data.datasets[0].data = [memPercent, 100 - memPercent];
        charts.memoryChart.update();
    }

    // 磁盘图表
    if (charts.diskChart) {
        const diskPercent = systemStatus.disk.percent;
        charts.diskChart.data.datasets[0].data = [diskPercent, 100 - diskPercent];
        charts.diskChart.update();
    }
}

// 工具函数
function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}



function showLoadingState() {
    // 显示加载状态
    const loadingElements = document.querySelectorAll('.loading-state');
    loadingElements.forEach(el => el.style.display = 'flex');
}

function hideLoadingState() {
    // 隐藏加载状态
    const loadingElements = document.querySelectorAll('.loading-state');
    loadingElements.forEach(el => el.style.display = 'none');
}

function showNotification(message, type = 'info') {
    // 简单的通知显示
    console.log(`${type.toUpperCase()}: ${message}`);
}

// 自动刷新
function startAutoRefresh() {
    // 默认30秒刷新间隔
    const refreshTime = 30000;
    refreshInterval = setInterval(refreshDashboardData, refreshTime);
    console.log(`自动刷新已启动，间隔: ${refreshTime}ms`);
}

// 更新当前时间
function updateCurrentTime() {
    const timeElement = document.getElementById('currentTime');
    if (timeElement) {
        const now = new Date();
        timeElement.textContent = now.toLocaleTimeString('zh-CN', { hour12: false });
    }
}

// 事件处理函数
async function refreshDashboard() {
    // 刷新仪表板数据
    await refreshDashboardData();
    // 更新WireGuard详情
    await getWireGuardDetails();
}

function restartVPN() {
    // 实现重启VPN功能
    showNotification('重启VPN功能待实现', 'info');
}

function resetTraffic() {
    // 实现重置流量统计功能
    showNotification('重置流量统计功能待实现', 'info');
}

function refreshNetworkMetrics() {
    // 实现刷新网络指标功能
    showNotification('刷新网络指标功能待实现', 'info');
}

function showSystemDetails() {
    // 实现显示系统详细信息功能
    showNotification('系统详细信息功能待实现', 'info');
}

function showQuickActions() {
    const overlay = document.getElementById('quickActionsOverlay');
    if (overlay) {
        overlay.classList.add('show');
    }
}

function hideQuickActions() {
    const overlay = document.getElementById('quickActionsOverlay');
    if (overlay) {
        overlay.classList.remove('show');
    }
}

function showSettings() {
    const overlay = document.getElementById('settingsOverlay');
    if (overlay) {
        overlay.classList.add('show');
    }
}

function hideSettings() {
    const overlay = document.getElementById('settingsOverlay');
    if (overlay) {
        overlay.classList.remove('show');
    }
}

function showConfigUploadModal() {
    const overlay = document.getElementById('configUploadOverlay');
    if (overlay) {
        overlay.classList.add('show');
    }
}

function hideConfigUploadModal() {
    const overlay = document.getElementById('configUploadOverlay');
    if (overlay) {
        overlay.classList.remove('show');
    }
}

function logout() {
    // 使用封装的认证管理器清除token
    AuthManager.clearToken();
    
    // 重定向到登录页面
    window.location.href = '/login';
}

// 面板控制
function togglePanel(panelName) {
    const panel = document.querySelector(`[data-panel="${panelName}"]`);
    if (panel) {
        const content = panel.querySelector('.panel-content');
        const toggleBtn = panel.querySelector('.toggle i');
        
        if (content.style.display === 'none') {
            content.style.display = 'block';
            toggleBtn.className = 'fas fa-chevron-up';
        } else {
            content.style.display = 'none';
            toggleBtn.className = 'fas fa-chevron-down';
        }
    }
}

// 配置文件上传相关函数
function selectConfigFile() {
    document.getElementById('configFileInput').click();
}

// 接口相关函数
async function changeInterface() {
    const select = document.getElementById('interfaceSelect');
    currentInterface = select.value;
    console.log('切换到接口:', currentInterface);
    
    // 刷新仪表板数据
    await refreshDashboardData();
    // 更新WireGuard详情
    await getWireGuardDetails();
}

// WireGuard控制函数
function startVPN() {
    controlWireGuard('start', currentInterface);
}

function stopVPN() {
    controlWireGuard('stop', currentInterface);
}

function restartVPN() {
    controlWireGuard('restart', currentInterface);
}

// 控制WireGuard接口
async function controlWireGuard(action, interfaceName) {
    try {
        await WireGuardAPI.control(action, interfaceName);
        showSuccess(`${action}操作成功`);
        // 延迟刷新状态
        setTimeout(async () => {
            // 刷新仪表板数据
            await refreshDashboardData();
            // 更新WireGuard详情
            await getWireGuardDetails();
        }, 1000);
    } catch (error) {
        console.error('控制WireGuard失败:', error);
        showError('操作失败: ' + error.message);
    }
}

// 更新applyConfig函数，支持接口选择
function applyConfig() {
    const configContent = document.getElementById('previewContent').textContent;
    const interfaceName = document.getElementById('configInterface').value;
    
    if (!configContent.trim()) {
        showError('没有配置内容');
        return;
    }

    uploadWireGuardConfig(interfaceName, configContent);
}

// 上传WireGuard配置
async function uploadWireGuardConfig(interfaceName, configData) {
    try {
        await WireGuardAPI.uploadConfig(interfaceName, configData);
        showSuccess('配置上传成功');
        hideConfigUploadModal();
        
        // 等待配置生效
        setTimeout(async () => {
            // 刷新仪表板数据
            await refreshDashboardData();
            // 更新WireGuard详情
            await getWireGuardDetails();
        }, 1000);
    } catch (error) {
        console.error('配置上传失败:', error);
        showError('配置上传失败: ' + error.message);
    }
}

// 显示成功消息
function showSuccess(message) {
    // 这里可以实现一个简单的成功提示
    console.log('成功:', message);
    alert('成功: ' + message); // 临时使用alert，后续可以美化
}

// 显示错误消息
function showError(message) {
    // 这里可以实现一个简单的错误提示
    console.error('错误:', message);
    alert('错误: ' + message); // 临时使用alert，后续可以美化
}

function clearConfig() {
    document.getElementById('configPreview').style.display = 'none';
    document.getElementById('configDropZone').style.display = 'block';
}

// 文件拖拽处理
document.addEventListener('DOMContentLoaded', function() {
    const dropZone = document.getElementById('configDropZone');
    const fileInput = document.getElementById('configFileInput');
    
    if (dropZone) {
        dropZone.addEventListener('click', () => fileInput.click());
        
        dropZone.addEventListener('dragover', (e) => {
            e.preventDefault();
            dropZone.classList.add('drag-over');
        });
        
        dropZone.addEventListener('dragleave', () => {
            dropZone.classList.remove('drag-over');
        });
        
        dropZone.addEventListener('drop', (e) => {
            e.preventDefault();
            dropZone.classList.remove('drag-over');
            
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                handleConfigFile(files[0]);
            }
        });
    }
    
    if (fileInput) {
        fileInput.addEventListener('change', (e) => {
            if (e.target.files.length > 0) {
                handleConfigFile(e.target.files[0]);
            }
        });
    }
});

function handleConfigFile(file) {
    if (!file.name.endsWith('.conf')) {
        showNotification('请选择.conf格式的配置文件', 'error');
        return;
    }
    
    const reader = new FileReader();
    reader.onload = function(e) {
        const content = e.target.result;
        showConfigPreview(content);
    };
    reader.readAsText(file);
}

function showConfigPreview(content) {
    const preview = document.getElementById('configPreview');
    const previewContent = document.getElementById('previewContent');
    const dropZone = document.getElementById('configDropZone');
    
    if (preview && previewContent && dropZone) {
        previewContent.textContent = content;
        preview.style.display = 'block';
        dropZone.style.display = 'none';
    }
}

// 更新连接状态信息
function updateConnectionInfo() {
    if (dashboardData.vpn_status && dashboardData.system_status) {
        const serverInfoElement = document.getElementById('serverInfo');
        const uptimeElement = document.getElementById('uptimeInfo');
        
        if (serverInfoElement) {
            const status = dashboardData.vpn_status.status;
            const uptime = dashboardData.vpn_status.uptime;
            
            if (status === 'running') {
                serverInfoElement.textContent = `VPN运行中 • ${uptime}`;
                serverInfoElement.className = 'success';
            } else if (status === 'stopped') {
                serverInfoElement.textContent = 'VPN已停止';
                serverInfoElement.className = 'danger';
            } else if (status === 'configured') {
                serverInfoElement.textContent = 'VPN已配置';
                serverInfoElement.className = 'warning';
            } else {
                serverInfoElement.textContent = 'VPN状态未知';
                serverInfoElement.className = 'muted';
            }
        }
        
        if (uptimeElement) {
            const uptime = dashboardData.system_status.uptime;
            if (uptime > 0) {
                const uptimeText = formatUptime(uptime);
                uptimeElement.textContent = uptimeText;
            } else {
                uptimeElement.textContent = '--';
            }
        }
    }
}

// 更新最后更新时间
function updateLastUpdated() {
    if (dashboardData.last_updated) {
        const lastUpdatedElement = document.getElementById('lastUpdated');
        if (lastUpdatedElement) {
            const date = new Date(dashboardData.last_updated);
            const timeString = date.toLocaleTimeString('zh-CN', { hour12: false });
            lastUpdatedElement.textContent = `最后更新: ${timeString}`;
        }
    }
}

// 格式化运行时间
function formatUptime(seconds) {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (days > 0) {
        return `${days}天${hours}小时`;
    } else if (hours > 0) {
        return `${hours}小时${minutes}分钟`;
    } else {
        return `${minutes}分钟`;
    }
}

// 根据VPN状态更新按钮显示
function updateVPNButtons(vpnStatus) {
    const startBtn = document.getElementById('startBtn');
    const stopBtn = document.getElementById('stopBtn');
    
    if (!startBtn || !stopBtn) {
        console.warn('VPN控制按钮未找到');
        return;
    }
    
    // 根据状态只显示一个按钮
    switch (vpnStatus) {
        case 'running':
            // VPN运行中，只显示停止按钮
            startBtn.style.display = 'none';
            stopBtn.style.display = 'inline-block';
            stopBtn.textContent = '停止';
            stopBtn.className = 'control-btn stop';
            break;
        case 'stopped':
        case 'configured':
            // VPN停止或已配置，只显示启动按钮
            startBtn.style.display = 'inline-block';
            stopBtn.style.display = 'none';
            startBtn.textContent = '启动';
            startBtn.className = 'control-btn start';
            break;
        default:
            // 未知状态，隐藏所有按钮
            startBtn.style.display = 'none';
            stopBtn.style.display = 'none';
    }
    
    console.log(`VPN状态: ${vpnStatus}, 显示按钮: ${vpnStatus === 'running' ? '停止' : '启动'}`);
}

// 更新顶部导航栏状态
function updateTopStatus(vpnStatus) {
    const statusText = document.getElementById('statusText');
    const statusDot = document.querySelector('#mainStatus .status-dot');
    
    if (!statusText || !statusDot) {
        console.warn('顶部状态元素未找到');
        return;
    }
    
    // 根据VPN状态更新顶部显示
    switch (vpnStatus) {
        case 'running':
            statusText.textContent = '已连接';
            statusText.className = 'success';
            statusDot.className = 'status-dot connected';
            break;
        case 'stopped':
            statusText.textContent = '未连接';
            statusText.className = 'danger';
            statusDot.className = 'status-dot disconnected';
            break;
        case 'configured':
            statusText.textContent = '已配置';
            statusText.className = 'warning';
            statusDot.className = 'status-dot configured';
            break;
        default:
            statusText.textContent = '状态未知';
            statusText.className = 'muted';
            statusDot.className = 'status-dot unknown';
    }
    
    console.log(`顶部状态已更新: ${vpnStatus}`);
}

// WireGuard详情相关函数
function refreshWireGuardDetails() {
    getWireGuardDetails();
}

// 获取WireGuard详细信息
async function getWireGuardDetails() {
    try {
        console.log('开始获取WireGuard详情...');
        
        const data = await DashboardAPI.getStatus();
        
        // 如果API返回null（通常是认证问题），直接返回
        if (data === null) {
            console.warn('获取状态信息失败，可能是认证问题');
            return;
        }
        
        console.log('API响应数据:', data);
        
        console.log('开始更新WireGuard详情...');
        console.log('完整响应数据:', data);
        
        // 从status接口获取WireGuard状态
        const wireguardStatus = data.wireguard;
        const moduleInfo = data.module;
        
        if (wireguardStatus) {
            // 合并WireGuard状态和模块信息
            const combinedData = {
                ...wireguardStatus,
                module: moduleInfo
            };
            updateWireGuardDetails(combinedData);
            console.log('WireGuard详情更新完成');
        } else {
            console.warn('未找到WireGuard状态数据');
        }
    } catch (error) {
        console.error('获取WireGuard详情失败:', error);
    }
}

// 更新WireGuard详情显示
function updateWireGuardDetails(wgData) {
    console.log('更新WireGuard详情:', wgData);
    
    // 更新接口信息
    updateElement('wgInterface', wgData.interface || 'wg0');
    updateElement('wgStatus', getStatusText(wgData.status));
    updateElement('wgPublicKey', wgData.public_key || '--');
    updateElement('wgLanIP', wgData.module?.ip_address || '10.10.0.7/32');
    updateElement('wgListenPort', wgData.listen_port || '--');
    
    console.log('WireGuard接口信息:', {
        interface: wgData.interface,
        status: wgData.status,
        public_key: wgData.public_key,
        listen_port: wgData.listen_port
    });
    
    // 更新节点和连接统计
    if (wgData.peers && wgData.peers.length > 0) {
        const peer = wgData.peers[0]; // 取第一个对等节点
        
        console.log('WireGuard对端信息:', {
            public_key: peer.public_key,
            endpoint: peer.endpoint,
            allowed_ips: peer.allowed_ips,
            persistent_keepalive: peer.persistent_keepalive,
            transfer_rx: peer.transfer_rx,
            transfer_tx: peer.transfer_tx,
            latest_handshake: peer.latest_handshake
        });
        
        updateElement('wgNode', peer.public_key || '--');
        updateElement('wgPresharedKey', '已启用'); // 固定值
        updateElement('wgEndpoint', peer.endpoint || '--');
        updateElement('wgRoutedIPs', peer.allowed_ips ? peer.allowed_ips.join(', ') : (wgData.module?.allowed_ips || '--'));
        updateElement('wgKeepalive', peer.persistent_keepalive ? `每隔 ${peer.persistent_keepalive} 秒` : (wgData.module?.persistent_ka ? `每隔 ${wgData.module.persistent_ka} 秒` : '--'));
        updateElement('wgReceivedData', formatBytes(peer.transfer_rx || 0));
        updateElement('wgSentData', formatBytes(peer.transfer_tx || 0));
        updateElement('wgLatestHandshake', formatHandshakeTime(peer.latest_handshake));
    } else {
        // 如果没有对等节点，显示默认值
        updateElement('wgNode', '--');
        updateElement('wgPresharedKey', '--');
        updateElement('wgEndpoint', '--');
        updateElement('wgRoutedIPs', '--');
        updateElement('wgKeepalive', '--');
        updateElement('wgReceivedData', '--');
        updateElement('wgSentData', '--');
        updateElement('wgLatestHandshake', '--');
    }
    
    // 更新操作按钮
    updateWireGuardButton(wgData.status);
}

// 更新元素内容
function updateElement(elementId, value) {
    const element = document.getElementById(elementId);
    if (element) {
        element.textContent = value;
        
        // 根据状态设置样式
        if (elementId === 'wgStatus') {
            element.className = 'value status-with-dot ' + getStatusClass(value);
        }
    }
}

// 获取状态文本
function getStatusText(status) {
    switch (status) {
        case 'running': return '激活';
        case 'stopped': return '未激活';
        case 'configured': return '已配置';
        default: return '未知';
    }
}

// 获取状态样式类
function getStatusClass(status) {
    switch (status) {
        case '激活': return 'success';
        case '未激活': return 'danger';
        case '已配置': return 'warning';
        default: return 'muted';
    }
}

// 更新WireGuard操作按钮
function updateWireGuardButton(status) {
    const actionBtn = document.getElementById('wgActionBtn');
    if (!actionBtn) return;
    
    switch (status) {
        case 'running':
            actionBtn.innerHTML = '<i class="fas fa-stop"></i> 停用';
            actionBtn.className = 'action-btn primary';
            break;
        case 'stopped':
        case 'configured':
            actionBtn.innerHTML = '<i class="fas fa-play"></i> 启用';
            actionBtn.className = 'action-btn success';
            break;
        default:
            actionBtn.innerHTML = '<i class="fas fa-question"></i> 未知';
            actionBtn.className = 'action-btn';
    }
}

// WireGuard开关控制
function toggleWireGuard() {
    const actionBtn = document.getElementById('wgActionBtn');
    if (!actionBtn) return;
    
    const isRunning = actionBtn.innerHTML.includes('停用');
    const action = isRunning ? 'stop' : 'start';
    const confirmText = isRunning ? '确定要停用VPN吗？' : '确定要启用VPN吗？';
    
    if (confirm(confirmText)) {
        controlWireGuard(action, currentInterface);
    }
}

// 格式化握手时间
function formatHandshakeTime(handshakeTime) {
    if (!handshakeTime) return '--';
    
    const handshake = new Date(handshakeTime);
    const now = new Date();
    const diffMs = now - handshake;
    
    if (diffMs < 60000) { // 1分钟内
        return '刚刚';
    } else if (diffMs < 3600000) { // 1小时内
        const minutes = Math.floor(diffMs / 60000);
        return `${minutes} 分前`;
    } else if (diffMs < 86400000) { // 1天内
        const hours = Math.floor(diffMs / 3600000);
        return `${hours} 小时前`;
    } else {
        const days = Math.floor(diffMs / 86400000);
        return `${days} 天前`;
    }
}

// 格式化字节数
function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// 配置文件查看相关函数
function viewCurrentConfig() {
    console.log('查看当前配置文件...');
    getCurrentConfig();
}

// 从文件读取的配置显示函数
function updateConfigDisplayFromFile(configData) {
    console.log('从文件更新配置显示:', configData);
    
    // 更新配置文件名显示
    const filenameElement = document.querySelector('.config-filename');
    if (filenameElement) {
        filenameElement.textContent = configData.interface + '.conf';
    }
    
    // 更新配置内容显示
    const configElement = document.getElementById('configContent');
    if (configElement) {
        configElement.textContent = configData.config_content;
    }
    
    // 添加文件信息提示
    const configHeader = document.querySelector('.config-header');
    if (configHeader) {
        // 检查是否已经有文件信息提示
        let fileInfo = configHeader.querySelector('.file-info');
        if (!fileInfo) {
            fileInfo = document.createElement('div');
            fileInfo.className = 'file-info';
            fileInfo.style.cssText = 'font-size: 12px; color: #666; margin-top: 5px;';
            configHeader.appendChild(fileInfo);
        }
        fileInfo.innerHTML = `
            <span>文件路径: ${configData.config_path}</span>
            <span style="margin-left: 15px;">文件大小: ${configData.file_size} 字节</span>
        `;
    }
}

async function getCurrentConfig() {
    try {
        // 显示模态框
        document.getElementById('configModal').style.display = 'flex';
        
        // 获取当前选择的接口
        const interfaceSelect = document.getElementById('interfaceSelect');
        const currentInterface = interfaceSelect ? interfaceSelect.value : 'wg0';
        
        // 使用封装的API读取配置文件
        const data = await WireGuardAPI.getConfig(currentInterface);
        
        // 如果API返回null（通常是认证问题），直接返回
        if (data === null) {
            console.warn('读取配置文件失败，可能是认证问题');
            return;
        }
        
        console.log('配置文件读取成功:', data);
        updateConfigDisplayFromFile(data);
    } catch (error) {
        console.error('读取配置文件失败:', error);
        showNotification('读取配置文件失败: ' + error.message, 'error');
    }
}

function updateConfigDisplay(wgData) {
    console.log('更新配置显示:', wgData);
    
    // 构建配置文件内容
    let configContent = '';
    
    // 接口部分
    configContent += '[Interface]\n';
    configContent += `PrivateKey = <hidden>\n`;
    configContent += `Address = ${wgData.module?.ip_address || '10.10.0.7/32'}\n`;
    configContent += `ListenPort = ${wgData.listen_port || '51820'}\n`;
    if (wgData.module?.dns) {
        configContent += `DNS = ${wgData.module.dns}\n`;
    }
    configContent += '\n';
    
    // 对端部分
    if (wgData.peers && wgData.peers.length > 0) {
        const peer = wgData.peers[0];
        configContent += '[Peer]\n';
        configContent += `PublicKey = ${peer.public_key || '--'}\n`;
        configContent += `PresharedKey = <hidden>\n`;
        configContent += `Endpoint = ${peer.endpoint || '--'}\n`;
        configContent += `AllowedIPs = ${peer.allowed_ips ? peer.allowed_ips.join(', ') : (wgData.module?.allowed_ips || '--')}\n`;
        if (peer.persistent_keepalive) {
            configContent += `PersistentKeepalive = ${peer.persistent_keepalive}\n`;
        } else if (wgData.module?.persistent_ka) {
            configContent += `PersistentKeepalive = ${wgData.module.persistent_ka}\n`;
        }
        
        // 添加传输统计信息（注释形式）
        if (peer.transfer_rx || peer.transfer_tx) {
            configContent += `# Transfer: ${formatBytes(peer.transfer_rx || 0)} received, ${formatBytes(peer.transfer_tx || 0)} sent\n`;
        }
        
        // 添加最新握手时间（注释形式）
        if (peer.latest_handshake) {
            configContent += `# Latest Handshake: ${formatHandshakeTime(peer.latest_handshake)}\n`;
        }
    } else {
        configContent += '[Peer]\n';
        configContent += `PublicKey = --\n`;
        configContent += `PresharedKey = --\n`;
        configContent += `Endpoint = --\n`;
        configContent += `AllowedIPs = ${wgData.module?.allowed_ips || '--'}\n`;
        if (wgData.module?.persistent_ka) {
            configContent += `PersistentKeepalive = ${wgData.module.persistent_ka}\n`;
        }
    }
    
    // 更新显示
    const configElement = document.getElementById('configContent');
    if (configElement) {
        configElement.textContent = configContent;
    }
}

function closeConfigModal() {
    document.getElementById('configModal').style.display = 'none';
}

// 点击模态框外部关闭
document.addEventListener('DOMContentLoaded', function() {
    const modal = document.getElementById('configModal');
    if (modal) {
        modal.addEventListener('click', function(e) {
            if (e.target === modal) {
                closeConfigModal();
            }
        });
    }
});

// 复制配置文件内容
function copyConfigContent() {
    const configContent = document.getElementById('configContent');
    if (configContent && configContent.textContent) {
        navigator.clipboard.writeText(configContent.textContent).then(() => {
            // 显示复制成功提示
            const copyBtn = document.querySelector('.copy-btn');
            const originalText = copyBtn.innerHTML;
            copyBtn.innerHTML = '<i class="fas fa-check"></i> 已复制';
            copyBtn.style.background = 'var(--success-color)';
            
            setTimeout(() => {
                copyBtn.innerHTML = originalText;
                copyBtn.style.background = 'var(--primary-color)';
            }, 2000);
        }).catch(err => {
            console.error('复制失败:', err);
            // 降级方案：使用传统方法
            const textArea = document.createElement('textarea');
            textArea.value = configContent.textContent;
            document.body.appendChild(textArea);
            textArea.select();
            document.execCommand('copy');
            document.body.removeChild(textArea);
            
            const copyBtn = document.querySelector('.copy-btn');
            const originalText = copyBtn.innerHTML;
            copyBtn.innerHTML = '<i class="fas fa-check"></i> 已复制';
            copyBtn.style.background = 'var(--success-color)';
            
            setTimeout(() => {
                copyBtn.innerHTML = originalText;
                copyBtn.style.background = 'var(--primary-color)';
            }, 2000);
        });
    }
}
