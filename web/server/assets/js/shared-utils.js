// =====================================================
// EITEC VPN - 精简版共享工具函数
// =====================================================
//
// 📋 功能概述：
// - 只保留必要的工具函数，删除无用代码
// - 简化模态框管理，专注核心功能
//
// 📦 导出的全局函数：
// - formatBytes() - 字节数格式化
// - validateNetworkFormat() - 网段格式验证  
// - checkInterfaceEditPermission() - 接口编辑权限检查
// - ModalManager - 模态框管理器
//
// =====================================================

// 格式化字节数
function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
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
            alert(`⚠️ 无法执行${operation}\n\n接口 "${interfaceInfo.name}" 当前状态为：${statusText}\n\n为了安全操作，请先停止该接口后再进行${operation}。`);
            return false;
        }
        
        return true;
    } catch (error) {
        console.error('检查接口状态失败:', error);
        alert('无法检查接口状态，建议先停止相关接口再进行操作');
        return false;
    }
}

// 精简版模态框管理器
const ModalManager = {
    // 显示模态框
    show(modalElement) {
        if (!modalElement) return null;
        
        // 清理任何残留状态
        this.cleanup();
        
        // 创建 Bootstrap 模态框实例
        const modal = new bootstrap.Modal(modalElement, {
            backdrop: 'static',
            keyboard: true
        });
        
        // 显示模态框
        modal.show();
        
        return modal;
    },
    
    // 清理模态框状态
    cleanup() {
        // 移除所有遮罩层
        const backdrops = document.querySelectorAll('.modal-backdrop');
        backdrops.forEach(backdrop => backdrop.remove());
        
        // 恢复 body 状态
        document.body.classList.remove('modal-open');
        document.body.style.removeProperty('padding-right');
        document.body.style.removeProperty('overflow');
    }
};

// 全局导出
window.formatBytes = formatBytes;
window.validateNetworkFormat = validateNetworkFormat;
window.checkInterfaceEditPermission = checkInterfaceEditPermission;
window.ModalManager = ModalManager;

console.log('✅ 精简版共享工具函数已加载');
