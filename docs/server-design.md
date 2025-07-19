# EITEC VPN 服务端设计文档

## 📋 概述

EITEC VPN 服务端是整个VPN管理系统的核心控制节点，提供统一的管理界面、API接口和数据存储服务。采用现代化的深色主题设计，实现多模块管理、用户权限控制、系统监控和配置管理等核心功能。

## 🏗️ 架构设计

### 后端架构
```
internal/server/
├── handlers/              # HTTP处理器层
│   ├── auth_handler.go        # 认证处理器
│   ├── config_handler.go      # 配置管理处理器
│   ├── dashboard_handler.go   # 仪表盘处理器
│   ├── module_handler.go      # 模块管理处理器
│   └── user_handler.go        # 用户管理处理器
├── middleware/            # 中间件层
│   ├── auth.go               # 认证中间件
│   └── security.go           # 安全中间件
├── routes/                # 路由定义
│   └── routes.go             # 主路由配置
└── services/              # 业务逻辑层
    ├── config_service.go     # 配置服务
    ├── dashboard_service.go  # 仪表盘服务
    └── module_service.go     # 模块服务
```

### 前端架构
```
web/server/
├── assets/                # 自定义资源
│   ├── css/style.css         # 自定义样式
│   └── js/main.js            # 自定义脚本
├── static/                # 第三方静态资源
│   ├── css/                  # Bootstrap、FontAwesome
│   ├── js/                   # Bootstrap、ECharts
│   └── webfonts/             # 字体文件
└── templates/             # HTML模板
    ├── layout/base.html      # 基础布局模板
    ├── index.html            # 主仪表盘
    ├── login.html            # 登录页面
    ├── dashboard.html        # 数据分析页面
    ├── modules.html          # 模块管理页面
    ├── users.html            # 用户管理页面
    ├── config.html           # 系统配置页面
    └── 404.html              # 错误页面
```

### 共享模块架构
```
internal/shared/
├── auth/                  # 认证授权模块
│   ├── jwt.go                # JWT令牌服务
│   ├── session.go            # 会话管理
│   └── user_service.go       # 用户服务
├── config/                # 配置模块
│   └── config.go             # 配置管理
├── database/              # 数据库模块
│   └── database.go           # 数据库连接和操作
├── models/                # 数据模型
│   └── models.go             # 数据结构定义
├── utils/                 # 工具库
│   └── utils.go              # 通用工具函数
└── wireguard/             # WireGuard模块
    └── wireguard.go          # WireGuard配置生成
```

## 🎨 设计理念

### 视觉设计系统
- **深色主题**：专业级的深色配色方案，减少眼部疲劳
- **现代化UI**：卡片式布局、圆角设计、渐变效果
- **一致性**：统一的CSS变量系统，确保整个系统视觉一致
- **响应式**：适配桌面、平板、手机等多种设备

### 用户体验设计
- **直观导航**：清晰的面包屑导航和页面结构
- **实时反馈**：操作状态提示、加载动画、错误提示
- **高效操作**：批量操作、快捷键支持、智能搜索
- **个性化**：用户偏好设置、自定义仪表盘

### 安全设计原则
- **多重认证**：JWT + Session双重认证机制
- **权限控制**：基于角色的访问控制(RBAC)
- **数据保护**：敏感数据加密存储
- **审计日志**：完整的操作日志记录

## 🔧 核心功能模块

### 1. 用户认证与授权

#### JWT认证系统
```go
type JWTClaims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}
```

#### 权限管理
- **Super Admin**：超级管理员，完全权限
- **Admin**：管理员，大部分管理权限
- **Guest**：访客，只读权限

#### 会话管理
- 分布式会话存储
- 自动过期机制
- 并发登录控制

### 2. 模块管理系统

#### 模块生命周期
```
创建模块 → 生成配置 → 分发部署 → 状态监控 → 流量统计 → 维护管理
```

#### 功能特性
- **动态配置生成**：自动为每个模块生成WireGuard配置
- **状态实时监控**：在线/离线/警告状态跟踪
- **流量统计分析**：上传下载流量、连接时长统计
- **批量操作支持**：批量创建、删除、状态更新
- **地理位置管理**：模块位置信息和地图展示

#### API接口设计
```go
// 模块管理接口
GET    /api/v1/modules              # 获取模块列表
POST   /api/v1/modules              # 创建新模块
GET    /api/v1/modules/:id          # 获取模块详情
PUT    /api/v1/modules/:id          # 更新模块信息
DELETE /api/v1/modules/:id          # 删除模块
GET    /api/v1/modules/:id/config   # 获取模块配置
GET    /api/v1/modules/:id/logs     # 获取模块日志
```

### 3. 仪表盘系统

#### 数据可视化
- **实时统计**：模块数量、在线率、流量统计
- **图表展示**：ECharts图表库，流量趋势、状态分布
- **地理分布**：模块地理位置可视化
- **性能监控**：系统资源使用情况

#### 核心指标
```javascript
// 仪表盘统计数据结构
{
    "module_stats": {
        "total": 16,
        "online": 9,
        "offline": 3,
        "warning": 2,
        "unconfigured": 2,
        "online_rate": 56.25
    },
    "traffic_stats": {
        "total_rx": 1073741824,
        "total_tx": 536870912,
        "today_total": 268435456
    },
    "system_stats": {
        "uptime": "7天 12小时",
        "user_count": 5,
        "server_status": "running"
    }
}
```

### 4. 系统配置管理

#### 配置分类
- **服务器配置**：端点地址、端口、HTTPS设置
- **WireGuard配置**：网络段、DNS、MTU等参数
- **安全配置**：JWT密钥、会话超时、访问控制
- **网络配置**：IP池管理、NAT设置、防火墙规则

#### 配置验证
```go
func (cs *ConfigService) ValidateNetworkSettings(network, ipStart, ipEnd string) error {
    // 网络段格式验证
    _, ipnet, err := net.ParseCIDR(network)
    if err != nil {
        return fmt.Errorf("网络段格式无效: %v", err)
    }
    
    // IP范围验证
    startIP := net.ParseIP(ipStart)
    endIP := net.ParseIP(ipEnd)
    
    if !ipnet.Contains(startIP) || !ipnet.Contains(endIP) {
        return fmt.Errorf("IP范围不在指定网络段内")
    }
    
    return nil
}
```

### 5. 用户管理系统

#### 用户数据模型
```go
type User struct {
    ID           uint      `json:"id"`
    Username     string    `json:"username"`
    Password     string    `json:"-"`
    Role         UserRole  `json:"role"`
    IsActive     bool      `json:"is_active"`
    LastLoginAt  *time.Time `json:"last_login_at"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

#### 密码安全
- **Bcrypt加密**：密码哈希存储
- **强度要求**：最小长度、复杂度检查
- **重置机制**：安全的密码重置流程

## 🎯 前端界面设计

### 1. 主仪表盘 (`index.html`)

#### 布局结构
```css
.main-content {
    display: grid;
    grid-template-columns: repeat(12, 1fr);
    grid-template-rows: auto auto auto auto;
    grid-gap: 1.2rem;
}

/* 响应式网格 */
.stat-card { grid-column: span 3; }        /* 统计卡片 */
.module-card { grid-column: span 8; grid-row: span 2; }  /* 模块管理 */
.system-card { grid-column: span 4; }      /* 系统信息 */
```

#### 核心组件
- **状态卡片**：模块统计、流量统计、系统状态、警告信息
- **数据图表**：ECharts流量趋势图、状态分布饼图
- **模块列表**：实时更新的模块状态表格
- **系统监控**：CPU、内存、磁盘使用率

### 2. 登录页面 (`login.html`)

#### 视觉特效
- **动态背景**：50个浮动粒子营造科技感
- **玻璃态设计**：毛玻璃背景 + 背景模糊
- **光泽动画**：shimmer效果增强视觉吸引力
- **微动效果**：输入框聚焦上移、按钮光波扫过

#### 安全特性
- **表单验证**：前端实时验证 + 后端安全检查
- **错误提示**：震动动画 + 详细错误信息
- **防暴力破解**：登录尝试次数限制

### 3. 模块管理页面 (`modules.html`)

#### 功能特性
- **高级搜索**：按名称、位置、状态筛选
- **批量操作**：多选删除、状态批量更新
- **实时状态**：WebSocket实时状态更新
- **配置下载**：一键生成并下载WireGuard配置

#### 表格设计
```javascript
// 模块状态渲染
function renderModuleStatus(status) {
    const statusConfig = {
        'online': { color: '#10b981', text: '在线', icon: 'fas fa-circle' },
        'offline': { color: '#ef4444', text: '离线', icon: 'fas fa-times-circle' },
        'warning': { color: '#f59e0b', text: '警告', icon: 'fas fa-exclamation-triangle' }
    };
    
    const config = statusConfig[status] || statusConfig.offline;
    return `<span style="color: ${config.color}">
        <i class="${config.icon}"></i> ${config.text}
    </span>`;
}
```

## 🔌 API接口架构

### RESTful API设计

#### 统一响应格式
```json
{
    "code": 200,
    "message": "success",
    "data": {},
    "timestamp": "2024-01-15T10:30:00Z"
}
```

#### 错误处理
```json
{
    "code": 400,
    "message": "请求参数无效",
    "errors": {
        "username": "用户名不能为空",
        "password": "密码长度至少6位"
    }
}
```

### 认证机制

#### JWT认证流程
```
1. 用户登录 → 验证凭据 → 生成JWT + RefreshToken
2. API请求 → 携带JWT → 验证有效性 → 处理请求
3. JWT过期 → 使用RefreshToken → 刷新JWT → 继续请求
```

#### API密钥认证
```go
// API密钥中间件
func APIKeyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("X-API-Key")
        if apiKey == "" {
            c.JSON(401, gin.H{"error": "API密钥缺失"})
            c.Abort()
            return
        }
        
        // 验证API密钥
        if !validateAPIKey(apiKey) {
            c.JSON(401, gin.H{"error": "API密钥无效"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

## 🛡️ 安全架构

### 认证安全
- **密码策略**：强密码要求、定期更新提醒
- **多因素认证**：支持TOTP、短信验证
- **会话安全**：安全的会话ID生成、自动过期

### 数据安全
- **传输加密**：HTTPS强制加密传输
- **存储加密**：敏感数据AES加密存储
- **SQL注入防护**：GORM ORM防止SQL注入

### 访问控制
```go
// 基于角色的访问控制
func RequireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("user_role")
        if !hasPermission(userRole, role) {
            c.JSON(403, gin.H{"error": "权限不足"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## 📊 性能优化

### 前端优化
- **资源压缩**：CSS/JS文件压缩、图片优化
- **懒加载**：图表和表格数据按需加载
- **缓存策略**：静态资源长缓存、API数据短缓存
- **CDN加速**：静态资源CDN分发

### 后端优化
- **数据库优化**：索引优化、查询性能调优
- **缓存机制**：Redis缓存热点数据
- **连接池**：数据库连接池、HTTP连接复用
- **异步处理**：后台任务队列处理耗时操作

### 监控指标
```go
// 性能监控中间件
func PerformanceMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        latency := time.Since(start)
        // 记录请求延迟、状态码等指标
        metrics.RecordHTTPRequest(
            c.Request.Method,
            c.Request.URL.Path,
            c.Writer.Status(),
            latency,
        )
    }
}
```

## 🚀 部署架构

### 容器化部署
```dockerfile
# Dockerfile示例
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server/

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/web ./web
CMD ["./server"]
```

### 环境配置
```yaml
# docker-compose.yml
version: '3.8'
services:
  eitec-vpn-server:
    image: eitec-vpn-server:latest
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis
  
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: eitec_vpn
      POSTGRES_USER: eitec
      POSTGRES_PASSWORD: secure_password
  
  redis:
    image: redis:7-alpine
```

### 负载均衡
```nginx
# Nginx配置示例
upstream eitec_vpn_backend {
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
}

server {
    listen 443 ssl;
    server_name vpn.example.com;
    
    ssl_certificate /etc/ssl/certs/vpn.crt;
    ssl_certificate_key /etc/ssl/private/vpn.key;
    
    location / {
        proxy_pass http://eitec_vpn_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 🔧 开发工具链

### 开发环境
```bash
# 本地开发启动
go run ./cmd/server/

# 热重载开发（需要air工具）
air

# 数据库迁移
go run scripts/migrate.go

# 创建测试数据
go run scripts/seed.go
```

### 测试策略
```go
// 单元测试示例
func TestCreateModule(t *testing.T) {
    // 设置测试数据库
    db := setupTestDB()
    defer cleanupTestDB(db)
    
    service := NewModuleService(db)
    
    // 测试创建模块
    module, err := service.CreateModule("Test Module", "Beijing")
    assert.NoError(t, err)
    assert.Equal(t, "Test Module", module.Name)
    assert.Equal(t, "Beijing", module.Location)
}
```

### CI/CD流程
```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: 1.21
      - run: go test ./...
      - run: go build ./cmd/server/
  
  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - name: Deploy to production
        run: echo "Deploying to production..."
```

## 📈 监控与日志

### 应用监控
```go
// 健康检查端点
func HealthCheck(c *gin.Context) {
    status := gin.H{
        "status": "ok",
        "timestamp": time.Now().Unix(),
        "version": "2.0.0",
        "database": checkDatabase(),
        "redis": checkRedis(),
        "wireguard": checkWireGuard(),
    }
    
    c.JSON(200, status)
}
```

### 日志系统
```go
// 结构化日志
log := logrus.WithFields(logrus.Fields{
    "user_id": userID,
    "action": "create_module",
    "module_name": moduleName,
    "ip_address": clientIP,
})

log.Info("模块创建成功")
```

### 指标收集
- **业务指标**：模块数量、用户活跃度、流量统计
- **系统指标**：CPU、内存、磁盘、网络使用率
- **应用指标**：请求量、响应时间、错误率

## 🔮 未来扩展计划

### 功能扩展
- [ ] **多租户支持**：企业级多租户架构
- [ ] **插件系统**：可扩展的插件架构
- [ ] **API网关**：统一API管理和限流
- [ ] **配置中心**：动态配置管理
- [ ] **消息队列**：异步任务处理
- [ ] **数据分析**：高级数据分析和报表

### 技术升级
- [ ] **微服务架构**：服务拆分和治理
- [ ] **Kubernetes部署**：云原生部署方案
- [ ] **GraphQL API**：更灵活的API查询
- [ ] **WebAssembly**：前端性能优化
- [ ] **机器学习**：智能运维和预测分析

### 安全增强
- [ ] **零信任架构**：端到端安全验证
- [ ] **区块链审计**：不可篡改的审计日志
- [ ] **量子加密**：量子安全的加密算法
- [ ] **生物识别**：多模态生物识别认证

## 📝 总结

EITEC VPN 服务端通过现代化的技术栈和精心设计的架构，提供了完整的VPN管理解决方案。系统采用分层架构设计，确保了代码的可维护性和扩展性；深色主题的现代化界面提供了优秀的用户体验；完善的安全机制保障了系统和数据的安全性；丰富的API接口支持了多样化的集成需求。

整个系统设计遵循了软件工程的最佳实践，包括：
- **模块化设计**：清晰的模块划分和接口定义
- **安全优先**：多层次的安全防护机制
- **性能导向**：多种性能优化策略
- **可观测性**：完善的监控和日志系统
- **可扩展性**：为未来扩展预留的接口和架构

服务端作为整个EITEC VPN系统的控制中心，为管理员提供了强大而易用的管理工具，为VPN网络的稳定运行提供了可靠的技术保障。 