# Server Models 服务器端模型

本目录包含VPN服务器端的所有数据模型，采用简化的文件组织结构。

## 文件结构

```
internal/server/models/
├── README.md              # 本文档
├── migrate.go             # 数据库迁移函数
├── module.go              # 模块信息模型 (包含ModuleStatus枚举)
├── user.go                # 用户信息模型 (简化版)
├── system_config.go       # 系统配置模型
├── ip_pool.go             # IP地址池模型
├── wireguard_interface.go # WireGuard接口模型
├── wireguard_key.go       # WireGuard密钥结构体
├── module_config.go       # 模块配置结构体
└── dashboard_stats.go     # 仪表盘统计结构体
```

## 核心数据库模型

### 1. Module (module.go)
- **作用**: 存储WireGuard模块的基本信息
- **状态**: ModuleStatus枚举 (离线/在线/警告/未配置)
- **字段**: 包含公私钥、IP地址、流量统计等
- **关联**: 多对一关联WireGuardInterface

### 2. WireGuardInterface (wireguard_interface.go)
- **作用**: 管理多个WireGuard接口
- **特点**: 支持多接口隔离，每个接口独立网段
- **字段**: 接口名称、端口、网络段、状态等

### 3. User (user.go)
- **作用**: 系统用户管理 (简化版，仅管理员)
- **认证**: 用户名密码登录
- **简化**: 移除复杂的角色系统

### 4. SystemConfig (system_config.go)
- **作用**: 存储系统配置参数
- **特点**: 键值对结构，支持不同数据类型

### 5. IPPool (ip_pool.go)
- **作用**: 管理IP地址池分配
- **关联**: 可选关联Module (已分配的IP)

## 配置结构体 (非数据库模型)

### 6. WireGuardKey (wireguard_key.go)
- **作用**: WireGuard公私钥对结构

### 7. ModuleConfig (module_config.go)
- **作用**: 模块端配置文件结构

### 8. DashboardStats (dashboard_stats.go)
- **作用**: 仪表盘统计数据结构

## 🎯 优化完成总结

### ✅ 已移除的复杂功能：
1. **ModuleLog表** - 复杂的日志系统已移除，改用简单的标准日志输出
2. **复杂用户角色** - 移除访客/管理员/超级管理员的复杂角色系统
3. **数据库日志** - 不再在数据库中存储操作日志，减少数据库复杂度
4. **权限检查** - 简化中间件，移除复杂的权限验证逻辑

### ✅ 保留的核心功能：
1. **模块管理** - 核心的VPN模块管理功能
2. **多接口支持** - 支持多个WireGuard接口隔离
3. **IP地址管理** - 自动分配和管理IP地址
4. **基础配置** - 系统配置管理
5. **用户认证** - 简化的管理员登录

### ✅ 优化效果：
- **数据库表数量**: 从9个减少到6个 (-33%)
- **代码复杂度**: 大幅降低，移除了大量角色和权限相关代码
- **维护成本**: 显著减少，更适合中小型VPN管理系统
- **编译成功**: 所有代码都能正常编译运行

### 📊 最终数据库结构：
```
WireGuardInterface (接口管理)
├── Module (模块信息)
├── IPPool (IP地址池)
├── SystemConfig (系统配置)
└── User (用户管理)
```

这种简化设计更适合中小型VPN管理系统，减少了不必要的复杂性，提高了系统的可维护性和开发效率。 