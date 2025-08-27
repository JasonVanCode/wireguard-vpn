# EiTec VPN 启动脚本使用说明

本目录包含了 EiTec VPN 项目的系统启动和管理脚本，支持全路径执行和当前目录执行。

## 📁 脚本文件

- **`start-server.sh`** - 启动服务端
- **`start-module.sh`** - 启动模块端  
- **`stop-all.sh`** - 停止所有服务
- **`status.sh`** - 检查服务状态

## 🚀 使用方法

### 1. 启动服务端

```bash
# 全路径执行
/absolute/path/to/eitec-vpn/scripts/start-server.sh

# 当前目录执行
cd /path/to/eitec-vpn
./scripts/start-server.sh

# 或者从项目根目录执行
cd /path/to/eitec-vpn
scripts/start-server.sh
```

### 2. 启动模块端

```bash
# 全路径执行
/absolute/path/to/eitec-vpn/scripts/start-module.sh

# 当前目录执行
cd /path/to/eitec-vpn
./scripts/start-module.sh

# 或者从项目根目录执行
cd /path/to/eitec-vpn
scripts/start-module.sh
```

### 3. 停止所有服务

```bash
# 全路径执行
/absolute/path/to/eitec-vpn/scripts/stop-all.sh

# 当前目录执行
cd /path/to/eitec-vpn
./scripts/stop-all.sh
```

### 4. 检查服务状态

```bash
# 全路径执行
/absolute/path/to/eitec-vpn/scripts/status.sh

# 当前目录执行
cd /path/to/eitec-vpn
./scripts/status.sh
```

## ⚠️ 注意事项

1. **前置条件**: 运行启动脚本前，请确保已经构建了项目：
   ```bash
   make build-local    # 本地构建
   # 或者
   make build-server   # 构建服务端
   make build-module   # 构建模块端
   ```

2. **配置文件**: 确保配置文件存在：
   - `configs/server.yaml` - 服务端配置
   - `configs/module.yaml` - 模块端配置

3. **权限要求**: 脚本会自动检测并停止已运行的进程，无需手动干预

4. **日志文件**: 启动后日志会保存在 `logs/` 目录下：
   - `logs/server.log` - 服务端日志
   - `logs/module.log` - 模块端日志

## 🔧 脚本特性

- **智能路径检测**: 自动识别项目根目录，支持任意位置执行
- **进程管理**: 自动检测并管理重复进程
- **状态检查**: 提供详细的服务状态信息
- **错误处理**: 完善的错误检查和提示信息
- **日志管理**: 自动创建日志目录和文件

## 📋 常用命令

```bash
# 查看服务端日志
tail -f logs/server.log

# 查看模块端日志  
tail -f logs/module.log

# 查看进程状态
ps aux | grep eitec-vpn

# 查看端口占用
netstat -tlnp | grep eitec-vpn
```

## 🆘 故障排除

如果启动失败，请检查：

1. 项目是否正确构建（`bin/` 目录下是否有二进制文件）
2. 配置文件是否存在且格式正确
3. 端口是否被其他服务占用
4. 查看日志文件获取详细错误信息

## 📞 支持

如有问题，请查看项目文档或提交 Issue。
