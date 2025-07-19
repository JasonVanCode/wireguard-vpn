# 模块端数据模型

本目录包含模块端的数据模型定义。采用简化设计原则，只保留核心必需的数据结构。

## 📋 模型列表

### 1. LocalModule (`local_module.go`)
**模块配置模型** - 存储模块的基本配置信息
- 模块ID、服务器地址、API密钥
- 配置状态、同步时间等
- 这是模块端的核心配置表

### 2. LocalUser (`local_user.go`)  
**用户认证模型** - 管理模块端的用户认证
- 用户名、密码（加密存储）
- 用户角色（admin/viewer）
- 防止未授权访问模块管理界面

## 🎯 设计原则

### 简化理念
- **最小化存储** - 只存储必需的配置数据
- **实时获取** - 状态、监控数据通过API实时获取
- **职责清晰** - 模块端专注配置管理，不做复杂的数据分析

### 不包含的功能
以下功能通过API实时获取，不需要本地存储：
- ❌ 性能监控数据
- ❌ 连接历史记录  
- ❌ 流量统计历史
- ❌ 系统信息缓存
- ❌ WireGuard详细状态
- ❌ 操作日志记录

## 🔧 使用方式

### 数据库迁移
```go
import "eitec-vpn/internal/module/models"

// 自动迁移表结构
err := models.AutoMigrate(db)
```

### 模型使用示例
```go
// 保存模块配置
module := &models.LocalModule{
    ModuleID:  1001,
    ServerURL: "https://vpn-server.example.com",
    APIKey:    "your-api-key",
    Status:    models.StatusConfigured,
}
db.Create(module)

// 用户认证
user := &models.LocalUser{
    Username: "admin",
    Password: hashedPassword,
    Role:     models.RoleAdmin,
}
db.Create(user)
```

## 📈 优势

1. **简单高效** - 只有2个表，易于理解和维护
2. **快速启动** - 数据库初始化迅速
3. **资源节省** - 减少存储空间和内存占用
4. **实时准确** - 状态数据始终是最新的
5. **易于部署** - 依赖关系简单，部署容易

这种设计确保模块端专注于核心功能：配置管理和用户认证，其他复杂功能由服务器端统一处理。 