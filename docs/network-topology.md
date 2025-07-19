# 网络拓扑详细说明

## 整体架构图

```
                    Internet (公网)
                         |
                    ┌─────────┐
                    │ Router  │ (NAT/Firewall)
                    │ Gateway │
                    └─────────┘
                         |
        ┌────────────────┼────────────────┐
        │                                 │
   ┌─────────┐                      ┌─────────┐
   │VPN Server│                     │External │
   │         │                      │ Users   │
   │10.8.0.1 │                      │         │
   │Port:51820│                      │10.8.0.10+│
   └─────────┘                      └─────────┘
        │                                 │
        │         VPN Tunnel              │
        │    (WireGuard Encrypted)        │
        └─────────────┬───────────────────┘
                      │
              ┌───────────────┐
              │ Factory Module│
              │   (Gateway)   │
              │   10.8.0.2    │
              └───────────────┘
                      │
              ┌───────────────┐
              │Factory Network│
              │192.168.1.0/24 │
              └───────────────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
   ┌─────────┐   ┌─────────┐   ┌─────────┐
   │ Device1 │   │ Device2 │   │ Device3 │
   │192.168. │   │192.168. │   │192.168. │
   │  1.100  │   │  1.101  │   │  1.102  │
   └─────────┘   └─────────┘   └─────────┘
```

## 网络流量路径

### 外部用户访问工厂设备流程

1. **用户发起连接**
   ```
   User PC (10.8.0.10) → VPN Server (10.8.0.1)
   ```

2. **VPN服务器转发**
   ```
   VPN Server (10.8.0.1) → Factory Module (10.8.0.2)
   ```

3. **工厂模块转发到内网**
   ```
   Factory Module (10.8.0.2) → Factory Device (192.168.1.100)
   ```

4. **完整路径**
   ```
   User PC → Internet → VPN Server → VPN Tunnel → Factory Module → Factory LAN → Target Device
   ```

## IP地址分配表

| 组件类型 | IP地址/网段 | 说明 | 配置位置 |
|---------|------------|------|----------|
| VPN网段 | 10.8.0.0/24 | WireGuard内部网络 | 所有配置文件 |
| VPN服务器 | 10.8.0.1 | VPN网关服务器 | 服务端配置 |
| 工厂模块 | 10.8.0.2 | 工厂网关模块 | 模块端配置 |
| 外部用户1 | 10.8.0.10 | 第一个用户 | 用户配置文件 |
| 外部用户2 | 10.8.0.11 | 第二个用户 | 用户配置文件 |
| 管理员 | 10.8.0.20 | 管理员用户 | 管理员配置 |
| 工厂内网 | 192.168.1.0/24 | 工厂局域网 | 模块端路由 |
| 工厂网关 | 192.168.1.1 | 工厂路由器 | 内网设备 |
| 工厂设备 | 192.168.1.100+ | 工厂内部设备 | 内网设备 |

## 端口配置

| 服务 | 端口 | 协议 | 说明 |
|------|------|------|------|
| WireGuard | 51820 | UDP | VPN隧道通信 |
| SSH | 22 | TCP | 远程管理 |
| HTTP | 80 | TCP | Web服务 |
| HTTPS | 443 | TCP | 安全Web服务 |
| EITEC Server | 8080 | TCP | 服务端管理界面 |
| EITEC Module | 8081 | TCP | 模块端管理界面 |

## 防火墙规则

### 服务端防火墙规则
```bash
# 允许WireGuard端口
ufw allow 51820/udp

# 允许SSH管理
ufw allow ssh

# 允许Web管理界面
ufw allow 8080/tcp

# 启用防火墙
ufw --force enable
```

### 模块端防火墙规则
```bash
# 允许内网访问
ufw allow from 192.168.1.0/24

# 允许VPN网段访问
ufw allow from 10.8.0.0/24

# 允许SSH管理
ufw allow ssh

# 允许Web管理界面
ufw allow 8081/tcp

# 启用防火墙
ufw --force enable
```

## 路由配置

### 服务端路由
- 默认路由：通过公网接口
- VPN路由：10.8.0.0/24 通过 wg0 接口
- 工厂网络路由：192.168.1.0/24 通过 10.8.0.2 (工厂模块)

### 模块端路由
- 默认路由：通过工厂网关 (192.168.1.1)
- VPN路由：10.8.0.0/24 通过 wg0 接口
- 内网路由：192.168.1.0/24 通过本地接口

### 用户端路由
- VPN路由：10.8.0.0/24 通过 wg0 接口
- 工厂网络路由：192.168.1.0/24 通过 wg0 接口
- 其他流量：保持原有路由 (分离隧道)

## NAT配置

### 服务端NAT
```bash
# 转发VPN流量到公网
iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
iptables -A FORWARD -i wg0 -j ACCEPT
```

### 模块端NAT
```bash
# 转发VPN流量到内网
iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
iptables -A FORWARD -i wg0 -j ACCEPT
iptables -A FORWARD -o wg0 -j ACCEPT
```

## 安全考虑

### 1. 网络隔离
- VPN网段与内网分离
- 用户只能访问授权的内网资源
- 防火墙限制不必要的端口

### 2. 访问控制
- 基于IP的访问控制
- 用户特定的AllowedIPs配置
- 定期密钥轮换

### 3. 监控和日志
- WireGuard连接日志
- 流量统计监控
- 异常连接告警

## 性能优化

### 1. MTU优化
```bash
# 设置合适的MTU值
ip link set dev wg0 mtu 1420
```

### 2. 内核参数优化
```bash
# 优化网络性能
echo 'net.core.rmem_max = 26214400' >> /etc/sysctl.conf
echo 'net.core.rmem_default = 26214400' >> /etc/sysctl.conf
echo 'net.core.wmem_max = 26214400' >> /etc/sysctl.conf
echo 'net.core.wmem_default = 26214400' >> /etc/sysctl.conf
sysctl -p
```

### 3. CPU优化
- 使用多核CPU处理加密
- 启用硬件加速 (如果支持)

## 故障排除流程

### 1. 连接问题
```bash
# 检查WireGuard状态
wg show

# 检查路由表
ip route show

# 测试连通性
ping 10.8.0.1
ping 192.168.1.1
```

### 2. 性能问题
```bash
# 检查流量统计
wg show all dump

# 检查网络接口状态
ip link show

# 检查系统负载
top
iotop
```

### 3. 安全问题
```bash
# 检查防火墙状态
ufw status verbose

# 检查连接日志
journalctl -u wg-quick@wg0

# 检查系统日志
tail -f /var/log/syslog
``` 