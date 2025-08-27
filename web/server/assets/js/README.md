# API管理系统使用说明

## 概述

新的API管理系统统一管理所有HTTP请求，消除了重复代码，提供了统一的错误处理、认证管理和响应处理。

## 文件结构

- `api.js` - 核心API管理类
- `interface-management.js` - 接口管理功能（已重构）
- `module-management.js` - 模块管理功能（已重构）
- `user-management.js` - 用户管理功能（已重构）
- `dashboard-core.js` - 核心仪表板功能（已重构）
- 其他功能模块文件

## 核心特性

### 1. 统一的API类

```javascript
// 基础API类
class API {
    // 自动处理认证token
    // 统一的错误处理
    // 统一的响应处理
}

// 专用API类
class WireGuardAPI extends API {
    // WireGuard接口相关API
}

class ModuleAPI extends API {
    // 模块相关API
}

class UserAPI extends API {
    // 用户相关API
}
```

### 2. 全局API实例

```javascript
// 全局可用的API实例
window.api = {
    wireguard: new WireGuardAPI(),
    modules: new ModuleAPI(),
    users: new UserAPI(),
    system: new SystemAPI(),
    auth: new AuthAPI()
};
```

### 3. 便捷的API助手

```javascript
// 全局可用的API助手
window.apiHelper = {
    handleError(),      // 统一错误处理
    handleSuccess(),    // 统一成功处理
    confirm(),          // 确认对话框
    showLoading(),      // 显示加载状态
    hideLoading()       // 隐藏加载状态
};
```

## 使用方法

### 基本API调用

```javascript
// 获取接口列表
try {
    const result = await api.wireguard.getInterfaces();
    const interfaces = result.data || result;
    // 处理数据
} catch (error) {
    apiHelper.handleError(error, '获取接口列表失败');
}

// 创建接口
try {
    const result = await api.wireguard.createInterface(interfaceData);
    apiHelper.handleSuccess('接口创建成功！');
} catch (error) {
    apiHelper.handleError(error, '创建接口失败');
}

// 创建模块
try {
    const result = await api.modules.createModule(moduleData);
    apiHelper.handleSuccess('模块创建成功！');
} catch (error) {
    apiHelper.handleError(error, '创建模块失败');
}

// 删除模块
try {
    await api.modules.deleteModule(moduleId);
    apiHelper.handleSuccess('模块删除成功！');
} catch (error) {
    apiHelper.handleError(error, '删除模块失败');
}
```

### 加载状态管理

```javascript
// 显示加载状态
apiHelper.showLoading('加载中...');

try {
    // API调用
    const result = await api.wireguard.getInterfaces();
    // 处理结果
} catch (error) {
    // 错误处理
} finally {
    // 隐藏加载状态
    apiHelper.hideLoading();
}
```

### 错误处理

```javascript
try {
    const result = await api.wireguard.getInterfaces();
} catch (error) {
    // 自动处理认证错误
    // 自动显示错误消息
    // 自动跳转登录页（如果是认证错误）
    apiHelper.handleError(error, '自定义错误消息');
}
```

### 确认操作

```javascript
// 使用Promise风格的确认对话框
const confirmed = await apiHelper.confirm('确定要删除吗？', '确认删除');
if (confirmed) {
    // 执行删除操作
    await api.wireguard.deleteInterface(id);
}
```

## API端点映射

### WireGuard接口

| 方法 | 端点 | 描述 |
|------|------|------|
| `getInterfaces()` | `/api/v1/system/wireguard-interfaces` | 获取所有接口 |
| `createInterface(data)` | `/api/v1/interfaces` | 创建新接口 |
| `startInterface(id)` | `/api/v1/interfaces/{id}/start` | 启动接口 |
| `stopInterface(id)` | `/api/v1/interfaces/{id}/stop` | 停止接口 |
| `deleteInterface(id)` | `/api/v1/interfaces/{id}` | 删除接口 |
| `getInterfaceConfig(id)` | `/api/v1/interfaces/{id}/config` | 获取配置 |
| `getNetworkInterfaces()` | `/api/v1/system/network-interfaces` | 获取网络接口 |

### 模块管理

| 方法 | 端点 | 描述 |
|------|------|------|
| `getModules()` | `/api/v1/modules` | 获取模块列表 |
| `createModule(data)` | `/api/v1/modules` | 创建模块 |
| `updateModule(id, data)` | `/api/v1/modules/{id}` | 更新模块 |
| `deleteModule(id)` | `/api/v1/modules/{id}` | 删除模块 |
| `getModuleConfig(id)` | `/api/v1/modules/{id}/config` | 获取配置 |
| `downloadModuleConfig(id)` | `/api/v1/modules/{id}/config/download` | 下载配置 |

### 用户管理

| 方法 | 端点 | 描述 |
|------|------|------|
| `getUsers()` | `/api/v1/users` | 获取用户列表 |
| `createUser(data)` | `/api/v1/users` | 创建用户 |
| `updateUser(id, data)` | `/api/v1/users/{id}` | 更新用户 |
| `deleteUser(id)` | `/api/v1/users/{id}` | 删除用户 |
| `getUserConfig(id)` | `/api/v1/users/{id}/config` | 获取配置 |
| `downloadUserConfig(id)` | `/api/v1/users/{id}/config/download` | 下载配置 |

## 重构前后对比

### 重构前（重复代码）

```javascript
// 每个API调用都要重复这些代码
const token = localStorage.getItem('access_token');
const response = await fetch('/api/v1/interfaces', {
    method: 'POST',
    headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
});

if (response.ok) {
    const result = await response.json();
    // 处理成功
} else {
    // 处理失败
}
```

### 重构后（简洁代码）

```javascript
// 一行代码完成API调用
const result = await api.wireguard.createInterface(data);
// 自动处理认证、错误、响应

// 模块管理同样简洁
const result = await api.modules.createModule(moduleData);
await api.modules.deleteModule(moduleId);

// 用户管理同样简洁
const result = await api.users.createUser(userData);
await api.users.updateUser(userId, updateData);
await api.users.deleteUser(userId);

// 仪表板数据加载
await api.system.getDashboardStats();
await api.auth.logout();
```

## 优势

1. **代码复用**: 消除了重复的fetch代码
2. **统一管理**: 所有API端点集中管理
3. **自动认证**: 自动处理token和认证头
4. **错误处理**: 统一的错误处理和用户提示
5. **加载状态**: 统一的加载状态管理
6. **类型安全**: 更好的代码提示和错误检查
7. **维护性**: 修改API端点只需要改一个地方

## 注意事项

1. 确保在HTML中先加载`api.js`，再加载其他功能模块
2. 所有API调用都应该使用try-catch包装
3. 使用`apiHelper`提供的工具函数处理常见操作
4. 错误处理会自动处理认证问题，无需手动处理

## 扩展

如需添加新的API端点，只需在相应的API类中添加新方法：

```javascript
class WireGuardAPI extends API {
    // 添加新的API方法
    async getInterfaceStats(interfaceId) {
        return this.get(`/interfaces/${interfaceId}/stats`);
    }
}
```
