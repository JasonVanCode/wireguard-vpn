// =====================================================
// EITEC VPN - 共享工具函数
// =====================================================
//
// 📋 功能概述：
// - 提供所有模块共用的工具函数和通用逻辑
// - 包含数据格式化、网络验证、API请求等基础功能
// - 负责模态框管理和状态检查等通用操作
//
// 🔗 依赖关系：
// - 无外部依赖，必须最先加载
// - 被所有其他JavaScript模块依赖
//
// 📦 导出的全局函数：
// - formatBytes() - 字节数格式化
// - formatDateTime() - 日期时间格式化
// - validateNetworkFormat() - 网段格式验证
// - checkInterfaceEditPermission() - 接口编辑权限检查
// - apiRequest() - 通用API请求封装
// - safeCloseModal() - 安全关闭模态框
//
// 📏 文件大小：5.6KB (原文件的 5.4%)
// =====================================================

// 格式化字节数
function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// 格式化日期时间
function formatDateTime(dateStr) {
    if (!dateStr) return '--';
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN');
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

// 获取状态对应的CSS类
function getStatusClass(status) {
    switch(status) {
        case '在线': return 'online';
        case '离线': return 'offline';
        case '警告': return 'error';
        case '未配置': return 'offline';
        default: return 'offline';
    }
}

// 获取状态文本
function getStatusText(status) {
    return status || '未知';
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

// 通用API请求函数
async function apiRequest(url, options = {}) {
    const token = localStorage.getItem('access_token');
    
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json',
            ...(token && { 'Authorization': `Bearer ${token}` })
        }
    };
    
    const mergedOptions = {
        ...defaultOptions,
        ...options,
        headers: {
            ...defaultOptions.headers,
            ...options.headers
        }
    };
    
    try {
        const response = await fetch(url, mergedOptions);
        
        if (response.status === 401) {
            // Token 过期，跳转到登录页
            localStorage.removeItem('access_token');
            localStorage.removeItem('refresh_token');
            window.location.href = '/login';
            return null;
        }
        
        return response;
    } catch (error) {
        console.error('API请求失败:', error);
        throw error;
    }
}

// 全局模态框管理器
const ModalManager = {
    activeModals: new Set(),
    
    // 安全地显示模态框
    show(modalElement) {
        if (!modalElement) return null;
        
        const modalId = modalElement.id;
        console.log(`[ModalManager] 显示模态框: ${modalId}`);
        
        // 清理任何残留状态
        this.forceCleanup();
        
        // 创建 Bootstrap 模态框实例
        const modal = new bootstrap.Modal(modalElement, {
            backdrop: 'static',
            keyboard: true
        });
        
        // 记录活跃模态框
        this.activeModals.add(modalId);
        
        // 监听关闭事件 - 只监听一次
        const handleHidden = () => {
            console.log(`[ModalManager] 模态框已关闭: ${modalId}`);
            this.cleanup(modalId);
            modalElement.removeEventListener('hidden.bs.modal', handleHidden);
        };
        
        modalElement.addEventListener('hidden.bs.modal', handleHidden);
        
        // 显示模态框
        modal.show();
        
        return modal;
    },
    
    // 清理指定模态框
    cleanup(modalId) {
        console.log(`[ModalManager] 清理模态框: ${modalId}`);
        
        // 从活跃列表中移除
        this.activeModals.delete(modalId);
        
        // 延迟清理，确保 Bootstrap 动画完成
        setTimeout(() => {
            this.forceCleanup();
        }, 200);
    },
    
    // 强制清理所有模态框状态
    forceCleanup() {
        console.log(`[ModalManager] 强制清理模态框状态`);
        
        // 移除所有遮罩层
        const backdrops = document.querySelectorAll('.modal-backdrop');
        backdrops.forEach(backdrop => {
            console.log(`[ModalManager] 移除遮罩层:`, backdrop);
            backdrop.remove();
        });
        
        // 恢复 body 状态
        document.body.classList.remove('modal-open');
        document.body.style.removeProperty('padding-right');
        document.body.style.removeProperty('overflow');
        
        // 确保没有遗留的模态框显示状态
        const modals = document.querySelectorAll('.modal.show');
        modals.forEach(modal => {
            if (!this.activeModals.has(modal.id)) {
                console.log(`[ModalManager] 移除遗留模态框:`, modal.id);
                modal.classList.remove('show');
                modal.style.display = 'none';
            }
        });
    },
    
    // 关闭所有模态框
    closeAll() {
        console.log(`[ModalManager] 关闭所有模态框`);
        
        this.activeModals.forEach(modalId => {
            const modalElement = document.getElementById(modalId);
            if (modalElement) {
                const modal = bootstrap.Modal.getInstance(modalElement);
                if (modal) {
                    modal.hide();
                }
            }
        });
        
        // 清空活跃列表
        this.activeModals.clear();
        
        // 强制清理
        setTimeout(() => this.forceCleanup(), 300);
    }
};

// 安全地关闭模态框（向后兼容）
function safeCloseModal(modalElement) {
    // 这个函数现在不做任何事情，因为 ModalManager.show() 已经处理了清理
    console.log(`[ModalManager] safeCloseModal 被调用，但不再需要 - 使用 ModalManager.show() 代替`);
}

// 全局导出工具函数
window.formatBytes = formatBytes;
window.formatDateTime = formatDateTime;
window.validateNetworkFormat = validateNetworkFormat;
window.getStatusClass = getStatusClass;
window.getStatusText = getStatusText;
window.checkInterfaceEditPermission = checkInterfaceEditPermission;
window.apiRequest = apiRequest;
window.safeCloseModal = safeCloseModal;
window.ModalManager = ModalManager;

// 添加全局错误处理
window.addEventListener('error', function(e) {
    console.error('[Global Error]', e.error);
    // 如果发生错误，强制清理模态框状态
    ModalManager.forceCleanup();
});

// 页面卸载时清理所有模态框
window.addEventListener('beforeunload', function() {
    ModalManager.closeAll();
});

// 应急恢复功能 - 按 Escape 键三次快速清理页面状态
let escapeKeyCount = 0;
let escapeTimer = null;

window.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') {
        escapeKeyCount++;
        
        // 重置计时器
        if (escapeTimer) {
            clearTimeout(escapeTimer);
        }
        
        // 3秒内按下3次ESC键
        escapeTimer = setTimeout(() => {
            escapeKeyCount = 0;
        }, 3000);
        
        if (escapeKeyCount >= 3) {
            console.log('[应急恢复] 检测到连续3次ESC键，强制清理页面状态');
            ModalManager.forceCleanup();
            escapeKeyCount = 0;
            
            // 显示恢复提示
            const notification = document.createElement('div');
            notification.style.cssText = `
                position: fixed;
                top: 20px;
                right: 20px;
                background: #10b981;
                color: white;
                padding: 15px 20px;
                border-radius: 8px;
                z-index: 9999;
                font-size: 14px;
                box-shadow: 0 4px 12px rgba(0,0,0,0.3);
            `;
            notification.innerHTML = `
                <i class="fas fa-check-circle" style="margin-right: 8px;"></i>
                页面状态已恢复！模态框已强制清理。
            `;
            document.body.appendChild(notification);
            
            setTimeout(() => {
                notification.remove();
            }, 3000);
        }
    }
}); 