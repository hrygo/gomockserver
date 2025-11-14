// MongoDB 测试数据库初始化脚本

// 创建测试数据库
db = db.getSiblingDB('mockserver_test');

// 创建索引
db.projects.createIndex({ "workspace_id": 1 });
db.projects.createIndex({ "created_at": -1 });

db.environments.createIndex({ "project_id": 1 });
db.environments.createIndex({ "project_id": 1, "name": 1 }, { unique: true });

db.rules.createIndex({ "project_id": 1, "environment_id": 1 });
db.rules.createIndex({ "project_id": 1, "environment_id": 1, "enabled": 1 });
db.rules.createIndex({ "project_id": 1, "environment_id": 1, "priority": -1 });

db.request_logs.createIndex({ "project_id": 1, "timestamp": -1 });
db.request_logs.createIndex({ "timestamp": -1 }, { expireAfterSeconds: 259200 }); // 3天自动删除

print('Test database initialized successfully');
