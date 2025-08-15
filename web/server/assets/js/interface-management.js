// =====================================================
// EITEC VPN - 接口管理功能
// =====================================================
//
// 📋 功能概述：
// - 管理WireGuard网络接口的完整生命周期
// - 处理接口创建、启动、停止、删除等核心操作
// - 提供接口配置预览和模板应用功能
//
// 🔗 依赖关系：
// - 依赖：shared-utils.js (工具函数)
// - 依赖：bootstrap (模态框管理)
// - 为 module-management.js 和 user-management.js 提供基础服务
//
// 📦 主要功能：
// - showInterfaceManager() - 接口管理界面
// - showCreateInterfaceModal() - 创建接口对话框
// - submitCreateInterface() - 提交接口创建
// - startInterface() / stopInterface() - 接口启停控制
// - deleteInterface() - 删除接口
// - showInterfaceConfig() - 查看接口配置
// - updateInterfaceConfigPreview() - 配置预览
//
// 📏 文件大小：30.0KB (原文件的 28.7%)
// =====================================================

// 显示接口管理器
async function showInterfaceManager() {
    const modalElement = document.getElementById('interfaceManagerModal');
    if (!modalElement) {
        console.error('找不到 interfaceManagerModal 元素');
        return;
    }
    
    // 使用新的模态框管理器
    ModalManager.show(modalElement);
    
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/system/wireguard-interfaces', {
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
                    let wgPrefix = '';
                    
                    // 使用实时WireGuard状态
                    if (iface.is_active !== undefined) {
                        if (iface.is_active) {
                            statusClass = 'success';
                            statusText = '[接口] 运行中';
                            statusIcon = 'fas fa-play-circle';
                        } else {
                            statusClass = 'secondary';
                            statusText = '[接口] 未运行';
                            statusIcon = 'fas fa-stop-circle';
                        }
                        
                        // 检查配置文件状态
                        if (!iface.config_exists) {
                            statusClass = 'warning';
                            statusText = '[接口] 配置缺失';
                            statusIcon = 'fas fa-exclamation-triangle';
                        }
                    } else {
                        // 降级到数据库状态 (兼容性)
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
                            <td style="border: none; padding: 12px 16px; color: #e2e8f0; font-size: 13px;">
                                ${iface.peer_count !== undefined ? `[实时] ${iface.active_peers || 0}/${iface.peer_count || 0}` : `${iface.total_peers || 0}/${iface.max_peers || 0}`}
                            </td>
                            <td style="border: none; padding: 12px 16px;">
                                <div style="display: flex; gap: 0.25rem;">
                    `;
                    
                    // 根据实时状态显示不同的操作按钮
                    const isRunning = iface.is_active !== undefined ? iface.is_active : (iface.status === 1);
                    if (isRunning) { // 运行中
                        content += `
                            <button class="btn btn-sm btn-outline-warning" onclick="stopInterface(${iface.id})" title="停止接口">
                                <i class="fas fa-stop"></i>
                            </button>
                        `;
                    } else { // 已停止
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

// 显示创建接口模态框
function showCreateInterfaceModal() {
    const modalElement = document.getElementById('createInterfaceModal');
    if (!modalElement) {
        console.error('找不到 createInterfaceModal 元素');
        return;
    }
    
    // 设置配置模板选择器事件
    const templateSelect = document.getElementById('configTemplate');
    if (templateSelect) {
        templateSelect.addEventListener('change', function() {
            applyInterfaceTemplate(this.value);
        });
    }
    
    // 使用新的模态框管理器
    ModalManager.show(modalElement);
    
    // 预填充建议的配置
    suggestInterfaceConfig();
}

// 建议接口配置
async function suggestInterfaceConfig() {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/system/wireguard-interfaces', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.ok) {
            const result = await response.json();
            const interfaces = result.data || result;
            
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
    const networkInterface = document.getElementById('interfaceNetworkInterface')?.value || 'eth0';
    
    const previewDiv = document.getElementById('interfaceConfigPreview');
    if (previewDiv) {
        const configExample = `[Interface]
Address = ${network.replace(/0\/24$/, '1/24')}
ListenPort = ${port}
MTU = 1420
SaveConfig = true
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -A FORWARD -o %i -j ACCEPT; iptables -t nat -A POSTROUTING -o ${networkInterface} -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -D FORWARD -o %i -j ACCEPT; iptables -t nat -D POSTROUTING -o ${networkInterface} -j MASQUERADE
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
    const networkInterface = formData.get('network_interface') || 'eth0';
    
    // 构建请求数据 - 参考配置文档的服务端配置
    const interfaceData = {
        name: interfaceName,
        description: formData.get('description') || '',
        network: networkCIDR,
        listen_port: port,
        dns: formData.get('dns') || '8.8.8.8,8.8.4.4',
        max_peers: parseInt(formData.get('max_peers')) || 50,
        mtu: parseInt(formData.get('mtu')) || 1420,
        network_interface: networkInterface.trim(),
        // 使用验证成功的规则格式：简洁且使用%i占位符，动态网络接口
        post_up: formData.get('post_up') || `iptables -A FORWARD -i %i -j ACCEPT; iptables -A FORWARD -o %i -j ACCEPT; iptables -t nat -A POSTROUTING -o ${networkInterface} -j MASQUERADE`,
        // 对应的清理规则
        post_down: `iptables -D FORWARD -i %i -j ACCEPT; iptables -D FORWARD -o %i -j ACCEPT; iptables -t nat -D POSTROUTING -o ${networkInterface} -j MASQUERADE`,
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
                if (typeof refreshSystemConfig === 'function') {
                    refreshSystemConfig();
                }
            }
            
            // 刷新接口管理界面
            if (document.getElementById('interfaceManagerModal').classList.contains('show')) {
                showInterfaceManager();
            }
            
            // 刷新主页面数据
            if (typeof loadAllData === 'function') {
                loadAllData();
            }
            
        } else {
            alert('创建接口失败：' + (result.message || '未知错误'));
        }
        
    } catch (error) {
        console.error('创建接口失败:', error);
        alert('创建接口失败：网络错误');
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
            if (typeof loadAllData === 'function') {
                loadAllData();
            }
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
            const modalElement = document.getElementById('userVPNModal');
            if (modalElement) {
                ModalManager.show(modalElement);
            }
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

// 处理网络接口预设按钮点击
function handleInterfacePresetClick(event) {
    if (event.target.classList.contains('interface-preset')) {
        const interfaceName = event.target.getAttribute('data-interface');
        const input = document.getElementById('interfaceNetworkInterface');
        if (input) {
            input.value = interfaceName;
            // 更新预览
            updateInterfaceConfigPreview();
        }
    }
}

// 自动检测网络接口
async function detectNetworkInterface() {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/system/network-interfaces', {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
        
        if (response.ok) {
            const data = await response.json();
            if (data.success && data.data && data.data.length > 0) {
                // 优先选择默认路由的接口
                const defaultInterface = data.data.find(iface => iface.is_default) || data.data[0];
                const input = document.getElementById('interfaceNetworkInterface');
                if (input) {
                    input.value = defaultInterface.name;
                    updateInterfaceConfigPreview();
                    
                    // 显示检测结果
                    alert(`已检测到网络接口：${defaultInterface.name}`);
                }
            } else {
                alert('未检测到可用的网络接口，请手动输入');
            }
        } else {
            alert('检测网络接口失败，请手动输入');
        }
    } catch (error) {
        console.error('检测网络接口时出错:', error);
        alert('检测网络接口失败，请手动输入');
    }
}

// 初始化网络接口相关事件监听器
function initNetworkInterfaceHandlers() {
    // 预设按钮点击事件
    document.addEventListener('click', handleInterfacePresetClick);
    
    // 自动检测按钮点击事件
    const detectButton = document.getElementById('detectNetworkInterface');
    if (detectButton) {
        detectButton.addEventListener('click', detectNetworkInterface);
    }
    
    // 网络接口输入变化时更新预览
    const interfaceInput = document.getElementById('interfaceNetworkInterface');
    if (interfaceInput) {
        interfaceInput.addEventListener('input', updateInterfaceConfigPreview);
    }
}

// 页面加载时初始化
document.addEventListener('DOMContentLoaded', function() {
    // 延迟初始化，确保DOM完全加载
    setTimeout(initNetworkInterfaceHandlers, 100);
});

// 全局导出接口管理函数
window.showInterfaceManager = showInterfaceManager;
window.showCreateInterfaceModal = showCreateInterfaceModal;
window.suggestInterfaceConfig = suggestInterfaceConfig;
window.applyInterfaceTemplate = applyInterfaceTemplate;
window.updateInterfaceConfigPreview = updateInterfaceConfigPreview;
window.submitCreateInterface = submitCreateInterface;
window.startInterface = startInterface;
window.stopInterface = stopInterface;
window.deleteInterface = deleteInterface;
window.showInterfaceConfig = showInterfaceConfig;
window.downloadInterfaceConfig = downloadInterfaceConfig; 

// =====================================================
// 接口-模块卡片网格渲染
// =====================================================

// 渲染接口-模块卡片网格
async function renderInterfaceModuleGrid() {
    const gridContainer = document.getElementById('interfaceModuleGrid');
    if (!gridContainer) {
        console.error('找不到 interfaceModuleGrid 元素');
        return;
    }
    
    try {
        const token = localStorage.getItem('access_token');
        
        // 获取带状态的接口数据（包含模块信息）
        const interfacesResponse = await fetch('/api/v1/system/wireguard-interfaces', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (!interfacesResponse.ok) {
            throw new Error('获取接口数据失败');
        }
        
        const interfaces = (await interfacesResponse.json()).data || [];
        // 从接口数据中提取模块信息
        const modules = [];
        interfaces.forEach(iface => {
            if (iface.modules && Array.isArray(iface.modules)) {
                modules.push(...iface.modules);
            }
        });
        
        renderGridWithData(interfaces, modules);
        
    } catch (error) {
        console.error('渲染接口-模块网格失败:', error);
        
        // 如果API调用失败，尝试使用演示数据
        if (typeof generateDemoData === 'function') {
            console.log('使用演示数据展示布局效果');
            const demoData = generateDemoData();
            renderGridWithData(demoData.interfaces, demoData.modules);
        } else {
            gridContainer.innerHTML = `
                <div class="empty-interface">
                    <i class="fas fa-exclamation-triangle"></i>
                    <h4>加载失败</h4>
                    <p>无法加载接口和模块信息，请检查网络连接</p>
                    <button class="card-action-btn primary" onclick="renderInterfaceModuleGrid()">
                        <i class="fas fa-sync-alt"></i> 重试
                    </button>
                </div>
            `;
        }
    }
}

// 使用数据渲染网格
function renderGridWithData(interfaces, modules) {
    const gridContainer = document.getElementById('interfaceModuleGrid');
    
    // 创建接口到模块的映射
    const interfaceModuleMap = new Map();
    interfaces.forEach(iface => {
        interfaceModuleMap.set(iface.id, {
            interface: iface,
            modules: modules.filter(module => module.interface_id === iface.id)
        });
    });
    
    // 渲染卡片网格
    if (interfaceModuleMap.size === 0) {
        gridContainer.innerHTML = `
            <div class="empty-interface">
                <i class="fas fa-network-wired"></i>
                <h4>暂无WireGuard接口</h4>
                <p>创建第一个接口来开始管理您的VPN网络</p>
                <button class="card-action-btn primary" onclick="showCreateInterfaceModal()">
                    <i class="fas fa-plus"></i> 创建接口
                </button>
            </div>
        `;
        return;
    }
    
    let gridHTML = '';
    interfaceModuleMap.forEach((data, interfaceId) => {
        const { interface: iface, modules: ifaceModules } = data;
        gridHTML += renderInterfaceModuleCard(iface, ifaceModules);
    });
    
    gridContainer.innerHTML = gridHTML;
    
    // 绑定事件监听器
    bindCardEventListeners();
}

// 渲染单个接口-模块卡片
function renderInterfaceModuleCard(iface, modules) {
    // 使用实时WireGuard状态（后端已加WG前缀）
    let statusClass, statusText, statusIcon;
    
    if (iface.is_active !== undefined) {
        if (iface.is_active) {
            statusClass = 'running';
            statusText = '[接口] 运行中';
            statusIcon = 'fas fa-play-circle';
        } else {
            statusClass = 'stopped';
            statusText = '[接口] 未运行';
            statusIcon = 'fas fa-stop-circle';
        }
        
        if (!iface.config_exists) {
            statusClass = 'error';
            statusText = '[接口] 配置缺失';
            statusIcon = 'fas fa-exclamation-triangle';
        }
    } else {
        // 降级到原来的逻辑
        statusClass = getInterfaceStatusClass(iface.status);
        statusText = getInterfaceStatusText(iface.status);
        statusIcon = getInterfaceStatusIcon(iface.status);
    }
    
    let modulesHTML = '';
    if (modules.length === 0) {
        modulesHTML = `
            <div class="module-section">
                <div class="module-header">
                    <div class="module-title">
                        <i class="fas fa-server"></i>
                        未分配模块
                    </div>
                    <div class="module-status unknown">无模块</div>
                </div>
                <div class="module-details">
                    <div class="module-detail-item">
                        <div class="module-detail-label">状态</div>
                        <div class="module-detail-value">等待分配</div>
                    </div>
                    <div class="module-detail-item">
                        <div class="module-detail-label">用户数</div>
                        <div class="module-detail-value">0</div>
                    </div>
                </div>
                <div class="user-stats">
                    <div class="user-count">
                        <i class="fas fa-users"></i>
                        <span class="user-count-text">0 个用户</span>
                    </div>
                    <div class="user-actions">
                        <button class="card-action-btn primary compact" onclick="showAddModuleModal()" title="添加模块">
                            <i class="fas fa-plus"></i>
                        </button>
                    </div>
                </div>
            </div>
        `;
    } else {
        modules.forEach(module => {
            // 使用实时WireGuard模块状态（后端已加WG前缀）
            let moduleStatusClass, moduleStatusText, moduleStatusIcon;
            
            if (module.is_online !== undefined) {
                if (module.is_online) {
                    moduleStatusClass = 'online';
                    moduleStatusText = '[模块] 在线';
                    moduleStatusIcon = 'fas fa-circle';
                } else {
                    moduleStatusClass = 'offline';
                    moduleStatusText = '[模块] 离线';
                    moduleStatusIcon = 'fas fa-circle';
                }
            } else {
                // 降级到数据库状态
                moduleStatusClass = getModuleStatusClass(module.status);
                moduleStatusText = getModuleStatusText(module.status);
                moduleStatusIcon = getModuleStatusIcon(module.status);
            }
            
            modulesHTML += `
                <div class="module-section">
                    <div class="module-header">
                        <div class="module-title">
                            <i class="fas fa-server"></i>
                            ${module.name}
                        </div>
                        <div class="module-status-actions" style="display: flex; align-items: center; justify-content: space-between;">
                            <div class="module-status ${moduleStatusClass}">
                                <i class="${moduleStatusIcon}"></i>
                                ${moduleStatusText}
                            </div>
                            <div class="module-quick-actions" style="display: flex; gap: 4px;">
                                <button class="btn btn-xs btn-outline-primary" onclick="showAddUserModal(${module.id})" title="添加用户" style="padding: 2px 6px; font-size: 10px; border-radius: 3px;">
                                    <i class="fas fa-user-plus"></i>
                                </button>
                                <button class="btn btn-xs btn-outline-info" onclick="downloadModuleConfig(${module.id})" title="下载模块配置" style="padding: 2px 6px; font-size: 10px; border-radius: 3px;">
                                    <i class="fas fa-download"></i>
                                </button>
                            </div>
                        </div>
                    </div>
                    <div class="module-details">
                        <div class="module-detail-item">
                            <div class="module-detail-label">位置</div>
                            <div class="module-detail-value">${module.location || '未知'}</div>
                        </div>
                        <div class="module-detail-item">
                            <div class="module-detail-label">WireGuard IP</div>
                            <div class="module-detail-value">${module.ip_address || '未分配'}</div>
                        </div>
                        <div class="module-detail-item">
                            <div class="module-detail-label">内网IP</div>
                            <div class="module-detail-value">${module.local_ip || '未配置'}</div>
                        </div>

                        <div class="module-detail-item">
                            <div class="module-detail-label">最后心跳</div>
                            <div class="module-detail-value">${formatLastHeartbeat(module.last_heartbeat)}</div>
                        </div>
                        <div class="module-detail-item">
                            <div class="module-detail-label">流量</div>
                            <div class="module-detail-value">${formatTraffic(module.total_rx_bytes, module.total_tx_bytes)}</div>
                        </div>
                    </div>
                    
                    <!-- 模块用户标题移到外面 -->
                    <div class="user-section-title" style="display: flex; align-items: center; margin-top: 12px; margin-bottom: 8px; color: #f1f5f9; font-weight: 600; font-size: 12px;">
                        <i class="fas fa-users" style="margin-right: 8px; color: #60a5fa; font-size: 14px;"></i>
                        <span>模块用户: ${module.users ? module.users.length : 0}个</span>
                        ${module.users && module.users.length > 0 ? `
                            <span style="margin-left: 12px; color: #94a3b8; font-size: 11px;">
                                (${module.users.filter(u => u.is_active).length}个在线)
                            </span>
                        ` : ''}
                    </div>
                    
                    <div class="user-stats">
                        ${module.users && module.users.length > 0 ? `
                            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 10px; width: 100%;">
                                ${module.users.map((user, index) => `
                                    <div class="user-item" style="display: flex; align-items: center; justify-content: space-between; padding: 10px 12px; background: rgba(15, 23, 42, 0.6); border-radius: 6px; font-size: 11px; border: 1px solid rgba(30, 41, 59, 0.6); box-sizing: border-box;">
                                        <div class="user-info" style="display: flex; flex-direction: column; flex: 1; min-width: 0; margin-right: 8px;">
                                            <div class="user-name" style="color: #f1f5f9; font-weight: 500; margin-bottom: 4px; font-size: 12px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">
                                                ${user.username || '未知用户'}
                                            </div>
                                            <div class="user-details" style="display: flex; flex-direction: column; gap: 2px; color: #94a3b8; font-size: 10px;">
                                                <span class="user-status">
                                                    状态: <span style="color: ${user.is_active ? '#10b981' : '#6b7280'}; font-weight: 500;">${user.is_active ? '在线' : '离线'}</span>
                                                </span>
                                                <span class="user-ip">
                                                    IP: <span style="color: #34d399; font-family: monospace; font-weight: 500;">${user.ip_address || '未分配'}</span>
                                                </span>
                                            </div>
                                        </div>
                                        <div class="user-actions" style="display: flex; gap: 4px; flex-shrink: 0;">
                                            <button class="btn btn-xs btn-outline-info" onclick="downloadUserConfig(${user.id})" title="下载 ${user.username} 的配置" style="padding: 4px 8px; font-size: 9px; border-radius: 3px;">
                                                <i class="fas fa-download"></i>
                                            </button>
                                        </div>
                                    </div>
                                `).join('')}
                            </div>
                        ` : `
                            <div class="user-section-title" style="display: flex; align-items: center; margin-bottom: 8px; color: #f1f5f9; font-weight: 600; font-size: 12px;">
                                <i class="fas fa-users" style="margin-right: 6px; color: #60a5fa;"></i>
                                模块用户 (0个)
                            </div>
                            <div class="no-users" style="background: rgba(15, 23, 42, 0.6); border-radius: 8px; padding: 16px; text-align: center; width: 100%; border: 1px solid rgba(30, 41, 59, 0.6);">
                                <div style="color: #94a3b8; font-size: 11px; margin-bottom: 10px;">
                                    <i class="fas fa-user-plus" style="margin-right: 6px;"></i>
                                    该模块暂无用户
                                </div>
                                <button class="btn btn-xs btn-outline-primary" onclick="showAddUserModal(${module.id})" style="padding: 6px 12px; font-size: 10px;">
                                    <i class="fas fa-plus me-1"></i>添加用户
                                </button>
                            </div>
                        `}
                    </div>
                </div>
            `;
        });
    }
    
    return `
        <div class="interface-module-card" data-interface-id="${iface.id}">
            <!-- 接口头部 -->
            <div class="interface-header">
                <div class="interface-title">
                    <i class="fas fa-ethernet"></i>
                    ${iface.name}
                </div>
                <div class="interface-status ${statusClass}">
                    <i class="${statusIcon}"></i>
                    ${statusText}
                </div>
            </div>
            
            <!-- 接口详情 -->
            <div class="interface-details">
                <div class="interface-detail-item">
                    <div class="interface-detail-label">网络段</div>
                    <div class="interface-detail-value">${iface.network || '未配置'}</div>
                </div>
                <div class="interface-detail-item">
                    <div class="interface-detail-label">监听端口</div>
                    <div class="interface-detail-value">${iface.listen_port || '未配置'}</div>
                </div>
                <div class="interface-detail-item">
                    <div class="interface-detail-label">连接数</div>
                    <div class="interface-detail-value">
                        ${iface.peer_count !== undefined ? 
                            `[实时] ${iface.active_peers || 0}/${iface.peer_count || 0}` : 
                            `${iface.total_peers || 0}/${iface.max_peers || 0}`}
                    </div>
                </div>
                <div class="interface-detail-item">
                    <div class="interface-detail-label">描述</div>
                    <div class="interface-detail-value">${iface.description || '无描述'}</div>
                </div>
            </div>
            
            <!-- 模块信息 -->
            ${modulesHTML}
            
            <!-- 操作按钮 -->
            <div class="card-actions">
                <button class="card-action-btn" onclick="showInterfaceConfig(${iface.id})">
                    <i class="fas fa-cog"></i> 配置
                </button>
                <button class="card-action-btn" onclick="showInterfaceManager()">
                    <i class="fas fa-tools"></i> 管理
                </button>
                ${(iface.is_active !== undefined ? iface.is_active : (iface.status === 1)) ? 
                    `<button class="card-action-btn danger" onclick="stopInterface(${iface.id})">
                        <i class="fas fa-stop"></i> 停止
                    </button>` :
                    `<button class="card-action-btn primary" onclick="startInterface(${iface.id})">
                        <i class="fas fa-play"></i> 启动
                    </button>`
                }
                <button class="card-action-btn danger" onclick="deleteInterface(${iface.id})">
                    <i class="fas fa-trash"></i> 删除
                </button>
            </div>
        </div>
    `;
}

// 获取接口状态样式类
function getInterfaceStatusClass(status) {
    switch (status) {
        case 1: return 'running';
        case 2: return 'running'; // 2 也表示运行中
        case 0: return 'stopped';
        default: return 'error';
    }
}

// 获取接口状态文本
function getInterfaceStatusText(status) {
    switch (status) {
        case 1: return '运行中';
        case 2: return '运行中'; // 2 也表示运行中
        case 0: return '已停止';
        default: return '错误';
    }
}

// 获取接口状态图标
function getInterfaceStatusIcon(status) {
    switch (status) {
        case 1: return 'fas fa-play-circle';
        case 2: return 'fas fa-play-circle'; // 2 也表示运行中
        case 0: return 'fas fa-stop-circle';
        default: return 'fas fa-exclamation-triangle';
    }
}

// 获取模块状态样式类
function getModuleStatusClass(status) {
    switch (status) {
        case 1: return 'online';
        case 0: return 'offline';
        default: return 'unknown';
    }
}

// 获取模块状态文本
function getModuleStatusText(status) {
    switch (status) {
        case 1: return '在线';
        case 0: return '离线';
        default: return '未知';
    }
}

// 获取模块状态图标
function getModuleStatusIcon(status) {
    switch (status) {
        case 1: return 'fas fa-circle';
        case 0: return 'fas fa-circle';
        default: return 'fas fa-question-circle';
    }
}

// 格式化最后心跳时间
function formatLastHeartbeat(timestamp) {
    if (!timestamp) return '从未';
    
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now - date;
    
    if (diff < 60000) return '刚刚';
    if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`;
    return `${Math.floor(diff / 86400000)}天前`;
}

// 格式化流量
function formatTraffic(inBytes, outBytes) {
    const inMB = (inBytes || 0) / (1024 * 1024);
    const outMB = (outBytes || 0) / (1024 * 1024);
    return `${inMB.toFixed(1)}MB / ${outMB.toFixed(1)}MB`;
}

// 格式化模块用户信息
function formatModuleUsers(module) {
    if (!module.users || module.users.length === 0) {
        return '无用户';
    }
    
    const userInfo = module.users.slice(0, 2).map(user => {
        const statusIcon = user.is_active ? '🟢' : '🔘';
        const name = user.username || user.email || '未知用户';
        const ip = user.ip_address ? ` (${user.ip_address})` : '';
        return `${statusIcon} ${name}${ip}`;
    });
    
    const displayInfo = userInfo.join(', ');
    const moreCount = module.users.length > 2 ? ` +${module.users.length - 2}更多` : '';
    
    return `${displayInfo}${moreCount}`;
}

// 绑定卡片事件监听器
function bindCardEventListeners() {
    // 这里可以添加卡片相关的交互事件
    console.log('接口-模块卡片事件监听器已绑定');
}

// 刷新接口-模块网格
function refreshInterfaceModuleGrid() {
    renderInterfaceModuleGrid();
}

// 绑定卡片事件监听器
function bindCardEventListeners() {
    // 这里可以添加卡片相关的交互事件
    console.log('接口-模块卡片事件监听器已绑定');
}

// 全局导出接口-模块网格函数
window.renderInterfaceModuleGrid = renderInterfaceModuleGrid;
window.refreshInterfaceModuleGrid = refreshInterfaceModuleGrid; 