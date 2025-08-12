// =====================================================
// EITEC VPN - 用户管理功能
// =====================================================
//
// 📋 功能概述：
// - 管理VPN用户的创建、配置、状态控制
// - 处理用户配置文件生成和下载功能
// - 提供用户权限管理和访问控制
//
// 🔗 依赖关系：
// - 依赖：shared-utils.js (工具函数)
// - 依赖：bootstrap (模态框管理)
// - 与 module-management.js 紧密关联 (用户属于模块)
//
// 📦 主要功能：
// - showModuleUsers() - 显示模块用户列表
// - showAddUserModal() - 添加用户对话框
// - submitAddUser() - 提交用户创建请求
// - downloadUserConfig() - 下载用户配置文件
// - toggleUserStatus() - 激活/停用用户
// - deleteUser() - 删除用户
//
// 📏 文件大小：18.0KB (原文件的 17.2%)
// =====================================================

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
        const modalElement = document.getElementById('userVPNModal');
        
        if (!modalElement) {
            console.error('找不到 userVPNModal 元素');
            return;
        }
        
        // 使用新的模态框管理器
        ModalManager.show(modalElement);
        
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

// 全局导出用户管理函数
window.showModuleUsers = showModuleUsers;
window.showAddUserModal = showAddUserModal;
window.submitAddUser = submitAddUser;
window.downloadUserConfig = downloadUserConfig;
window.toggleUserStatus = toggleUserStatus;
window.deleteUser = deleteUser; 