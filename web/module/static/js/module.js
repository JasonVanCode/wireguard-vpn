// 全局变量
let updateInterval;
let currentView = 'dashboard'; // 'dashboard' 或 'config'

// 页面加载时初始化
document.addEventListener('DOMContentLoaded', function() {
    updateTime();
    // 先验证token，然后再加载数据
    checkAuthAndLoadData();
    
    // 每秒更新时间
    setInterval(updateTime, 1000);
    
    // 初始化配置界面
    initializeConfigForm();
    
    // 创建背景粒子效果
    createParticles();
});

// 检查认证状态并加载数据
async function checkAuthAndLoadData() {
    const token = getAuthToken();
    if (!token) {
        showConfigView();
        return;
    }

    try {
        // 先验证token是否有效
        const response = await fetch('/api/v1/auth/verify', {
            headers: { 'Authorization': 'Bearer ' + token }
        });
        
        if (response.ok) {
            // Token有效，检查配置状态
            const configResponse = await fetch('/api/v1/config/status');
            const configData = await configResponse.json();
            
            if (!configData.configured) {
                // 模块未配置，显示配置界面
                showConfigView();
                return;
            }
            
            // 配置已就绪，显示管理界面
            showDashboardView();
            loadAllData();
            // 设置定时刷新 - 降低频率减少卡顿
            updateInterval = setInterval(loadAllData, 60000);
        } else {
            // Token无效，显示配置界面
            showConfigView();
        }
    } catch (error) {
        console.error('认证验证失败:', error);
        showConfigView();
    }
}

// 显示管理界面
function showDashboardView() {
    currentView = 'dashboard';
    document.getElementById('dashboardView').classList.remove('d-none');
    document.getElementById('configView').classList.add('d-none');
    document.getElementById('particles').style.display = 'none';
}

// 显示配置界面
function showConfigView() {
    currentView = 'config';
    document.getElementById('dashboardView').classList.add('d-none');
    document.getElementById('configView').classList.remove('d-none');
    document.getElementById('particles').style.display = 'block';
    
    // 清除定时器
    if (updateInterval) {
        clearInterval(updateInterval);
    }
}

// 切换到配置界面
function switchToConfig() {
    showConfigView();
}

// 切换到管理界面
function switchToDashboard() {
    checkAuthAndLoadData();
}

function updateTime() {
    const now = new Date();
    const timeString = now.toLocaleTimeString('zh-CN', { hour12: false });
    const timeEl = document.getElementById('currentTime');
    if (timeEl) {
        timeEl.textContent = timeString;
    }
}

async function loadAllData() {
    try {
        // 并行加载所有数据
        await Promise.all([
            loadModuleStatus(),
            loadWireGuardStatus(),
            loadSystemStatus(),
            loadTrafficStats()
        ]);
    } catch (error) {
        console.error('加载数据失败:', error);
    }
}

// 认证相关函数
function getAuthToken() {
    return localStorage.getItem('module_token');
}

function createAuthHeaders() {
    const token = getAuthToken();
    const headers = { 'Content-Type': 'application/json' };
    if (token) {
        headers['Authorization'] = 'Bearer ' + token;
    }
    return headers;
}

function handleAuthError() {
    // 防止重复重定向
    if (currentView === 'config') {
        return;
    }
    
    // 清除localStorage和cookie
    localStorage.removeItem('module_token');
    document.cookie = 'module_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
    
    // 清除定时器
    if (updateInterval) {
        clearInterval(updateInterval);
    }
    
    showConfigView();
}

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
            updateModuleStatus(result.data);
        }
    } catch (error) {
        console.error('获取模块状态失败:', error);
    }
}

async function loadWireGuardStatus() {
    try {
        const headers = createAuthHeaders();
        const response = await fetch('/api/v1/status', { headers });
        if (response.status === 401) {
            handleAuthError();
            return;
        }
        if (response.ok) {
            const result = await response.json();
            updateWireGuardStatus(result.data.wireguard);
        }
    } catch (error) {
        console.error('获取WireGuard状态失败:', error);
    }
}

async function loadSystemStatus() {
    try {
        const headers = createAuthHeaders();
        const response = await fetch('/api/v1/status', { headers });
        if (response.status === 401) {
            handleAuthError();
            return;
        }
        if (response.ok) {
            const result = await response.json();
            updateSystemStatus(result.data.system);
        }
    } catch (error) {
        console.error('获取系统状态失败:', error);
    }
}

async function loadTrafficStats() {
    try {
        const headers = createAuthHeaders();
        const response = await fetch('/api/v1/stats', { headers });
        if (response.status === 401) {
            handleAuthError();
            return;
        }
        if (response.ok) {
            const result = await response.json();
            updateTrafficStats(result.data);
        }
    } catch (error) {
        console.error('获取流量统计失败:', error);
    }
}

function updateModuleStatus(data) {
    const statusEl = document.getElementById('moduleStatus');
    const statusTextEl = document.getElementById('moduleStatusText');
    
    if (!statusEl || !statusTextEl) return;
    
    if (data.wireguard && data.wireguard.status === 'running') {
        statusEl.className = 'status-indicator status-online';
        statusTextEl.textContent = '运行中';
        
        const startBtn = document.getElementById('startBtn');
        const stopBtn = document.getElementById('stopBtn');
        if (startBtn) startBtn.style.display = 'none';
        if (stopBtn) stopBtn.style.display = 'flex';
    } else {
        statusEl.className = 'status-indicator status-offline';
        statusTextEl.textContent = '已停止';
        
        const startBtn = document.getElementById('startBtn');
        const stopBtn = document.getElementById('stopBtn');
        if (startBtn) startBtn.style.display = 'flex';
        if (stopBtn) stopBtn.style.display = 'none';
    }
}

function updateWireGuardStatus(wgData) {
    if (!wgData) return;
    
    const wgStatusEl = document.getElementById('wgStatus');
    const wgStatusDetailEl = document.getElementById('wgStatusDetail');
    
    if (wgStatusEl) {
        wgStatusEl.textContent = wgData.status === 'running' ? '运行中' : '已停止';
    }
    if (wgStatusDetailEl) {
        wgStatusDetailEl.textContent = wgData.status === 'running' ? '正常运行' : '服务已停止';
    }
    
    // 更新连接质量
    if (wgData.peers && wgData.peers.length > 0) {
        const peer = wgData.peers[0];
        const handshakeTime = new Date(peer.latest_handshake);
        const timeDiff = (Date.now() - handshakeTime.getTime()) / 1000;
        
        const qualityEl = document.getElementById('connectionQuality');
        const handshakeEl = document.getElementById('lastHandshake');
        
        if (qualityEl) {
            if (timeDiff < 120) { // 2分钟内
                qualityEl.textContent = '良好';
            } else if (timeDiff < 300) { // 5分钟内
                qualityEl.textContent = '一般';
            } else {
                qualityEl.textContent = '较差';
            }
        }
        
        if (handshakeEl) {
            handshakeEl.textContent = formatTimeAgo(handshakeTime);
        }
    } else {
        const qualityEl = document.getElementById('connectionQuality');
        const handshakeEl = document.getElementById('lastHandshake');
        if (qualityEl) qualityEl.textContent = '无连接';
        if (handshakeEl) handshakeEl.textContent = '--';
    }
    
    // 更新详细信息
    updateWireGuardDetails(wgData);
}

function updateWireGuardDetails(wgData) {
    const detailsEl = document.getElementById('wireGuardDetails');
    if (!detailsEl) return;
    
    if (!wgData || wgData.status !== 'running') {
        detailsEl.innerHTML = `
            <div style="text-align: center; color: var(--text-secondary); padding: 2rem;">
                <i class="fas fa-exclamation-triangle" style="font-size: 2rem; margin-bottom: 1rem; color: var(--warning-color);"></i>
                <div>WireGuard 未运行</div>
            </div>
        `;
        return;
    }
    
    let peersHtml = '';
    if (wgData.peers && wgData.peers.length > 0) {
        peersHtml = wgData.peers.map(peer => `
            <div class="info-grid" style="margin-bottom: 1rem;">
                <div class="info-item">
                    <div class="info-label">公钥</div>
                    <div class="info-value">${peer.public_key.substring(0, 20)}...</div>
                </div>
                <div class="info-item">
                    <div class="info-label">端点</div>
                    <div class="info-value">${peer.endpoint || '--'}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">接收流量</div>
                    <div class="info-value">${formatBytes(peer.transfer_rx || 0)}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">发送流量</div>
                    <div class="info-value">${formatBytes(peer.transfer_tx || 0)}</div>
                </div>
            </div>
        `).join('');
    }
    
    detailsEl.innerHTML = `
        <div class="info-grid" style="margin-bottom: 1.5rem;">
            <div class="info-item">
                <div class="info-label">接口</div>
                <div class="info-value">${wgData.interface || 'wg0'}</div>
            </div>
            <div class="info-item">
                <div class="info-label">监听端口</div>
                <div class="info-value">${wgData.listen_port || '--'}</div>
            </div>
            <div class="info-item">
                <div class="info-label">公钥</div>
                <div class="info-value">${wgData.public_key ? wgData.public_key.substring(0, 20) + '...' : '--'}</div>
            </div>
            <div class="info-item">
                <div class="info-label">连接数</div>
                <div class="info-value">${wgData.peers ? wgData.peers.length : 0}</div>
            </div>
        </div>
        ${peersHtml ? `<div style="border-top: 1px solid var(--border-color); padding-top: 1rem;"><h6 style="color: var(--text-primary); margin-bottom: 1rem;">连接信息</h6>${peersHtml}</div>` : ''}
    `;
}

function updateSystemStatus(sysData) {
    if (!sysData) return;
    
    // 更新内存使用率
    const memPercent = sysData.memory.percent || 0;
    const memUsageEl = document.getElementById('memoryUsage');
    const memProgressEl = document.getElementById('memoryProgress');
    
    if (memUsageEl) memUsageEl.textContent = memPercent.toFixed(1) + '%';
    if (memProgressEl) memProgressEl.style.width = memPercent + '%';
    
    // 更新磁盘使用率
    const diskPercent = sysData.disk.percent || 0;
    const diskUsageEl = document.getElementById('diskUsage');
    const diskProgressEl = document.getElementById('diskProgress');
    
    if (diskUsageEl) diskUsageEl.textContent = diskPercent.toFixed(1) + '%';
    if (diskProgressEl) diskProgressEl.style.width = diskPercent + '%';
    
    // 更新系统信息
    const uptimeEl = document.getElementById('systemUptime');
    const loadEl = document.getElementById('loadAverage');
    
    if (uptimeEl) uptimeEl.textContent = formatUptime(sysData.uptime || 0);
    if (loadEl) loadEl.textContent = sysData.load_average || '--';
    
    // 更新健康状态
    let healthStatus = '健康';
    if (memPercent > 90 || diskPercent > 95) {
        healthStatus = '警告';
    }
    const healthEl = document.getElementById('healthStatus');
    if (healthEl) healthEl.textContent = healthStatus;
}

function updateTrafficStats(trafficData) {
    if (!trafficData) return;
    
    const totalBytes = (trafficData.rx_bytes || 0) + (trafficData.tx_bytes || 0);
    const totalTrafficEl = document.getElementById('totalTraffic');
    const trafficRateEl = document.getElementById('trafficRate');
    
    if (totalTrafficEl) totalTrafficEl.textContent = formatBytes(totalBytes);
    if (trafficRateEl) trafficRateEl.textContent = `↓${formatBytes(trafficData.rx_bytes || 0)} ↑${formatBytes(trafficData.tx_bytes || 0)}`;
}

async function startWireGuard() {
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
        if (response.ok) {
            setTimeout(loadAllData, 1000); // 1秒后刷新状态
        } else {
            alert('启动失败');
        }
    } catch (error) {
        console.error('启动WireGuard失败:', error);
        alert('启动失败');
    }
}

async function stopWireGuard() {
    if (confirm('确定要停止WireGuard服务吗？')) {
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
            if (response.ok) {
                setTimeout(loadAllData, 1000); // 1秒后刷新状态
            } else {
                alert('停止失败');
            }
        } catch (error) {
            console.error('停止WireGuard失败:', error);
            alert('停止失败');
        }
    }
}



// 下载配置文件
async function downloadConfig() {
    try {
        const headers = createAuthHeaders();
        const response = await fetch('/api/v1/config', { headers });
        if (response.status === 401) {
            handleAuthError();
            return;
        }
        if (response.ok) {
            const result = await response.json();
            const config = result.data.config || '';
            
            if (!config.trim()) {
                alert('没有可下载的配置');
                return;
            }
            
            // 创建下载链接
            const blob = new Blob([config], { type: 'text/plain' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'wireguard-module.conf';
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
        } else {
            alert('获取配置失败');
        }
    } catch (error) {
        console.error('下载配置失败:', error);
        alert('下载配置失败');
    }
}

function refreshData() {
    loadAllData();
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
        
        // 无论后端是否成功，都清除本地token并跳转到配置界面
        handleAuthError();
    } catch (error) {
        console.error('退出登录请求失败:', error);
        // 即使请求失败也要清除本地token
        handleAuthError();
    }
}

// 配置表单相关函数
function initializeConfigForm() {
    const configForm = document.getElementById('configForm');
    if (!configForm) return;
    
    // 初始化文件上传
    initializeFileUpload();
    
    configForm.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const submitBtn = this.querySelector('.btn-config');
        const originalText = submitBtn.innerHTML;
        
        // 清除之前的提示
        const alertContainer = document.getElementById('alertContainer');
        if (alertContainer) alertContainer.innerHTML = '';
        
        // 检查是否有上传的配置
        if (!window.uploadedConfig) {
            showAlert('danger', '请先上传配置文件');
            return;
        }
        
        // 显示加载状态
        submitBtn.classList.add('loading');
        
        // 从配置文件中提取信息
        const configInfo = parseWireGuardConfig(window.uploadedConfig);
        
        const configData = {
            module_id: 1, // 默认模块ID
            api_key: '', // 可选
            server_url: extractServerFromEndpoint(configInfo.peer.Endpoint || ''),
            config_data: window.uploadedConfig
        };
        
        try {
            const response = await fetch('/api/v1/configure', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(configData)
            });
            
            const result = await response.json();
            
            if (response.ok) {
                // 配置成功
                submitBtn.classList.remove('loading');
                submitBtn.classList.add('success');
                submitBtn.innerHTML = '<i class="fas fa-check me-2"></i>配置成功';
                
                showAlert('success', '模块配置成功！正在重定向到管理界面...');
                
                // 延迟跳转到管理界面
                setTimeout(() => {
                    showDashboardView();
                    checkAuthAndLoadData();
                }, 2000);
            } else {
                // 显示错误信息
                showAlert('danger', result.message || '配置失败，请检查配置文件');
                resetButton();
            }
        } catch (error) {
            console.error('Configuration error:', error);
            showAlert('danger', '网络连接失败，请检查网络后重试');
            resetButton();
        }
        
        function resetButton() {
            submitBtn.classList.remove('loading');
            submitBtn.innerHTML = originalText;
        }
    });
}

// 初始化文件上传功能
function initializeFileUpload() {
    const fileUploadArea = document.getElementById('fileUploadArea');
    const fileInput = document.getElementById('configFile');
    const submitBtn = document.getElementById('submitBtn');
    
    if (!fileUploadArea || !fileInput) return;
    
    // 点击上传区域触发文件选择
    fileUploadArea.addEventListener('click', () => {
        fileInput.click();
    });
    
    // 点击上传链接触发文件选择
    const uploadLink = fileUploadArea.querySelector('.upload-link');
    if (uploadLink) {
        uploadLink.addEventListener('click', (e) => {
            e.stopPropagation();
            fileInput.click();
        });
    }
    
    // 文件选择事件
    fileInput.addEventListener('change', handleFileSelect);
    
    // 拖拽事件
    fileUploadArea.addEventListener('dragover', (e) => {
        e.preventDefault();
        fileUploadArea.classList.add('dragover');
    });
    
    fileUploadArea.addEventListener('dragleave', (e) => {
        e.preventDefault();
        fileUploadArea.classList.remove('dragover');
    });
    
    fileUploadArea.addEventListener('drop', (e) => {
        e.preventDefault();
        fileUploadArea.classList.remove('dragover');
        
        const files = e.dataTransfer.files;
        if (files.length > 0) {
            handleFile(files[0]);
        }
    });
}

// 处理文件选择
function handleFileSelect(event) {
    const file = event.target.files[0];
    if (file) {
        handleFile(file);
    }
}

// 处理文件
function handleFile(file) {
    // 检查文件类型
    if (!file.name.endsWith('.conf')) {
        showAlert('danger', '请选择.conf格式的配置文件');
        return;
    }
    
    // 读取文件内容
    const reader = new FileReader();
    reader.onload = (e) => {
        const content = e.target.result;
        
        // 验证配置格式
        if (validateWireGuardConfig(content)) {
            window.uploadedConfig = content;
            showConfigPreview(content);
            document.getElementById('submitBtn').disabled = false;
        } else {
            showAlert('danger', '配置文件格式不正确，请检查文件内容');
        }
    };
    
    reader.readAsText(file);
}

// 验证WireGuard配置
function validateWireGuardConfig(config) {
    // 基本格式验证
    if (!config.includes('[Interface]') || !config.includes('[Peer]')) {
        return false;
    }
    
    if (!config.includes('PrivateKey') || !config.includes('PublicKey')) {
        return false;
    }
    
    return true;
}

// 显示配置预览
function showConfigPreview(config) {
    const previewContainer = document.getElementById('configPreview');
    const previewContent = document.getElementById('previewContent');
    const fileUploadArea = document.getElementById('fileUploadArea');
    
    if (!previewContainer || !previewContent) return;
    
    // 解析配置
    const configInfo = parseWireGuardConfig(config);
    
    // 显示预览
    previewContent.innerHTML = `
        <div style="font-family: 'Monaco', 'Menlo', monospace; font-size: 0.875rem; line-height: 1.5;">
            <div style="margin-bottom: 1rem;">
                <strong>接口:</strong> ${configInfo.interface.Address || '--'}<br>
                <strong>DNS:</strong> ${configInfo.interface.DNS || '--'}
            </div>
            <div>
                <strong>对端:</strong> ${configInfo.peer.Endpoint || '--'}<br>
                <strong>允许IP:</strong> ${configInfo.peer.AllowedIPs || '--'}
            </div>
        </div>
    `;
    
    // 隐藏上传区域，显示预览
    fileUploadArea.style.display = 'none';
    previewContainer.style.display = 'block';
}

// 清除上传的文件
function clearUploadedFile() {
    window.uploadedConfig = null;
    
    const fileUploadArea = document.getElementById('fileUploadArea');
    const previewContainer = document.getElementById('configPreview');
    const fileInput = document.getElementById('configFile');
    const submitBtn = document.getElementById('submitBtn');
    
    // 重置界面
    if (fileUploadArea) fileUploadArea.style.display = 'block';
    if (previewContainer) previewContainer.style.display = 'none';
    if (fileInput) fileInput.value = '';
    if (submitBtn) submitBtn.disabled = true;
}

function validateConfigData(configData) {
    // 验证模块ID
    if (!configData.module_id || configData.module_id <= 0) {
        showAlert('danger', '请输入有效的模块ID');
        return false;
    }

    // 验证API密钥
    if (configData.api_key && configData.api_key.trim().length < 10) {
        showAlert('danger', 'API密钥长度至少为10个字符');
        return false;
    }

    // 验证服务器地址
    try {
        new URL(configData.server_url);
    } catch (e) {
        showAlert('danger', '请输入有效的服务器地址');
        return false;
    }

    // 验证WireGuard配置
    const config = configData.config_data;
    if (!config.includes('[Interface]') || !config.includes('[Peer]')) {
        showAlert('danger', 'WireGuard配置格式不正确：缺少必要的 [Interface] 或 [Peer] 部分');
        return false;
    }

    if (!config.includes('PrivateKey') || !config.includes('PublicKey')) {
        showAlert('danger', 'WireGuard配置格式不正确：缺少必要的密钥信息');
        return false;
    }

    if (!config.includes('Address')) {
        showAlert('danger', 'WireGuard配置格式不正确：缺少地址信息');
        return false;
    }

    return true;
}

function showAlert(type, message) {
    const alertContainer = document.getElementById('alertContainer');
    if (!alertContainer) return;
    
    const alertClass = type === 'success' ? 'alert-success' : 'alert-danger';
    const iconClass = type === 'success' ? 'fas fa-check-circle' : 'fas fa-exclamation-triangle';
    
    alertContainer.innerHTML = `
        <div class="alert-custom ${alertClass}">
            <i class="${iconClass}"></i>
            ${message}
        </div>
    `;
    
    // 5秒后自动隐藏错误提示（成功提示不自动隐藏）
    if (type === 'danger') {
        setTimeout(() => {
            const alert = alertContainer.querySelector('.alert-custom');
            if (alert) {
                alert.style.opacity = '0';
                setTimeout(() => {
                    alertContainer.innerHTML = '';
                }, 300);
            }
        }, 5000);
    }
}

// 创建背景粒子效果
function createParticles() {
    const particlesContainer = document.getElementById('particles');
    if (!particlesContainer) return;
    
    const particleCount = 30;
    
    for (let i = 0; i < particleCount; i++) {
        const particle = document.createElement('div');
        particle.className = 'particle';
        particle.style.left = Math.random() * 100 + '%';
        particle.style.top = Math.random() * 100 + '%';
        particle.style.animationDelay = Math.random() * 6 + 's';
        particle.style.animationDuration = (Math.random() * 3 + 3) + 's';
        particlesContainer.appendChild(particle);
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

// 解析WireGuard配置文件
function parseWireGuardConfig(config) {
    const result = {
        interface: {},
        peer: {}
    };
    
    const lines = config.split('\n');
    let currentSection = null;
    
    for (const line of lines) {
        const trimmedLine = line.trim();
        
        if (trimmedLine === '[Interface]') {
            currentSection = 'interface';
            continue;
        } else if (trimmedLine === '[Peer]') {
            currentSection = 'peer';
            continue;
        }
        
        if (currentSection && trimmedLine.includes('=')) {
            const [key, value] = trimmedLine.split('=').map(s => s.trim());
            result[currentSection][key] = value;
        }
    }
    
    return result;
}

// 从Endpoint提取服务器地址
function extractServerFromEndpoint(endpoint) {
    if (!endpoint) return '';
    
    // 移除端口号，只保留域名或IP
    const parts = endpoint.split(':');
    if (parts.length >= 2) {
        return parts.slice(0, -1).join(':'); // 处理IPv6的情况
    }
    return endpoint;
}

// 删除旧的验证函数，不再需要
function validateConfigData(configData) {
    return true; // 简化版本，文件上传时已经验证过
}

// 键盘快捷键
document.addEventListener('keydown', function(e) {
    if (e.key === 'Enter' && e.ctrlKey && currentView === 'config') {
        const form = document.getElementById('configForm');
        if (form) form.dispatchEvent(new Event('submit'));
    }
}); 