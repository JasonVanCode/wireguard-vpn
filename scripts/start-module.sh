#!/bin/bash

# EiTec VPN 模块端启动脚本
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

echo "=== EiTec VPN 模块端启动脚本 ==="
echo "项目根目录: $PROJECT_ROOT"
echo "脚本目录: $SCRIPT_DIR"

# 检查二进制文件
MODULE_BINARY="$PROJECT_ROOT/bin/eitec-vpn-module"
CONFIG_FILE="$PROJECT_ROOT/configs/module.yaml"

# 检查二进制文件是否存在
if [ ! -f "$MODULE_BINARY" ]; then
    echo "错误: 模块端二进制文件不存在: $MODULE_BINARY"
    echo "请先运行 'make build-module' 或 'make build-local' 构建项目"
    exit 1
fi

# 检查配置文件
if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 配置文件不存在: $CONFIG_FILE"
    exit 1
fi

# 检查是否已有进程在运行
if pgrep -f "eitec-vpn-module" > /dev/null; then
    echo "警告: 检测到模块端进程已在运行"
    echo "正在停止现有进程..."
    pkill -f "eitec-vpn-module"
    sleep 2
fi

# 创建日志目录
mkdir -p "$PROJECT_ROOT/logs"

# 启动模块端
echo "正在启动模块端..."
echo "二进制文件: $MODULE_BINARY"
echo "配置文件: $CONFIG_FILE"
echo "日志文件: $PROJECT_ROOT/logs/module.log"

# 后台启动模块端
nohup "$MODULE_BINARY" --config "$CONFIG_FILE" > "$PROJECT_ROOT/logs/module.log" 2>&1 &

# 获取进程ID
MODULE_PID=$!
echo "模块端已启动，进程ID: $MODULE_PID"

# 等待服务启动
sleep 3

# 检查服务状态
if kill -0 $MODULE_PID 2>/dev/null; then
    echo "✅ 模块端启动成功!"
    echo "进程ID: $MODULE_PID"
    echo "日志文件: $PROJECT_ROOT/logs/module.log"
    echo ""
    echo "常用命令:"
    echo "  查看日志: tail -f $PROJECT_ROOT/logs/module.log"
    echo "  停止服务: kill $MODULE_PID"
    echo "  查看进程: ps aux | grep eitec-vpn-module"
else
    echo "❌ 模块端启动失败!"
    echo "请检查日志文件: $PROJECT_ROOT/logs/module.log"
    exit 1
fi
