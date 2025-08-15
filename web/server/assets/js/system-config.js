// =====================================================
// EITEC VPN - 系统配置管理功能  
// =====================================================
//
// 📋 功能概述：
// - 提供系统级WireGuard配置管理界面
// - 显示系统状态概览和接口统计信息
// - 处理系统配置导出、日志查看等高级功能
//
// 🔗 依赖关系：
// - 依赖：shared-utils.js (工具函数)
// - 依赖：bootstrap (模态框管理)
// - 调用 interface-management.js 的函数进行接口操作
//
// 📦 主要功能：
// - showWireGuardConfig() - 显示系统配置总览
// - refreshSystemConfig() - 刷新系统状态
// - exportSystemConfig() - 导出系统配置
// - viewSystemLogs() - 查看系统日志
// - initializeWireGuard() - 初始化WireGuard服务
// - 系统统计和状态监控显示
//
// 📏 文件大小：25.5KB (原文件的 24.4%)
// =====================================================

// WireGuard配置管理 - 重新设计为系统级管理面板
async function showWireGuardConfig() {
    const modalElement = document.getElementById('wireGuardConfigModal');
    if (!modalElement) {
        console.error('找不到 wireGuardConfigModal 元素');
        return;
    }
    
    // 使用新的模态框管理器
    ModalManager.show(modalElement);
    
    try {
        const token = localStorage.getItem('access_token');
        
        // 获取系统配置（服务器信息）
        const configResponse = await fetch('/api/v1/config', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        // 获取带状态的接口信息
        const interfaceStatsResponse = await fetch('/api/v1/system/wireguard-interfaces', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (configResponse.ok && interfaceStatsResponse.ok) {
            const configResult = await configResponse.json();
            const interfaceStatsResult = await interfaceStatsResponse.json();
            const config = configResult.data;
            const interfaces = interfaceStatsResult.data || [];
            
            // 计算统计信息
            const interfaceStats = {
                total_interfaces: interfaces.length,
                active_interfaces: interfaces.filter(iface => iface.status === 1).length,
                total_capacity: interfaces.reduce((sum, iface) => sum + (iface.max_peers || 0), 0),
                used_capacity: interfaces.reduce((sum, iface) => sum + (iface.total_peers || 0), 0)
            };
            
            let content = `
                <!-- 系统状态概览 -->
                <div class="row mb-4">
                    <div class="col-md-12">
                        <div class="card" style="background: rgba(30, 41, 59, 0.8); border: 1px solid rgba(100, 116, 139, 0.3);">
                            <div class="card-header" style="background: rgba(15, 23, 42, 0.6); border-bottom: 1px solid rgba(100, 116, 139, 0.3);">
                                <h6 class="mb-0" style="color: #f1f5f9;"><i class="fas fa-server me-2"></i>系统状态概览</h6>
                            </div>
                            <div class="card-body">
                                <div class="row">
                                    <div class="col-md-3">
                                        <div class="text-center">
                                            <div class="h4" style="color: #3b82f6;">${interfaceStats.total_interfaces}</div>
                                            <div style="color: #94a3b8;">总接口数</div>
                                        </div>
                                    </div>
                                    <div class="col-md-3">
                                        <div class="text-center">
                                            <div class="h4" style="color: #10b981;">${interfaceStats.active_interfaces}</div>
                                            <div style="color: #94a3b8;">运行中</div>
                                        </div>
                                    </div>
                                    <div class="col-md-3">
                                        <div class="text-center">
                                            <div class="h4" style="color: #06b6d4;">${interfaceStats.total_capacity}</div>
                                            <div style="color: #94a3b8;">总容量</div>
                                        </div>
                                    </div>
                                    <div class="col-md-3">
                                        <div class="text-center">
                                            <div class="h4" style="color: #f59e0b;">${interfaceStats.used_capacity}</div>
                                            <div style="color: #94a3b8;">已使用</div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- 服务器配置 -->
                <div class="row mb-4">
                    <div class="col-md-6">
                        <div class="card" style="background: rgba(30, 41, 59, 0.8); border: 1px solid rgba(100, 116, 139, 0.3);">
                            <div class="card-header" style="background: rgba(15, 23, 42, 0.6); border-bottom: 1px solid rgba(100, 116, 139, 0.3);">
                                <h6 class="mb-0" style="color: #f1f5f9;"><i class="fas fa-server me-2"></i>服务器信息</h6>
                            </div>
                            <div class="card-body">
                                <div class="mb-3">
                                    <label class="form-label" style="color: #e2e8f0;">服务器名称</label>
                                    <input type="text" class="form-control" value="${config.server.name}" readonly 
                                           style="background: rgba(15, 23, 42, 0.6); border: 1px solid rgba(100, 116, 139, 0.3); color: #f1f5f9;">
                                </div>
                                <div class="mb-3">
                                    <label class="form-label" style="color: #e2e8f0;">外网端点</label>
                                    <input type="text" class="form-control" value="${config.server.endpoint || '未配置'}" readonly 
                                           style="background: rgba(15, 23, 42, 0.6); border: 1px solid rgba(100, 116, 139, 0.3); color: #f1f5f9;">
                                </div>
                                <div class="mb-0">
                                    <label class="form-label" style="color: #e2e8f0;">Web管理端口</label>
                                    <input type="number" class="form-control" value="${config.server.web_port}" readonly 
                                           style="background: rgba(15, 23, 42, 0.6); border: 1px solid rgba(100, 116, 139, 0.3); color: #f1f5f9;">
                                </div>
                            </div>
                        </div>
                    </div>
                    <div class="col-md-6">
                        <div class="card" style="background: rgba(30, 41, 59, 0.8); border: 1px solid rgba(100, 116, 139, 0.3);">
                            <div class="card-header" style="background: rgba(15, 23, 42, 0.6); border-bottom: 1px solid rgba(100, 116, 139, 0.3);">
                                <h6 class="mb-0" style="color: #f1f5f9;"><i class="fas fa-chart-bar me-2"></i>系统统计</h6>
                            </div>
                            <div class="card-body">
                                <div class="mb-3">
                                    <div class="d-flex justify-content-between" style="color: #e2e8f0;">
                                        <span>接口使用率</span>
                                        <span>${interfaceStats.active_interfaces}/${interfaceStats.total_interfaces}</span>
                                    </div>
                                    <div class="progress mt-1" style="background: rgba(15, 23, 42, 0.6);">
                                        <div class="progress-bar bg-success" style="width: ${interfaceStats.total_interfaces > 0 ? (interfaceStats.active_interfaces / interfaceStats.total_interfaces * 100) : 0}%"></div>
                                    </div>
                                </div>
                                <div class="mb-3">
                                    <div class="d-flex justify-content-between" style="color: #e2e8f0;">
                                        <span>连接使用率</span>
                                        <span>${interfaceStats.used_capacity}/${interfaceStats.total_capacity}</span>
                                    </div>
                                    <div class="progress mt-1" style="background: rgba(15, 23, 42, 0.6);">
                                        <div class="progress-bar bg-info" style="width: ${interfaceStats.total_capacity > 0 ? (interfaceStats.used_capacity / interfaceStats.total_capacity * 100) : 0}%"></div>
                                    </div>
                                </div>
                                <div class="mb-0">
                                    <div class="d-flex justify-content-between" style="color: #e2e8f0;">
                                        <span>系统状态</span>
                                        <span class="badge ${interfaceStats.active_interfaces > 0 ? 'bg-success' : 'bg-warning'}">
                                            ${interfaceStats.active_interfaces > 0 ? '运行中' : '待启动'}
                                        </span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- 接口概览 -->
                <div class="row mb-4">
                    <div class="col-md-12">
                        <div class="card" style="background: rgba(30, 41, 59, 0.8); border: 1px solid rgba(100, 116, 139, 0.3);">
                            <div class="card-header" style="background: rgba(15, 23, 42, 0.6); border-bottom: 1px solid rgba(100, 116, 139, 0.3);">
                                <h6 class="mb-0" style="color: #f1f5f9;"><i class="fas fa-list me-2"></i>接口概览</h6>
                            </div>
                            <div class="card-body" style="padding: 0;">
                                <div class="table-responsive">
                                    <table style="width: 100%; margin: 0; background: transparent; color: #e2e8f0;">
                                        <thead>
                                            <tr style="background: rgba(15, 23, 42, 0.8); border-bottom: 2px solid rgba(100, 116, 139, 0.4);">
                                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">接口名称</th>
                                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">状态</th>
                                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">网络段</th>
                                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">端口</th>
                                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">连接数</th>
                                                <th style="color: #f1f5f9; border: none; padding: 12px 16px; font-weight: 600; text-align: left;">操作</th>
                                            </tr>
                                        </thead>
                                        <tbody style="background: transparent;">
                                            ${interfaceStats.interfaces.map(iface => `
                                                <tr style="background: rgba(30, 41, 59, 0.3); border-bottom: 1px solid rgba(100, 116, 139, 0.2); transition: background-color 0.2s ease;" 
                                                    onmouseover="this.style.background='rgba(30, 41, 59, 0.6)'" 
                                                    onmouseout="this.style.background='rgba(30, 41, 59, 0.3)'">
                                                    <td style="border: none; padding: 12px 16px;">
                                                        <div style="display: flex; align-items: center;">
                                                            <i class="fas fa-ethernet" style="color: #60a5fa; margin-right: 8px; font-size: 14px;"></i>
                                                            <div>
                                                                <div style="color: #f1f5f9; font-size: 14px; font-weight: 600;">${iface.name}</div>
                                                                <small style="color: #94a3b8;">${iface.description || '无描述'}</small>
                                                            </div>
                                                        </div>
                                                    </td>
                                                    <td style="border: none; padding: 12px 16px;">
                                                        <span class="badge bg-${iface.status === 1 ? 'success' : 'secondary'}">${iface.status === 1 ? '运行中' : '已停止'}</span>
                                                    </td>
                                                    <td style="border: none; padding: 12px 16px;">
                                                        <span style="background: rgba(15, 23, 42, 0.8); color: #a78bfa; padding: 4px 8px; border-radius: 6px; font-family: 'Courier New', monospace; font-size: 13px; font-weight: 500;">${iface.network}</span>
                                                        <br>
                                                        <small style="color: #94a3b8; margin-top: 4px; display: inline-block;">服务器IP: ${iface.server_ip}</small>
                                                    </td>
                                                    <td style="border: none; padding: 12px 16px;">
                                                        <span style="background: rgba(15, 23, 42, 0.8); color: #34d399; padding: 4px 8px; border-radius: 6px; font-family: 'Courier New', monospace; font-size: 13px; font-weight: 500;">${iface.listen_port}</span>
                                                    </td>
                                                    <td style="border: none;">
                                                        <div style="position: relative;">
                                                            <div class="progress" style="height: 20px; background: rgba(15, 23, 42, 0.6);">
                                                                <div class="progress-bar bg-info" role="progressbar" 
                                                                     style="width: ${(iface.total_peers / iface.max_peers * 100).toFixed(1)}%">
                                                                </div>
                                                            </div>
                                                            <div style="position: absolute; top: 0; left: 0; right: 0; bottom: 0; display: flex; align-items: center; justify-content: center;">
                                                                <span style="color: #f1f5f9; font-weight: 600; font-size: 0.8rem; text-shadow: 1px 1px 2px rgba(0,0,0,0.8);">${iface.total_peers}/${iface.max_peers}</span>
                                                            </div>
                                                        </div>
                                                    </td>
                                                    <td style="border: none; padding: 12px 16px;">
                                                        <div class="btn-group btn-group-sm">
                                                            ${iface.status === 1 ? 
                                                                `<button onclick="stopInterface(${iface.id})" 
                                                                        style="background: transparent; border: 1px solid #f59e0b; color: #fbbf24; padding: 4px 8px; border-radius: 4px; font-size: 12px; cursor: pointer; margin-right: 4px; transition: all 0.2s ease;"
                                                                        onmouseover="this.style.background='rgba(245, 158, 11, 0.1)'"
                                                                        onmouseout="this.style.background='transparent'">
                                                                    <i class="fas fa-stop" style="margin-right: 4px;"></i>停止
                                                                </button>` :
                                                                `<button onclick="startInterface(${iface.id})" 
                                                                        style="background: transparent; border: 1px solid #10b981; color: #34d399; padding: 4px 8px; border-radius: 4px; font-size: 12px; cursor: pointer; margin-right: 4px; transition: all 0.2s ease;"
                                                                        onmouseover="this.style.background='rgba(16, 185, 129, 0.1)'"
                                                                        onmouseout="this.style.background='transparent'">
                                                                    <i class="fas fa-play" style="margin-right: 4px;"></i>启动
                                                                </button>`
                                                            }
                                                            <button onclick="viewInterfaceConfig(${iface.id})" 
                                                                    style="background: transparent; border: 1px solid #06b6d4; color: #22d3ee; padding: 4px 8px; border-radius: 4px; font-size: 12px; cursor: pointer; transition: all 0.2s ease;"
                                                                    onmouseover="this.style.background='rgba(6, 182, 212, 0.1)'"
                                                                    onmouseout="this.style.background='transparent'">
                                                                <i class="fas fa-eye" style="margin-right: 4px;"></i>查看
                                                            </button>
                                                        </div>
                                                    </td>
                                                </tr>
                                            `).join('')}
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- 快捷操作 -->
                <div class="row">
                    <div class="col-md-12">
                        <div class="d-flex justify-content-between">
                            <div>
                                <button class="btn btn-primary me-2" onclick="showInterfaceManager()">
                                    <i class="fas fa-ethernet me-1"></i>接口管理
                                </button>
                                <button class="btn btn-success me-2" onclick="createNewInterface()">
                                    <i class="fas fa-plus me-1"></i>创建接口
                                </button>
                                <button class="btn btn-secondary me-2" onclick="refreshSystemConfig()">
                                    <i class="fas fa-sync-alt me-1"></i>刷新状态
                                </button>
                            </div>
                            <div>
                                <button class="btn btn-info me-2" onclick="exportSystemConfig()">
                                    <i class="fas fa-download me-1"></i>导出配置
                                </button>
                                <button class="btn btn-warning me-2" onclick="viewSystemLogs()">
                                    <i class="fas fa-file-alt me-1"></i>系统日志
                                </button>
                                <button class="btn btn-outline-danger" onclick="systemSettings()">
                                    <i class="fas fa-cog me-1"></i>系统设置
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            `;
            
            document.getElementById('wgConfigContent').innerHTML = content;
        } else {
            document.getElementById('wgConfigContent').innerHTML = '<div class="alert alert-danger">加载系统信息失败</div>';
        }
    } catch (error) {
        console.error('加载系统配置失败:', error);
        document.getElementById('wgConfigContent').innerHTML = '<div class="alert alert-danger">网络错误</div>';
    }
}

// 查看接口详情
function viewInterfaceDetails(interfaceId) {
    // 关闭当前模态框
    const currentModal = bootstrap.Modal.getInstance(document.getElementById('wireGuardConfigModal'));
    if (currentModal) {
        currentModal.hide();
    }
    
    // 打开接口管理器并定位到指定接口
    setTimeout(() => {
        showInterfaceManager();
        // 可以在这里添加高亮显示指定接口的逻辑
    }, 300);
}

// 查看接口配置（在系统配置中）
function viewInterfaceConfig(interfaceId) {
    showInterfaceConfig(interfaceId);
}

// 刷新系统配置
function refreshSystemConfig() {
    showWireGuardConfig();
}

// 创建新接口（从系统配置中）
function createNewInterface() {
    showCreateInterfaceModal();
}

// 导出系统配置
async function exportSystemConfig() {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/config/export', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.ok) {
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'eitec-vpn-config.json';
            a.click();
            window.URL.revokeObjectURL(url);
        } else {
            alert('导出配置失败');
        }
    } catch (error) {
        console.error('导出配置失败:', error);
        alert('导出配置失败');
    }
}

// 查看系统日志
function viewSystemLogs() {
    alert('系统日志功能开发中...');
}

// 系统设置
function systemSettings() {
    alert('系统设置功能开发中...');
}

// 初始化WireGuard
async function initializeWireGuard() {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/config/wireguard/init', {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        const result = await response.json();
        if (response.ok) {
            alert('WireGuard初始化成功！');
            showWireGuardConfig(); // 刷新配置显示
        } else {
            alert('初始化失败：' + result.message);
        }
    } catch (error) {
        console.error('初始化WireGuard失败:', error);
        alert('网络错误，请重试');
    }
}

// 下载服务器配置
async function downloadServerConfig() {
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/config/wireguard/server-config', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (response.ok) {
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'wg0.conf';
            a.click();
            window.URL.revokeObjectURL(url);
        } else {
            alert('下载失败');
        }
    } catch (error) {
        console.error('下载配置失败:', error);
        alert('网络错误，请重试');
    }
}

// 应用WireGuard配置
async function applyWireGuardConfig() {
    if (!confirm('确定要应用WireGuard配置吗？这将重启WireGuard服务。')) {
        return;
    }
    
    try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/config/wireguard/apply', {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        const result = await response.json();
        if (response.ok) {
            alert('配置应用成功！');
        } else {
            alert('应用失败：' + result.message);
        }
    } catch (error) {
        console.error('应用配置失败:', error);
        alert('网络错误，请重试');
    }
}

// 用户VPN管理（已经整合到模块管理中）
function showUserVPNManager() {
    alert('用户VPN管理功能已整合到模块管理中！\n\n请在模块列表中点击"管理用户"按钮来管理模块的用户VPN配置。');
}

// 全局导出系统配置管理函数
window.showWireGuardConfig = showWireGuardConfig;
window.viewInterfaceDetails = viewInterfaceDetails;
window.viewInterfaceConfig = viewInterfaceConfig;
window.refreshSystemConfig = refreshSystemConfig;
window.createNewInterface = createNewInterface;
window.exportSystemConfig = exportSystemConfig;
window.viewSystemLogs = viewSystemLogs;
window.systemSettings = systemSettings;
window.initializeWireGuard = initializeWireGuard;
window.downloadServerConfig = downloadServerConfig;
window.applyWireGuardConfig = applyWireGuardConfig;
window.showUserVPNManager = showUserVPNManager; 