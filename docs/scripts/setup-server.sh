#!/bin/bash

# WireGuard服务端部署脚本
# 适用于: Ubuntu 20.04+, Debian 11+, CentOS 8+

set -e

echo "=== WireGuard服务端部署脚本 ==="

# 检查是否为root用户
if [[ $EUID -ne 0 ]]; then
   echo "错误: 请使用root用户运行此脚本"
   exit 1
fi

# 检测操作系统
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$NAME
    VER=$VERSION_ID
else
    echo "错误: 无法检测操作系统"
    exit 1
fi

echo "检测到操作系统: $OS $VER"

# 安装WireGuard
echo "正在安装WireGuard..."
if [[ "$OS" == *"Ubuntu"* ]] || [[ "$OS" == *"Debian"* ]]; then
    apt update
    apt install -y wireguard wireguard-tools ufw
elif [[ "$OS" == *"CentOS"* ]] || [[ "$OS" == *"Red Hat"* ]]; then
    yum install -y epel-release
    yum install -y wireguard-tools firewalld
else
    echo "错误: 不支持的操作系统"
    exit 1
fi

# 启用IP转发
echo "配置IP转发..."
echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.conf
sysctl -p

# 生成服务端密钥
echo "生成服务端密钥..."
cd /etc/wireguard
wg genkey | tee server_private.key | wg pubkey > server_public.key
chmod 600 server_private.key

# 获取网络接口名称
INTERFACE=$(ip route | grep default | awk '{print $5}' | head -n1)
echo "检测到网络接口: $INTERFACE"

# 读取生成的密钥
SERVER_PRIVATE_KEY=$(cat server_private.key)
SERVER_PUBLIC_KEY=$(cat server_public.key)

# 创建配置文件
echo "创建WireGuard配置文件..."
cat > /etc/wireguard/wg0.conf << EOF
[Interface]
PrivateKey = $SERVER_PRIVATE_KEY
Address = 10.8.0.1/24
ListenPort = 51820
PostUp = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o $INTERFACE -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o $INTERFACE -j MASQUERADE

# 工厂模块端配置 (需要手动添加公钥)
[Peer]
# PublicKey = MODULE_PUBLIC_KEY_HERE
# AllowedIPs = 10.8.0.2/32, 192.168.1.0/24
# PersistentKeepalive = 25

# 外部用户配置 (需要手动添加)
# [Peer]
# PublicKey = USER_PUBLIC_KEY_HERE
# AllowedIPs = 10.8.0.10/32
EOF

# 配置防火墙
echo "配置防火墙..."
if [[ "$OS" == *"Ubuntu"* ]] || [[ "$OS" == *"Debian"* ]]; then
    ufw allow 51820/udp
    ufw allow ssh
    ufw --force enable
elif [[ "$OS" == *"CentOS"* ]] || [[ "$OS" == *"Red Hat"* ]]; then
    firewall-cmd --permanent --add-port=51820/udp
    firewall-cmd --permanent --add-service=ssh
    firewall-cmd --reload
fi

# 启动WireGuard服务
echo "启动WireGuard服务..."
systemctl enable wg-quick@wg0
systemctl start wg-quick@wg0

# 显示服务端信息
echo ""
echo "=== 服务端部署完成 ==="
echo "服务端公钥: $SERVER_PUBLIC_KEY"
echo "服务端IP: 10.8.0.1"
echo "监听端口: 51820"
echo "配置文件: /etc/wireguard/wg0.conf"
echo ""
echo "下一步操作:"
echo "1. 记录服务端公钥，配置客户端时需要使用"
echo "2. 编辑 /etc/wireguard/wg0.conf 添加客户端配置"
echo "3. 使用 'systemctl restart wg-quick@wg0' 重启服务"
echo "4. 使用 'wg show' 查看连接状态"
echo ""
echo "客户端配置模板已保存到 /etc/wireguard/client-template.conf"

# 创建客户端配置模板
cat > /etc/wireguard/client-template.conf << EOF
[Interface]
PrivateKey = CLIENT_PRIVATE_KEY_HERE
Address = 10.8.0.10/32
DNS = 8.8.8.8, 8.8.4.4

[Peer]
PublicKey = $SERVER_PUBLIC_KEY
Endpoint = $(curl -s ifconfig.me):51820
AllowedIPs = 10.8.0.0/24, 192.168.1.0/24
PersistentKeepalive = 25
EOF

echo "客户端配置模板已创建完成!" 