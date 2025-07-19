# EITEC VPN 模块端设计文档

## 📋 概述

EITEC VPN 模块端是独立运行在VPN节点服务器上的管理界面，与服务器端保持一致的设计风格，提供WireGuard配置、状态监控和系统管理功能。

## 🏗️ 架构设计

### 后端架构
```
internal/module/
├── handlers/         # HTTP处理器（暂未实现）
├── routes/          # 路由定义
│   └── routes.go    # API路由和HTML路由
├── services/        # 业务逻辑层
│   ├── module_service.go   # 模块配置和管理
│   └── status_service.go   # 状态监控和统计
```

### 前端架构
```
web/module/
├── static/          # 静态资源
│   ├── css/        # 样式文件（复制自server端）
│   ├── js/         # JavaScript文件
│   └── webfonts/   # 字体文件
└── templates/       # HTML模板
    ├── config.html    # 配置页面
    ├── dashboard.html # 管理界面
    └── demo.html      # 演示页面
```

## 🎨 设计理念

### 统一的视觉风格
- **深色主题**：与server端完全一致的CSS变量和配色方案
- **现代化UI**：玻璃态效果、动画过渡、响应式设计
- **模块化组件**：可复用的卡片、按钮、表单组件

### 用户体验
- **渐进式配置**：从配置页面到管理界面的平滑过渡
- **实时反馈**：状态更新、错误提示、成功确认
- **直观操作**：清晰的按钮标识、状态指示器

## 🔧 功能模块

### 1. 配置管理模块 (`config.html`)

#### 功能特性
- **动态背景**：30个浮动粒子营造科技感
- **玻璃态设计**：毛玻璃背景 + backdrop-filter
- **表单验证**：实时验证模块ID、API密钥、配置格式
- **步骤引导**：3步配置流程说明
- **动画效果**：shimmer光泽、脉冲动画

#### 技术实现
```javascript
// 配置验证
function validateConfig(configData) {
    // 验证模块ID
    if (!configData.module_id || configData.module_id <= 0) {
        showAlert('danger', '请输入有效的模块ID');
        return false;
    }
    
    // 验证WireGuard配置格式
    const config = configData.config_data;
    if (!config.includes('[Interface]') || !config.includes('[Peer]')) {
        showAlert('danger', 'WireGuard配置格式不正确');
        return false;
    }
    
    return true;
}
```

### 2. 监控管理模块 (`dashboard.html`)

#### 功能特性
- **实时监控**：WireGuard状态、系统资源、网络流量
- **控制面板**：启动/停止/重启WireGuard服务
- **网格布局**：响应式卡片布局，适配不同屏幕
- **自动刷新**：30秒间隔的数据更新

#### API接口
```javascript
// 主要数据获取
await Promise.all([
    fetch('/api/v1/status'),  // WireGuard和系统状态
    fetch('/api/v1/stats'),   // 流量统计
]);

// WireGuard控制
fetch('/api/v1/wireguard/start', { method: 'POST' });
fetch('/api/v1/wireguard/stop', { method: 'POST' });
```

## 🔌 API接口设计

### 状态接口
- `GET /api/v1/status` - 获取WireGuard和系统状态
- `POST /api/v1/status` - 更新状态（暂未使用）

### 配置接口
- `GET /api/v1/config` - 获取WireGuard配置
- `POST /api/v1/config` - 更新WireGuard配置
- `POST /api/v1/configure` - 初始化模块配置

### 统计接口
- `GET /api/v1/stats` - 获取流量统计

### 控制接口
- `POST /api/v1/wireguard/start` - 启动WireGuard
- `POST /api/v1/wireguard/stop` - 停止WireGuard
- `POST /api/v1/wireguard/restart` - 重启WireGuard

### HTML路由
- `GET /` - 根据配置状态显示dashboard或config页面
- `GET /config` - 强制显示配置页面

## 📱 响应式设计

### 桌面端 (>768px)
```css
.main-content {
    display: grid;
    grid-template-columns: repeat(12, 1fr);
    grid-gap: 1.2rem;
}
.card { grid-column: span 6; }
```

### 移动端 (≤768px)
```css
.main-content {
    grid-template-columns: 1fr;
}
.card { grid-column: span 1; }
```

## 🎯 数据流设计

### 配置流程
```
用户输入 → 前端验证 → API提交 → 后端处理 → 文件写入 → 重定向到管理界面
```

### 监控流程
```
页面加载 → 并行API请求 → 数据解析 → UI更新 → 定时刷新(30s)
```

### 控制流程
```
用户操作 → API调用 → WireGuard命令 → 状态反馈 → 延迟刷新(1s)
```

## 🔒 安全考虑

### 输入验证
- API密钥长度检查（≥10字符）
- 服务器URL格式验证
- WireGuard配置格式检查

### 权限控制
- 配置文件权限：0600（仅所有者读写）
- WireGuard操作需要适当的系统权限

## 🚀 部署要求

### 系统依赖
- WireGuard工具：`wg`, `wg-quick`
- 系统权限：能够执行WireGuard命令
- 网络访问：能够连接到服务器API

### 文件权限
```bash
/etc/wireguard/wg0.conf     # 0600
/etc/eitec-vpn/module.info  # 0600
```

## 📊 性能优化

### 前端优化
- **并行API请求**：同时获取多个数据源
- **本地资源**：避免CDN依赖，提高加载速度
- **CSS优化**：使用CSS变量，减少重复代码

### 后端优化
- **缓存机制**：适当缓存系统状态信息
- **异步处理**：非阻塞的WireGuard操作

## 🔧 开发指南

### 本地开发
```bash
# 编译模块
go build ./cmd/module/

# 运行模块（需要配置文件）
./module

# 访问界面
http://localhost:8081
```

### 添加新功能
1. 在 `internal/module/services/` 添加业务逻辑
2. 在 `internal/module/routes/` 添加API路由
3. 在 `web/module/templates/` 添加前端界面
4. 更新API文档和测试

## 📈 未来扩展

### 功能扩展
- [ ] 配置备份和恢复
- [ ] 日志查看界面
- [ ] 性能监控图表
- [ ] 自动配置更新
- [ ] 多语言支持

### 技术优化
- [ ] WebSocket实时通信
- [ ] PWA离线支持
- [ ] 性能监控指标
- [ ] 自动化测试覆盖

## 📝 总结

EITEC VPN 模块端通过统一的设计语言和现代化的技术栈，提供了完整的VPN节点管理解决方案。设计重点关注用户体验、系统稳定性和视觉一致性，为运维人员提供了便捷的管理工具。 