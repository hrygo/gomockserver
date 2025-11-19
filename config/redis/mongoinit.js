// MongoDB初始化脚本 - MockServer
// 在MongoDB启动后创建基础集合和索引

// 切换到mockserver数据库
db = db.getSiblingDB('mockserver');

print('正在初始化MongoDB数据库...');

// 创建项目集合
if (!db.projects.getIndexes().some(index => index.name === 'name_1')) {
    db.projects.createIndex({ "name": 1 }, { unique: true });
    print('已创建projects集合的name索引');
}

// 创建规则集合
if (!db.rules.getIndexes().some(index => index.name === 'project_id_1')) {
    db.rules.createIndex({ "projectId": 1 });
    print('已创建rules集合的projectId索引');
}

// 创建缓存相关的集合（如果需要）
if (!db.cache_stats.getIndexes().some(index => index.name === 'timestamp_-1')) {
    db.cache_stats.createIndex({ "timestamp": -1 });
    print('已创建cache_stats集合的timestamp索引');
}

// 插入一些初始化数据
db.system.insertOne({
    key: "initialized",
    value: true,
    timestamp: new Date(),
    version: "1.0.0"
});

print('MongoDB初始化完成！');