// 主要JavaScript文件

// 全局变量
let currentToken = localStorage.getItem('authToken');
let refreshTimer;

// 检查认证状态
function checkAuth() {
    if (!currentToken) {
        window.location.href = '/login';
        return false;
    }
    return true;
}

// 创建认证头
function createAuthHeaders() {
    return {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${currentToken}`
    };
}

// 处理认证错误
function handleAuthError() {
    localStorage.removeItem('authToken');
    window.location.href = '/login';
}

// 页面加载时检查认证
if (!checkAuth()) {
    // 如果认证失败，页面会重定向，不需要继续执行
}

// 初始化页面
document.addEventListener('DOMContentLoaded', function() {
    updateTime();
    setInterval(updateTime, 1000);
    
    loadData();
    
    // 设置定时刷新
    refreshTimer = setInterval(refreshData, 30000); // 30秒刷新一次
});

// 更新时间显示
function updateTime() {
    const now = new Date();
    const timeString = now.toLocaleTimeString('zh-CN', {
        hour12: false,
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
    document.getElementById('currentTime').textContent = timeString;
}

// 加载所有数据
async function loadData() {
    await Promise.all([
        loadModuleStatus(),
        loadWireGuardStatus(),
        loadSystemStatus(),
        loadTrafficStats()
    ]);
}

// 刷新数据
function refreshData() {
    loadData();
}

// 加载模块状态
async function loadModuleStatus() {
    try {
        const headers = createAuthHeaders();
        const response = await fetch('/api/v1/status', { headers });
        if (response.status === 401) {
            handleAuthError();
            return;
        }
        if (response.ok) {
            const result = await response.json();
            if (result.success) {
                updateModuleStatus(result.data);
            }
        }
    } catch (error) {
        console.error('加载模块状态失败:', error);
        updateModuleStatus({ status: 'error', message: '连接失败' });
    }
}

// 更新模块状态显示
function updateModuleStatus(data) {
    const statusElement = document.getElementById('moduleStatus');
    const statusTextElement = document.getElementById('moduleStatusText');
    const locationElement = document.getElementById('moduleLocation');
    
    if (data.status === 'online') {
        statusElement.className = 'status-indicator online';
        statusTextElement.textContent = '在线';
        locationElement.textContent = data.location || '未知位置';
    } else if (data.status === 'offline') {
        statusElement.className = 'status-indicator offline';
        statusTextElement.textContent = '离线';
        locationElement.textContent = '--';
    } else {
        statusElement.className = 'status-indicator unconfigured';
        statusTextElement.textContent = '未配置';
        locationElement.textContent = '--';
    }
}

// 加载WireGuard状态
async function loadWireGuardStatus() {
    try {
        const headers = createAuthHeaders();
        const response = await fetch('/api/v1/wireguard/status', { headers });
        if (response.status === 401) {
            handleAuthError();
            return;
        }
        if (response.ok) {
            const result = await response.json();
            if (result.success) {
                updateWireGuardStatus(result.data);
                updateWireGuardDetails(result.data);
            }
        }
    } catch (error) {
        console.error('加载WireGuard状态失败:', error);
        updateWireGuardStatus({ status: 'error', message: '获取状态失败' });
    }
}

// 更新WireGuard状态显示
function updateWireGuardStatus(data) {
    const statusElement = document.getElementById('wgStatus');
    const detailElement = document.getElementById('wgStatusDetail');
    const qualityElement = document.getElementById('connectionQuality');
    const handshakeElement = document.getElementById('lastHandshake');
    
    if (data.status === 'active') {
        statusElement.textContent = '运行中';
        statusElement.style.color = 'var(--success-color)';
        detailElement.textContent = '连接正常';
        
        // 更新连接质量
        if (data.last_handshake) {
            const lastHandshake = new Date(data.last_handshake);
            const timeDiff = (Date.now() - lastHandshake.getTime()) / 1000;
            
            if (timeDiff < 300) { // 5分钟内
                qualityElement.textContent = '优秀';
                qualityElement.style.color = 'var(--success-color)';
            } else if (timeDiff < 900) { // 15分钟内
                qualityElement.textContent = '良好';
                qualityElement.style.color = 'var(--warning-color)';
            } else {
                qualityElement.textContent = '较差';
                qualityElement.style.color = 'var(--danger-color)';
            }
            
            handshakeElement.textContent = formatTimeAgo(lastHandshake);
        } else {
            qualityElement.textContent = '未知';
            handshakeElement.textContent = '--';
        }
    } else if (data.status === 'inactive') {
        statusElement.textContent = '已停止';
        statusElement.style.color = 'var(--text-secondary)';
        detailElement.textContent = '服务未运行';
        qualityElement.textContent = '--';
        handshakeElement.textContent = '--';
    } else {
        statusElement.textContent = '错误';
        statusElement.style.color = 'var(--danger-color)';
        detailElement.textContent = data.message || '状态异常';
        qualityElement.textContent = '--';
        handshakeElement.textContent = '--';
    }
    
    // 更新控制按钮状态
    updateControlButtons(data.status);
}

// 更新控制按钮状态
function updateControlButtons(status) {
    const startBtn = document.getElementById('startBtn');
    const stopBtn = document.getElementById('stopBtn');
    
    if (status === 'active') {
        startBtn.style.display = 'none';
        stopBtn.style.display = 'flex';
    } else {
        startBtn.style.display = 'flex';
        stopBtn.style.display = 'none';
    }
}

// 更新WireGuard详细信息
function updateWireGuardDetails(data) {
    const detailsElement = document.getElementById('wireGuardDetails');
    
    if (data.status === 'active' && data.interface_info) {
        const info = data.interface_info;
        detailsElement.innerHTML = `
            <div class="info-grid">
                <div class="info-item">
                    <div class="info-label">接口名称</div>
                    <div class="info-value">${info.name || '--'}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">IP地址</div>
                    <div class="info-value">${info.address || '--'}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">监听端口</div>
                    <div class="info-value">${info.listen_port || '--'}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">公钥</div>
                    <div class="info-value" style="font-size: 0.75rem; word-break: break-all;">${info.public_key || '--'}</div>
                </div>
            </div>
            ${data.peers && data.peers.length > 0 ? `
                <div style="margin-top: 1.5rem;">
                    <div class="section-header">
                        <i class="fas fa-users"></i>
                        <span>对等节点</span>
                    </div>
                    ${data.peers.map(peer => `
                        <div class="info-grid" style="margin-bottom: 1rem; padding: 1rem; background: rgba(255,255,255,0.02); border-radius: 8px;">
                            <div class="info-item">
                                <div class="info-label">公钥</div>
                                <div class="info-value" style="font-size: 0.75rem; word-break: break-all;">${peer.public_key || '--'}</div>
                            </div>
                            <div class="info-item">
                                <div class="info-label">端点</div>
                                <div class="info-value">${peer.endpoint || '--'}</div>
                            </div>
                            <div class="info-item">
                                <div class="info-label">允许的IP</div>
                                <div class="info-value">${peer.allowed_ips || '--'}</div>
                            </div>
                            <div class="info-item">
                                <div class="info-label">最后握手</div>
                                <div class="info-value">${peer.last_handshake ? formatTimeAgo(new Date(peer.last_handshake)) : '--'}</div>
                            </div>
                        </div>
                    `).join('')}
                </div>
            ` : ''}
        `;
    } else {
        detailsElement.innerHTML = `
            <div style="text-align: center; color: var(--text-secondary); padding: 2rem;">
                <i class="fas fa-info-circle" style="font-size: 2rem; margin-bottom: 1rem; opacity: 0.5;"></i>
                <div>WireGuard 服务未运行</div>
            </div>
        `;
    }
}

// 加载系统状态
async function loadSystemStatus() {
    try {
        const headers = createAuthHeaders();
        const response = await fetch('/api/v1/system/status', { headers });
        if (response.status === 401) {
            handleAuthError();
            return;
        }
        if (response.ok) {
            const result = await response.json();
            if (result.success) {
                updateSystemStatus(result.data);
            }
        }
    } catch (error) {
        console.error('加载系统状态失败:', error);
    }
}

// 更新系统状态显示
function updateSystemStatus(data) {
    // 更新健康状态
    const healthElement = document.getElementById('healthStatus');
    if (data.health_score >= 80) {
        healthElement.textContent = '健康';
        healthElement.style.color = 'var(--success-color)';
    } else if (data.health_score >= 60) {
        healthElement.textContent = '一般';
        healthElement.style.color = 'var(--warning-color)';
    } else {
        healthElement.textContent = '异常';
        healthElement.style.color = 'var(--danger-color)';
    }
    
    // 更新系统信息
    if (data.memory) {
        const memUsage = ((data.memory.used / data.memory.total) * 100).toFixed(1);
        document.getElementById('memoryUsage').textContent = `${memUsage}%`;
        document.getElementById('memoryProgress').style.width = `${memUsage}%`;
    }
    
    if (data.disk) {
        const diskUsage = ((data.disk.used / data.disk.total) * 100).toFixed(1);
        document.getElementById('diskUsage').textContent = `${diskUsage}%`;
        document.getElementById('diskProgress').style.width = `${diskUsage}%`;
    }
    
    if (data.uptime) {
        document.getElementById('systemUptime').textContent = formatUptime(data.uptime);
    }
    
    if (data.load_average) {
        document.getElementById('loadAverage').textContent = data.load_average.join(', ');
    }
}

// 加载流量统计
async function loadTrafficStats() {
    try {
        const headers = createAuthHeaders();
        const response = await fetch('/api/v1/traffic/stats', { headers });
        if (response.status === 401) {
            handleAuthError();
            return;
        }
        if (response.ok) {
            const result = await response.json();
            if (result.success) {
                updateTrafficStats(result.data);
            }
        }
    } catch (error) {
        console.error('加载流量统计失败:', error);
    }
}

// 更新流量统计显示
function updateTrafficStats(data) {
    const totalElement = document.getElementById('totalTraffic');
    const rateElement = document.getElementById('trafficRate');
    
    if (data.total_bytes) {
        totalElement.textContent = formatBytes(data.total_bytes);
    } else {
        totalElement.textContent = '0 B';
    }
    
    if (data.rate_in && data.rate_out) {
        rateElement.textContent = `↓ ${formatBytes(data.rate_in)}/s ↑ ${formatBytes(data.rate_out)}/s`;
    } else {
        rateElement.textContent = '无数据';
    }
}

// WireGuard 控制函数
async function startWireGuard() {
    if (!confirm('确定要启动 WireGuard 服务吗？')) {
        return;
    }
    
    try {
        const headers = createAuthHeaders();
        const response = await fetch('/api/v1/wireguard/start', {
            method: 'POST',
            headers
        });
        
        if (response.status === 401) {
            handleAuthError();
            return;
        }
        
        const result = await response.json();
        if (result.success) {
            alert('WireGuard 服务启动成功');
            loadWireGuardStatus(); // 刷新状态
        } else {
            alert('启动失败: ' + (result.message || '未知错误'));
        }
    } catch (error) {
        console.error('启动WireGuard失败:', error);
        alert('启动失败，请检查网络连接');
    }
}

async function stopWireGuard() {
    if (!confirm('确定要停止 WireGuard 服务吗？')) {
        return;
    }
    
    try {
        const headers = createAuthHeaders();
        const response = await fetch('/api/v1/wireguard/stop', {
            method: 'POST',
            headers
        });
        
        if (response.status === 401) {
            handleAuthError();
            return;
        }
        
        const result = await response.json();
        if (result.success) {
            alert('WireGuard 服务停止成功');
            loadWireGuardStatus(); // 刷新状态
        } else {
            alert('停止失败: ' + (result.message || '未知错误'));
        }
    } catch (error) {
        console.error('停止WireGuard失败:', error);
        alert('停止失败，请检查网络连接');
    }
}

function refreshWireGuardStatus() {
    loadWireGuardStatus();
}

function logout() {
    if (confirm('确定要退出登录吗？')) {
        performLogout();
    }
}

async function performLogout() {
    try {
        const headers = createAuthHeaders();
        const response = await fetch('/api/v1/auth/logout', {
            method: 'POST',
            headers
        });
        
        // 无论后端是否成功，都清除本地token并跳转到登录页
        handleAuthError();
    } catch (error) {
        console.error('退出登录请求失败:', error);
        // 即使请求失败也要清除本地token
        handleAuthError();
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

function formatUptime(seconds) {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (days > 0) {
        return `${days}天 ${hours}小时`;
    } else if (hours > 0) {
        return `${hours}小时 ${minutes}分钟`;
    } else {
        return `${minutes}分钟`;
    }
}

function formatTimeAgo(time) {
    const diff = (Date.now() - time.getTime()) / 1000;
    
    if (diff < 60) {
        return '刚刚';
    } else if (diff < 3600) {
        return Math.floor(diff / 60) + '分钟前';
    } else if (diff < 86400) {
        return Math.floor(diff / 3600) + '小时前';
    } else {
        return Math.floor(diff / 86400) + '天前';
    }
}