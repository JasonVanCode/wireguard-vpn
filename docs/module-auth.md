# EITEC VPN 模块端身份验证设计

## 📋 概述

为了增强EITEC VPN模块端的安全性，我们为模块端添加了完整的身份验证功能。现在所有的管理操作都需要先登录才能访问，确保只有授权用户才能管理模块。

## 🔐 认证架构

### 认证流程
```
1. 用户访问模块端 → 检查是否已登录
2. 未登录用户 → 重定向到登录页面
3. 用户输入凭据 → 后端验证用户名密码
4. 验证成功 → 生成JWT token → 返回给前端
5. 前端保存token → 后续请求携带token
6. 后端验证token → 允许访问受保护资源
```

### JWT Token机制
- **算法**：HS256 (HMAC SHA-256)
- **有效期**：普通登录24小时，"记住我"7天
- **存储**：前端localStorage
- **传输**：HTTP Authorization Header (Bearer token)

## 🎨 登录界面设计

### 视觉特色
- **深色主题**：与server端保持一致的专业深色风格
- **动态背景**：30个浮动粒子营造科技感
- **玻璃态设计**：毛玻璃背景效果 + 背景模糊
- **Shimmer动画**：卡片光泽扫过效果
- **微动效果**：输入框聚焦上移、按钮光波动画

### 用户体验
- **自动聚焦**：页面加载后自动聚焦用户名输入框
- **键盘支持**：Enter键快速登录
- **错误提示**：震动动画 + 详细错误信息
- **加载状态**：登录过程中的视觉反馈
- **记住登录**：可选择延长登录有效期

### 界面元素
```html
<!-- Logo区域 -->
<div class="logo-section">
    <div class="logo-icon">
        <i class="fas fa-server"></i>  <!-- 服务器图标 -->
    </div>
    <h2>模块管理</h2>
    <p>EITEC VPN Module</p>
</div>

<!-- 登录表单 -->
<form>
    <input type="text" placeholder="用户名" required>
    <input type="password" placeholder="密码" required>
    <checkbox> 记住登录状态
    <button type="submit">登录</button>
</form>
```

## 🔧 技术实现

### 后端认证中间件
```go
func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 获取Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            // 页面请求重定向到登录页，API请求返回401
            if strings.HasPrefix(c.Request.URL.Path, "/api/") {
                c.JSON(401, gin.H{"message": "未授权访问"})
            } else {
                c.Redirect(302, "/login")
            }
            c.Abort()
            return
        }

        // 验证JWT token
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        token, err := jwt.ParseWithClaims(tokenString, &ModuleClaims{}, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            // 处理token无效的情况
            handleInvalidToken(c)
            return
        }

        // 保存用户信息到上下文
        if claims, ok := token.Claims.(*ModuleClaims); ok {
            c.Set("username", claims.Username)
        }

        c.Next()
    }
}
```

### 前端认证处理
```javascript
// 创建带认证header的请求
function createAuthHeaders() {
    const token = localStorage.getItem('module_token');
    const headers = { 'Content-Type': 'application/json' };
    if (token) {
        headers['Authorization'] = 'Bearer ' + token;
    }
    return headers;
}

// 处理认证失败
function handleAuthError() {
    localStorage.removeItem('module_token');
    window.location.href = '/login';
}

// API请求示例
async function loadModuleStatus() {
    try {
        const headers = createAuthHeaders();
        const response = await fetch('/api/v1/status', { headers });
        if (response.status === 401) {
            handleAuthError();
            return;
        }
        // 处理正常响应...
    } catch (error) {
        console.error('请求失败:', error);
    }
}
```

## 🔒 安全特性

### 密码安全
- **简单验证**：当前使用硬编码用户名密码（admin/admin123）
- **生产环境建议**：使用环境变量或配置文件管理凭据
- **未来增强**：支持bcrypt密码哈希、密码复杂度要求

### Token安全
- **签名验证**：使用HMAC-SHA256签名确保token完整性
- **有效期控制**：普通登录24小时，记住登录7天
- **自动过期**：token过期后自动清除并重定向登录

### 传输安全
- **HTTPS建议**：生产环境应启用HTTPS加密传输
- **Header传输**：token通过Authorization header传输
- **本地存储**：使用localStorage安全存储token

## 🛡️ 访问控制

### 路由保护
```go
// 公开路由（无需认证）
router.GET("/login", loginPage)
router.GET("/health", healthCheck)
router.POST("/api/v1/auth/login", handleLogin)

// 受保护路由（需要认证）
router.GET("/", authMiddleware(), homePage)
router.GET("/config", authMiddleware(), configPage)

// 受保护API（需要认证）
api := router.Group("/api/v1", authMiddleware())
{
    api.GET("/status", getStatus)
    api.POST("/wireguard/start", startWireGuard)
    api.POST("/wireguard/stop", stopWireGuard)
    // ... 其他API
}
```

### 权限级别
- **当前实现**：单一用户级别（admin）
- **未来扩展**：可支持多级权限（只读、操作员、管理员）

## 📱 用户交互

### 登录流程
1. **访问检查**：用户访问任何页面时检查登录状态
2. **自动跳转**：未登录自动跳转到登录页面
3. **凭据输入**：用户输入用户名和密码
4. **验证反馈**：实时显示登录状态和错误信息
5. **登录成功**：跳转到原访问页面或主页面

### 状态保持
- **Token验证**：页面加载时验证存储的token有效性
- **自动登录**：有效token自动跳转到主页面
- **过期处理**：token过期时自动清除并提示重新登录

### 错误处理
```javascript
// 网络错误
catch (error) {
    showError('网络错误，请稍后重试');
}

// 认证错误
if (response.status === 401) {
    showError('用户名或密码错误');
}

// 服务器错误
if (response.status === 500) {
    showError('服务器内部错误');
}
```

## ⚙️ 配置管理

### 默认用户配置
```go
// 默认用户 (生产环境应使用配置文件)
var defaultUsers = map[string]string{
    "admin": "admin123",
}

// JWT密钥 (生产环境应使用更安全的密钥)
var jwtSecret = []byte("eitec-module-jwt-secret-key")
```

### 环境变量支持
```bash
# 建议的环境变量配置
export MODULE_ADMIN_USER="admin"
export MODULE_ADMIN_PASS="your_secure_password"
export MODULE_JWT_SECRET="your_jwt_secret_key_32_chars_long"
export MODULE_SESSION_TIMEOUT="24h"
```

### 配置文件示例
```yaml
# module-auth.yaml
auth:
  users:
    - username: admin
      password: $2a$10$encrypted_password_hash
  jwt:
    secret: your_jwt_secret_key
    expiry: 24h
    remember_expiry: 168h  # 7 days
  security:
    max_login_attempts: 5
    lockout_duration: 15m
```

## 🚀 部署建议

### 生产环境安全
1. **更改默认密码**：替换默认的admin/admin123
2. **使用强密钥**：生成32字符以上的随机JWT密钥
3. **启用HTTPS**：确保传输加密
4. **定期轮换**：定期更换JWT密钥和用户密码
5. **访问日志**：记录登录尝试和访问日志

### 监控建议
```go
// 登录监控
func logLoginAttempt(username, ip string, success bool) {
    log.Printf("Login attempt: user=%s, ip=%s, success=%t", 
        username, ip, success)
}

// 访问监控
func logAPIAccess(username, method, path string) {
    log.Printf("API access: user=%s, method=%s, path=%s", 
        username, method, path)
}
```

## 🔮 未来增强计划

### 安全增强
- [ ] **密码哈希**：使用bcrypt替代明文密码
- [ ] **多因素认证**：支持TOTP、短信验证
- [ ] **失败限制**：登录失败次数限制和临时锁定
- [ ] **会话管理**：支持会话失效和强制登出
- [ ] **角色权限**：细粒度的权限控制系统

### 用户体验
- [ ] **密码重置**：支持安全的密码重置流程
- [ ] **用户管理**：支持多用户管理界面
- [ ] **登录历史**：显示最近登录记录
- [ ] **设备管理**：管理和撤销设备登录状态

### 集成功能
- [ ] **LDAP集成**：支持企业LDAP认证
- [ ] **SSO支持**：单点登录集成
- [ ] **API密钥**：支持API密钥认证方式
- [ ] **Webhook通知**：登录事件通知

## 📝 总结

通过添加完整的身份验证功能，EITEC VPN模块端现在具备了：

1. **安全访问控制**：确保只有授权用户能访问管理功能
2. **现代化登录界面**：与系统整体风格一致的美观界面
3. **JWT token认证**：标准化的无状态认证机制
4. **用户友好体验**：流畅的登录流程和错误处理
5. **可扩展架构**：为未来功能增强预留接口

这个认证系统为模块端提供了基础而必要的安全保护，同时保持了良好的用户体验和代码可维护性。 