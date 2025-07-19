-- EITEC VPN 测试数据脚本
-- 清空现有数据
DELETE FROM modules;
DELETE FROM wire_guard_interfaces;
DELETE FROM ip_pools;
DELETE FROM system_configs;
DELETE FROM users;

-- 重置自增ID
DELETE FROM sqlite_sequence WHERE name IN ('modules', 'wire_guard_interfaces', 'ip_pools', 'system_configs', 'users');

-- 1. 插入WireGuard接口数据
INSERT INTO wire_guard_interfaces (
    id, name, description, network, server_ip, listen_port, 
    public_key, private_key, status, max_peers, dns, mtu, 
    post_up, post_down, save_config, total_peers, active_peers, 
    total_traffic, created_at, updated_at
) VALUES 
-- 主接口 - 生产环境
(1, 'wg0', '主接口 - 生产环境', '10.10.0.0/24', '10.10.0.1', 51820,
 'K8zZGtlFJdGvB9p8XQmN7wR5vL2sM3nO4pQ6rT7uV8w=', 'sA1bC2dE3fG4hI5jK6lM7nO8pQ9rS0tU1vW2xY3zA4B=',
 1, 100, '8.8.8.8,8.8.4.4', 1420,
 'iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE',
 'iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE',
 1, 8, 6, 1048576000, datetime('now'), datetime('now')),

-- 北京节点
(2, 'wg1', '北京节点专用', '10.11.0.0/24', '10.11.0.1', 51821,
 'L9aZHumGKeHwC0q9YRnO8xS6wM3tN4oP5qR7sU8vW9x=', 'tB2cD3eF4gH5iJ6kL7mN8oP9qR0sT1uV2wX3yZ4aB5C=',
 1, 50, '8.8.8.8,8.8.4.4', 1420,
 'iptables -A FORWARD -i wg1 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE',
 'iptables -D FORWARD -i wg1 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE',
 1, 12, 10, 2097152000, datetime('now'), datetime('now')),

-- 上海节点
(3, 'wg2', '上海节点专用', '10.12.0.0/24', '10.12.0.1', 51822,
 'M0bZIvnHLfIxD1r0ZSoP9yT7xN4uO5pQ6rS8tV9wX0y=', 'uC3dE4fG5hI6jK7lM8nO9pQ0rS1tU2vW3xY4zA5bC6D=',
 1, 50, '8.8.8.8,8.8.4.4', 1420,
 'iptables -A FORWARD -i wg2 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE',
 'iptables -D FORWARD -i wg2 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE',
 1, 5, 4, 524288000, datetime('now'), datetime('now')),

-- 测试环境
(4, 'wg99', '测试环境专用', '10.99.0.0/24', '10.99.0.1', 51899,
 'N1cZJwoIMgJyE2s1aTpQ0zU8yO5vP6qR7sT9uW0xY1z=', 'vD4eF5gH6iJ7kL8mN9oP0qR1sT2uV3wX4yZ5aB6cD7E=',
 0, 10, '8.8.8.8,8.8.4.4', 1420,
 'iptables -A FORWARD -i wg99 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE',
 'iptables -D FORWARD -i wg99 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE',
 1, 2, 1, 10485760, datetime('now'), datetime('now'));

-- 2. 插入模块数据
INSERT INTO modules (
    id, name, location, description, interface_id, public_key, private_key,
    ip_address, status, last_seen, total_tx_bytes, total_rx_bytes,
    latest_handshake, allowed_ips, persistent_ka, created_at, updated_at
) VALUES 
-- wg0 接口的模块
(1, 'beijing-node-01', '北京-阿里云', '北京主节点服务器', 1,
 'A1bC2dE3fG4hI5jK6lM7nO8pQ9rS0tU1vW2xY3zA4B=', 'K8zZGtlFJdGvB9p8XQmN7wR5vL2sM3nO4pQ6rT7uV8w=',
 '10.10.0.2', 1, datetime('now', '-2 minutes'), 524288000, 1048576000,
 datetime('now', '-1 minute'), '192.168.1.0/24', 25, datetime('now', '-2 hours'), datetime('now')),

(2, 'shanghai-node-01', '上海-腾讯云', '上海主节点服务器', 1,
 'B2cD3eF4gH5iJ6kL7mN8oP9qR0sT1uV2wX3yZ4aB5C=', 'L9aZHumGKeHwC0q9YRnO8xS6wM3tN4oP5qR7sU8vW9x=',
 '10.10.0.3', 1, datetime('now', '-1 minute'), 1048576000, 2097152000,
 datetime('now', '-30 seconds'), '192.168.2.0/24', 25, datetime('now', '-3 hours'), datetime('now')),

(3, 'guangzhou-node-01', '广州-华为云', '广州节点服务器', 1,
 'C3dE4fG5hI6jK7lM8nO9pQ0rS1tU2vW3xY4zA5bC6D=', 'M0bZIvnHLfIxD1r0ZSoP9yT7xN4uO5pQ6rS8tV9wX0y=',
 '10.10.0.4', 1, datetime('now', '-5 minutes'), 268435456, 536870912,
 datetime('now', '-2 minutes'), '192.168.3.0/24', 25, datetime('now', '-1 hour'), datetime('now')),

(4, 'shenzhen-node-01', '深圳-百度云', '深圳备用节点', 1,
 'D4eF5gH6iJ7kL8mN9oP0qR1sT2uV3wX4yZ5aB6cD7E=', 'N1cZJwoIMgJyE2s1aTpQ0zU8yO5vP6qR7sT9uW0xY1z=',
 '10.10.0.5', 2, datetime('now', '-30 minutes'), 134217728, 268435456,
 datetime('now', '-15 minutes'), '192.168.4.0/24', 25, datetime('now', '-4 hours'), datetime('now')),

(5, 'hangzhou-node-01', '杭州-金山云', '杭州节点服务器', 1,
 'E5fG6hI7jK8lM9nO0pQ1rS2tU3vW4xY5zA6bC7dE8F=', 'O2dZKxpJNhKzF3t2bUqR1aV9zP6wQ7rS8tU0vX1yZ2a=',
 '10.10.0.6', 1, datetime('now', '-3 minutes'), 67108864, 134217728,
 datetime('now', '-1 minute'), '192.168.5.0/24', 25, datetime('now', '-2 hours'), datetime('now')),

(6, 'nanjing-node-01', '南京-京东云', '南京节点服务器', 1,
 'F6gH7iJ8kL9mN0oP1qR2sT3uV4wX5yZ6aB7cD8eF9G=', 'P3eZLyqKOiLaG4u3cVrS2bW0aQ7xR8sT9uV1wY2zA3b=',
 '10.10.0.7', 0, datetime('now', '-45 minutes'), 33554432, 67108864,
 datetime('now', '-30 minutes'), '192.168.6.0/24', 25, datetime('now', '-5 hours'), datetime('now')),

(7, 'chengdu-node-01', '成都-移动云', '成都节点服务器', 1,
 'G7hI8jK9lM0nO1pQ2rS3tU4vW5xY6zA7bC8dE9fG0H=', 'Q4fZMzrLPjMbH5v4dWsT3cX1bR8yS9tU0vW2xZ3aB4c=',
 '10.10.0.8', 3, datetime('now', '-1 hour'), 16777216, 33554432,
 datetime('now', '-45 minutes'), '192.168.7.0/24', 25, datetime('now', '-6 hours'), datetime('now')),

-- wg1 接口的模块
(8, 'beijing-edge-01', '北京-边缘节点1', '北京边缘计算节点', 2,
 'H8iJ9kL0mN1oP2qR3sT4uV5wX6yZ7aB8cD9eF0gH1I=', 'R5gZN0sMP kNcI6w5eXtU4dY2cS9zT0uV1wX3yZ4aB5d=',
 '10.11.0.2', 1, datetime('now', '-1 minute'), 1073741824, 2147483648,
 datetime('now', '-30 seconds'), '192.168.11.0/24', 25, datetime('now', '-1 hour'), datetime('now')),

(9, 'beijing-edge-02', '北京-边缘节点2', '北京边缘计算节点备份', 2,
 'I9jK0lM1nO2pQ3rS4tU5vW6xY7zA8bC9dE0fG1hI2J=', 'S6hZO1tNQlOdJ7x6fYuV5eZ3dT0aU1vW2xY4zA5bC6e=',
 '10.11.0.3', 1, datetime('now', '-2 minutes'), 536870912, 1073741824,
 datetime('now', '-1 minute'), '192.168.12.0/24', 25, datetime('now', '-2 hours'), datetime('now')),

-- wg2 接口的模块
(10, 'shanghai-edge-01', '上海-边缘节点1', '上海边缘计算节点', 3,
 'J0kL1mN2oP3qR4sT5uV6wX7yZ8aB9cD0eF1gH2iJ3K=', 'T7iZP2uOQmPeK8y7gZvW6fA4eU1bV2wX3yZ5aB6cD7f=',
 '10.12.0.2', 1, datetime('now', '-3 minutes'), 268435456, 536870912,
 datetime('now', '-2 minutes'), '192.168.21.0/24', 25, datetime('now', '-1 hour'), datetime('now')),

-- wg99 测试接口的模块
(11, 'test-node-01', '测试环境-节点1', '开发测试节点', 4,
 'K1lM2nO3pQ4rS5tU6vW7xY8zA9bC0dE1fG2hI3jK4L=', 'U8jZQ3vPRnQfL9z8hAwX7gB5fV2cW3xY4zA6bC7dE8g=',
 '10.99.0.2', 1, datetime('now', '-10 minutes'), 10485760, 20971520,
 datetime('now', '-5 minutes'), '192.168.99.0/24', 25, datetime('now', '-30 minutes'), datetime('now'));

-- 3. 插入IP池数据
-- wg0 网络的IP池
INSERT INTO ip_pools (network, ip_address, is_used, module_id) VALUES 
('10.10.0.0/24', '10.10.0.2', 1, 1),
('10.10.0.0/24', '10.10.0.3', 1, 2),
('10.10.0.0/24', '10.10.0.4', 1, 3),
('10.10.0.0/24', '10.10.0.5', 1, 4),
('10.10.0.0/24', '10.10.0.6', 1, 5),
('10.10.0.0/24', '10.10.0.7', 1, 6),
('10.10.0.0/24', '10.10.0.8', 1, 7),
('10.10.0.0/24', '10.10.0.9', 0, NULL),
('10.10.0.0/24', '10.10.0.10', 0, NULL);

-- wg1 网络的IP池
INSERT INTO ip_pools (network, ip_address, is_used, module_id) VALUES 
('10.11.0.0/24', '10.11.0.2', 1, 8),
('10.11.0.0/24', '10.11.0.3', 1, 9),
('10.11.0.0/24', '10.11.0.4', 0, NULL),
('10.11.0.0/24', '10.11.0.5', 0, NULL);

-- wg2 网络的IP池
INSERT INTO ip_pools (network, ip_address, is_used, module_id) VALUES 
('10.12.0.0/24', '10.12.0.2', 1, 10),
('10.12.0.0/24', '10.12.0.3', 0, NULL),
('10.12.0.0/24', '10.12.0.4', 0, NULL);

-- wg99 测试网络的IP池
INSERT INTO ip_pools (network, ip_address, is_used, module_id) VALUES 
('10.99.0.0/24', '10.99.0.2', 1, 11),
('10.99.0.0/24', '10.99.0.3', 0, NULL);

-- 4. 插入系统配置
INSERT INTO system_configs (key, value, type, comment) VALUES 
('server.public_key', 'K8zZGtlFJdGvB9p8XQmN7wR5vL2sM3nO4pQ6rT7uV8w=', 'string', '服务器WireGuard公钥'),
('server.private_key', 'sA1bC2dE3fG4hI5jK6lM7nO8pQ9rS0tU1vW2xY3zA4B=', 'string', '服务器WireGuard私钥'),
('server.endpoint', 'vpn.eitec.com:51820', 'string', '服务器公网地址:端口'),
('server.name', 'EITEC VPN Server', 'string', '服务器名称'),
('server.web_port', '8070', 'string', 'Web管理端口'),
('wg.network', '10.10.0.0/24', 'string', 'WireGuard网络段'),
('wg.dns', '8.8.8.8,8.8.4.4', 'string', 'DNS服务器'),
('wg.interface', 'wg0', 'string', '默认WireGuard接口'),
('wg.listen_port', '51820', 'string', '默认监听端口'),
('wg.mtu', '1420', 'string', 'MTU大小'),
('network.ip_pool_start', '10.10.0.2', 'string', 'IP池起始地址'),
('network.ip_pool_end', '10.10.0.254', 'string', 'IP池结束地址'),
('network.default_gateway', '10.10.0.1', 'string', '默认网关'),
('network.enable_nat', 'true', 'string', '启用NAT转发'),
('security.max_peers', '100', 'string', '最大连接数'),
('security.enable_logging', 'true', 'string', '启用日志记录'),
('monitoring.heartbeat_interval', '60', 'string', '心跳检测间隔(秒)'),
('monitoring.offline_timeout', '300', 'string', '离线超时时间(秒)');

-- 5. 插入用户数据
INSERT INTO users (username, password, is_active, created_at, updated_at) VALUES 
('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMye.Uo/QqvJ6.jqKCL8eDH6VJjGPNg4/0u', 1, datetime('now'), datetime('now')); -- 密码: admin123

-- 查询验证数据
SELECT 'WireGuard接口数量:' as info, COUNT(*) as count FROM wire_guard_interfaces
UNION ALL
SELECT '模块数量:', COUNT(*) FROM modules
UNION ALL
SELECT 'IP池数量:', COUNT(*) FROM ip_pools
UNION ALL
SELECT '系统配置数量:', COUNT(*) FROM system_configs
UNION ALL
SELECT '用户数量:', COUNT(*) FROM users; 