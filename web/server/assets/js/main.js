// EITEC VPN 管理平台 - 主要 JavaScript 文件

// 全局变量
let isLoggedIn = false;

// 全局配置
const API_BASE_URL = '/api/v1';

// 工具函数
function formatBytes(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatDateTime(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN');
}

// API 请求辅助函数
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
        const response = await fetch(API_BASE_URL + url, mergedOptions);
        
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

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    // 设置当前页面的导航高亮
    highlightCurrentNavigation();
    
    // 检查登录状态
    checkAuthStatus();
    
    // 初始化事件监听器
    initEventListeners();
});

function highlightCurrentNavigation() {
    const currentPath = window.location.pathname;
    const navLinks = document.querySelectorAll('.sidebar .nav-link');
    
    navLinks.forEach(link => {
        link.classList.remove('active');
        if (link.getAttribute('href') === currentPath) {
            link.classList.add('active');
        }
    });
}

// 检查认证状态
function checkAuthStatus() {
    const token = localStorage.getItem('access_token');
    if (!token && window.location.pathname !== '/login') {
        window.location.href = '/login';
        return;
    }
    isLoggedIn = !!token;
}

// 初始化事件监听器
function initEventListeners() {
    // 登录表单
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }
    
    // 登出按钮
    const logoutButtons = document.querySelectorAll('[onclick="logout()"]');
    logoutButtons.forEach(btn => {
        btn.removeAttribute('onclick');
        btn.addEventListener('click', logout);
    });
}

// 处理登录
async function handleLogin(event) {
    event.preventDefault();
    
    const form = event.target;
    const username = form.querySelector('input[name="username"]').value;
    const password = form.querySelector('input[name="password"]').value;
    
    if (!username || !password) {
        showError('请输入用户名和密码');
        return;
    }
    
    try {
        showLoading('登录中...');
        
        const response = await fetch('/api/v1/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, password })
        });
        
        const data = await response.json();
        
        if (response.ok && data.access_token) {
            // 保存令牌
            localStorage.setItem('access_token', data.access_token);
            if (data.refresh_token) {
                localStorage.setItem('refresh_token', data.refresh_token);
            }
            
            showSuccess('登录成功，正在跳转...');
            setTimeout(() => {
                window.location.href = '/';
            }, 1000);
        } else {
            showError(data.message || '登录失败，请检查用户名和密码');
        }
    } catch (error) {
        console.error('登录错误:', error);
        showError('网络错误，请稍后重试');
    } finally {
        hideLoading();
    }
}

// 登出
function logout() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    showSuccess('已退出登录');
    setTimeout(() => {
        window.location.href = '/login';
    }, 1000);
}

// 显示通知消息
function showNotification(message, type = 'info') {
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type} alert-dismissible fade show position-fixed`;
    alertDiv.style.cssText = 'top: 20px; right: 20px; z-index: 9999; min-width: 300px;';
    alertDiv.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    `;
    
    document.body.appendChild(alertDiv);
    
    // 5秒后自动消失
    setTimeout(() => {
        if (alertDiv.parentNode) {
            alertDiv.remove();
        }
    }, 5000);
}

// 确认对话框
function confirmAction(message, callback) {
    if (confirm(message)) {
        callback();
    }
}

// 显示加载状态
function showLoading(message = '加载中...') {
    // 移除现有的加载提示
    hideLoading();
    
    const loading = document.createElement('div');
    loading.id = 'loadingToast';
    loading.className = 'position-fixed top-50 start-50 translate-middle bg-primary text-white px-4 py-2 rounded';
    loading.style.zIndex = '9999';
    loading.innerHTML = `
        <div class="d-flex align-items-center">
            <div class="spinner-border spinner-border-sm me-2" role="status">
                <span class="visually-hidden">Loading...</span>
            </div>
            ${message}
        </div>
    `;
    
    document.body.appendChild(loading);
}

// 隐藏加载状态
function hideLoading() {
    const loading = document.getElementById('loadingToast');
    if (loading) {
        loading.remove();
    }
}

// 显示成功消息
function showSuccess(message) {
    showToast(message, 'success');
}

// 显示错误消息
function showError(message) {
    showToast(message, 'danger');
}

// 显示信息消息
function showInfo(message) {
    showToast(message, 'info');
}

// 通用toast显示函数
function showToast(message, type = 'info') {
    const toastId = 'toast-' + Date.now();
    const toast = document.createElement('div');
    toast.id = toastId;
    toast.className = `position-fixed top-0 start-50 translate-middle-x mt-3 alert alert-${type} alert-dismissible fade show`;
    toast.style.zIndex = '9999';
    toast.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    `;
    
    document.body.appendChild(toast);
    
    // 3秒后自动消失
    setTimeout(() => {
        if (document.getElementById(toastId)) {
            toast.remove();
        }
    }, 3000);
}

// 表格排序
function sortTable(table, column, direction = 'asc') {
    const tbody = table.querySelector('tbody');
    const rows = Array.from(tbody.querySelectorAll('tr'));
    
    const sortedRows = rows.sort((a, b) => {
        const aVal = a.children[column].textContent.trim();
        const bVal = b.children[column].textContent.trim();
        
        if (direction === 'asc') {
            return aVal.localeCompare(bVal, 'zh-CN', { numeric: true });
        } else {
            return bVal.localeCompare(aVal, 'zh-CN', { numeric: true });
        }
    });
    
    // 清空tbody并重新添加排序后的行
    tbody.innerHTML = '';
    sortedRows.forEach(row => tbody.appendChild(row));
}

// 导出为 CSV
function exportToCSV(data, filename) {
    const csv = data.map(row => 
        row.map(field => 
            typeof field === 'string' && field.includes(',') 
                ? `"${field}"` 
                : field
        ).join(',')
    ).join('\n');
    
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    
    if (link.download !== undefined) {
        const url = URL.createObjectURL(blob);
        link.setAttribute('href', url);
        link.setAttribute('download', filename);
        link.style.visibility = 'hidden';
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    }
}

// 页面离开前的确认
window.addEventListener('beforeunload', function(e) {
    // 如果有未保存的更改，显示确认对话框
    const hasUnsavedChanges = document.querySelector('.form-modified');
    if (hasUnsavedChanges) {
        e.preventDefault();
        e.returnValue = '';
    }
});

// 表单修改跟踪
function trackFormChanges(form) {
    const inputs = form.querySelectorAll('input, select, textarea');
    
    inputs.forEach(input => {
        input.addEventListener('change', function() {
            form.classList.add('form-modified');
        });
    });
    
    form.addEventListener('submit', function() {
        form.classList.remove('form-modified');
    });
}

// 格式化文件大小
function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// 格式化日期时间
function formatDateTime(dateString) {
    if (!dateString) return '--';
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return '--';
    
    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

// 格式化相对时间
function formatRelativeTime(dateString) {
    if (!dateString) return '--';
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return '--';
    
    const now = new Date();
    const diff = now - date;
    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);
    
    if (seconds < 60) return `${seconds}秒前`;
    if (minutes < 60) return `${minutes}分钟前`;
    if (hours < 24) return `${hours}小时前`;
    if (days < 30) return `${days}天前`;
    
    return formatDateTime(dateString);
}

// 验证IP地址
function isValidIP(ip) {
    const ipRegex = /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
    return ipRegex.test(ip);
}

// 验证CIDR网络
function isValidCIDR(cidr) {
    const cidrRegex = /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\/([0-9]|[1-2][0-9]|3[0-2])$/;
    return cidrRegex.test(cidr);
}

// 复制到剪贴板
async function copyToClipboard(text) {
    try {
        await navigator.clipboard.writeText(text);
        showSuccess('已复制到剪贴板');
    } catch (err) {
        console.error('复制失败:', err);
        // 降级方案
        const textArea = document.createElement('textarea');
        textArea.value = text;
        document.body.appendChild(textArea);
        textArea.select();
        try {
            document.execCommand('copy');
            showSuccess('已复制到剪贴板');
        } catch (err) {
            showError('复制失败');
        }
        document.body.removeChild(textArea);
    }
}

// 防抖函数
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// 节流函数
function throttle(func, limit) {
    let inThrottle;
    return function() {
        const args = arguments;
        const context = this;
        if (!inThrottle) {
            func.apply(context, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

// 导出常用函数到全局作用域
window.apiRequest = apiRequest;
window.showLoading = showLoading;
window.hideLoading = hideLoading;
window.showSuccess = showSuccess;
window.showError = showError;
window.showInfo = showInfo;
window.formatFileSize = formatFileSize;
window.formatDateTime = formatDateTime;
window.formatRelativeTime = formatRelativeTime;
window.isValidIP = isValidIP;
window.isValidCIDR = isValidCIDR;
window.copyToClipboard = copyToClipboard;
window.logout = logout; 