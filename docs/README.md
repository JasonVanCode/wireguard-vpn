# EITEC VPN 文档中心

欢迎使用 EITEC VPN 文档中心！这里包含了完整的 WireGuard 内网穿透解决方案的配置和部署指南。

## 📚 文档目录

### 🚀 快速开始
- **[快速部署指南](DEPLOYMENT_GUIDE.md)** - 30分钟内完成完整部署
- **[WireGuard配置指南](wireguard-config.md)** - 详细的配置说明和最佳实践

### 🔧 配置文件
- **[服务端配置模板](configs/server-wg0.conf)** - WireGuard服务端配置模板
- **[模块端配置模板](configs/module-wg0.conf)** - 工厂模块端配置模板  
- **[用户端配置模板](configs/user-client.conf)** - 客户端配置模板

### 🛠️ 部署脚本
- **[服务端部署脚本](scripts/setup-server.sh)** - 自动化服务端部署
- **[模块端部署脚本](scripts/setup-module.sh)** - 自动化模块端部署

### 📊 网络架构
- **[网络拓扑详解](network-topology.md)** - 完整的网络架构说明

## 🎯 解决方案概述

EITEC VPN 是一个基于 WireGuard 的内网穿透解决方案，主要用于：

- **工厂远程访问**: 外部用户安全访问工厂内网设备
- **设备监控**: 远程监控和管理工厂设备
- **数据采集**: 安全的数据传输通道
- **技术支持**: 远程技术支持和故障排除

## 🏗️ 系统架构

```
[外部用户] ←→ [VPN服务器] ←→ [工厂模块] ←→ [工厂内网]
```

### 核心组件

1. **VPN服务器** - 部署在公网的WireGuard服务器
2. **工厂模块** - 部署在工厂内网的网关设备
3. **管理系统** - Web界面管理和监控
4. **用户客户端** - 各平台的VPN客户端

## 🚀 快速部署

### 第一步: 服务端部署
```bash
wget https://raw.githubusercontent.com/your-repo/eitec-vpn/main/docs/scripts/setup-server.sh
chmod +x setup-server.sh
sudo ./setup-server.sh
```

### 第二步: 模块端部署
```bash
wget https://raw.githubusercontent.com/your-repo/eitec-vpn/main/docs/scripts/setup-module.sh
chmod +x setup-module.sh
sudo ./setup-module.sh "服务端公钥" "服务端IP" "内网网段"
```

### 第三步: 配置用户
参考 [快速部署指南](DEPLOYMENT_GUIDE.md) 中的用户配置章节。

## 📋 系统要求

### 服务端要求
- Linux服务器 (Ubuntu 20.04+, Debian 11+, CentOS 8+)
- 公网IP地址
- Root权限
- 开放UDP 51820端口

### 模块端要求
- Linux设备 (支持WireGuard)
- 部署在目标内网中
- Root权限
- 能访问互联网

### 客户端支持
- Windows 10/11
- macOS 10.14+
- Linux (各发行版)
- Android 5.0+
- iOS 12.0+

## 🔒 安全特性

- **现代加密**: ChaCha20-Poly1305加密算法
- **密钥管理**: 自动密钥轮换和管理
- **访问控制**: 基于IP的精细访问控制
- **审计日志**: 完整的连接和访问日志
- **网络隔离**: VPN网段与内网分离

## 📊 网络规划

| 网段 | 用途 | 示例 |
|------|------|------|
| 10.8.0.0/24 | VPN内部网段 | 10.8.0.1-254 |
| 192.168.1.0/24 | 工厂内网 | 192.168.1.1-254 |
| 10.8.0.1 | VPN服务器 | 固定分配 |
| 10.8.0.2 | 工厂模块 | 固定分配 |
| 10.8.0.10+ | 外部用户 | 动态分配 |

## 🛠️ 管理工具

### Web管理界面
- **服务端管理**: `http://server-ip:8080`
- **模块端管理**: `http://module-ip:8081`

### 命令行工具
```bash
# 查看连接状态
wg show

# 重启服务
systemctl restart wg-quick@wg0

# 查看日志
journalctl -u wg-quick@wg0 -f
```

## 📞 技术支持

### 常见问题
1. **连接问题** - 检查防火墙和网络配置
2. **性能问题** - 优化MTU和内核参数
3. **安全问题** - 检查密钥和访问控制

### 故障排除
- 查看 [WireGuard配置指南](wireguard-config.md) 的故障排除章节
- 使用提供的测试脚本进行诊断
- 检查系统日志和WireGuard状态

### 获取帮助
- 查看详细文档
- 运行诊断脚本
- 检查日志文件
- 联系技术支持

## 📈 性能优化

### 网络优化
- MTU调优
- 内核参数优化
- 硬件加速支持

### 监控指标
- 连接数量
- 流量统计
- 延迟监控
- 错误率统计

## 🔄 维护和更新

### 定期维护
- 密钥轮换
- 日志清理
- 性能监控
- 安全审计

### 系统更新
- WireGuard版本更新
- 操作系统补丁
- 配置优化
- 功能增强

---

## 📖 相关链接

- [WireGuard官方文档](https://www.wireguard.com/)
- [项目GitHub仓库](https://github.com/your-repo/eitec-vpn)
- [技术支持](mailto:support@example.com)

**最后更新**: 2024年1月

**版本**: v2.1 