// =====================================================
// EITEC VPN - 模块管理功能
// =====================================================
//
// 📋 功能概述：
// - 管理WireGuard模块的创建、配置、删除等操作
// - 处理模块与WireGuard接口的关联关系
// - 提供模块配置下载和状态管理功能
//
// 🔗 依赖关系：
// - 依赖：shared-utils.js (工具函数)
// - 依赖：bootstrap (模态框管理)
// - 与 interface-management.js 有业务关联
//
// 📦 主要功能：
// - showAddModuleModal() - 显示添加模块对话框
// - submitAddModule() - 提交模块创建请求
// - loadWireGuardInterfaces() - 加载可用接口列表
// - downloadModuleConfig() - 下载模块配置文件
// - deleteModule() - 删除指定模块
// - updateModulesTable() - 刷新模块网格显示
//
// 📏 文件大小：12.3KB (原文件的 11.8%)
// =====================================================

// 模块管理功能
async function showAddModuleModal() {
    const modalElement = document.getElementById('addModuleModal');
    if (!modalElement) {
        console.error('找不到 addModuleModal 元素');
        return;
    }
    
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
    
    // 使用新的模态框管理器
    ModalManager.show(modalElement);
}

// 加载WireGuard接口列表（使用带状态的接口API）
async function loadWireGuardInterfaces() {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/system/wireguard-interfaces', {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        if (response.ok) {
            const result = await response.json();
            console.log('Interfaces loaded (with status):', result);
            const interfaces = result.data || result.interfaces || [];
            const select = document.getElementById('moduleInterface');
            
            // 清空现有选项
            select.innerHTML = '<option value="">选择WireGuard接口</option>';
            
            // 添加接口选项（包含实时状态信息）
            interfaces.forEach(iface => {
                const option = document.createElement('option');
                option.value = iface.id;
                
                // 显示接口状态信息
                const statusText = iface.status === 1 ? '运行中' : '已停止';
                option.textContent = `${iface.name} - ${iface.description} (${iface.network}) [${statusText}]`;
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
            
            console.log(`加载了 ${interfaces.length} 个接口（包含状态信息）`);
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

// 提交添加模块
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
        persistent_keepalive: parseInt(formData.get('persistent_keepalive')) || 25
    };

    console.log('提交的模块数据:', data);

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
            // 触发主页面数据刷新
            if (typeof loadAllData === 'function') {
                loadAllData();
            }
        } else {
            alert('创建失败：' + result.message);
        }
    } catch (error) {
        console.error('创建模块失败:', error);
        alert('网络错误，请重试');
    }
}

// 下载模块配置
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
            
            // 从响应头获取后端设置的文件名
            let fileName = 'module_config.conf'; // 默认文件名
            const contentDisposition = response.headers.get('Content-Disposition');
            if (contentDisposition) {
                const match = contentDisposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/);
                if (match && match[1]) {
                    fileName = match[1].replace(/['"]/g, '');
                }
            }
            
            a.download = fileName;
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

// 编辑模块
function editModule(id) {
    alert('编辑模块功能开发中...');
}

// 删除模块
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
            alert('模块删除成功！');
            
            // 触发主页面数据刷新
            if (typeof loadAllData === 'function') {
                console.log('删除成功，开始刷新数据...');
                await loadAllData();
                
                // 确保模块表格得到更新
                setTimeout(() => {
                    console.log('延迟刷新确保数据同步...');
                    loadAllData();
                }, 1000);
            }
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

// 刷新模块显示 - 使用新的网格布局
function updateModulesTable(modules) {
    // 使用新的网格布局刷新
    if (typeof refreshInterfaceModuleGrid === 'function') {
        refreshInterfaceModuleGrid();
    } else {
        console.warn('refreshInterfaceModuleGrid 函数不可用');
    }
}

// 获取接口显示信息
function getInterfaceDisplay(module) {
    // 支持新的数据格式
    return module.interface || module.interface_name || '--';
}

// 全局导出模块管理函数
window.showAddModuleModal = showAddModuleModal;
window.loadWireGuardInterfaces = loadWireGuardInterfaces;
window.submitAddModule = submitAddModule;
window.downloadModuleConfig = downloadModuleConfig;
window.editModule = editModule;
window.deleteModule = deleteModule;
window.updateModulesTable = updateModulesTable; 