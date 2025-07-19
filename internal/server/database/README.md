# Server Database 服务器端数据库

本目录包含VPN服务器端的数据库连接、迁移和管理功能。

## 文件结构

```
internal/server/database/
├── README.md          # 本文档
└── database.go        # 数据库管理主文件
```

## 核心功能

### 1. 数据库连接管理
- **InitDatabase()**: 标准数据库初始化
- **InitDatabaseWithOptions()**: 带选项的数据库初始化
- **Close()**: 关闭数据库连接

### 2. 数据库迁移
- **AutoMigrate()**: 自动迁移所有模型的表结构
- 支持的模型：Module, ModuleLog, User, SystemConfig, IPPool

### 3. 默认数据初始化
- **InitDefaultData()**: 初始化系统默认数据
- **initDefaultAdmin()**: 创建默认管理员账户 (admin/admin123)
- **initIPPool()**: 初始化IP地址池 (10.10.0.2-10.10.0.254)
- **initSystemConfig()**: 初始化系统配置项

### 4. IP地址池管理
- **GetAvailableIP()**: 获取可用IP地址
- **AllocateIP()**: 分配IP地址给模块
- **ReleaseIP()**: 释放IP地址

### 5. 系统配置管理
- **GetSystemConfig()**: 获取系统配置值
- **SetSystemConfig()**: 设置系统配置值

## 技术特性

### 数据库类型
- **SQLite**: 使用GORM ORM框架
- **日志模式**: 默认Silent，可通过`DB_DEBUG=true`启用详细日志

### 初始化逻辑
- 自动创建数据库目录
- 检测新数据库并自动初始化默认数据
- 支持强制重新初始化 (`forceInit=true`)

### 全局变量
- **DB**: 全局GORM数据库实例

## 使用示例

### 基本使用
```go
import "eitec-vpn/internal/server/database"

// 初始化数据库
err := database.InitDatabase("data/server.db")
if err != nil {
    log.Fatal(err)
}

// 使用全局DB实例
var modules []models.Module
database.DB.Find(&modules)

// 关闭数据库
defer database.Close()
```

### IP地址管理
```go
// 获取可用IP
ip, err := database.GetAvailableIP()
if err != nil {
    return err
}

// 分配IP给模块
err = database.AllocateIP(ip, moduleID)
if err != nil {
    return err
}

// 释放IP
err = database.ReleaseIP(ip)
```

### 配置管理
```go
// 获取配置
endpoint, err := database.GetSystemConfig("server.endpoint")
if err != nil {
    return err
}

// 设置配置
err = database.SetSystemConfig("server.endpoint", "vpn.example.com:51820")
```

## 架构说明

这个database包专门为服务器端设计：

- **服务器专用**: 与 `internal/server/models` 紧密集成
- **数据持久化**: 负责所有服务器端数据的存储和管理
- **业务逻辑**: 包含IP分配、配置管理等服务器端特有的业务逻辑
- **模块端分离**: 模块端使用独立的本地存储，不依赖此数据库

## 默认配置

### 管理员账户
- 用户名: `admin`
- 密码: `admin123` (生产环境请及时修改)
- 角色: 管理员

### IP地址池
- 网络段: `10.10.0.0/24`
- 可用IP: `10.10.0.2` - `10.10.0.254`
- 总数: 253个IP地址

### 系统配置项
- `server.public_key`: 服务器WireGuard公钥
- `server.private_key`: 服务器WireGuard私钥  
- `server.endpoint`: 服务器公网地址:端口
- `wg.network`: WireGuard网络段
- `wg.dns`: DNS服务器 