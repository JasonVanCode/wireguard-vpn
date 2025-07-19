#!/bin/bash

# WireGuard模块端部署脚本
# 适用于: Ubuntu 20.04+, Debian 11+, CentOS 8+

set -e

echo "=== WireGuard模块端部署脚本 ==="

# 检查是否为root用户
if [[ $EUID -ne 0 ]]; then
   echo "错误: 请使用root用户运行此脚本"
   exit 1
fi

# 获取参数
if [ $# -lt 2 ]; then
    echo "用法: $0 <服务端公钥> <服务端IP或域名> [内网网段]"
    echo "示例: $0 'server_public_key_here' '203.0.113.1' '192.168.1.0/24'"
    exit 1
fi

SERVER_PUBLIC_KEY="$1"
SERVER_ENDPOINT="$2"
INTERNAL_NETWORK="${3:-192.168.1.0/24}"

echo "服务端公钥: $SERVER_PUBLIC_KEY"
echo "服务端地址: $SERVER_ENDPOINT"
echo "内网网段: $INTERNAL_NETWORK"

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

# 生成模块端密钥
echo "生成模块端密钥..."
cd /etc/wireguard
wg genkey | tee module_private.key | wg pubkey > module_public.key
chmod 600 module_private.key

# 获取网络接口名称
INTERFACE=$(ip route | grep default | awk '{print $5}' | head -n1)
echo "检测到网络接口: $INTERFACE"

# 读取生成的密钥
MODULE_PRIVATE_KEY=$(cat module_private.key)
MODULE_PUBLIC_KEY=$(cat module_public.key)

# 创建配置文件
echo "创建WireGuard配置文件..."
cat > /etc/wireguard/wg0.conf << EOF
[Interface]
PrivateKey = $MODULE_PRIVATE_KEY
Address = 10.8.0.2/32
DNS = 8.8.8.8, 8.8.4.4

# 配置NAT转发，实现内网穿透
PostUp = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -A FORWARD -o wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o $INTERFACE -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -D FORWARD -o wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o $INTERFACE -j MASQUERADE

[Peer]
PublicKey = $SERVER_PUBLIC_KEY
Endpoint = $SERVER_ENDPOINT:51820
AllowedIPs = 10.8.0.0/24
PersistentKeepalive = 25
EOF

# 配置防火墙
echo "配置防火墙..."
if [[ "$OS" == *"Ubuntu"* ]] || [[ "$OS" == *"Debian"* ]]; then
    # 允许内网访问
    ufw allow from $INTERNAL_NETWORK
    # 允许VPN网段访问
    ufw allow from 10.8.0.0/24
    ufw allow ssh
    ufw --force enable
elif [[ "$OS" == *"CentOS"* ]] || [[ "$OS" == *"Red Hat"* ]]; then
    firewall-cmd --permanent --add-rich-rule="rule family='ipv4' source address='$INTERNAL_NETWORK' accept"
    firewall-cmd --permanent --add-rich-rule="rule family='ipv4' source address='10.8.0.0/24' accept"
    firewall-cmd --permanent --add-service=ssh
    firewall-cmd --reload
fi

# 启动WireGuard服务
echo "启动WireGuard服务..."
systemctl enable wg-quick@wg0
systemctl start wg-quick@wg0

# 检查服务状态
sleep 2
if systemctl is-active --quiet wg-quick@wg0; then
    echo "WireGuard服务启动成功"
else
    echo "警告: WireGuard服务启动失败，请检查配置"
fi

# 显示模块端信息
echo ""
echo "=== 模块端部署完成 ==="
echo "模块端公钥: $MODULE_PUBLIC_KEY"
echo "模块端VPN IP: 10.8.0.2"
echo "内网网段: $INTERNAL_NETWORK"
echo "配置文件: /etc/wireguard/wg0.conf"
echo ""
echo "重要: 请将以下信息添加到服务端配置中:"
echo ""
echo "[Peer]"
echo "PublicKey = $MODULE_PUBLIC_KEY"
echo "AllowedIPs = 10.8.0.2/32, $INTERNAL_NETWORK"
echo "PersistentKeepalive = 25"
echo ""
echo "然后在服务端执行: systemctl restart wg-quick@wg0"
echo ""
echo "测试命令:"
echo "wg show                    # 查看连接状态"
echo "ping 10.8.0.1             # 测试到服务端的连接"
echo "ip route show             # 查看路由表"

# 创建测试脚本
cat > /etc/wireguard/test-connection.sh << 'EOF'
#!/bin/bash
echo "=== WireGuard连接测试 ==="
echo "1. WireGuard状态:"
wg show

echo ""
echo "2. 测试到服务端连接:"
ping -c 3 10.8.0.1

echo ""
echo "3. 路由表:"
ip route show | grep wg0

echo ""
echo "4. 防火墙状态:"
if command -v ufw >/dev/null 2>&1; then
    ufw status
elif command -v firewall-cmd >/dev/null 2>&1; then
    firewall-cmd --list-all
fi
EOF

chmod +x /etc/wireguard/test-connection.sh
echo "测试脚本已创建: /etc/wireguard/test-connection.sh" 