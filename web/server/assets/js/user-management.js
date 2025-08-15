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
// - (已删除) showModuleUsers() - 用户信息已集成到卡片显示
// - showAddUserModal() - 添加用户对话框
// - submitAddUser() - 提交用户创建请求
// - downloadUserConfig() - 下载用户配置文件
// - toggleUserStatus() - 激活/停用用户
// - deleteUser() - 删除用户
//
// 📏 文件大小：18.0KB (原文件的 17.2%)
// =====================================================

// 注意：showModuleUsers 函数已删除
// 用户信息现在直接在接口-模块卡片中显示
// 如需管理用户，请使用卡片中的用户管理功能

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
        
        // 添加用户是安全操作，不需要停止接口
        // 只是获取接口信息用于验证模块有效性
        const interfaceResponse = await fetch(`/api/v1/interfaces/${module.interface_id}`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (!interfaceResponse.ok) {
            alert('无法获取接口信息，请检查模块配置');
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
                                    <option value="">自动生成（推荐）</option>
                                    <option value="10.10.0.0/24,192.168.50.0/24">VPN网段+内网穿透</option>
                                    <option value="10.10.0.0/24">仅VPN网段</option>
                                    <option value="0.0.0.0/0">全网访问</option>
                                </select>
                                <div class="form-text" style="color: #94a3b8;">
                                    推荐选择"自动生成"，系统将智能组合VPN网段和模块内网段
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
                </div>
            </div>
        </div>
    `;
    
    document.getElementById('userVPNContent').innerHTML = content;
    
    // 显示模态框
    const modalElement = document.getElementById('userVPNModal');
    if (!modalElement) {
        console.error('找不到 userVPNModal 元素');
        return;
    }
    
    // 使用模态框管理器显示
    ModalManager.show(modalElement);
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
            alert('用户创建成功！请查看接口卡片中的用户信息。');
            // 关闭模态框
            const modalElement = document.getElementById('userVPNModal');
            if (modalElement) {
                const modal = bootstrap.Modal.getInstance(modalElement);
                if (modal) {
                    modal.hide();
                }
            }
            // 刷新主页面数据以显示新用户
            if (typeof loadAllData === 'function') {
                loadAllData();
            }
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
            
            // 从响应头获取后端设置的文件名
            let fileName = 'user_config.conf'; // 默认文件名
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
window.showAddUserModal = showAddUserModal;
window.submitAddUser = submitAddUser;
window.downloadUserConfig = downloadUserConfig;
window.toggleUserStatus = toggleUserStatus;
window.deleteUser = deleteUser; 