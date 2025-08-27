/**
 * 统一API管理类
 * 负责管理所有HTTP请求，统一处理认证、错误和响应
 */
class API {
    constructor() {
        this.baseURL = '/api/v1';
        this.tokenKey = 'access_token';
        this.defaultHeaders = {
            'Content-Type': 'application/json'
        };
    }

    /**
     * 获取认证token
     */
    getAuthToken() {
        return localStorage.getItem(this.tokenKey);
    }

    /**
     * 构建请求头
     */
    buildHeaders(customHeaders = {}) {
        const token = this.getAuthToken();
        const headers = { ...this.defaultHeaders, ...customHeaders };
        
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }
        
        return headers;
    }

    /**
     * 检查响应状态
     */
    async checkResponse(response) {
        if (!response.ok) {
            let errorMessage = '请求失败';
            try {
                const errorData = await response.json();
                errorMessage = errorData.message || errorData.error || errorMessage;
            } catch (e) {
                errorMessage = `HTTP ${response.status}: ${response.statusText}`;
            }
            throw new Error(errorMessage);
        }
        return response;
    }

    /**
     * 处理API响应
     */
    async handleResponse(response) {
        await this.checkResponse(response);
        try {
            const data = await response.json();
            console.log(`API响应成功 [${response.url}]:`, data);
            return data;
        } catch (e) {
            console.error('响应数据解析失败:', e);
            throw new Error('响应数据解析失败');
        }
    }

    /**
     * 通用请求方法
     */
    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: this.buildHeaders(options.headers),
            ...options
        };

        try {
            const response = await fetch(url, config);
            return await this.handleResponse(response);
        } catch (error) {
            console.error(`API请求失败 [${endpoint}]:`, error);
            throw error;
        }
    }

    /**
     * GET请求
     */
    async get(endpoint, options = {}) {
        return this.request(endpoint, { method: 'GET', ...options });
    }

    /**
     * POST请求
     */
    async post(endpoint, data, options = {}) {
        console.log(`发送POST请求到 ${endpoint}:`, data);
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(data),
            ...options
        });
    }

    /**
     * PUT请求
     */
    async put(endpoint, data, options = {}) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(data),
            ...options
        });
    }

    /**
     * DELETE请求
     */
    async delete(endpoint, options = {}) {
        return this.request(endpoint, { method: 'DELETE', ...options });
    }

    /**
     * 检查认证状态
     */
    isAuthenticated() {
        return !!this.getAuthToken();
    }

    /**
     * 清除认证信息
     */
    clearAuth() {
        localStorage.removeItem(this.tokenKey);
    }
}

/**
 * WireGuard接口相关API
 */
class WireGuardAPI extends API {
    /**
     * 获取所有WireGuard接口列表
     */
    async getInterfaces() {
        return this.get('/system/wireguard-interfaces');
    }



    /**
     * 创建新接口
     */
    async createInterface(interfaceData) {
        return this.post('/interfaces', interfaceData);
    }

    /**
     * 启动接口
     */
    async startInterface(interfaceId) {
        return this.put(`/interfaces/${interfaceId}/start`);
    }

    /**
     * 停止接口
     */
    async stopInterface(interfaceId) {
        return this.put(`/interfaces/${interfaceId}/stop`);
    }

    /**
     * 删除接口
     */
    async deleteInterface(interfaceId) {
        return this.delete(`/interfaces/${interfaceId}`);
    }

    /**
     * 获取接口配置
     */
    async getInterfaceConfig(interfaceId) {
        return this.get(`/interfaces/${interfaceId}/config`);
    }

    /**
     * 获取网络接口列表
     */
    async getNetworkInterfaces() {
        return this.get('/system/network-interfaces');
    }
}

/**
 * 模块相关API
 */
class ModuleAPI extends API {
    /**
     * 获取模块列表
     */
    async getModules() {
        return this.get('/modules');
    }



    /**
     * 创建模块
     */
    async createModule(moduleData) {
        return this.post('/modules', moduleData);
    }

    /**
     * 更新模块
     */
    async updateModule(moduleId, moduleData) {
        return this.put(`/modules/${moduleId}`, moduleData);
    }

    /**
     * 删除模块
     */
    async deleteModule(moduleId) {
        return this.delete(`/modules/${moduleId}`);
    }

    /**
     * 获取模块配置
     */
    async getModuleConfig(moduleId) {
        return this.get(`/modules/${moduleId}/config`);
    }

    /**
     * 下载模块配置
     */
    async downloadModuleConfig(moduleId) {
        return this.get(`/modules/${moduleId}/config/download`);
    }
}

/**
 * 用户相关API
 */
class UserAPI extends API {
    /**
     * 获取用户列表
     */
    async getUsers() {
        return this.get('/users');
    }

    /**
     * 创建用户
     */
    async createUser(userData) {
        return this.post('/users', userData);
    }

    /**
     * 更新用户
     */
    async updateUser(userId, userData) {
        return this.put(`/users/${userId}`, userData);
    }

    /**
     * 删除用户
     */
    async deleteUser(userId) {
        return this.delete(`/users/${userId}`);
    }

    /**
     * 获取用户配置
     */
    async getUserConfig(userId) {
        return this.get(`/users/${userId}/config`);
    }

    /**
     * 下载用户配置
     */
    async downloadUserConfig(userId) {
        return this.get(`/users/${userId}/config/download`);
    }
}

/**
 * 用户VPN相关API
 */
class UserVPNAPI extends API {
    /**
     * 创建用户VPN
     */
    async createUserVPN(userData) {
        return this.post('/user-vpn', userData);
    }

    /**
     * 获取用户VPN信息
     */
    async getUserVPN(userId) {
        return this.get(`/user-vpn/${userId}`);
    }

    /**
     * 更新用户VPN
     */
    async updateUserVPN(userId, userData) {
        return this.put(`/user-vpn/${userId}`, userData);
    }

    /**
     * 删除用户VPN
     */
    async deleteUserVPN(userId) {
        return this.delete(`/user-vpn/${userId}`);
    }

    /**
     * 生成用户VPN配置
     */
    async generateUserVPNConfig(userId) {
        return this.get(`/user-vpn/${userId}/config`);
    }

    /**
     * 获取模块的用户VPN列表
     */
    async getUserVPNsByModule(moduleId, page = 1, pageSize = 20) {
        return this.get(`/modules/${moduleId}/users?page=${page}&page_size=${pageSize}`);
    }

    /**
     * 获取模块的用户VPN统计
     */
    async getUserVPNStats(moduleId) {
        return this.get(`/modules/${moduleId}/user-stats`);
    }

    /**
     * 下载用户VPN配置文件
     */
    async downloadUserConfig(userId) {
        try {
            // 使用fetch直接下载文件（因为需要处理blob和响应头）
            const response = await fetch(`${this.baseURL}/user-vpn/${userId}/config`, {
                headers: this.buildHeaders()
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
                
                return { success: true, fileName };
            } else {
                throw new Error('下载失败');
            }
        } catch (error) {
            console.error('下载用户配置失败:', error);
            throw error;
        }
    }
}

/**
 * 系统相关API
 */
class SystemAPI extends API {
    /**
     * 获取系统状态
     */
    async getSystemStatus() {
        return this.get('/system/status');
    }

    /**
     * 获取系统配置
     */
    async getSystemConfig() {
        return this.get('/system/config');
    }

    /**
     * 更新系统配置
     */
    async updateSystemConfig(configData) {
        return this.put('/system/config', configData);
    }

    /**
     * 获取仪表板统计
     */
    async getDashboardStats() {
        return this.get('/dashboard/stats');
    }
}

/**
 * 认证相关API
 */
class AuthAPI extends API {
    /**
     * 用户登录
     */
    async login(credentials) {
        return this.post('/auth/login', credentials);
    }

    /**
     * 用户登出
     */
    async logout() {
        return this.post('/auth/logout');
    }

    /**
     * 刷新token
     */
    async refreshToken() {
        return this.post('/auth/refresh');
    }

    /**
     * 获取当前用户信息
     */
    async getCurrentUser() {
        return this.get('/auth/me');
    }
}

/**
 * 全局API实例
 */
const api = {
    wireguard: new WireGuardAPI(),
    modules: new ModuleAPI(),
    users: new UserAPI(),
    userVPN: new UserVPNAPI(),
    system: new SystemAPI(),
    auth: new AuthAPI()
};

/**
 * 便捷的API调用方法
 */
const apiHelper = {
    /**
     * 统一错误处理
     */
    handleError(error, defaultMessage = '操作失败') {
        console.error('API错误:', error);
        const message = error.message || defaultMessage;
        
        // 如果是认证错误，跳转到登录页
        if (message.includes('认证') || message.includes('token') || message.includes('401')) {
            api.auth.clearAuth();
            window.location.href = '/login';
            return;
        }
        
        // 显示错误消息
        if (typeof showAlert === 'function') {
            showAlert(message, 'error');
        } else {
            alert(message);
        }
        
        return message;
    },

    /**
     * 统一成功处理
     */
    handleSuccess(message = '操作成功') {
        if (typeof showAlert === 'function') {
            showAlert(message, 'success');
        } else {
            console.log(message);
        }
    },

    /**
     * 确认操作
     */
    confirm(message, title = '确认操作') {
        return new Promise((resolve) => {
            if (typeof showConfirmDialog === 'function') {
                showConfirmDialog(message, title).then(resolve);
            } else {
                resolve(confirm(message));
            }
        });
    },

    /**
     * 显示加载状态
     */
    showLoading(message = '加载中...') {
        if (typeof showLoading === 'function') {
            showLoading(message);
        }
    },

    /**
     * 隐藏加载状态
     */
    hideLoading() {
        if (typeof hideLoading === 'function') {
            hideLoading();
        }
    }
};

// 导出到全局
window.api = api;
window.apiHelper = apiHelper;

// 兼容性导出（保持原有代码可用）
window.API = API;
window.WireGuardAPI = WireGuardAPI;
window.ModuleAPI = ModuleAPI;
window.UserAPI = UserAPI;
window.UserVPNAPI = UserVPNAPI;
window.SystemAPI = SystemAPI;
window.AuthAPI = AuthAPI;
