#!/bin/bash

# EiTec VPN 服务状态检查脚本
# 支持全路径执行和当前目录执行

set -e

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=== EiTec VPN 服务状态检查 ==="
echo "项目根目录: $PROJECT_ROOT"
echo "检查时间: $(date)"
echo ""

# 检查服务端状态
echo "🔍 服务端状态检查:"
SERVER_PID=$(pgrep -f "eitec-vpn-server" | head -1)
if [ -n "$SERVER_PID" ]; then
    echo "✅ 服务端正在运行 (PID: $SERVER_PID)"
    echo "   进程信息:"
    ps -p $SERVER_PID -o pid,ppid,cmd,etime,pcpu,pmem --no-headers | sed 's/^/   /'
    
    # 检查端口占用
    SERVER_PORT=$(netstat -tlnp 2>/dev/null | grep "$SERVER_PID" | grep LISTEN | awk '{print $4}' | cut -d: -f2 | head -1)
    if [ -n "$SERVER_PORT" ]; then
        echo "   监听端口: $SERVER_PORT"
    fi
else
    echo "❌ 服务端未运行"
fi

echo ""

# 检查模块端状态
echo "🔍 模块端状态检查:"
MODULE_PID=$(pgrep -f "eitec-vpn-module" | head -1)
if [ -n "$MODULE_PID" ]; then
    echo "✅ 模块端正在运行 (PID: $MODULE_PID)"
    echo "   进程信息:"
    ps -p $MODULE_PID -o pid,ppid,cmd,etime,pcpu,pmem --no-headers | sed 's/^/   /'
    
    # 检查端口占用
    MODULE_PORT=$(netstat -tlnp 2>/dev/null | grep "$MODULE_PID" | grep LISTEN | awk '{print $4}' | cut -d: -f2 | head -1)
    if [ -n "$MODULE_PORT" ]; then
        echo "   监听端口: $MODULE_PORT"
    fi
else
    echo "❌ 模块端未运行"
fi

echo ""

# 检查日志文件
echo "📋 日志文件状态:"
if [ -f "$PROJECT_ROOT/logs/server.log" ]; then
    SERVER_LOG_SIZE=$(du -h "$PROJECT_ROOT/logs/server.log" | cut -f1)
    echo "   服务端日志: $PROJECT_ROOT/logs/server.log ($SERVER_LOG_SIZE)"
else
    echo "   服务端日志: 不存在"
fi

if [ -f "$PROJECT_ROOT/logs/module.log" ]; then
    MODULE_LOG_SIZE=$(du -h "$PROJECT_ROOT/logs/module.log" | cut -f1)
    echo "   模块端日志: $PROJECT_ROOT/logs/module.log ($MODULE_LOG_SIZE)"
else
    echo "   模块端日志: 不存在"
fi

echo ""

# 检查配置文件
echo "⚙️  配置文件状态:"
if [ -f "$PROJECT_ROOT/configs/server.yaml" ]; then
    echo "   服务端配置: ✅ $PROJECT_ROOT/configs/server.yaml"
else
    echo "   服务端配置: ❌ 不存在"
fi

if [ -f "$PROJECT_ROOT/configs/module.yaml" ]; then
    echo "   模块端配置: ✅ $PROJECT_ROOT/configs/module.yaml"
else
    echo "   模块端配置: ❌ 不存在"
fi

echo ""

# 检查二进制文件
echo "🔧 二进制文件状态:"
if [ -f "$PROJECT_ROOT/bin/eitec-vpn-server" ]; then
    SERVER_BIN_SIZE=$(du -h "$PROJECT_ROOT/bin/eitec-vpn-server" | cut -f1)
    echo "   服务端二进制: ✅ $PROJECT_ROOT/bin/eitec-vpn-server ($SERVER_BIN_SIZE)"
else
    echo "   服务端二进制: ❌ 不存在"
fi

if [ -f "$PROJECT_ROOT/bin/eitec-vpn-module" ]; then
    MODULE_BIN_SIZE=$(du -h "$PROJECT_ROOT/bin/eitec-vpn-module" | cut -f1)
    echo "   模块端二进制: ✅ $PROJECT_ROOT/bin/eitec-vpn-module ($MODULE_BIN_SIZE)"
else
    echo "   模块端二进制: ❌ 不存在"
fi

echo ""
echo "状态检查完成"
