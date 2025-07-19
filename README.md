# WireGuard 集中管理平台 v2.1

基于 Go + Gin 的 WireGuard 集中管理解决方案，支持多设备统一管控和远程运维。

> **v2.1 重大更新**: 采用Laravel风格的模型架构，实现了服务器端和模块端的完全分离，提升了代码的可维护性和扩展性。

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                    WireGuard 集中管理平台                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐         ┌─────────────────┐                │
│  │   管理后台       │   HTTP  │   服务器平台     │                │
│  │  (Web界面)      │◄────────┤  (VPS公网IP)    │                │
│  │   管理员        │         │  WG Server      │                │
│  └─────────────────┘         └─────────┬───────┘                │
│                                       │ VPN隧道                │
│  ┌─────────────────┐                   │                       │
│  │   运维客户端     │   WireGuard      │                       │
│  │  (WG Client)    │◄──────────────────┤                       │
│  │   运维人员      │                   │                       │
│  └─────────────────┘                   │                       │
│                               ┌────────▼─────────┐               │
│                               │                  │               │
│                               │  模块集群        │               │
│                               │                  │               │
│  ┌─────────────────┐          │ ┌──────────────┐ │               │
│  │   模块应用      │   HTTP   │ │   树莓派A     │ │               │
│  │  (Web界面)     │◄─────────┤ │  WG Client   │ │               │
│  │   本地管理      │          │ │  工厂设备    │ │               │
│  └─────────────────┘          │ └──────────────┘ │               │
│                               │ ┌──────────────┐ │               │
│                               │ │   树莓派B     │ │               │
│                               │ │  WG Client   │ │               │
│                               │ │  办公室设备   │ │               │
│                               │ └──────────────┘ │               │
│                               └──────────────────┘               │
└─────────────────────────────────────────────────────────────────┘
```

## ✨ 核心特性

### 🎛️ 服务器管理平台
- **统一仪表盘**: 实时监控所有接入设备状态
- **设备管理**: 可视化设备列表，支持搜索和筛选
- **配置自动化**: 一键生成设备和运维配置
- **访问控制**: 灵活的权限管理和用户认证
- **状态监控**: 实时连接状态和流量统计

### 📱 模块端应用  
- **简化配置**: 一键导入服务器生成的配置
- **状态显示**: 清晰的连接状态和网络信息
- **本地管理**: 基础的设备信息和网络配置
- **轻量设计**: 适配树莓派等低功耗设备

## 🚀 快速开始

### 环境要求
- Go 1.21+
- WireGuard 工具
- Linux 系统 (推荐 Ubuntu/Debian)

### 1. 服务器端部署

```bash
# 下载并编译
git clone <repo-url>
cd eitec-vpn
make build-server

# 安装 WireGuard
sudo apt update && sudo apt install wireguard

# 生成服务器密钥
wg genkey | tee server.key | wg pubkey > server.pub

# 配置并启动
sudo ./bin/eitec-vpn-server --config config/server.yaml
```

### 2. 模块端部署

```bash
# 在树莓派上编译
make build-module

# 配置并启动
sudo ./bin/eitec-vpn-module --config config/module.yaml
```

## 🏛️ 架构设计

### 分层架构
项目采用清晰的分层架构设计，实现了服务器端和模块端的完全分离：

```
┌─────────────────────────────────────────────────────────────┐
│                     架构分层图                                │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │   服务器端       │    │   模块端        │                │
│  │  (数据库驱动)    │    │  (文件驱动)     │                │
│  └─────────────────┘    └─────────────────┘                │
│         │                        │                         │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │ server/models   │    │ module/models   │                │
│  │ server/database │    │ (本地存储)      │                │
│  │ server/handlers │    │ module/handlers │                │
│  │ server/services │    │ module/services │                │
│  └─────────────────┘    └─────────────────┘                │
│         │                        │                         │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                shared/ (共享组件)                        │ │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐       │ │
│  │  │  auth   │ │response │ │ config  │ │  utils  │       │ │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘       │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 模型设计 (Laravel风格)
- **服务器端模型**: `internal/server/models/` - 每个数据库表一个文件
- **模块端模型**: `internal/module/models/` - 每个本地配置一个文件
- **单一职责**: 每个模型文件专注于一个业务实体
- **清晰命名**: 文件名直接反映模型用途

## 📁 项目结构

```
eitec-vpn/
├── cmd/                          # 应用入口
│   ├── server/                   # 服务器端主程序
│   │   └── main.go
│   └── module/                   # 模块端主程序
│       └── main.go
├── internal/                     # 内部包
│   ├── server/                   # 服务器端业务逻辑
│   │   ├── models/              # 服务器端数据模型 (Laravel风格)
│   │   │   ├── module.go        # 模块信息模型
│   │   │   ├── user.go          # 用户管理模型
│   │   │   ├── system_config.go # 系统配置模型
│   │   │   └── ...              # 其他模型文件
│   │   ├── database/            # 数据库连接和管理
│   │   │   └── database.go
│   │   ├── handlers/            # HTTP 处理器
│   │   ├── middleware/          # 中间件
│   │   ├── services/            # 业务服务
│   │   └── routes/              # 路由定义
│   ├── module/                   # 模块端业务逻辑
│   │   ├── models/              # 模块端本地模型
│   │   │   ├── local_module.go  # 本地模块配置
│   │   │   ├── local_log.go     # 本地日志记录
│   │   │   └── ...              # 其他本地模型
│   │   ├── handlers/            # 模块端处理器
│   │   ├── services/            # 模块端服务
│   │   └── routes/              # 模块端路由
│   └── shared/                   # 真正共享的组件
│       ├── auth/                # 认证服务
│       ├── response/            # 统一响应处理
│       ├── config/              # 配置管理
│       ├── wireguard/           # WireGuard 工具
│       └── utils/               # 工具函数
├── web/                          # 前端资源
│   ├── server/                   # 服务器端前端
│   │   ├── templates/           # HTML 模板
│   │   ├── static/              # 静态资源
│   │   └── assets/              # 编译后资源
│   └── module/                   # 模块端前端
│       ├── templates/
│       └── static/
├── configs/                      # 配置文件
│   ├── server.yaml              # 服务器配置
│   └── module.yaml              # 模块配置
├── scripts/                      # 部署脚本
│   ├── install-server.sh
│   ├── install-module.sh
│   └── docker/
├── docs/                         # 文档
└── Makefile                      # 构建脚本
```

## 🛠️ 技术特性

### 架构优势
- **分层清晰**: 服务器端和模块端完全分离，职责明确
- **Laravel风格**: 每个模型一个文件，符合现代开发习惯
- **响应统一**: 统一的HTTP响应格式和错误处理机制
- **类型安全**: 强类型定义，减少运行时错误
- **易于维护**: 清晰的文件组织，便于查找和修改

### 开发特性
- **热重载**: 开发模式下支持代码热重载
- **API优先**: RESTful API设计，支持前后端分离
- **插件化**: 模块化设计，便于功能扩展
- **文档完整**: 每个包都有详细的README说明
- **单元测试**: 完整的测试覆盖(规划中)

### 部署特性
- **单文件部署**: 编译后的二进制文件无外部依赖
- **跨平台**: 支持Linux、Windows、macOS
- **容器化**: 支持Docker部署和Kubernetes编排
- **配置灵活**: YAML配置文件，支持环境变量覆盖

## ⚙️ 配置说明

### 服务器配置 (configs/server.yaml)
```yaml
app:
  name: "EiTec VPN Server"
  port: 8080
  mode: "release"
  secret: "your-jwt-secret-key"
  
wireguard:
  interface: "wg0"
  port: 51820
  network: "10.10.0.0/24"
  dns: "8.8.8.8,8.8.4.4"
  
database:
  type: "sqlite"
  path: "data/eitec-vpn.db"
  
auth:
  admin_username: "admin"
  admin_password: "admin123"
  session_timeout: 24h
```

### 模块配置 (configs/module.yaml)
```yaml
app:
  name: "EiTec VPN Module"
  port: 8080
  secret: "your-jwt-secret-key"
  
module:
  name: "默认模块"
  location: "未设置"
  
wireguard:
  interface: "wg0"
```

## 🔗 API 接口

### 服务器端 API

#### 认证相关
- `POST /api/auth/login` - 管理员登录
- `POST /api/auth/logout` - 退出登录
- `GET /api/auth/profile` - 获取用户信息

#### 模块管理
- `GET /api/modules` - 获取模块列表
- `POST /api/modules` - 创建新模块
- `GET /api/modules/:id` - 获取模块详情
- `PUT /api/modules/:id` - 更新模块信息
- `DELETE /api/modules/:id` - 删除模块
- `GET /api/modules/:id/config` - 获取模块配置
- `GET /api/modules/:id/status` - 获取模块状态

#### 系统监控
- `GET /api/dashboard` - 获取仪表盘数据
- `GET /api/system/status` - 获取系统状态

### 模块端 API

#### 配置管理
- `POST /api/config/apply` - 应用配置
- `GET /api/config/current` - 获取当前配置
- `GET /api/status` - 获取运行状态

## 🎨 用户界面

### 服务器管理后台
- **现代化设计**: 基于 TailwindCSS 的响应式界面
- **实时更新**: WebSocket 实时状态更新
- **直观操作**: 拖拽式配置和一键操作

### 模块管理界面
- **简洁明了**: 专注核心功能的精简界面
- **移动友好**: 适配手机和平板设备
- **操作简单**: 一键配置和状态查看

## 🚦 状态说明

| 状态 | 图标 | 说明 |
|------|------|------|
| 在线 | 🟢 | 设备正常连接，最近握手 < 2分钟 |
| 离线 | 🔴 | 设备未连接或握手超时 > 5分钟 |
| 警告 | 🟡 | 连接不稳定，握手间隔 2-5分钟 |
| 未配置 | ⚪ | 设备已创建但未配置 |

## 📊 监控指标

- **连接状态**: 实时连接数量和状态分布
- **流量统计**: 上行/下行流量统计
- **握手监控**: 最新握手时间和频率
- **地理分布**: 设备地理位置分布(可选)

## 🔒 安全特性

- **JWT 认证**: 安全的会话管理
- **密钥管理**: 自动密钥生成和轮换
- **访问控制**: 基于角色的权限管理
- **审计日志**: 完整的操作日志记录

## 👨‍💻 开发指南

### 项目构建
```bash
# 安装依赖
go mod download

# 构建所有组件
make build

# 分别构建
make build-server  # 构建服务器端
make build-module  # 构建模块端

# 开发模式运行
make dev-server    # 开发模式运行服务器
make dev-module    # 开发模式运行模块
```

### 代码结构规范
```bash
# 添加新的服务器端模型
internal/server/models/your_model.go

# 添加新的模块端配置
internal/module/models/your_config.go

# 添加共享工具函数
internal/shared/utils/your_utils.go

# 添加API处理器
internal/server/handlers/your_handler.go
```

### 数据库迁移
```bash
# 创建迁移
# 在 internal/server/models/migrate.go 中添加新模型

# 强制重新初始化数据库
./bin/eitec-vpn-server --init

# 查看数据库状态
sqlite3 data/eitec-vpn.db ".tables"
```

### API开发
```go
// 使用统一响应格式
import "eitec-vpn/internal/shared/response"

func YourHandler(c *gin.Context) {
    // 成功响应
    response.Success(c, data)
    
    // 错误响应
    response.BadRequest(c, "参数错误")
    response.InternalError(c, "服务器错误")
}
```

## 📝 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件 