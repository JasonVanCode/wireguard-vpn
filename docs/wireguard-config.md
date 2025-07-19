# WireGuard 内网穿透配置指南

## 概述

本文档提供了完整的WireGuard VPN配置方案，实现工厂内网穿透功能。通过这个配置，外部用户可以通过VPN安全地访问工厂内网资源。

## 网络拓扑

```
[外部用户] ←→ [公网] ←→ [VPN服务器] ←→ [VPN隧道] ←→ [工厂模块] ←→ [工厂内网]
    |                     |                              |              |
10.0.0.0/24          公网IP                        10.8.0.2      192.168.1.0/24
                   10.8.0.1                                    (工厂局域网)
```

## IP地址规划

| 组件 | 网段/IP | 说明 |
|------|---------|------|
| VPN网段 | 10.8.0.0/24 | WireGuard VPN内部网段 |
| 服务端VPN IP | 10.8.0.1 | VPN服务器在VPN网段的IP |
| 模块端VPN IP | 10.8.0.2 | 工厂模块在VPN网段的IP |
| 工厂内网 | 192.168.1.0/24 | 工厂局域网网段 |
| 外部用户VPN IP | 10.8.0.10+ | 外部用户VPN IP范围 |

## 服务端配置

### 1. 服务端WireGuard配置文件

文件路径: `/etc/wireguard/wg0.conf`

```ini
[Interface]
# 服务端私钥 (使用 wg genkey 生成)
PrivateKey = SERVER_PRIVATE_KEY_HERE
# 服务端在VPN网段的IP
Address = 10.8.0.1/24
# WireGuard监听端口
ListenPort = 51820
# 数据包转发
PostUp = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

# 工厂模块端配置
[Peer]
# 模块端公钥
PublicKey = MODULE_PUBLIC_KEY_HERE
# 模块端允许的IP范围 (VPN IP + 工厂内网)
AllowedIPs = 10.8.0.2/32, 192.168.1.0/24
# 保持连接
PersistentKeepalive = 25

# 外部用户1配置
[Peer]
PublicKey = USER1_PUBLIC_KEY_HERE
AllowedIPs = 10.8.0.10/32

# 外部用户2配置
[Peer]
PublicKey = USER2_PUBLIC_KEY_HERE
AllowedIPs = 10.8.0.11/32
```

### 2. 服务端系统配置

```bash
# 启用IP转发
echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.conf
sysctl -p

# 配置防火墙
ufw allow 51820/udp
ufw allow ssh
ufw --force enable

# 启动WireGuard
systemctl enable wg-quick@wg0
systemctl start wg-quick@wg0
```

## 模块端配置

### 1. 模块端WireGuard配置文件

文件路径: `/etc/wireguard/wg0.conf`

```ini
[Interface]
# 模块端私钥
PrivateKey = MODULE_PRIVATE_KEY_HERE
# 模块端在VPN网段的IP
Address = 10.8.0.2/32
# DNS服务器 (可选)
DNS = 8.8.8.8, 8.8.4.4

# 启动后配置路由和NAT
PostUp = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -A FORWARD -o wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -D FORWARD -o wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

[Peer]
# 服务端公钥
PublicKey = SERVER_PUBLIC_KEY_HERE
# 服务端地址和端口
Endpoint = YOUR_SERVER_PUBLIC_IP:51820
# 允许的IP范围 (VPN网段)
AllowedIPs = 10.8.0.0/24
# 保持连接
PersistentKeepalive = 25
```

### 2. 模块端系统配置

```bash
# 启用IP转发
echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.conf
sysctl -p

# 配置防火墙 (允许内网访问)
ufw allow from 192.168.1.0/24
ufw allow from 10.8.0.0/24
ufw --force enable

# 启动WireGuard
systemctl enable wg-quick@wg0
systemctl start wg-quick@wg0
```

## 外部用户配置

### 用户端配置文件示例

```ini
[Interface]
# 用户私钥
PrivateKey = USER_PRIVATE_KEY_HERE
# 用户在VPN网段的IP
Address = 10.8.0.10/32
# DNS服务器
DNS = 8.8.8.8

[Peer]
# 服务端公钥
PublicKey = SERVER_PUBLIC_KEY_HERE
# 服务端地址
Endpoint = YOUR_SERVER_PUBLIC_IP:51820
# 允许访问的网段 (VPN网段 + 工厂内网)
AllowedIPs = 10.8.0.0/24, 192.168.1.0/24
# 保持连接
PersistentKeepalive = 25
```

## 密钥生成

### 生成密钥对

```bash
# 生成私钥
wg genkey | tee private.key | wg pubkey > public.key

# 或者一次性生成
wg genkey | tee server_private.key | wg pubkey > server_public.key
wg genkey | tee module_private.key | wg pubkey > module_public.key
wg genkey | tee user1_private.key | wg pubkey > user1_public.key
```

## 配置验证

### 1. 检查连接状态

```bash
# 服务端检查
wg show

# 模块端检查  
wg show

# 检查路由
ip route show
```

### 2. 连通性测试

```bash
# 从外部用户测试
ping 10.8.0.1  # 测试到VPN服务器
ping 10.8.0.2  # 测试到工厂模块
ping 192.168.1.1  # 测试到工厂内网网关
```

## 故障排除

### 常见问题

1. **连接不上VPN**
   - 检查防火墙是否开放51820端口
   - 确认服务端公网IP正确
   - 检查密钥是否匹配

2. **能连VPN但无法访问内网**
   - 检查IP转发是否启用
   - 确认iptables规则正确
   - 检查AllowedIPs配置

3. **连接不稳定**
   - 调整PersistentKeepalive值
   - 检查网络MTU设置
   - 确认NAT类型

### 调试命令

```bash
# 查看WireGuard日志
journalctl -u wg-quick@wg0 -f

# 查看网络连接
ss -tulpn | grep 51820

# 查看路由表
ip route show table all

# 查看iptables规则
iptables -L -v -n
iptables -t nat -L -v -n
```

## 安全建议

1. **定期更换密钥**
2. **限制AllowedIPs范围**
3. **使用强密码保护配置文件**
4. **启用日志监控**
5. **定期更新WireGuard版本**

## 性能优化

1. **调整MTU大小**
2. **优化内核参数**
3. **使用SSD存储**
4. **监控带宽使用** 