


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
        apiHelper.showLoading('加载接口列表...');
        const result = await api.wireguard.getInterfaces();
        console.log('API返回结果:', result); // 调试信息
        
        let interfaces = result.data || result;
        console.log('处理后的interfaces:', interfaces); // 调试信息
        
        // 确保interfaces是数组
        if (!Array.isArray(interfaces)) {
            console.warn('接口数据格式异常:', interfaces);
            interfaces = [];
        }
        
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
        
        apiHelper.hideLoading();
    } catch (error) {
        console.error('获取接口建议配置失败:', error);
        apiHelper.hideLoading();
        apiHelper.handleError(error, '获取接口建议配置失败');
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
        apiHelper.showLoading('创建接口中...');
        console.log('准备发送接口创建请求:', interfaceData);
        
        const result = await api.wireguard.createInterface(interfaceData);
        
        console.log('响应内容:', result);
        
        // 关闭模态框
        const modal = bootstrap.Modal.getInstance(document.getElementById('createInterfaceModal'));
        modal.hide();
        
        // 清空表单
        form.reset();
        
        // 显示成功消息
        apiHelper.handleSuccess(`接口创建成功！\n\n配置信息：\n- 接口名称：${interfaceName}\n- 网络段：${networkCIDR}\n- 监听端口：${port}\n- 已按照标准配置生成防火墙规则`);
        
        // 刷新系统配置显示
        if (document.getElementById('wireGuardConfigModal').classList.contains('show')) {
            if (typeof refreshSystemConfig === 'function') {
                refreshSystemConfig();
            }
        }
        
        // 刷新主页面数据
        if (typeof loadAllData === 'function') {
            loadAllData();
        }
        
        apiHelper.hideLoading();
        
    } catch (error) {
        console.error('创建接口失败:', error);
        apiHelper.hideLoading();
        apiHelper.handleError(error, '创建接口失败');
    }
}

// 启动接口
async function startInterface(interfaceId) {
    const confirmed = await apiHelper.confirm('确定要启动此接口吗？\n\n启动后接口将开始监听端口并可以接受连接。', '启动接口');
    if (!confirmed) {
        return;
    }
    
    try {
        apiHelper.showLoading('启动接口中...');
        await api.wireguard.startInterface(interfaceId);
        
        apiHelper.handleSuccess('接口启动成功！');
        
        apiHelper.hideLoading();
    } catch (error) {
        console.error('启动接口失败:', error);
        apiHelper.hideLoading();
        apiHelper.handleError(error, '启动接口失败');
    }
}

// 停止接口
async function stopInterface(interfaceId) {
    const confirmed = await apiHelper.confirm('确定要停止此接口吗？\n\n停止后所有连接将断开，可以安全地修改配置。', '停止接口');
    if (!confirmed) {
        return;
    }
    
    try {
        apiHelper.showLoading('停止接口中...');
        await api.wireguard.stopInterface(interfaceId);
        
        apiHelper.handleSuccess('接口停止成功！现在可以安全地修改配置了。');
        
        apiHelper.hideLoading();
    } catch (error) {
        console.error('停止接口失败:', error);
        apiHelper.hideLoading();
        apiHelper.handleError(error, '停止接口失败');
    }
}

// 删除接口
async function deleteInterface(interfaceId) {
	const confirmed = await apiHelper.confirm('⚠️ 危险操作：确定要删除此接口吗？\n\n删除后：\n- 接口配置将永久丢失\n- 关联的模块和用户将被删除\n- 此操作不可撤销', '删除接口');
	if (!confirmed) {
		return;
	}
	
	// 二次确认
	const confirmText = prompt('请输入 "DELETE" 确认删除操作：');
	if (confirmText !== 'DELETE') {
		alert('操作已取消');
		return;
	}
	
	try {
		apiHelper.showLoading('删除接口中...');
		await api.wireguard.deleteInterface(interfaceId);
		
		apiHelper.handleSuccess('接口删除成功！');
		
		// 刷新主页面数据
		if (typeof loadAllData === 'function') {
			loadAllData();
		}
		
		apiHelper.hideLoading();
	} catch (error) {
		console.error('删除接口失败:', error);
		apiHelper.hideLoading();
		apiHelper.handleError(error, '删除接口失败');
	}
}

// 删除模块
async function deleteModule(moduleId, moduleName) {
	// 显示详细的确认对话框
	const confirmed = await apiHelper.confirm(
		`⚠️ 危险操作：确定要删除模块 "${moduleName}" 吗？\n\n` +
		`删除后：\n` +
		`- 模块配置将永久丢失\n` +
		`- 关联的用户VPN配置将被删除\n` +
		`- 模块密钥将无法恢复\n` +
		`- 此操作不可撤销\n\n` +
		`请确认您真的要删除这个模块吗？`, 
		'删除模块'
	);
	
	if (!confirmed) {
		return;
	}
	
	// 二次确认 - 要求用户输入模块名称
	const confirmText = prompt(`请输入模块名称 "${moduleName}" 来确认删除操作：`);
	if (confirmText !== moduleName) {
		alert('模块名称不匹配，操作已取消');
		return;
	}
	
	try {
		apiHelper.showLoading('删除模块中...');
		await api.modules.deleteModule(moduleId);
		
		apiHelper.handleSuccess(`模块 "${moduleName}" 删除成功！\n\n模块配置已从系统中移除，相关用户VPN配置已同步清理，密钥已销毁。`);
		
		// 刷新主页面数据
		if (typeof loadAllData === 'function') {
			loadAllData();
		} else if (typeof renderInterfaceModuleGrid === 'function') {
			renderInterfaceModuleGrid();
		}
		
		apiHelper.hideLoading();
	} catch (error) {
		console.error('删除模块失败:', error);
		apiHelper.hideLoading();
		apiHelper.handleError(error, '删除模块失败');
	}
}

// 删除用户VPN
async function deleteUserVPN(userId, userName) {
	// 显示详细的确认对话框
	const confirmed = await apiHelper.confirm(
		`⚠️ 危险操作：确定要删除用户 "${userName}" 吗？\n\n` +
		`删除后：\n` +
		`- 用户VPN配置将永久丢失\n` +
		`- 用户密钥将无法恢复\n` +
		`- 用户将无法连接VPN\n` +
		`- 此操作不可撤销\n\n` +
		`请确认您真的要删除这个用户吗？`, 
		'删除用户'
	);
	
	if (!confirmed) {
		return;
	}
	
	// 二次确认 - 要求用户输入用户名
	const confirmText = prompt(`请输入用户名 "${userName}" 来确认删除操作：`);
	if (confirmText !== userName) {
		alert('用户名不匹配，操作已取消');
		return;
	}
	
	try {
		apiHelper.showLoading('删除用户中...');
		await api.userVPN.deleteUserVPN(userId);
		
		apiHelper.handleSuccess(`用户 "${userName}" 删除成功！\n\n用户VPN配置已从系统中移除，相关密钥已清理。`);
		
		// 刷新主页面数据
		if (typeof loadAllData === 'function') {
			loadAllData();
		} else if (typeof renderInterfaceModuleGrid === 'function') {
			renderInterfaceModuleGrid();
		}
		
		apiHelper.hideLoading();
	} catch (error) {
		console.error('删除用户失败:', error);
		apiHelper.hideLoading();
		apiHelper.handleError(error, '删除用户失败');
	}
}



// 查看接口配置
async function showInterfaceConfig(interfaceId) {
    try {
        apiHelper.showLoading('获取接口配置...');
        const result = await api.wireguard.getInterfaceConfig(interfaceId);
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
        
        apiHelper.hideLoading();
    } catch (error) {
        console.error('获取接口配置失败:', error);
        apiHelper.hideLoading();
        apiHelper.handleError(error, '获取接口配置失败');
    }
}

// 下载接口配置文件
async function downloadInterfaceConfig(interfaceId) {
    try {
        apiHelper.showLoading('准备下载配置...');
        const result = await api.wireguard.getInterfaceConfig(interfaceId);
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
        
        apiHelper.hideLoading();
        apiHelper.handleSuccess('配置文件下载成功！');
    } catch (error) {
        console.error('下载配置失败:', error);
        apiHelper.hideLoading();
        apiHelper.handleError(error, '下载配置文件失败');
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
        apiHelper.showLoading('检测网络接口...');
        const data = await api.wireguard.getNetworkInterfaces();
        
        if (data.success && data.data && data.data.length > 0) {
            // 优先选择默认路由的接口
            const defaultInterface = data.data.find(iface => iface.is_default) || data.data[0];
            const input = document.getElementById('interfaceNetworkInterface');
            if (input) {
                input.value = defaultInterface.name;
                updateInterfaceConfigPreview();
                
                // 显示检测结果
                apiHelper.handleSuccess(`已检测到网络接口：${defaultInterface.name}`);
            }
        } else {
            alert('未检测到可用的网络接口，请手动输入');
        }
        
        apiHelper.hideLoading();
    } catch (error) {
        console.error('检测网络接口时出错:', error);
        apiHelper.hideLoading();
        apiHelper.handleError(error, '检测网络接口失败，请手动输入');
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

window.showCreateInterfaceModal = showCreateInterfaceModal;
window.suggestInterfaceConfig = suggestInterfaceConfig;
window.applyInterfaceTemplate = applyInterfaceTemplate;
window.updateInterfaceConfigPreview = updateInterfaceConfigPreview;
window.submitCreateInterface = submitCreateInterface;
window.startInterface = startInterface;
window.stopInterface = stopInterface;
window.deleteInterface = deleteInterface;
window.deleteModule = deleteModule;
window.deleteUserVPN = deleteUserVPN;
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
        apiHelper.showLoading('加载接口数据...');
        
        // 获取带状态的接口数据（包含模块信息）
        const result = await api.wireguard.getInterfaces();
        const interfaces = result.data || [];
        
        // 从接口数据中提取模块信息
        const modules = [];
        interfaces.forEach(iface => {
            if (iface.modules && Array.isArray(iface.modules)) {
                modules.push(...iface.modules);
            }
        });
        
        renderGridWithData(interfaces, modules);
        apiHelper.hideLoading();
        
    } catch (error) {
        console.error('渲染接口-模块网格失败:', error);
        apiHelper.hideLoading();
        
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
                                <button class="btn btn-xs btn-outline-danger" onclick="deleteModule(${module.id}, '${module.name}')" title="删除模块" style="padding: 2px 6px; font-size: 10px; border-radius: 3px;">
                                    <i class="fas fa-trash"></i>
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
                            <div class="module-detail-value">${formatLastHeartbeat(module.last_seen || module.latest_handshake)}</div>
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
                                ${module.users.map((user, index) => {
                                    // 全部从users获取实时wg show数据（状态和心跳时间）
                                    const isOnline = user.is_active; // 实时的wg show状态
                                    const lastHeartbeat = user.last_seen || user.latest_handshake; // 实时的wg show心跳时间
                                    
                                    return `
                                    <div class="user-item" style="display: flex; align-items: center; justify-content: space-between; padding: 10px 12px; background: rgba(15, 23, 42, 0.6); border-radius: 6px; font-size: 11px; border: 1px solid rgba(30, 41, 59, 0.6); box-sizing: border-box;">
                                        <div class="user-info" style="display: flex; flex-direction: column; flex: 1; min-width: 0; margin-right: 8px;">
                                            <div class="user-name" style="color: #f1f5f9; font-weight: 500; margin-bottom: 4px; font-size: 12px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">
                                                ${user.username || '未知用户'}
                                            </div>
                                            <div class="user-details" style="display: flex; flex-direction: column; gap: 2px; color: #94a3b8; font-size: 10px;">
                                                <span class="user-status">
                                                    状态: <span style="color: ${isOnline ? '#10b981' : '#6b7280'}; font-weight: 500;">${isOnline ? '在线' : '离线'}</span>
                                                </span>
                                                <span class="user-ip">
                                                    IP: <span style="color: #34d399; font-family: monospace; font-weight: 500;">${user.ip_address || '未分配'}</span>
                                                </span>
                                                <span class="user-heartbeat">
                                                    心跳: <span style="color: #fbbf24; font-weight: 500;">${formatLastHeartbeat(lastHeartbeat)}</span>
                                                </span>
                                            </div>
                                        </div>
                                        <div class="user-actions" style="display: flex; gap: 4px; flex-shrink: 0;">
                                            <button class="btn btn-xs btn-outline-info" onclick="downloadUserConfig(${user.id})" title="下载 ${user.username} 的配置" style="padding: 4px 8px; font-size: 9px; border-radius: 3px;">
                                                <i class="fas fa-download"></i>
                                            </button>
                                            <button class="btn btn-xs btn-outline-danger" onclick="deleteUserVPN(${user.id}, '${user.username}')" title="删除用户 ${user.username}" style="padding: 4px 8px; font-size: 9px; border-radius: 3px;">
                                                <i class="fas fa-trash"></i>
                                            </button>
                                        </div>
                                    </div>
                                    `;
                                }).join('')}
                            </div>
                        ` : `
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
    if (!timestamp || timestamp === null) return '从未';
    
    try {
        const date = new Date(timestamp);
        // 检查日期是否有效
        if (isNaN(date.getTime())) return '无效时间';
        
        const now = new Date();
        const diff = now - date;
        
        // 处理未来时间（可能的时区问题）
        if (diff < 0) {
            return '刚刚';
        }
        
        if (diff < 60000) return '刚刚'; // 1分钟内
        if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`; // 1小时内
        if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`; // 1天内
        if (diff < 604800000) return `${Math.floor(diff / 86400000)}天前`; // 1周内
        
        // 超过1周显示具体日期
        return date.toLocaleDateString('zh-CN', {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    } catch (error) {
        console.warn('格式化心跳时间出错:', error, 'timestamp:', timestamp);
        return '解析失败';
    }
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

// 下载用户配置文件
async function downloadUserConfig(userId) {
    try {
        apiHelper.showLoading('准备下载配置...');
        
        // 使用封装的API方法
        await api.userVPN.downloadUserConfig(userId);
        
        apiHelper.hideLoading();
        apiHelper.handleSuccess('用户配置文件下载成功！');
    } catch (error) {
        console.error('下载用户配置失败:', error);
        apiHelper.hideLoading();
        apiHelper.handleError(error, '下载用户配置文件失败');
    }
}

// 全局导出downloadUserConfig方法
window.downloadUserConfig = downloadUserConfig; 