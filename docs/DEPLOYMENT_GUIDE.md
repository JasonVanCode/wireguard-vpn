# WireGuard 内网穿透快速部署指南

## 🚀 快速开始

这个指南将帮助你在30分钟内完成WireGuard VPN的部署，实现工厂内网穿透功能。

## 📋 准备工作

### 服务端要求
- 具有公网IP的Linux服务器 (Ubuntu 20.04+, Debian 11+, CentOS 8+)
- Root权限
- 开放UDP 51820端口

### 模块端要求  
- 部署在工厂内网的Linux设备
- Root权限
- 能够访问互联网

### 网络规划
- VPN网段: `10.8.0.0/24`
- 服务端VPN IP: `10.8.0.1`
- 模块端VPN IP: `10.8.0.2`
- 工厂内网: `192.168.1.0/24` (根据实际情况调整)
- 用户VPN IP: `10.8.0.10+`

## 🔧 部署步骤

### 第一步: 部署服务端

1. **登录服务端**
   ```bash
   ssh root@your-server-ip
   ```

2. **下载并运行部署脚本**
   ```bash
   wget https://raw.githubusercontent.com/your-repo/eitec-vpn/main/docs/scripts/setup-server.sh
   chmod +x setup-server.sh
   ./setup-server.sh
   ```

3. **记录服务端信息**
   脚本执行完成后，记录显示的服务端公钥，例如:
   ```
   服务端公钥: AbCdEf1234567890...
   ```

### 第二步: 部署模块端

1. **登录模块端设备**
   ```bash
   ssh root@factory-module-ip
   ```

2. **下载并运行部署脚本**
   ```bash
   wget https://raw.githubusercontent.com/your-repo/eitec-vpn/main/docs/scripts/setup-module.sh
   chmod +x setup-module.sh
   
   # 替换参数为实际值
   ./setup-module.sh "服务端公钥" "服务端公网IP" "192.168.1.0/24"
   ```

3. **记录模块端公钥**
   脚本执行完成后，记录显示的模块端公钥。

### 第三步: 配置服务端Peer

1. **编辑服务端配置**
   ```bash
   nano /etc/wireguard/wg0.conf
   ```

2. **添加模块端配置**
   在文件末尾添加:
   ```ini
   [Peer]
   PublicKey = 模块端公钥
   AllowedIPs = 10.8.0.2/32, 192.168.1.0/24
   PersistentKeepalive = 25
   ```

3. **重启服务端WireGuard**
   ```bash
   systemctl restart wg-quick@wg0
   ```

### 第四步: 验证连接

1. **检查服务端状态**
   ```bash
   wg show
   ```

2. **检查模块端状态**
   ```bash
   wg show
   ping 10.8.0.1
   ```

3. **运行连接测试**
   ```bash
   /etc/wireguard/test-connection.sh
   ```

## 👥 添加用户

### 生成用户密钥
```bash
# 在任意设备上执行
wg genkey | tee user1_private.key | wg pubkey > user1_public.key
cat user1_private.key  # 记录私钥
cat user1_public.key   # 记录公钥
```

### 添加用户到服务端
编辑 `/etc/wireguard/wg0.conf`，添加:
```ini
[Peer]
PublicKey = 用户公钥
AllowedIPs = 10.8.0.10/32
```

### 创建用户配置文件
```ini
[Interface]
PrivateKey = 用户私钥
Address = 10.8.0.10/32
DNS = 8.8.8.8

[Peer]
PublicKey = 服务端公钥
Endpoint = 服务端公网IP:51820
AllowedIPs = 10.8.0.0/24, 192.168.1.0/24
PersistentKeepalive = 25
```

## 📱 客户端软件

### Windows
- [WireGuard for Windows](https://www.wireguard.com/install/)

### macOS
- [WireGuard for macOS](https://apps.apple.com/app/wireguard/id1451685025)

### Android
- [WireGuard for Android](https://play.google.com/store/apps/details?id=com.wireguard.android)

### iOS
- [WireGuard for iOS](https://apps.apple.com/app/wireguard/id1441195209)

### Linux
```bash
sudo apt install wireguard
```

## 🔍 故障排除

### 常见问题

1. **连接超时**
   - 检查防火墙设置
   - 确认服务端公网IP正确
   - 验证端口51820是否开放

2. **能连VPN但无法访问内网**
   - 检查模块端IP转发是否启用
   - 验证iptables规则
   - 确认内网网段配置正确

3. **频繁断线**
   - 调整PersistentKeepalive值
   - 检查NAT类型
   - 优化MTU设置

### 调试命令
```bash
# 查看WireGuard状态
wg show

# 查看日志
journalctl -u wg-quick@wg0 -f

# 查看路由
ip route show

# 测试连通性
ping 10.8.0.1
ping 192.168.1.1
```

## 🛡️ 安全建议

1. **定期更换密钥**
2. **限制用户访问范围**
3. **启用日志监控**
4. **使用强密码保护配置文件**
5. **定期更新WireGuard版本**

## 📊 监控和维护

### 查看连接状态
```bash
wg show
```

### 查看流量统计
```bash
wg show all dump
```

### 重启服务
```bash
systemctl restart wg-quick@wg0
```

### 查看日志
```bash
journalctl -u wg-quick@wg0
```

## 🔄 自动化管理

考虑使用项目中的EITEC VPN管理系统来自动化管理WireGuard配置:

1. **启动服务端管理系统**
   ```bash
   cd /path/to/eitec-vpn
   ./bin/eitec-vpn-server
   ```

2. **启动模块端管理系统**
   ```bash
   cd /path/to/eitec-vpn
   ./bin/eitec-vpn-module
   ```

3. **通过Web界面管理**
   - 服务端: `http://server-ip:8080`
   - 模块端: `http://module-ip:8081`

## 📞 支持

如果遇到问题，请:
1. 查看日志文件
2. 运行测试脚本
3. 检查网络配置
4. 参考故障排除章节

---

**注意**: 请根据实际网络环境调整IP地址段和网络接口名称。 