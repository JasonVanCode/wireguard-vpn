#!/bin/bash

# EiTec VPN 服务端启动脚本
# 支持全路径执行和当前目录执行

set -e

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 检查是否在项目根目录
if [ ! -f "$PROJECT_ROOT/go.mod" ]; then
    echo "错误: 无法找到项目根目录，请确保在正确的项目目录中运行"
    exit 1
fi

echo "=== EiTec VPN 服务端启动脚本 ==="
echo "项目根目录: $PROJECT_ROOT"
echo "脚本目录: $SCRIPT_DIR"

# 检查二进制文件
SERVER_BINARY="$PROJECT_ROOT/bin/eitec-vpn-server"
CONFIG_FILE="$PROJECT_ROOT/configs/server.yaml"

# 检查二进制文件是否存在
if [ ! -f "$SERVER_BINARY" ]; then
    echo "错误: 服务端二进制文件不存在: $SERVER_BINARY"
    echo "请先运行 'make build-server' 或 'make build-local' 构建项目"
    exit 1
fi

# 检查配置文件
if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 配置文件不存在: $CONFIG_FILE"
    exit 1
fi

# 检查是否已有进程在运行
if pgrep -f "eitec-vpn-server" > /dev/null; then
    echo "警告: 检测到服务端进程已在运行"
    echo "正在停止现有进程..."
    pkill -f "eitec-vpn-server"
    sleep 2
fi

# 创建日志目录
mkdir -p "$PROJECT_ROOT/logs"

# 启动服务端
echo "正在启动服务端..."
echo "二进制文件: $SERVER_BINARY"
echo "配置文件: $CONFIG_FILE"
echo "日志文件: $PROJECT_ROOT/logs/server.log"

# 后台启动服务端
nohup "$SERVER_BINARY" --config "$CONFIG_FILE" > "$PROJECT_ROOT/logs/server.log" 2>&1 &

# 获取进程ID
SERVER_PID=$!
echo "服务端已启动，进程ID: $SERVER_PID"

# 等待服务启动
sleep 3

# 检查服务状态
if kill -0 $SERVER_PID 2>/dev/null; then
    echo "✅ 服务端启动成功!"
    echo "进程ID: $SERVER_PID"
    echo "日志文件: $PROJECT_ROOT/logs/server.log"
    echo ""
    echo "常用命令:"
    echo "  查看日志: tail -f $PROJECT_ROOT/logs/server.log"
    echo "  停止服务: kill $SERVER_PID"
    echo "  查看进程: ps aux | grep eitec-vpn-server"
else
    echo "❌ 服务端启动失败!"
    echo "请检查日志文件: $PROJECT_ROOT/logs/server.log"
    exit 1
fi
