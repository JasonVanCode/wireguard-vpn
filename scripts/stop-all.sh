#!/bin/bash

# EiTec VPN 停止所有服务脚本
# 支持全路径执行和当前目录执行

set -e

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=== EiTec VPN 停止所有服务 ==="
echo "项目根目录: $PROJECT_ROOT"

# 停止服务端进程
if pgrep -f "eitec-vpn-server" > /dev/null; then
    echo "正在停止服务端进程..."
    pkill -f "eitec-vpn-server"
    sleep 2
    
    # 强制停止（如果还在运行）
    if pgrep -f "eitec-vpn-server" > /dev/null; then
        echo "强制停止服务端进程..."
        pkill -9 -f "eitec-vpn-server"
    fi
    echo "✅ 服务端已停止"
else
    echo "服务端进程未运行"
fi

# 停止模块端进程
if pgrep -f "eitec-vpn-module" > /dev/null; then
    echo "正在停止模块端进程..."
    pkill -f "eitec-vpn-module"
    sleep 2
    
    # 强制停止（如果还在运行）
    if pgrep -f "eitec-vpn-module" > /dev/null; then
        echo "强制停止模块端进程..."
        pkill -9 -f "eitec-vpn-module"
    fi
    echo "✅ 模块端已停止"
else
    echo "模块端进程未运行"
fi

# 检查是否还有相关进程
REMAINING=$(pgrep -f "eitec-vpn" | wc -l)
if [ "$REMAINING" -eq 0 ]; then
    echo "✅ 所有服务已成功停止"
else
    echo "⚠️  仍有 $REMAINING 个相关进程在运行"
    pgrep -f "eitec-vpn" | xargs ps -o pid,cmd
fi

echo ""
echo "服务状态检查完成"
