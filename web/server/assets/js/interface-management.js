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