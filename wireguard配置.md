好的，非常荣幸能为您在我们共同排查和解决问题的基础上，整理出这份完整的配置说明文档。

这份文档将详细解释如何从零开始，生成并理解您现在这套完整、健壮、且功能强大的WireGuard网络配置。

-----

## WireGuard “内网穿透” (LAN-to-LAN) 完整配置指南

### 1\. 目标与场景

本指南旨在搭建一个包含三个节点的WireGuard网络，并实现“内网穿透”功能。

  * **节点一：云服务器 (Server)**

      * **VPN IP**: `10.50.0.1`
      * 作为中心节点，负责转发所有客户端之间的流量。
      * 拥有一个公网IP地址。

  * **节点二：模块端 (Module/Gateway)**

      * **VPN IP**: `10.50.0.2`
      * 通常是一个位于家庭或办公室局域网中的设备（如树莓派）。
      * 它背后的局域网是 `192.168.50.0/24`。
      * **目标**：允许其他VPN客户端访问它所在的这个局域网。

  * **节点三：电脑端 (PC/Roaming Client)**

      * **VPN IP**: `10.50.0.3`
      * 一个普通的漫游客户端（如笔记本电脑、手机）。
      * **目标**：能够访问VPN内的所有设备，并且能“穿透”到模块端背后的 `192.168.50.0/24` 局域网。

### 2\. 准备工作

在开始配置前，需要完成以下准备：

#### 2.1 密钥生成

每个节点都需要自己的一对**私钥 (PrivateKey)** 和**公钥 (PublicKey)**。总共需要生成3对。在每个设备的Linux终端上执行以下命令：

```bash
# 生成私钥并保存到 privatekey 文件
wg genkey | tee privatekey
# 从私钥生成对应的公钥并保存到 publickey 文件
cat privatekey | wg pubkey > publickey
```

请将这3对密钥（6个文件）安全地保存好，并明确标记哪个属于哪个设备。

#### 2.2 服务器准备

1.  **开启内核IP转发**：这是让服务器能作为路由器的总开关。
    ```bash
    # 编辑配置文件
    sudo nano /etc/sysctl.conf
    # 确保文件中有以下内容（去掉#号注释或新增一行）
    net.ipv4.ip_forward=1
    # 使配置生效
    sudo sysctl -p
    ```
2.  **防火墙放行端口**：在您的云服务商（如阿里云/腾讯云）的**安全组**，以及服务器自身的防火墙（如UFW/宝塔面板）中，放行WireGuard的监听端口。在本例中是 **`51824/udp`**。

-----

### 3\. 配置文件生成详解

下面我们来逐一生成每个节点的配置文件。

#### 3.1 【服务端】配置文件 (`/etc/wireguard/wg5.conf`)

这是整个网络的中枢，配置最为关键。

```ini
[Interface]
# 使用服务器的【私钥】
PrivateKey = [此处填写服务器的私钥]
# 定义整个VPN子网，/24表示范围是10.50.0.1 - 10.50.0.254
Address = 10.50.0.1/24
# 服务器监听的公网端口
ListenPort = 51824
# MTU值，通常设为1420以避免分片问题
MTU = 1420
SaveConfig = true
# 启动时执行的防火墙命令
# 1. MASQUERADE: 允许VPN客户端通过服务器访问外网（如果需要）
# 2. INPUT ACCEPT: 允许WireGuard端口的入站流量
# 3. FORWARD ACCEPT: 允许客户端之间互相通信（关键！）
PostUp = iptables -t nat -A POSTROUTING -s 10.50.0.0/24 -o eth0 -j MASQUERADE; iptables -A INPUT -p udp -m udp --dport 51824 -j ACCEPT; iptables -I FORWARD 1 -i %i -j ACCEPT
# 关闭时移除对应的防火墙规则
PostDown = iptables -t nat -D POSTROUTING -s 10.50.0.0/24 -o eth0 -j MASQUERADE; iptables -D INPUT -p udp -m udp --dport 51824 -j ACCEPT; iptables -D FORWARD -i %i -j ACCEPT

[Peer]
# 描述：模块端/树莓派
# 使用模块端的【公钥】
PublicKey = [此处填写模块端的公钥]
# 为此Peer生成一个预共享密钥，增强安全性
PresharedKey = [此处填写为模块端生成的预共享密钥]
# AllowedIPs的含义：
# 1. 10.50.0.2/32: 服务器只接受来自10.50.0.2的流量从此Peer进入
# 2. 192.168.50.0/24: 告诉服务器，要访问这个局域网，必须把流量转发给此Peer
AllowedIPs = 10.50.0.2/32, 192.168.50.0/24
PersistentKeepalive = 25

[Peer]
# 描述：电脑端
# 使用电脑端的【公钥】
PublicKey = [此处填写电脑端的公钥]
# 为此Peer生成另一个预共享密钥
PresharedKey = [此处填写为电脑端生成的预共享密钥]
# AllowedIPs的含义：
# 这是一个普通漫游客户端，服务器只需知道它的VPN IP地址即可
AllowedIPs = 10.50.0.3/32
PersistentKeepalive = 25
```

#### 3.2 【模块端/网关】配置文件 (`/etc/wireguard/wg0.conf`)

这个客户端扮演着“网关”的角色，需要做一些额外的配置。

```ini
[Interface]
# 使用模块端的【私钥】
PrivateKey = [此处填写模块端的私钥]
# 客户端地址用/32
Address = 10.50.0.2/32
# 启动时执行的防火墙/路由命令
# 1. ip route replace: 解决源地址选择错误的问题，强制发包时源IP为10.50.0.2
# 2. iptables -I FORWARD: 解决本地Docker等软件的防火墙冲突
# 3. iptables -t nat: 实现内网穿透的核心，伪装地址
PostUp = ip route replace 10.50.0.0/24 dev %i src 10.50.0.2; iptables -I FORWARD 1 -o %i -j ACCEPT; iptables -I FORWARD 1 -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -s 10.50.0.0/24 -o wlan0 -j MASQUERADE
# 关闭时移除对应的规则
PostDown = iptables -D FORWARD -o %i -j ACCEPT; iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -s 10.50.0.0/24 -o wlan0 -j MASQUERADE

[Peer]
# 描述：服务器
# 使用服务器的【公钥】
PublicKey = [此处填写服务器的公钥]
# 使用与服务端为本设备配置的同一个预共享密钥
PresharedKey = [此处填写为模块端生成的预共享密钥]
# 服务器的公网IP和端口
Endpoint = [服务器公网IP]:51824
# AllowedIPs的含义：
# 告诉本设备，所有发往VPN网络的流量，都从隧道走
AllowedIPs = 10.50.0.0/24
PersistentKeepalive = 25
```

#### 3.3 【电脑端/漫游客户端】配置文件

这是最简单的一个配置，它只是一个纯粹的客户端。

```ini
[Interface]
# 使用电脑端的【私钥】
PrivateKey = [此处填写电脑端的私钥]
# 客户端地址用/32
Address = 10.50.0.3/32
# DNS (可选): 将DNS请求也通过VPN发送，防止DNS污染
DNS = 8.8.8.8, 8.8.4.4

[Peer]
# 描述：服务器
# 使用服务器的【公钥】
PublicKey = [此处填写服务器的公钥]
# 使用与服务端为本设备配置的同一个预共享密钥
PresharedKey = [此处填写为电脑端生成的预共享密钥]
# 服务器的公网IP和端口
Endpoint = [服务器公网IP]:51824
# AllowedIPs的含义：
# 1. 10.50.0.0/24: 告诉本设备，所有VPN内部流量都从隧道走
# 2. 192.168.50.0/24: 告诉本设备，要访问模块端的局域网，也要从隧道走
AllowedIPs = 10.50.0.0/24, 192.168.50.0/24
PersistentKeepalive = 25
```

-----

### 4\. 启动与验证

1.  将每个配置文件放置到对应设备的 `/etc/wireguard/` 目录下（接口名即文件名，如 `wg5.conf`）。
2.  **先启动服务端**：`sudo wg-quick up wg5`
3.  **再启动客户端**：`sudo wg-quick up wg0`
4.  **验证**：
      * 在每个设备上用 `sudo wg show` 检查是否有 `latest handshake`。
      * 在电脑端 `ping 10.50.0.1` (服务器) 和 `ping 10.50.0.2` (模块端)。
      * 在电脑端 `ping 192.168.50.241` (模块端的局域网IP)，验证内网穿透。

至此，整个网络的搭建就全部完成了。这份文档详细记录了每一步的逻辑和原因，希望能对您有所帮助。