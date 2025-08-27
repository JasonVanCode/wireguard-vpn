// EITEC VPN 模块管理中心 - API封装

// API基础配置
const API_CONFIG = {
    baseURL: '/api/v1',
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json'
    }
};

// 认证token管理
class AuthManager {
    static getToken() {
        // 优先从localStorage获取
        let token = localStorage.getItem('module_token') || sessionStorage.getItem('module_token');
        
        // 如果没有从存储中获取到，尝试从cookie获取
        if (!token) {
            const cookies = document.cookie.split(';');
            for (let cookie of cookies) {
                const [name, value] = cookie.trim().split('=');
                if (name === 'module_token') {
                    token = value;
                    break;
                }
            }
        }
        
        return token;
    }
    
    static setToken(token, remember = true) {
        if (remember) {
            localStorage.setItem('module_token', token);
        } else {
            sessionStorage.setItem('module_token', token);
        }
    }
    
    static clearToken() {
        localStorage.removeItem('module_token');
        sessionStorage.removeItem('module_token');
    }
    
    static isAuthenticated() {
        return !!this.getToken();
    }
}

// 统一的API请求函数
class API {
    static async request(endpoint, options = {}) {
        const token = AuthManager.getToken();
        
        if (!token && !options.skipAuth) {
            // 如果没有token，直接清除并跳转登录，不抛出错误
            console.warn('未找到认证token，跳转到登录页面');
            AuthManager.clearToken();
            setTimeout(() => {
                window.location.href = '/login';
            }, 1000);
            return null;
        }
        
        const url = `${API_CONFIG.baseURL}${endpoint}`;
        const config = {
            method: options.method || 'GET',
            headers: {
                ...API_CONFIG.headers,
                ...options.headers
            },
            ...options
        };
        
        // 添加认证头
        if (token && !options.skipAuth) {
            config.headers['Authorization'] = `Bearer ${token}`;
        }
        
        try {
            const response = await fetch(url, config);
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            
            if (data.code === 200) {
                return data.data;
            } else {
                throw new Error(data.message || `API错误: ${data.code}`);
            }
        } catch (error) {
            console.error(`API请求失败 [${endpoint}]:`, error);
            
            // 如果是认证错误，清除token并跳转登录
            if (error.message.includes('401') || error.message.includes('403')) {
                AuthManager.clearToken();
                setTimeout(() => {
                    window.location.href = '/login';
                }, 1000);
            }
            
            throw error;
        }
    }
    
    // GET请求
    static async get(endpoint, options = {}) {
        return this.request(endpoint, { ...options, method: 'GET' });
    }
    
    // POST请求
    static async post(endpoint, data, options = {}) {
        return this.request(endpoint, {
            ...options,
            method: 'POST',
            body: JSON.stringify(data)
        });
    }
    
    // PUT请求
    static async put(endpoint, data, options = {}) {
        return this.request(endpoint, {
            ...options,
            method: 'PUT',
            body: JSON.stringify(data)
        });
    }
    
    // DELETE请求
    static async delete(endpoint, options = {}) {
        return this.request(endpoint, { ...options, method: 'DELETE' });
    }
}

// 具体的API端点函数
const DashboardAPI = {
    // 获取仪表板统计数据
    async getStats() {
        return await API.get('/dashboard/stats');
    },
    
    // 获取状态信息
    async getStatus() {
        return await API.get('/status');
    }
};

const WireGuardAPI = {
    // 控制WireGuard接口
    async control(action, interfaceName) {
        return await API.post('/wireguard/control', {
            action: action,
            interface: interfaceName
        });
    },
    
    // 上传WireGuard配置
    async uploadConfig(interfaceName, configData) {
        return await API.post('/wireguard/config/upload', {
            interface: interfaceName,
            config_data: configData
        });
    },
    
    // 读取WireGuard配置
    async getConfig(interfaceName) {
        return await API.get(`/wireguard/config/${interfaceName}`);
    }
};

// 导出API对象
window.API = API;
window.AuthManager = AuthManager;
window.DashboardAPI = DashboardAPI;
window.WireGuardAPI = WireGuardAPI;
