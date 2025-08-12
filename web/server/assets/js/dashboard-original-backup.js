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

function updateTime() {
    const now = new Date();
    const timeString = now.toLocaleTimeString('zh-CN', { hour12: false });
    document.getElementById('currentTime').textContent = timeString;
}

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

// 生成时间标签（用作备用）
function generateTimeLabels() {
    const labels = [];
    const now = new Date();
    for (let i = 11; i >= 0; i--) {
        const time = new Date(now.getTime() - i * 5 * 60 * 1000);
        labels.push(time.getHours().toString().padStart(2, '0') + ':' + 
                   time.getMinutes().toString().padStart(2, '0'));
    }
    return labels;
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

function getChineseStatusName(status) {
    switch(status) {
        case 'online': return '在线';
        case 'offline': return '离线';
        case 'warning': return '警告';
        case 'unconfigured': return '未配置';
        default: return status;
    }
}

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

function updateModulesTable(modules) {
    const tbody = document.getElementById('modulesTableBody');
    
    // 清空现有内容
    tbody.innerHTML = '';

    // 检查数据状态
    if (!modules) {
        // 数据还在加载中
        tbody.innerHTML = `
            <tr>
                <td colspan="7" style="text-align: center; color: var(--text-secondary); padding: 2rem;">
                    <div class="loading">
                        <div class="spinner"></div>
                        加载中...
                    </div>
                </td>
            </tr>`;
        return;
    }

    if (modules.length === 0) {
        // 数据为空
        tbody.innerHTML = `
            <tr>
                <td colspan="7" style="text-align: center; color: var(--text-secondary); padding: 2rem;">
                    <div style="opacity: 0.7;">
                        <i class="fas fa-server" style="font-size: 2rem; margin-bottom: 1rem; opacity: 0.5;"></i>
                        <div style="font-size: 1.1rem; margin-bottom: 0.5rem;">暂无模块数据</div>
                        <div style="font-size: 0.9rem; opacity: 0.8;">
                            <a href="#" onclick="showAddModuleModal()" style="color: var(--primary-color); text-decoration: none;">
                                <i class="fas fa-plus me-1"></i>点击添加第一个模块
                            </a>
                        </div>
                    </div>
                </td>
            </tr>`;
        return;
    }

    // 渲染模块数据
    modules.forEach(module => {
        const row = tbody.insertRow();
        row.innerHTML = `
            <td style="font-weight: 600;">${module.name || '未知'}</td>
            <td><span class="status-badge status-${getStatusClass(module.status)}">${getStatusText(module.status)}</span></td>
            <td>${module.location || '--'}</td>
            <td>${module.ip_address || '--'}</td>
            <td>${formatDateTime(module.last_seen)}</td>
            <td>${formatBytes(module.total_traffic || 0)}</td>
            <td>
                <div class="table-actions">
                    <button class="btn btn-sm btn-outline-primary" onclick="downloadModuleConfig('${module.id}')" title="下载模块配置">
                        <i class="fas fa-download"></i>
                    </button>
                    <button class="btn btn-sm btn-outline-info" onclick="showModuleUsers('${module.id}')" title="管理用户">
                        <i class="fas fa-users"></i>
                    </button>
                    <button class="btn btn-sm btn-outline-danger" onclick="deleteModule('${module.id}')" title="删除模块">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </td>
        `;
    });
}

function getStatusClass(status) {
    switch(status) {
        case '在线': return 'online';
        case '离线': return 'offline';
        case '警告': return 'error';
        case '未配置': return 'offline';
        default: return 'offline';
    }
}

function getStatusText(status) {
    // API返回的已经是中文状态，直接返回
    return status || '未知';
}

function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatDateTime(dateStr) {
    if (!dateStr) return '--';
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN');
}

async function switchTimeRange(range) {
    currentTimeRange = range;
    document.querySelectorAll('.chart-control-btn').forEach(btn => btn.classList.remove('active'));
    event.target.classList.add('active');
    
    // 更新图表数据
    await updateTrafficChart(range);
}

async function refreshAllData() {
    await loadAllData();
}

function logout() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    window.location.href = '/login';
}

// 模块管理功能
async function showAddModuleModal() {
    const modal = new bootstrap.Modal(document.getElementById('addModuleModal'));
    
    // 添加模态框关闭事件监听器
    const modalElement = document.getElementById('addModuleModal');
    modalElement.addEventListener('hidden.bs.modal', function () {
        document.body.classList.remove('modal-open');
        document.body.style.removeProperty('padding-right');
        const backdrops = document.querySelectorAll('.modal-backdrop');
        backdrops.forEach(backdrop => backdrop.remove());
    }, { once: true });
    
    // 设置自定义网段输入的事件监听
    const allowedIPsSelect = document.getElementById('moduleAllowedIPs');
    const customInput = document.getElementById('moduleAllowedIPsCustom');
    
    if (allowedIPsSelect && customInput) {
        allowedIPsSelect.addEventListener('change', function() {
            if (this.value === '') {
                customInput.classList.remove('d-none');
                customInput.required = true;
                customInput.focus();
            } else {
                customInput.classList.add('d-none');
                customInput.required = false;
                customInput.value = '';
            }
        });
    }
    
    loadWireGuardInterfaces(); // 加载接口列表
    modal.show();
}

// 加载WireGuard接口列表
async function loadWireGuardInterfaces() {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/interfaces', {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (response.ok) {
            const result = await response.json();
            console.log('Interfaces loaded:', result);
            const interfaces = result.data || result.interfaces || [];
            const select = document.getElementById('moduleInterface');
            
            // 清空现有选项
            select.innerHTML = '<option value="">选择WireGuard接口</option>';
            
            // 添加接口选项
            interfaces.forEach(iface => {
                const option = document.createElement('option');
                option.value = iface.id;
                option.textContent = `${iface.name} - ${iface.description} (${iface.network})`;
                option.dataset.maxPeers = iface.max_peers;
                option.dataset.totalPeers = iface.total_peers;
                
                // 如果接口已满，禁用选项
                if (iface.total_peers >= iface.max_peers) {
                    option.disabled = true;
                    option.textContent += ' [已满]';
                } else {
                    option.textContent += ` [${iface.total_peers}/${iface.max_peers}]`;
                }
                
                select.appendChild(option);
            });
            
            console.log(`加载了 ${interfaces.length} 个接口`);
        } else {
            console.error('加载接口列表失败:', response.status, response.statusText);
            const select = document.getElementById('moduleInterface');
            select.innerHTML = '<option value="">加载接口失败，请刷新重试</option>';
        }
    } catch (error) {
        console.error('加载接口列表失败:', error);
        const select = document.getElementById('moduleInterface');
        select.innerHTML = '<option value="">网络错误，请刷新重试</option>';
    }
}

async function submitAddModule() {
    const form = document.getElementById('addModuleForm');
    const formData = new FormData(form);
    
    // 处理网段配置 - 支持自定义输入
    let allowedIPs = formData.get('allowed_ips');
    const customAllowedIPs = document.getElementById('moduleAllowedIPsCustom').value.trim();
    
    if (allowedIPs === '' && customAllowedIPs) {
        allowedIPs = customAllowedIPs;
    } else if (!allowedIPs) {
        allowedIPs = '192.168.50.0/24'; // 默认使用配置文档中的网段
    }
    
    // 收集所有表单数据
    const data = {
        name: formData.get('name'),
        location: formData.get('location'),
        description: formData.get('description') || '',
        interface_id: parseInt(formData.get('interface_id')),
        allowed_ips: allowedIPs,
        local_ip: formData.get('local_ip') || '', // 模块内网IP地址
        persistent_keepalive: parseInt(formData.get('persistent_keepalive')) || 25,
        dns: formData.get('dns') || '8.8.8.8,8.8.4.4',
        auto_generate_keys: document.getElementById('autoGenerateKeys').checked,
        auto_assign_ip: document.getElementById('autoAssignIP').checked,
        config_template: formData.get('config_template') || 'default'
    };

    console.log('提交的模块数据:', data); // 添加调试日志

    // 验证必填字段
    if (!data.name || !data.location || !data.interface_id) {
        alert('请填写模块名称、位置并选择WireGuard接口');
        return;
    }
    
    // 验证网段格式
    if (!data.allowed_ips || !validateNetworkFormat(data.allowed_ips)) {
        alert('请选择或输入有效的网段格式（如：192.168.50.0/24）');
        return;
    }

    // 验证保活间隔
    if (data.persistent_keepalive < 0 || data.persistent_keepalive > 300) {
        alert('保活间隔必须在0-300秒之间');
        return;
    }
    
    // 🔒 安全检查：检查接口状态是否允许添加模块
    if (!await checkInterfaceEditPermission(data.interface_id, '添加模块')) {
        return;
    }

    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/modules', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify(data)
        });

        const result = await response.json();
        if (response.ok) {
            alert(`模块创建成功！\n\n配置信息：\n- 模块名称：${data.name}\n- 内网网段：${data.allowed_ips}\n- 配置已自动生成并分配到指定接口`);
            bootstrap.Modal.getInstance(document.getElementById('addModuleModal')).hide();
            form.reset();
            loadAllData(); // 刷新数据
        } else {
            alert('创建失败：' + result.message);
        }
    } catch (error) {
        console.error('创建模块失败:', error);
        alert('网络错误，请重试');
    }
}

// 验证网段格式
function validateNetworkFormat(networks) {
    if (!networks) return false;
    
    const networkList = networks.split(',').map(n => n.trim());
    const cidrRegex = /^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/;
    
    return networkList.every(network => {
        if (!cidrRegex.test(network)) return false;
        
        const [ip, mask] = network.split('/');
        const maskNum = parseInt(mask);
        
        // 验证子网掩码范围
        if (maskNum < 8 || maskNum > 30) return false;
        
        // 验证IP地址格式
        const ipParts = ip.split('.').map(part => parseInt(part));
        return ipParts.every(part => part >= 0 && part <= 255);
    });
}

// WireGuard配置管理 - 重新设计为系统级管理面板
async function showWireGuardConfig() {
    const modal = new bootstrap.Modal(document.getElementById('wireGuardConfigModal'));
    
    // 添加模态框关闭事件监听器
    const modalElement = document.getElementById('wireGuardConfigModal');
    modalElement.addEventListener('hidden.bs.modal', function () {
        document.body.classList.remove('modal-open');
        document.body.style.removeProperty('padding-right');
        const backdrops = document.querySelectorAll('.modal-backdrop');
        backdrops.forEach(backdrop => backdrop.remove());
    }, { once: true });
    
    modal.show();
    
    try {
        const token = localStorage.getItem('access_token');
        
        // 获取系统配置（服务器信息）
        const configResponse = await fetch('/api/v1/config', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        // 获取接口统计信息
        const interfaceStatsResponse = await fetch('/api/v1/interfaces/stats', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (configResponse.ok && interfaceStatsResponse.ok) {
            const configResult = await configResponse.json();
            const interfaceStatsResult = await interfaceStatsResponse.json();
            const config = configResult.data;
            const interfaceStats = interfaceStatsResult.data;
            
            let content = `
                <!-- 系统状态概览 -->
                <div class="row mb-4">
                    <div class="col-md-12">
                        <div class="card" style="background: rgba(30, 41, 59, 0.8); border: 1px solid rgba(100, 116, 139, 0.3);">
                            <div class="card-header" style="background: rgba(15, 23, 42, 0.6); border-bottom: 1px solid rgba(100, 116, 139, 0.3);">
                                <h6 class="mb-0" style="color: #f1f5f9;"><i class="fas fa-server me-2"></i>系统状态概览</h6>
                            </div>
                            <div class="card-body">
                                <div class="row">
                                    <div class="col-md-3">
                                        <div class="text-center">
                                            <div class="h4" style="color: #3b82f6;">${interfaceStats.total_interfaces}</div>
                                            <div style="color: #94a3b8;">总接口数</div>
                                        </div>
                                    </div>
                                    <div class="col-md-3">
                                        <div class="text-center">
                                            <div class="h4" style="color: #10b981;">${interfaceStats.active_interfaces}</div>
                                            <div style="color: #94a3b8;">运行中</div>
                                        </div>
                                    </div>
                                    <div class="col-md-3">
                                        <div class="text-center">
                                            <div class="h4" style="color: #06b6d4;">${interfaceStats.total_capacity}</div>
                                            <div style="color: #94a3b8;">总容量</div>
                                        </div>
                                    </div>
                                    <div class="col-md-3">
                                        <div class="text-center">
                                            <div class="h4" style="color: #f59e0b;">${interfaceStats.used_capacity}</div>
                                            <div style="color: #94a3b8;">已使用</div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- 服务器配置 -->
                <div class="row mb-4">
                    <div class="col-md-6">
                        <div class="card" style="background: rgba(30, 41, 59, 0.8); border: 1px solid rgba(100, 116, 139, 0.3);">
                            <div class="card-header" style="background: rgba(15, 23, 42, 0.6); border-bottom: 1px solid rgba(100, 116, 139, 0.3);">
                                <h6 class="mb-0" style="color: #f1f5f9;"><i class="fas fa-server me-2"></i>服务器信息</h6>
                            </div>
                            <div class="card-body">
                                <div class="mb-3">
                                    <label class="form-label" style="color: #e2e8f0;">服务器名称</label>
                                    <input type="text" class="form-control" value="${config.server.name}" readonly 
                                           style="background: rgba(15, 23, 42, 0.6); border: 1px solid rgba(100, 116, 139, 0.3); color: #f1f5f9;">
                                </div>
                                <div class="mb-3">
                                    <label class="form-label" style="color: #e2e8f0;">外网端点</label>
                                    <input type="text" class="form-control" value="${config.server.endpoint || '未配置'}" readonly 
                                           style="background: rgba(15, 23, 42, 0.6); border: 1px solid rgba(100, 116, 139, 0.3); color: #f1f5f9;">
                                </div>
                                <div class="mb-0">
                                    <label class="form-label" style="color: #e2e8f0;">Web管理端口</label>
                                    <input type="number" class="form-control" value="${config.server.web_port}" readonly 
                                           style="background: rgba(15, 23, 42, 0.6); border: 1px solid rgba(100, 116, 139, 0.3); color: #f1f5f9;">
                                </div>
                            </div>
                        </div>
                    </div>
                    <div class="col-md-6">
                        <div class="card" style="background: rgba(30, 41, 59, 0.8); border: 1px solid rgba(100, 116, 139, 0.3);">
                            <div class="card-header" style="background: rgba(15, 23, 42, 0.6); border-bottom: 1px solid rgba(100, 116, 139, 0.3);">
                                <h6 class="mb-0" style="color: #f1f5f9;"><i class="fas fa-chart-bar me-2"></i>系统统计</h6>
                            </div>
                            <div class="card-body">
                                <div class="mb-3">
                                    <div class="d-flex justify-content-between" style="color: #e2e8f0;">
                                        <span>接口使用率</span>
                                        <span>${interfaceStats.active_interfaces}/${interfaceStats.total_interfaces}</span>
                                    </div>
                                    <div class="progress mt-1" style="background: rgba(15, 23, 42, 0.6);">
                                        <div class="progress-bar bg-success" style="width: ${interfaceStats.total_interfaces > 0 ? (interfaceStats.active_interfaces / interfaceStats.total_interfaces * 100) : 0}%"></div>
                                    </div>
                                </div>
                                <div class="mb-3">
                                    <div class="d-flex justify-content-between" style="color: #e2e8f0;">
                                        <span>连接使用率</span>
                                        <span>${interfaceStats.used_capacity}/${interfaceStats.total_capacity}</span>
                                    </div>
                                    <div class="progress mt-1" style="background: rgba(15, 23, 42, 0.6);">
                                        <div class="progress-bar bg-info" style="width: ${interfaceStats.total_capacity > 0 ? (interfaceStats.used_capacity / interfaceStats.total_capacity * 100) : 0}%"></div>
                                    </div>
                                </div>
                                <div class="mb-0">
                                    <div class="d-flex justify-content-between" style="color: #e2e8f0;">
                                        <span>系统状态</span>
                                        <span class="badge ${interfaceStats.active_interfaces > 0 ? 'bg-success' : 'bg-warning'}">
                                            ${interfaceStats.active_interfaces > 0 ? '运行中' : '待启动'}
                                        </span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- 接口概览 -->
                <div class="row mb-4">
                    <div class="col-md-12">
                        <div class="card" style="background: rgba(30, 41, 59, 0.8); border: 1px solid rgba(100, 116, 139, 0.3);">
                            <div class="card-header" style="background: rgba(15, 23, 42, 0.6); border-bottom: 1px solid rgba(100, 116, 139, 0.3);">
                                <h6 class="mb-0" style="color: #f1f5f9;"><i class="fas fa-list me-2"></i>接口概览</h6>
                            </div>
                            <div class="card-body" style="padding: 0;">
                                <div class="table-responsive">
                                    <table style="width: 100%; margin: 0; background: transparent; color: #e2e8f0;">
                                        <thead>
                                            <tr style="background: rgba(15, 23, 42, 0.8); border-bottom: 2px solid rgba(100, 116, 139, 0.4);">
                                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">接口名称</th>
                                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">状态</th>
                                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">网络段</th>
                                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">端口</th>
                                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">连接数</th>
                                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">操作</th>
                                            </tr>
                                        </thead>
                                        <tbody style="background: transparent;">
                                            ${interfaceStats.interfaces.map(iface => `
                                                <tr style="background: rgba(30, 41, 59, 0.3); border-bottom: 1px solid rgba(100, 116, 139, 0.2); transition: background-color 0.2s ease;" 
                                                    onmouseover="this.style.background='rgba(30, 41, 59, 0.6)'" 
                                                    onmouseout="this.style.background='rgba(30, 41, 59, 0.3)'">
                                                    <td style="border: none; padding: 12px 16px;">
                                                        <div style="display: flex; align-items: center;">
                                                            <i class="fas fa-ethernet" style="color: #60a5fa; margin-right: 8px; font-size: 14px;"></i>
                                                            <div>
                                                                <div style="color: #f1f5f9; font-size: 14px; font-weight: 600;">${iface.name}</div>
                                                                <small style="color: #94a3b8;">${iface.description || '无描述'}</small>
                                                            </div>
                                                        </div>
                                                    </td>
                                                    <td style="border: none; padding: 12px 16px;">
                                                        <span class="badge bg-${iface.status === 1 ? 'success' : 'secondary'}">${iface.status === 1 ? '运行中' : '已停止'}</span>
                                                    </td>
                                                    <td style="border: none; padding: 12px 16px;">
                                                        <span style="background: rgba(15, 23, 42, 0.8); color: #a78bfa; padding: 4px 8px; border-radius: 6px; font-family: 'Courier New', monospace; font-size: 13px; font-weight: 500;">${iface.network}</span>
                                                        <br>
                                                        <small style="color: #94a3b8; margin-top: 4px; display: inline-block;">服务器IP: ${iface.server_ip}</small>
                                                    </td>
                                                    <td style="border: none; padding: 12px 16px;">
                                                        <span style="background: rgba(15, 23, 42, 0.8); color: #34d399; padding: 4px 8px; border-radius: 6px; font-family: 'Courier New', monospace; font-size: 13px; font-weight: 500;">${iface.listen_port}</span>
                                                    </td>
                                                    <td style="border: none;">
                                                        <div style="position: relative;">
                                                            <div class="progress" style="height: 20px; background: rgba(15, 23, 42, 0.6);">
                                                                <div class="progress-bar bg-info" role="progressbar" 
                                                                     style="width: ${(iface.total_peers / iface.max_peers * 100).toFixed(1)}%">
                                                                </div>
                                                            </div>
                                                            <div style="position: absolute; top: 0; left: 0; right: 0; bottom: 0; display: flex; align-items: center; justify-content: center;">
                                                                <span style="color: #f1f5f9; font-weight: 600; font-size: 0.8rem; text-shadow: 1px 1px 2px rgba(0,0,0,0.8);">${iface.total_peers}/${iface.max_peers}</span>
                                                            </div>
                                                        </div>
                                                    </td>
                                                    <td style="border: none; padding: 12px 16px;">
                                                        <div class="btn-group btn-group-sm">
                                                            ${iface.status === 1 ? 
                                                                `<button onclick="stopInterface(${iface.id})" 
                                                                        style="background: transparent; border: 1px solid #f59e0b; color: #fbbf24; padding: 4px 8px; border-radius: 4px; font-size: 12px; cursor: pointer; margin-right: 4px; transition: all 0.2s ease;"
                                                                        onmouseover="this.style.background='rgba(245, 158, 11, 0.1)'"
                                                                        onmouseout="this.style.background='transparent'">
                                                                    <i class="fas fa-stop" style="margin-right: 4px;"></i>停止
                                                                </button>` :
                                                                `<button onclick="startInterface(${iface.id})" 
                                                                        style="background: transparent; border: 1px solid #10b981; color: #34d399; padding: 4px 8px; border-radius: 4px; font-size: 12px; cursor: pointer; margin-right: 4px; transition: all 0.2s ease;"
                                                                        onmouseover="this.style.background='rgba(16, 185, 129, 0.1)'"
                                                                        onmouseout="this.style.background='transparent'">
                                                                    <i class="fas fa-play" style="margin-right: 4px;"></i>启动
                                                                </button>`
                                                            }
                                                            <button onclick="viewInterfaceConfig(${iface.id})" 
                                                                    style="background: transparent; border: 1px solid #06b6d4; color: #22d3ee; padding: 4px 8px; border-radius: 4px; font-size: 12px; cursor: pointer; transition: all 0.2s ease;"
                                                                    onmouseover="this.style.background='rgba(6, 182, 212, 0.1)'"
                                                                    onmouseout="this.style.background='transparent'">
                                                                <i class="fas fa-eye" style="margin-right: 4px;"></i>查看
                                                            </button>
                                                        </div>
                                                    </td>
                                                </tr>
                                            `).join('')}
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- 快捷操作 -->
                <div class="row">
                    <div class="col-md-12">
                        <div class="d-flex justify-content-between">
                            <div>
                                <button class="btn btn-primary me-2" onclick="showInterfaceManager()">
                                    <i class="fas fa-ethernet me-1"></i>接口管理
                                </button>
                                <button class="btn btn-success me-2" onclick="createNewInterface()">
                                    <i class="fas fa-plus me-1"></i>创建接口
                                </button>
                                <button class="btn btn-secondary me-2" onclick="refreshSystemConfig()">
                                    <i class="fas fa-sync-alt me-1"></i>刷新状态
                                </button>
                            </div>
                            <div>
                                <button class="btn btn-info me-2" onclick="exportSystemConfig()">
                                    <i class="fas fa-download me-1"></i>导出配置
                                </button>
                                <button class="btn btn-warning me-2" onclick="viewSystemLogs()">
                                    <i class="fas fa-file-alt me-1"></i>系统日志
                                </button>
                                <button class="btn btn-outline-danger" onclick="systemSettings()">
                                    <i class="fas fa-cog me-1"></i>系统设置
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            `;
            
            document.getElementById('wgConfigContent').innerHTML = content;
        } else {
            document.getElementById('wgConfigContent').innerHTML = '<div class="alert alert-danger">加载系统信息失败</div>';
        }
    } catch (error) {
        console.error('加载系统配置失败:', error);
        document.getElementById('wgConfigContent').innerHTML = '<div class="alert alert-danger">网络错误</div>';
    }
}

// 获取接口状态样式
function getInterfaceStatusClass(status) {
    switch(status) {
        case 1: return 'bg-success';  // 运行中
        case 0: return 'bg-secondary'; // 停止
        case 2: return 'bg-danger';   // 错误
        case 3: return 'bg-warning';  // 启动中
        case 4: return 'bg-warning';  // 停止中
        default: return 'bg-secondary';
    }
}

// 获取接口状态文本
function getInterfaceStatusText(status) {
    switch(status) {
        case 1: return '运行中';
        case 0: return '已停止';
        case 2: return '错误';
        case 3: return '启动中';
        case 4: return '停止中';
        default: return '未知';
    }
}

// 查看接口详情
function viewInterfaceDetails(interfaceId) {
    // 关闭当前模态框
    const currentModal = bootstrap.Modal.getInstance(document.getElementById('wireGuardConfigModal'));
    if (currentModal) {
        currentModal.hide();
    }
    
    // 打开接口管理器并定位到指定接口
    setTimeout(() => {
        showInterfaceManager();
        // 可以在这里添加高亮显示指定接口的逻辑
    }, 300);
}

// 刷新系统配置
function refreshSystemConfig() {
    showWireGuardConfig();
}

// 导出系统配置
async function exportSystemConfig() {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/config/export', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.ok) {
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'eitec-vpn-config.json';
            a.click();
            window.URL.revokeObjectURL(url);
        } else {
            alert('导出配置失败');
        }
    } catch (error) {
        console.error('导出配置失败:', error);
        alert('导出配置失败');
    }
}

// 查看系统日志
function viewSystemLogs() {
    alert('系统日志功能开发中...');
}

// 系统设置
function systemSettings() {
    alert('系统设置功能开发中...');
}

async function initializeWireGuard() {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/config/wireguard/init', {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        const result = await response.json();
        if (response.ok) {
            alert('WireGuard初始化成功！');
            showWireGuardConfig(); // 刷新配置显示
        } else {
            alert('初始化失败：' + result.message);
        }
    } catch (error) {
        console.error('初始化WireGuard失败:', error);
        alert('网络错误，请重试');
    }
}

async function downloadServerConfig() {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/config/wireguard/server-config', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.ok) {
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'wg0.conf';
            a.click();
            window.URL.revokeObjectURL(url);
        } else {
            alert('下载失败');
        }
    } catch (error) {
        console.error('下载配置失败:', error);
        alert('网络错误，请重试');
    }
}

async function applyWireGuardConfig() {
    if (!confirm('确定要应用WireGuard配置吗？这将重启WireGuard服务。')) {
        return;
    }
    
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/config/wireguard/apply', {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        const result = await response.json();
        if (response.ok) {
            alert('配置应用成功！');
        } else {
            alert('应用失败：' + result.message);
        }
    } catch (error) {
        console.error('应用配置失败:', error);
        alert('网络错误，请重试');
    }
}

// 用户VPN管理
function showUserVPNManager() {
    // 这个功能已经整合到模块管理中，通过点击模块的"管理用户"按钮来访问
    alert('用户VPN管理功能已整合到模块管理中！\n\n请在模块列表中点击"管理用户"按钮来管理模块的用户VPN配置。');
}

// 模块操作功能
async function downloadModuleConfig(id) {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch(`/api/v1/modules/${id}/config`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.ok) {
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `module_${id}_config.conf`;
            a.click();
            window.URL.revokeObjectURL(url);
        } else {
            alert('下载失败');
        }
    } catch (error) {
        console.error('下载模块配置失败:', error);
        alert('网络错误，请重试');
    }
}

// generateUserConfig 函数已删除，统一使用管理用户功能

function editModule(id) {
    alert('编辑模块功能开发中...');
}

async function deleteModule(id) {
    if (!confirm('确定要删除此模块吗？此操作不可撤销！')) {
        return;
    }
    
    // 显示删除中状态
    const deleteBtn = event.target.closest('button');
    const originalContent = deleteBtn.innerHTML;
    deleteBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
    deleteBtn.disabled = true;
    
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch(`/api/v1/modules/${id}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        const result = await response.json();
        if (response.ok) {
            // 显示成功消息
            alert('模块删除成功！');
            
            // 立即刷新所有数据
            console.log('删除成功，开始刷新数据...');
            await loadAllData();
            
            // 确保模块表格得到更新
            setTimeout(() => {
                console.log('延迟刷新确保数据同步...');
                loadAllData();
            }, 1000);
            
        } else {
            alert('删除失败：' + result.message);
            // 恢复按钮状态
            deleteBtn.innerHTML = originalContent;
            deleteBtn.disabled = false;
        }
    } catch (error) {
        console.error('删除模块失败:', error);
        alert('网络错误，请重试');
        // 恢复按钮状态
        deleteBtn.innerHTML = originalContent;
        deleteBtn.disabled = false;
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

// 显示接口管理器
async function showInterfaceManager() {
    const modal = new bootstrap.Modal(document.getElementById('interfaceManagerModal'));
    
    // 添加模态框关闭事件监听器
    const modalElement = document.getElementById('interfaceManagerModal');
    modalElement.addEventListener('hidden.bs.modal', function () {
        document.body.classList.remove('modal-open');
        document.body.style.removeProperty('padding-right');
        const backdrops = document.querySelectorAll('.modal-backdrop');
        backdrops.forEach(backdrop => backdrop.remove());
    }, { once: true });
    
    modal.show();
    
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/interfaces', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.ok) {
            const result = await response.json();
            const interfaces = result.data || result;
            
            let content = `
                <div class="mb-3">
                    <h6 style="color: #f1f5f9;"><i class="fas fa-network-wired me-2"></i>WireGuard接口管理</h6>
                    <p style="color: #94a3b8;">管理系统中的所有WireGuard接口，每个接口对应不同的网络段和端口。</p>
                    <button class="btn btn-success btn-sm" onclick="showCreateInterfaceModal()">
                        <i class="fas fa-plus me-1"></i>创建新接口
                    </button>
                </div>
                
                <div class="table-responsive">
                    <table style="width: 100%; margin: 0; background: transparent; color: #e2e8f0;">
                        <thead>
                            <tr style="background: rgba(15, 23, 42, 0.8); border-bottom: 2px solid rgba(100, 116, 139, 0.4);">
                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">接口名称</th>
                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">状态</th>
                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">网络段</th>
                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">端口</th>
                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">连接数</th>
                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">操作</th>
                            </tr>
                        </thead>
                        <tbody style="background: transparent;">
            `;
            
            if (interfaces.length === 0) {
                content += `
                    <tr>
                        <td colspan="6" style="text-align: center; padding: 2rem; color: #94a3b8;">
                            <i class="fas fa-network-wired" style="font-size: 3rem; margin-bottom: 1rem; opacity: 0.5;"></i>
                            <div>暂无WireGuard接口</div>
                            <div style="margin-top: 0.5rem;">
                                <button class="btn btn-primary btn-sm" onclick="showCreateInterfaceModal()">
                                    <i class="fas fa-plus me-1"></i>创建第一个接口
                                </button>
                            </div>
                        </td>
                    </tr>
                `;
            } else {
                interfaces.forEach(iface => {
                    let statusClass = 'secondary';
                    let statusText = '未知';
                    let statusIcon = 'fas fa-question-circle';
                    
                    switch (iface.status) {
                        case 0: // Down
                            statusClass = 'secondary';
                            statusText = '已停止';
                            statusIcon = 'fas fa-stop-circle';
                            break;
                        case 1: // Up
                            statusClass = 'success';
                            statusText = '运行中';
                            statusIcon = 'fas fa-play-circle';
                            break;
                        case 2: // Error
                            statusClass = 'danger';
                            statusText = '错误';
                            statusIcon = 'fas fa-exclamation-circle';
                            break;
                        case 3: // Starting
                            statusClass = 'warning';
                            statusText = '启动中';
                            statusIcon = 'fas fa-spinner fa-spin';
                            break;
                        case 4: // Stopping
                            statusClass = 'warning';
                            statusText = '停止中';
                            statusIcon = 'fas fa-spinner fa-spin';
                            break;
                    }
                    
                    content += `
                        <tr style="background: rgba(30, 41, 59, 0.3); border-bottom: 1px solid rgba(100, 116, 139, 0.2); transition: background-color 0.2s ease;" 
                            onmouseover="this.style.background='rgba(30, 41, 59, 0.6)'" 
                            onmouseout="this.style.background='rgba(30, 41, 59, 0.3)'">
                            <td style="border: none; padding: 12px 16px;">
                                <div>
                                    <div style="color: #f1f5f9; font-size: 14px; font-weight: 600;">${iface.name}</div>
                                    <small style="color: #94a3b8;">${iface.description || '无描述'}</small>
                                </div>
                            </td>
                            <td style="border: none; padding: 12px 16px;">
                                <span class="badge bg-${statusClass}" style="font-size: 11px; padding: 4px 8px;">
                                    <i class="${statusIcon}" style="margin-right: 4px;"></i>${statusText}
                                </span>
                            </td>
                            <td style="border: none; padding: 12px 16px;">
                                <span style="background: rgba(15, 23, 42, 0.8); color: #34d399; padding: 4px 8px; border-radius: 6px; font-family: 'Courier New', monospace; font-size: 13px; font-weight: 500;">${iface.network}</span>
                            </td>
                            <td style="border: none; padding: 12px 16px; color: #e2e8f0; font-size: 13px;">${iface.listen_port}</td>
                            <td style="border: none; padding: 12px 16px; color: #e2e8f0; font-size: 13px;">${iface.total_peers || 0}/${iface.max_peers || 0}</td>
                            <td style="border: none; padding: 12px 16px;">
                                <div style="display: flex; gap: 0.25rem;">
                    `;
                    
                    // 根据状态显示不同的操作按钮
                    if (iface.status === 1) { // 运行中
                        content += `
                            <button class="btn btn-sm btn-outline-warning" onclick="stopInterface(${iface.id})" title="停止接口">
                                <i class="fas fa-stop"></i>
                            </button>
                        `;
                    } else if (iface.status === 0) { // 已停止
                        content += `
                            <button class="btn btn-sm btn-outline-success" onclick="startInterface(${iface.id})" title="启动接口">
                                <i class="fas fa-play"></i>
                            </button>
                        `;
                    }
                    
                    content += `
                                    <button class="btn btn-sm btn-outline-info" onclick="showInterfaceConfig(${iface.id})" title="查看配置">
                                        <i class="fas fa-file-code"></i>
                                    </button>
                                    <button class="btn btn-sm btn-outline-danger" onclick="deleteInterface(${iface.id})" title="删除接口">
                                        <i class="fas fa-trash"></i>
                                    </button>
                                </div>
                            </td>
                        </tr>
                    `;
                });
            }
            
            content += `
                        </tbody>
                    </table>
                </div>
                
                <div class="mt-3">
                    <div class="alert alert-info" style="background: rgba(59, 130, 246, 0.1); border: 1px solid rgba(59, 130, 246, 0.3); color: #f1f5f9;">
                        <i class="fas fa-info-circle me-2"></i>
                        <strong>操作说明：</strong>
                        <ul class="mb-0 mt-2" style="padding-left: 1.5rem;">
                            <li>🔴 <strong>重要</strong>：修改接口配置前请先停止相关接口</li>
                            <li>接口停止后可以安全地添加/删除模块和用户</li>
                            <li>配置完成后重新启动接口以应用新的配置</li>
                            <li>删除接口前请确保没有关联的模块</li>
                        </ul>
                    </div>
                </div>
            `;
            
            document.getElementById('interfaceManagerContent').innerHTML = content;
        } else {
            document.getElementById('interfaceManagerContent').innerHTML = `
                <div class="alert alert-danger">
                    <i class="fas fa-exclamation-triangle me-2"></i>
                    加载接口列表失败
                </div>
            `;
        }
    } catch (error) {
        console.error('加载接口管理失败:', error);
        document.getElementById('interfaceManagerContent').innerHTML = `
            <div class="alert alert-danger">
                <i class="fas fa-exclamation-triangle me-2"></i>
                网络错误，请重试
            </div>
        `;
    }
}

// 启动接口
async function startInterface(interfaceId) {
    if (!confirm('确定要启动此接口吗？\n\n启动后接口将开始监听端口并可以接受连接。')) {
        return;
    }
    
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch(`/api/v1/interfaces/${interfaceId}/start`, {
            method: 'PUT',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        const result = await response.json();
        if (response.ok) {
            alert('接口启动成功！');
            // 刷新接口管理界面
            if (document.getElementById('interfaceManagerModal').classList.contains('show')) {
                showInterfaceManager();
            }
        } else {
            alert('接口启动失败：' + (result.message || '未知错误'));
        }
    } catch (error) {
        console.error('启动接口失败:', error);
        alert('网络错误，请重试');
    }
}

// 停止接口
async function stopInterface(interfaceId) {
    if (!confirm('确定要停止此接口吗？\n\n停止后所有连接将断开，可以安全地修改配置。')) {
        return;
    }
    
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch(`/api/v1/interfaces/${interfaceId}/stop`, {
            method: 'PUT',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        const result = await response.json();
        if (response.ok) {
            alert('接口停止成功！现在可以安全地修改配置了。');
            // 刷新接口管理界面
            if (document.getElementById('interfaceManagerModal').classList.contains('show')) {
                showInterfaceManager();
            }
        } else {
            alert('接口停止失败：' + (result.message || '未知错误'));
        }
    } catch (error) {
        console.error('停止接口失败:', error);
        alert('网络错误，请重试');
    }
}

// 删除接口
async function deleteInterface(interfaceId) {
    if (!confirm('⚠️ 危险操作：确定要删除此接口吗？\n\n删除后：\n- 接口配置将永久丢失\n- 关联的模块和用户将被删除\n- 此操作不可撤销')) {
        return;
    }
    
    // 二次确认
    const confirmText = prompt('请输入 "DELETE" 确认删除操作：');
    if (confirmText !== 'DELETE') {
        alert('操作已取消');
        return;
    }
    
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch(`/api/v1/interfaces/${interfaceId}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        const result = await response.json();
        if (response.ok) {
            alert('接口删除成功！');
            // 刷新接口管理界面
            if (document.getElementById('interfaceManagerModal').classList.contains('show')) {
                showInterfaceManager();
            }
            // 刷新主页面数据
            loadAllData();
        } else {
            alert('接口删除失败：' + (result.message || '未知错误'));
        }
    } catch (error) {
        console.error('删除接口失败:', error);
        alert('网络错误，请重试');
    }
}

// 查看接口配置
async function showInterfaceConfig(interfaceId) {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch(`/api/v1/interfaces/${interfaceId}/config`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.ok) {
            const result = await response.json();
            const interfaceInfo = result.data.interface;
            const configContent = result.data.config;
            
            // 显示配置内容模态框
            const content = `
                <div class="mb-3">
                    <h6 style="color: #f1f5f9;">
                        <i class="fas fa-file-code me-2"></i>接口配置：${interfaceInfo.name}
                    </h6>
                    <p style="color: #94a3b8;">
                        网络段：${interfaceInfo.network} | 端口：${interfaceInfo.listen_port} | 
                        状态：<span class="badge bg-${interfaceInfo.status === 1 ? 'success' : 'secondary'}">${interfaceInfo.status === 1 ? '运行中' : '已停止'}</span>
                    </p>
                </div>
                
                <div class="mb-3">
                    <label class="form-label" style="color: #e2e8f0;">配置文件内容：</label>
                    <textarea class="form-control" rows="20" readonly
                              style="background: rgba(15, 23, 42, 0.8); color: #34d399; font-family: 'Courier New', monospace; font-size: 0.875rem;">${configContent}</textarea>
                </div>
                
                <div class="alert alert-info" style="background: rgba(59, 130, 246, 0.1); border: 1px solid var(--primary-color); color: var(--text-primary);">
                    <i class="fas fa-info-circle me-2"></i>
                    <strong>配置说明：</strong>
                    <ul class="mb-0 mt-2">
                        <li>此配置遵循您的配置文档标准</li>
                        <li>已包含所有模块的Peer配置</li>
                        <li>支持完整的内网穿透功能</li>
                        <li>配置文件路径：/etc/wireguard/${interfaceInfo.name}.conf</li>
                    </ul>
                </div>
                
                <div class="mt-3">
                    <button class="btn btn-primary" onclick="downloadInterfaceConfig(${interfaceId})">
                        <i class="fas fa-download me-1"></i>下载配置文件
                    </button>
                </div>
            `;
            
            document.getElementById('userVPNContent').innerHTML = content;
            const modal = new bootstrap.Modal(document.getElementById('userVPNModal'));
            modal.show();
        } else {
            alert('获取配置失败');
        }
    } catch (error) {
        console.error('获取接口配置失败:', error);
        alert('网络错误，请重试');
    }
}

// 下载接口配置文件
async function downloadInterfaceConfig(interfaceId) {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch(`/api/v1/interfaces/${interfaceId}/config`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.ok) {
            const result = await response.json();
            const interfaceInfo = result.data.interface;
            const configContent = result.data.config;
            
            // 创建下载链接
            const blob = new Blob([configContent], { type: 'text/plain' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${interfaceInfo.name}.conf`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
        } else {
            alert('下载失败');
        }
    } catch (error) {
        console.error('下载配置失败:', error);
        alert('网络错误，请重试');
    }
}

// 创建新接口
function createNewInterface() {
    const modal = new bootstrap.Modal(document.getElementById('createInterfaceModal'));
    
    // 添加模态框关闭事件监听器
    const modalElement = document.getElementById('createInterfaceModal');
    modalElement.addEventListener('hidden.bs.modal', function () {
        document.body.classList.remove('modal-open');
        document.body.style.removeProperty('padding-right');
        const backdrops = document.querySelectorAll('.modal-backdrop');
        backdrops.forEach(backdrop => backdrop.remove());
    }, { once: true });
    
    modal.show();
    
    // 预填充建议的配置
    suggestInterfaceConfig();
}

// 建议接口配置
async function suggestInterfaceConfig() {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/interfaces/stats', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.ok) {
            const result = await response.json();
            const interfaces = result.data.interfaces;
            
            // 建议下一个接口名称
            const existingNames = interfaces.map(iface => iface.name);
            let suggestedName = '';
            for (let i = 0; i < 10; i++) {
                const name = `wg${i}`;
                if (!existingNames.includes(name)) {
                    suggestedName = name;
                    break;
                }
            }
            
            // 建议下一个端口
            const existingPorts = interfaces.map(iface => iface.listen_port);
            let suggestedPort = 51820;
            while (existingPorts.includes(suggestedPort)) {
                suggestedPort++;
            }
            
            // 建议下一个网络段
            const existingNetworks = interfaces.map(iface => iface.network);
            let suggestedNetwork = '';
            for (let i = 10; i < 100; i++) {
                const network = `10.${i}.0.0/24`;
                if (!existingNetworks.includes(network)) {
                    suggestedNetwork = network;
                    break;
                }
            }
            
            // 填充建议值
            if (suggestedName) {
                document.getElementById('interfaceName').value = suggestedName;
            }
            if (suggestedPort) {
                document.getElementById('interfacePort').value = suggestedPort;
            }
            if (suggestedNetwork) {
                document.getElementById('interfaceNetwork').value = suggestedNetwork;
            }
            
            // 建议描述
            const descriptions = {
                'wg0': '主接口 - 生产环境',
                'wg1': '北京节点专用',
                'wg2': '上海节点专用',
                'wg3': '广州节点专用',
                'wg4': '深圳节点专用',
                'wg5': '杭州节点专用'
            };
            
            if (descriptions[suggestedName]) {
                document.getElementById('interfaceDescription').value = descriptions[suggestedName];
            }
        }
    } catch (error) {
        console.error('获取接口建议配置失败:', error);
    }
}

// 提交创建接口
async function submitCreateInterface() {
    const form = document.getElementById('createInterfaceForm');
    const formData = new FormData(form);
    
    // 验证必填字段
    const name = formData.get('name');
    const network = formData.get('network');
    const listenPort = formData.get('listen_port');
    
    if (!name || !network || !listenPort) {
        alert('请填写所有必填字段');
        return;
    }
    
    // 验证网络段格式
    const networkRegex = /^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/;
    if (!networkRegex.test(network)) {
        alert('网络段格式不正确，请使用CIDR格式（如 10.50.0.0/24）');
        return;
    }
    
    // 验证端口范围
    const port = parseInt(listenPort);
    if (port < 1024 || port > 65535) {
        alert('端口范围应在1024-65535之间');
        return;
    }
    
    // 根据配置文档生成PostUp和PostDown规则
    const networkCIDR = network.trim();
    const interfaceName = name.trim();
    
    // 构建请求数据 - 参考配置文档的服务端配置
    const interfaceData = {
        name: interfaceName,
        description: formData.get('description') || '',
        network: networkCIDR,
        listen_port: port,
        dns: formData.get('dns') || '8.8.8.8,8.8.4.4',
        max_peers: parseInt(formData.get('max_peers')) || 50,
        mtu: parseInt(formData.get('mtu')) || 1420,
        // 根据配置文档生成标准的PostUp规则
        post_up: formData.get('post_up') || `iptables -t nat -A POSTROUTING -s ${networkCIDR} -o eth0 -j MASQUERADE; iptables -A INPUT -p udp -m udp --dport ${port} -j ACCEPT; iptables -I FORWARD 1 -i ${interfaceName} -j ACCEPT`,
        // 根据配置文档生成标准的PostDown规则（自动生成，因为表单中没有这个字段）
        post_down: `iptables -t nat -D POSTROUTING -s ${networkCIDR} -o eth0 -j MASQUERADE; iptables -D INPUT -p udp -m udp --dport ${port} -j ACCEPT; iptables -D FORWARD -i ${interfaceName} -j ACCEPT`,
        auto_start: document.getElementById('autoStartInterface') ? document.getElementById('autoStartInterface').checked : false
    };
    
    // 验证数据完整性
    console.log('接口数据验证:', {
        name: interfaceData.name,
        network: interfaceData.network,
        listen_port: interfaceData.listen_port,
        dns: interfaceData.dns,
        max_peers: interfaceData.max_peers,
        mtu: interfaceData.mtu,
        auto_start: interfaceData.auto_start
    });
    
    try {
        const token = localStorage.getItem('access_token');
        console.log('准备发送接口创建请求:', interfaceData);
        
        const response = await fetch('/api/v1/interfaces', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(interfaceData)
        });
        
        console.log('接收到响应状态码:', response.status);
        console.log('响应是否OK:', response.ok);
        
        const result = await response.json();
        console.log('响应内容:', result);
        
        if (response.ok) {
            // 关闭模态框
            const modal = bootstrap.Modal.getInstance(document.getElementById('createInterfaceModal'));
            modal.hide();
            
            // 清空表单
            form.reset();
            
            // 显示成功消息
            alert(`接口创建成功！\n\n配置信息：\n- 接口名称：${interfaceName}\n- 网络段：${networkCIDR}\n- 监听端口：${port}\n- 已按照标准配置生成防火墙规则`);
            
            // 刷新系统配置显示
            if (document.getElementById('wireGuardConfigModal').classList.contains('show')) {
                refreshSystemConfig();
            }
            
            // 刷新接口管理界面
            if (document.getElementById('interfaceManagerModal').classList.contains('show')) {
                showInterfaceManager();
            }
            
            // 刷新主页面数据
            loadAllData();
            
        } else {
            alert('创建接口失败：' + (result.message || '未知错误'));
        }
        
    } catch (error) {
        console.error('创建接口失败:', error);
        alert('创建接口失败：网络错误');
    }
}

// 显示创建接口模态框，并初始化配置模板
function showCreateInterfaceModal() {
    const modal = new bootstrap.Modal(document.getElementById('createInterfaceModal'));
    
    // 添加模态框关闭事件监听器
    const modalElement = document.getElementById('createInterfaceModal');
    modalElement.addEventListener('hidden.bs.modal', function () {
        document.body.classList.remove('modal-open');
        document.body.style.removeProperty('padding-right');
        const backdrops = document.querySelectorAll('.modal-backdrop');
        backdrops.forEach(backdrop => backdrop.remove());
    }, { once: true });
    
    // 设置配置模板选择器事件
    const templateSelect = document.getElementById('configTemplate');
    if (templateSelect) {
        templateSelect.addEventListener('change', function() {
            applyInterfaceTemplate(this.value);
        });
    }
    
    modal.show();
}

// 应用接口配置模板
function applyInterfaceTemplate(templateType) {
    const nameInput = document.getElementById('interfaceName');
    const networkInput = document.getElementById('interfaceNetwork');
    const portInput = document.getElementById('interfacePort');
    const descInput = document.getElementById('interfaceDescription');
    
    // 获取当前已有接口数量来生成建议配置
    const currentCount = document.querySelectorAll('#interfaceManagerContent tbody tr').length || 0;
    
    switch (templateType) {
        case 'standard':
            nameInput.value = `wg${currentCount}`;
            networkInput.value = `10.${50 + currentCount}.0.0/24`;
            portInput.value = 51820 + currentCount;
            descInput.value = `标准配置接口${currentCount + 1}`;
            break;
        case 'high-capacity':
            nameInput.value = `wg${currentCount}`;
            networkInput.value = `10.${50 + currentCount}.0.0/24`;
            portInput.value = 51820 + currentCount;
            descInput.value = `高容量接口${currentCount + 1}`;
            document.getElementById('interfaceMaxPeers').value = 200;
            break;
        case 'lan-bridge':
            nameInput.value = `wg${currentCount}`;
            networkInput.value = `10.${50 + currentCount}.0.0/24`;
            portInput.value = 51820 + currentCount;
            descInput.value = `内网穿透专用接口${currentCount + 1}`;
            break;
        default:
            // 默认配置：参考配置文档的推荐设置
            nameInput.value = `wg${currentCount}`;
            networkInput.value = `10.${50 + currentCount}.0.0/24`;
            portInput.value = 51820 + currentCount;
            descInput.value = `WireGuard接口${currentCount + 1}`;
    }
    
    // 更新配置预览
    updateInterfaceConfigPreview();
}

// 更新接口配置预览
function updateInterfaceConfigPreview() {
    const name = document.getElementById('interfaceName')?.value || 'wgX';
    const network = document.getElementById('interfaceNetwork')?.value || '10.50.0.0/24';
    const port = document.getElementById('interfacePort')?.value || '51824';
    
    const previewDiv = document.getElementById('interfaceConfigPreview');
    if (previewDiv) {
        const configExample = `[Interface]
Address = ${network.replace(/0\/24$/, '1/24')}
ListenPort = ${port}
MTU = 1420
SaveConfig = true
PostUp = iptables -t nat -A POSTROUTING -s ${network} -o eth0 -j MASQUERADE; iptables -A INPUT -p udp -m udp --dport ${port} -j ACCEPT; iptables -I FORWARD 1 -i ${name} -j ACCEPT
PostDown = iptables -t nat -D POSTROUTING -s ${network} -o eth0 -j MASQUERADE; iptables -D INPUT -p udp -m udp --dport ${port} -j ACCEPT; iptables -D FORWARD -i ${name} -j ACCEPT
PrivateKey = [自动生成]

# 模块端和客户端将通过配置自动添加为 [Peer] 段`;
        
        previewDiv.innerHTML = `
            <div class="mt-3">
                <h6 style="color: var(--primary-color);">
                    <i class="fas fa-eye me-2"></i>配置预览
                </h6>
                <pre style="background: rgba(15, 23, 42, 0.8); color: #34d399; padding: 1rem; border-radius: 0.375rem; font-size: 0.875rem; overflow-x: auto;">${configExample}</pre>
                <div class="alert alert-info" style="background: rgba(59, 130, 246, 0.1); border: 1px solid var(--primary-color); color: var(--text-primary);">
                    <i class="fas fa-info-circle me-2"></i>
                    此配置遵循您的配置文档标准，支持完整的内网穿透功能
                </div>
            </div>
        `;
    }
}

// 模块用户管理功能
async function showModuleUsers(moduleId) {
    try {
        const token = localStorage.getItem('access_token');
        
        // 获取模块信息
        const moduleResponse = await fetch(`/api/v1/modules/${moduleId}`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        const moduleResult = await moduleResponse.json();
        const module = moduleResult.data || moduleResult;
        
        // 获取用户列表
        const usersResponse = await fetch(`/api/v1/modules/${moduleId}/users`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        const usersResult = await usersResponse.json();
        const users = usersResult.data || [];
        
        // 创建模态框内容
        let content = `
            <div class="d-flex justify-content-between align-items-center mb-3">
                <h6 style="color: #f1f5f9;"><i class="fas fa-network-wired me-2"></i>模块: ${module.name}</h6>
                <button class="btn btn-primary btn-sm" onclick="showAddUserModal(${moduleId})">
                    <i class="fas fa-user-plus me-1"></i>添加用户
                </button>
            </div>
            
            <div class="table-responsive">
                <table style="width: 100%; margin: 0; background: transparent; color: #e2e8f0;">
                    <thead>
                        <tr style="background: rgba(15, 23, 42, 0.8); border-bottom: 2px solid rgba(100, 116, 139, 0.4);">
                            <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">用户名</th>
                            <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">状态</th>
                            <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">IP地址</th>
                            <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">最后在线</th>
                            <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">流量</th>
                            <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">操作</th>
                        </tr>
                    </thead>
                    <tbody style="background: transparent;">
                        ${users && users.length > 0 ? users.map(user => `
                            <tr style="background: rgba(30, 41, 59, 0.3); border-bottom: 1px solid rgba(100, 116, 139, 0.2); transition: background-color 0.2s ease;" 
                                onmouseover="this.style.background='rgba(30, 41, 59, 0.6)'" 
                                onmouseout="this.style.background='rgba(30, 41, 59, 0.3)'">
                                <td style="border: none; padding: 12px 16px;">
                                    <div>
                                        <div style="color: #f1f5f9; font-size: 14px; font-weight: 600;">${user.username}</div>
                                        <small style="color: #94a3b8;">${user.email || '无邮箱'}</small>
                                    </div>
                                </td>
                                <td style="border: none; padding: 12px 16px;">
                                    <span class="badge bg-${user.status === 1 ? 'success' : 'secondary'}" style="font-size: 11px; padding: 4px 8px;">${user.status === 1 ? '在线' : '离线'}</span>
                                    ${!user.is_active ? '<span class="badge bg-warning ms-1" style="font-size: 11px; padding: 4px 8px;">已停用</span>' : ''}
                                </td>
                                <td style="border: none; padding: 12px 16px;">
                                    <span style="background: rgba(15, 23, 42, 0.8); color: #34d399; padding: 4px 8px; border-radius: 6px; font-family: 'Courier New', monospace; font-size: 13px; font-weight: 500;">${user.ip_address}</span>
                                </td>
                                <td style="border: none; padding: 12px 16px; color: #e2e8f0; font-size: 13px;">${user.last_seen ? formatDateTime(user.last_seen) : '从未连接'}</td>
                                <td style="border: none; padding: 12px 16px; color: #e2e8f0; font-size: 13px;">${formatBytes((user.total_rx_bytes || 0) + (user.total_tx_bytes || 0))}</td>
                                <td style="border: none; padding: 12px 16px;">
                                    <div style="display: flex; gap: 4px;">
                                        <button onclick="downloadUserConfig(${user.id})" 
                                                style="background: transparent; border: 1px solid #3b82f6; color: #60a5fa; padding: 4px 8px; border-radius: 4px; font-size: 12px; cursor: pointer; transition: all 0.2s ease;"
                                                onmouseover="this.style.background='rgba(59, 130, 246, 0.1)'"
                                                onmouseout="this.style.background='transparent'">
                                            <i class="fas fa-download"></i>
                                        </button>
                                        <button onclick="toggleUserStatus(${user.id}, ${!user.is_active})" 
                                                style="background: transparent; border: 1px solid #f59e0b; color: #fbbf24; padding: 4px 8px; border-radius: 4px; font-size: 12px; cursor: pointer; transition: all 0.2s ease;"
                                                onmouseover="this.style.background='rgba(245, 158, 11, 0.1)'"
                                                onmouseout="this.style.background='transparent'">
                                            <i class="fas fa-${user.is_active ? 'pause' : 'play'}"></i>
                                        </button>
                                        <button onclick="deleteUser(${user.id})" 
                                                style="background: transparent; border: 1px solid #ef4444; color: #f87171; padding: 4px 8px; border-radius: 4px; font-size: 12px; cursor: pointer; transition: all 0.2s ease;"
                                                onmouseover="this.style.background='rgba(239, 68, 68, 0.1)'"
                                                onmouseout="this.style.background='transparent'">
                                            <i class="fas fa-trash"></i>
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        `).join('') : `
                            <tr style="background: rgba(30, 41, 59, 0.3); border-bottom: 1px solid rgba(100, 116, 139, 0.2);">
                                <td colspan="6" style="border: none; padding: 2rem; text-align: center; color: #94a3b8;">
                                    <i class="fas fa-users" style="margin-right: 8px; font-size: 16px;"></i>暂无用户
                                </td>
                            </tr>
                        `}
                    </tbody>
                </table>
            </div>
        `;
        
        // 显示模态框
        document.getElementById('userVPNContent').innerHTML = content;
        const modal = new bootstrap.Modal(document.getElementById('userVPNModal'));
        
        // 添加模态框关闭事件监听器，确保遮罩层正确移除
        const modalElement = document.getElementById('userVPNModal');
        modalElement.addEventListener('hidden.bs.modal', function () {
            // 强制移除所有模态框相关的类和遮罩层
            document.body.classList.remove('modal-open');
            document.body.style.removeProperty('padding-right');
            const backdrops = document.querySelectorAll('.modal-backdrop');
            backdrops.forEach(backdrop => backdrop.remove());
        }, { once: true });
        
        modal.show();
        
    } catch (error) {
        console.error('加载模块用户失败:', error);
        alert('加载用户列表失败');
    }
}

// 显示添加用户模态框
async function showAddUserModal(moduleId) {
    // 🔒 安全检查：先获取模块信息，检查接口状态
    try {
        const token = localStorage.getItem('access_token');
        const moduleResponse = await fetch(`/api/v1/modules/${moduleId}`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (!moduleResponse.ok) {
            alert('无法获取模块信息');
            return;
        }
        
        const moduleResult = await moduleResponse.json();
        const module = moduleResult.data || moduleResult;
        
        // 检查模块关联的接口状态
        if (!await checkInterfaceEditPermission(module.interface_id, '添加用户')) {
            return;
        }
    } catch (error) {
        console.error('检查模块状态失败:', error);
        alert('无法检查模块状态，建议先停止相关接口再进行操作');
        return;
    }
    
    const content = `
        <div class="row">
            <div class="col-md-12">
                <h6 class="mb-3" style="color: #f1f5f9;"><i class="fas fa-user-plus me-2"></i>为模块添加用户</h6>
                <form id="addUserForm">
                    <input type="hidden" name="module_id" value="${moduleId}">
                    
                    <div class="row">
                        <div class="col-md-6">
                            <div class="mb-3">
                                <label class="form-label" style="color: #e2e8f0;">用户名 *</label>
                                <input type="text" class="form-control" name="username" required 
                                       style="background: rgba(15, 23, 42, 0.6); border-color: rgba(100, 116, 139, 0.3); color: #f1f5f9;">
                            </div>
                        </div>
                        <div class="col-md-6">
                            <div class="mb-3">
                                <label class="form-label" style="color: #e2e8f0;">邮箱</label>
                                <input type="email" class="form-control" name="email"
                                       style="background: rgba(15, 23, 42, 0.6); border-color: rgba(100, 116, 139, 0.3); color: #f1f5f9;">
                            </div>
                        </div>
                    </div>
                    
                    <div class="mb-3">
                        <label class="form-label" style="color: #e2e8f0;">描述</label>
                        <textarea class="form-control" name="description" rows="2"
                                  style="background: rgba(15, 23, 42, 0.6); border-color: rgba(100, 116, 139, 0.3); color: #f1f5f9;"></textarea>
                    </div>
                    
                    <div class="row">
                        <div class="col-md-6">
                            <div class="mb-3">
                                <label class="form-label" style="color: #e2e8f0;">允许访问网段</label>
                                <select class="form-control" name="allowed_ips"
                                        style="background: rgba(15, 23, 42, 0.6); border-color: rgba(100, 116, 139, 0.3); color: #f1f5f9;">
                                    <option value="10.50.0.0/24,192.168.50.0/24">VPN网段+内网穿透（推荐）</option>
                                    <option value="10.50.0.0/24">仅VPN网段</option>
                                    <option value="0.0.0.0/0">全网访问</option>
                                    <option value="192.168.0.0/16">本地网络</option>
                                </select>
                                <div class="form-text" style="color: #94a3b8;">
                                    根据配置文档，推荐选择"VPN网段+内网穿透"以实现完整的内网访问功能
                                </div>
                            </div>
                        </div>
                        <div class="col-md-6">
                            <div class="mb-3">
                                <label class="form-label" style="color: #e2e8f0;">最大设备数</label>
                                <input type="number" class="form-control" name="max_devices" value="1" min="1" max="10"
                                       style="background: rgba(15, 23, 42, 0.6); border-color: rgba(100, 116, 139, 0.3); color: #f1f5f9;">
                            </div>
                        </div>
                    </div>
                    
                    <div class="mb-3">
                        <label class="form-label" style="color: #e2e8f0;">过期时间</label>
                        <input type="datetime-local" class="form-control" name="expires_at"
                               style="background: rgba(15, 23, 42, 0.6); border-color: rgba(100, 116, 139, 0.3); color: #f1f5f9;">
                        <div class="form-text" style="color: #94a3b8;">留空表示永不过期</div>
                    </div>
                </form>
                
                <div class="mt-3">
                    <button class="btn btn-primary" onclick="submitAddUser()">
                        <i class="fas fa-plus me-1"></i>创建用户
                    </button>
                    <button class="btn btn-secondary" onclick="showModuleUsers(${moduleId})">
                        <i class="fas fa-arrow-left me-1"></i>返回
                    </button>
                </div>
            </div>
        </div>
    `;
    
    document.getElementById('userVPNContent').innerHTML = content;
}

// 提交添加用户
async function submitAddUser() {
    try {
        const form = document.getElementById('addUserForm');
        const formData = new FormData(form);
        const data = Object.fromEntries(formData.entries());
        
        // 转换数据类型
        data.module_id = parseInt(data.module_id);
        data.max_devices = parseInt(data.max_devices);
        
        if (data.expires_at) {
            data.expires_at = new Date(data.expires_at).toISOString();
        } else {
            delete data.expires_at;
        }
        
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/user-vpn', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify(data)
        });
        
        const result = await response.json();
        if (response.ok) {
            alert('用户创建成功！');
            showModuleUsers(data.module_id); // 返回用户列表
        } else {
            alert('创建失败：' + result.message);
        }
    } catch (error) {
        console.error('创建用户失败:', error);
        alert('网络错误，请重试');
    }
}

// 下载用户配置
async function downloadUserConfig(userId) {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch(`/api/v1/user-vpn/${userId}/config`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.ok) {
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `user_${userId}_config.conf`;
            a.click();
            window.URL.revokeObjectURL(url);
        } else {
            alert('下载失败');
        }
    } catch (error) {
        console.error('下载用户配置失败:', error);
        alert('网络错误，请重试');
    }
}

// 切换用户状态
async function toggleUserStatus(userId, activate) {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch(`/api/v1/user-vpn/${userId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ is_active: activate })
        });
        
        const result = await response.json();
        if (response.ok) {
            alert(activate ? '用户已激活' : '用户已停用');
            // 刷新当前显示的用户列表
            const currentModal = document.querySelector('#userVPNModal .modal-body');
            if (currentModal) {
                // 重新加载当前模块的用户列表
                location.reload(); // 简单的刷新，也可以优化为只刷新列表
            }
        } else {
            alert('操作失败：' + result.message);
        }
    } catch (error) {
        console.error('切换用户状态失败:', error);
        alert('网络错误，请重试');
    }
}

// 删除用户
async function deleteUser(userId) {
    if (!confirm('确定要删除此用户吗？此操作不可撤销！')) {
        return;
    }
    
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch(`/api/v1/user-vpn/${userId}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        const result = await response.json();
        if (response.ok) {
            alert('用户删除成功！');
            location.reload(); // 刷新页面
        } else {
            alert('删除失败：' + result.message);
        }
    } catch (error) {
        console.error('删除用户失败:', error);
        alert('网络错误，请重试');
    }
} 

// 检查接口状态是否允许修改
async function checkInterfaceEditPermission(interfaceId, operation = '操作') {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch(`/api/v1/interfaces/${interfaceId}`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (!response.ok) {
            throw new Error('无法获取接口状态');
        }
        
        const result = await response.json();
        const interfaceInfo = result.data || result;
        
        // 检查接口状态
        if (interfaceInfo.status === 1 || interfaceInfo.status === 3) { // 运行中或启动中
            const statusText = interfaceInfo.status === 1 ? '运行中' : '启动中';
            alert(`⚠️ 无法执行${operation}\n\n接口 "${interfaceInfo.name}" 当前状态为：${statusText}\n\n为了安全操作，请先停止该接口后再进行${operation}。\n\n建议步骤：\n1. 在接口管理中停止接口\n2. 完成${operation}\n3. 重新启动接口以应用新配置`);
            return false;
        }
        
        return true;
    } catch (error) {
        console.error('检查接口状态失败:', error);
        alert('无法检查接口状态，建议先停止相关接口再进行操作');
        return false;
    }
}