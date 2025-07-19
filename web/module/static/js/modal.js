// 模态框相关JavaScript文件

// 配置模态框相关函数
function openConfigModal() {
    document.getElementById('configModal').style.display = 'flex';
    loadCurrentConfig();
}

function closeConfigModal() {
    document.getElementById('configModal').style.display = 'none';
    clearConfigAlert();
}

// 点击模态框外部关闭
window.onclick = function(event) {
    const modal = document.getElementById('configModal');
    if (event.target === modal) {
        closeConfigModal();
    }
}

// 加载当前配置
async function loadCurrentConfig() {
    try {
        const headers = createAuthHeaders();
        const response = await fetch('/api/v1/config', { headers });
        if (response.status === 401) {
            handleAuthError();
            return;
        }
        if (response.ok) {
            const result = await response.json();
            if (result.success && result.data) {
                document.getElementById('moduleId').value = result.data.module_id || '';
                document.getElementById('apiKey').value = result.data.api_key || '';
                document.getElementById('serverUrl').value = result.data.server_url || '';
                document.getElementById('configData').value = result.data.config_data || '';
            }
        }
    } catch (error) {
        console.error('加载配置失败:', error);
    }
}

// 显示配置提示信息
function showConfigAlert(message, type = 'info') {
    const alertDiv = document.getElementById('configAlert');
    alertDiv.className = `alert alert-${type}`;
    alertDiv.textContent = message;
    alertDiv.style.display = 'block';
}

// 清除配置提示信息
function clearConfigAlert() {
    const alertDiv = document.getElementById('configAlert');
    alertDiv.style.display = 'none';
}

// 验证WireGuard配置
function validateWireGuardConfig(configData) {
    if (!configData || configData.trim() === '') {
        return { valid: false, message: '配置内容不能为空' };
    }

    const requiredSections = ['[Interface]', '[Peer]'];
    const requiredKeys = ['PrivateKey', 'Address'];
    
    // 检查必需的节
    for (const section of requiredSections) {
        if (!configData.includes(section)) {
            return { valid: false, message: `缺少必需的配置节: ${section}` };
        }
    }
    
    // 检查必需的键
    for (const key of requiredKeys) {
        if (!configData.includes(key)) {
            return { valid: false, message: `缺少必需的配置项: ${key}` };
        }
    }
    
    return { valid: true };
}

// 配置表单提交
document.addEventListener('DOMContentLoaded', function() {
    const configForm = document.getElementById('configForm');
    if (configForm) {
        configForm.addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const formData = new FormData(this);
            const configData = {
                module_id: parseInt(formData.get('module_id')),
                api_key: formData.get('api_key'),
                server_url: formData.get('server_url'),
                config_data: formData.get('config_data')
            };
            
            // 验证配置
            const validation = validateWireGuardConfig(configData.config_data);
            if (!validation.valid) {
                showConfigAlert(validation.message, 'error');
                return;
            }
            
            showConfigAlert('正在保存配置...', 'info');
            
            try {
                const headers = createAuthHeaders();
                const response = await fetch('/api/v1/config', {
                    method: 'POST',
                    headers,
                    body: JSON.stringify(configData)
                });
                
                if (response.status === 401) {
                    handleAuthError();
                    return;
                }
                
                const result = await response.json();
                if (result.success) {
                    showConfigAlert('配置保存成功！', 'success');
                    setTimeout(() => {
                        closeConfigModal();
                        refreshData(); // 刷新主页面数据
                    }, 1500);
                } else {
                    showConfigAlert(result.message || '配置保存失败', 'error');
                }
            } catch (error) {
                console.error('配置保存错误:', error);
                showConfigAlert('配置保存失败，请检查网络连接', 'error');
            }
        });
    }
});