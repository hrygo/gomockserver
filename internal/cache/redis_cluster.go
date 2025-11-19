package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisCluster Redis集群客户端
type RedisCluster struct {
	clients          []*redis.Client
	currentCluster   *redis.ClusterClient
	config           *RedisClusterConfig
	mu               sync.RWMutex
	healthChecker    *ClusterHealthChecker
	connectionPool   *ConnectionPool
	failureDetector  *FailureDetector
	loadBalancer     *LoadBalancer
	logger           *zap.Logger
	isHealthy        bool
	lastHealthCheck  time.Time
}

// RedisClusterConfig Redis集群配置
type RedisClusterConfig struct {
	// 集群节点配置
	Nodes            []RedisNode `json:"nodes"`
	Password         string      `json:"password"`
	Username         string      `json:"username"`

	// 连接配置
	PoolSize         int         `json:"pool_size"`
	MinIdleConns     int         `json:"min_idle_conns"`
	MaxIdleConns     int         `json:"max_idle_conns"`
	MaxRetries       int         `json:"max_retries"`
	DialTimeout      time.Duration `json:"dial_timeout"`
	ReadTimeout      time.Duration `json:"read_timeout"`
	WriteTimeout     time.Duration `json:"write_timeout"`
	PoolTimeout      time.Duration `json:"pool_timeout"`

	// 集群配置
	MaxRedirects     int         `json:"max_redirects"`
	RouteByLatency   bool        `json:"route_by_latency"`
	RouteRandomly    bool        `json:"route_randomly"`

	// 健康检查配置
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	HealthCheckTimeout  time.Duration `json:"health_check_timeout"`
	MaxFailures        int           `json:"max_failures"`
	FailureBackoff     time.Duration `json:"failure_backoff"`

	// 负载均衡配置
	LoadBalanceStrategy string `json:"load_balance_strategy"` // round_robin, weighted, least_connections

	// TLS配置
	TLSConfig         *TLSConfig `json:"tls_config"`

	// 分片配置
	ShardStrategy     string `json:"shard_strategy"` // consistent_hash, range, mod
	ReplicaFactor     int    `json:"replica_factor"`
}

// RedisNode Redis节点
type RedisNode struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Weight   int    `json:"weight"`
	Role     string `json:"role"` // master, slave
	IsAlive  bool   `json:"is_alive"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled            bool   `json:"enabled"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
	CertFile           string `json:"cert_file"`
	KeyFile            string `json:"key_file"`
	CAFile             string `json:"ca_file"`
}

// ClusterHealthChecker 集群健康检查器
type ClusterHealthChecker struct {
	config      *RedisClusterConfig
	nodes       map[string]*RedisNode
	mu          sync.RWMutex
	logger      *zap.Logger
	stopCh      chan struct{}
	isRunning   bool
}

// ConnectionPool 连接池
type ConnectionPool struct {
	config      *RedisClusterConfig
	connections map[string]*redis.Client
	mu          sync.RWMutex
	logger      *zap.Logger
}

// FailureDetector 故障检测器
type FailureDetector struct {
	config      *RedisClusterConfig
	failures    map[string]int
	lastFailure map[string]time.Time
	mu          sync.RWMutex
	logger      *zap.Logger
}

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	config      *RedisClusterConfig
	nodes       []*RedisNode
	currentIndex int
	mu          sync.RWMutex
	logger      *zap.Logger
}

// DefaultRedisClusterConfig 默认Redis集群配置
func DefaultRedisClusterConfig() *RedisClusterConfig {
	return &RedisClusterConfig{
		Nodes: []RedisNode{
			{Host: "localhost", Port: 6379, Weight: 1, Role: "master"},
			{Host: "localhost", Port: 6380, Weight: 1, Role: "slave"},
			{Host: "localhost", Port: 6381, Weight: 1, Role: "slave"},
		},
		PoolSize:         10,
		MinIdleConns:     5,
		MaxIdleConns:     20,
		MaxRetries:       3,
		DialTimeout:      5 * time.Second,
		ReadTimeout:      3 * time.Second,
		WriteTimeout:     3 * time.Second,
		PoolTimeout:      4 * time.Second,
		MaxRedirects:     3,
		RouteByLatency:   true,
		RouteRandomly:    false,
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		MaxFailures:        3,
		FailureBackoff:     60 * time.Second,
		LoadBalanceStrategy: "round_robin",
		TLSConfig:         &TLSConfig{Enabled: false},
		ShardStrategy:     "consistent_hash",
		ReplicaFactor:     1,
	}
}

// NewRedisCluster 创建Redis集群客户端
func NewRedisCluster(config *RedisClusterConfig, logger *zap.Logger) *RedisCluster {
	if config == nil {
		config = DefaultRedisClusterConfig()
	}

	cluster := &RedisCluster{
		config:          config,
		clients:         make([]*redis.Client, 0),
		logger:          logger.Named("redis_cluster"),
		isHealthy:       true,
		lastHealthCheck: time.Now(),
	}

	// 初始化健康检查器
	cluster.healthChecker = &ClusterHealthChecker{
		config:    config,
		nodes:     make(map[string]*RedisNode),
		logger:    logger.Named("health_checker"),
		stopCh:    make(chan struct{}),
		isRunning: false,
	}

	// 初始化连接池
	cluster.connectionPool = &ConnectionPool{
		config:      config,
		connections: make(map[string]*redis.Client),
		logger:      logger.Named("connection_pool"),
	}

	// 初始化故障检测器
	cluster.failureDetector = &FailureDetector{
		config:      config,
		failures:    make(map[string]int),
		lastFailure: make(map[string]time.Time),
		logger:      logger.Named("failure_detector"),
	}

	// 初始化负载均衡器
	cluster.loadBalancer = &LoadBalancer{
		config:       config,
		nodes:        make([]*RedisNode, 0),
		currentIndex: 0,
		logger:       logger.Named("load_balancer"),
	}

	// 初始化节点
	for _, node := range config.Nodes {
		cluster.healthChecker.nodes[node.Address()] = &node
		cluster.loadBalancer.nodes = append(cluster.loadBalancer.nodes, &node)
	}

	// 创建集群客户端
	cluster.createClusterClient()

	// 启动健康检查
	go cluster.healthChecker.Start()

	cluster.logger.Info("Redis cluster initialized",
		zap.Int("node_count", len(config.Nodes)),
		zap.String("load_balance_strategy", config.LoadBalanceStrategy),
		zap.String("shard_strategy", config.ShardStrategy),
	)

	return cluster
}

// Address 获取节点地址
func (n *RedisNode) Address() string {
	return fmt.Sprintf("%s:%d", n.Host, n.Port)
}

// createClusterClient 创建集群客户端
func (rc *RedisCluster) createClusterClient() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// 构建Redis集群选项
	clusterOptions := &redis.ClusterOptions{
		Addrs:           rc.getNodeAddresses(),
		Password:        rc.config.Password,
		Username:        rc.config.Username,
		MaxRedirects:    rc.config.MaxRedirects,
		RouteByLatency:  rc.config.RouteByLatency,
		RouteRandomly:   rc.config.RouteRandomly,
		DialTimeout:     rc.config.DialTimeout,
		ReadTimeout:     rc.config.ReadTimeout,
		WriteTimeout:    rc.config.WriteTimeout,
		PoolSize:        rc.config.PoolSize,
		MinIdleConns:    rc.config.MinIdleConns,
		PoolTimeout:     rc.config.PoolTimeout,
		MaxRetries:      rc.config.MaxRetries,
	}

	// TLS配置
	if rc.config.TLSConfig.Enabled {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: rc.config.TLSConfig.InsecureSkipVerify,
		}
		clusterOptions.TLSConfig = tlsConfig
	}

	// 创建集群客户端
	rc.currentCluster = redis.NewClusterClient(clusterOptions)

	// 创建单个客户端用于特定操作
	rc.createIndividualClients()
}

// createIndividualClients 创建单个客户端
func (rc *RedisCluster) createIndividualClients() {
	rc.clients = make([]*redis.Client, 0)

	for _, node := range rc.config.Nodes {
		clientOptions := &redis.Options{
			Addr:         node.Address(),
			Password:     rc.config.Password,
			Username:     rc.config.Username,
			DB:           0,
			DialTimeout:  rc.config.DialTimeout,
			ReadTimeout:  rc.config.ReadTimeout,
			WriteTimeout: rc.config.WriteTimeout,
			PoolSize:     rc.config.PoolSize / len(rc.config.Nodes),
			MinIdleConns: rc.config.MinIdleConns / len(rc.config.Nodes),
		PoolTimeout:  rc.config.PoolTimeout,
		MaxRetries:   rc.config.MaxRetries,
		}

		// TLS配置
		if rc.config.TLSConfig.Enabled {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: rc.config.TLSConfig.InsecureSkipVerify,
			}
			clientOptions.TLSConfig = tlsConfig
		}

		client := redis.NewClient(clientOptions)
		rc.clients = append(rc.clients, client)
		rc.connectionPool.connections[node.Address()] = client
	}
}

// getNodeAddresses 获取节点地址列表
func (rc *RedisCluster) getNodeAddresses() []string {
	addresses := make([]string, len(rc.config.Nodes))
	for i, node := range rc.config.Nodes {
		addresses[i] = node.Address()
	}
	return addresses
}

// Get 获取缓存值
func (rc *RedisCluster) Get(ctx context.Context, key string) (string, error) {
	// 选择最优节点
	node := rc.loadBalancer.SelectNode(key)
	if node == nil {
		return "", fmt.Errorf("no available nodes")
	}

	// 获取对应客户端
	client := rc.connectionPool.connections[node.Address()]
	if client == nil {
		return "", fmt.Errorf("client not available for node %s", node.Address())
	}

	// 记录成功或失败
	defer func() {
		if err := recover(); err != nil {
			rc.failureDetector.RecordFailure(node.Address())
		}
	}()

	// 尝试读取
	value, err := client.Get(ctx, key).Result()
	if err != nil {
		rc.failureDetector.RecordFailure(node.Address())
		// 如果失败，尝试从集群客户端读取
		return rc.currentCluster.Get(ctx, key).Result()
	}

	// 记录成功
	rc.failureDetector.RecordSuccess(node.Address())
	return value, nil
}

// Set 设置缓存值
func (rc *RedisCluster) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// 选择主节点
	node := rc.selectMasterNode()
	if node == nil {
		return fmt.Errorf("no master nodes available")
	}

	// 获取对应客户端
	client := rc.connectionPool.connections[node.Address()]
	if client == nil {
		return fmt.Errorf("client not available for node %s", node.Address())
	}

	// 执行设置
	err := client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		rc.failureDetector.RecordFailure(node.Address())
		// 如果失败，尝试从集群客户端写入
		return rc.currentCluster.Set(ctx, key, value, expiration).Err()
	}

	rc.failureDetector.RecordSuccess(node.Address())
	return nil
}

// Delete 删除缓存值
func (rc *RedisCluster) Delete(ctx context.Context, key string) error {
	// 选择主节点
	node := rc.selectMasterNode()
	if node == nil {
		return fmt.Errorf("no master nodes available")
	}

	client := rc.connectionPool.connections[node.Address()]
	if client == nil {
		return fmt.Errorf("client not available for node %s", node.Address())
	}

	err := client.Del(ctx, key).Err()
	if err != nil {
		rc.failureDetector.RecordFailure(node.Address())
		return rc.currentCluster.Del(ctx, key).Err()
	}

	rc.failureDetector.RecordSuccess(node.Address())
	return nil
}

// Exists 检查键是否存在
func (rc *RedisCluster) Exists(ctx context.Context, key string) (bool, error) {
	// 选择任意可用节点
	node := rc.loadBalancer.SelectNode(key)
	if node == nil {
		return false, fmt.Errorf("no available nodes")
	}

	client := rc.connectionPool.connections[node.Address()]
	if client == nil {
		return false, fmt.Errorf("client not available for node %s", node.Address())
	}

	result, err := client.Exists(ctx, key).Result()
	if err != nil {
		rc.failureDetector.RecordFailure(node.Address())
		clusterResult, clusterErr := rc.currentCluster.Exists(ctx, key).Result()
		return clusterResult > 0, clusterErr
	}

	rc.failureDetector.RecordSuccess(node.Address())
	return result > 0, nil
}

// selectMasterNode 选择主节点
func (rc *RedisCluster) selectMasterNode() *RedisNode {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	for i, node := range rc.config.Nodes {
		if node.Role == "master" && node.IsAlive {
			return &rc.config.Nodes[i]
		}
	}

	// 如果没有主节点，返回第一个可用节点
	for i, node := range rc.config.Nodes {
		if node.IsAlive {
			return &rc.config.Nodes[i]
		}
	}

	return nil
}

// GetClusterStats 获取集群统计信息
func (rc *RedisCluster) GetClusterStats(ctx context.Context) (*ClusterStats, error) {
	stats := &ClusterStats{
		NodeCount:    len(rc.config.Nodes),
		HealthyNodes: 0,
		UnhealthyNodes: 0,
		MasterNodes:   0,
		SlaveNodes:    0,
		Nodes:         make([]*NodeStats, 0),
		LastUpdate:    time.Now(),
	}

	// 收集各节点统计
	for _, node := range rc.config.Nodes {
		nodeStats := &NodeStats{
			Address: node.Address(),
			Role:    node.Role,
			Weight:  node.Weight,
			IsAlive: node.IsAlive,
		}

		// 获取客户端统计
		if client := rc.connectionPool.connections[node.Address()]; client != nil {
			poolStats := client.PoolStats()
			nodeStats.TotalConns = poolStats.TotalConns
			nodeStats.IdleConns = poolStats.IdleConns
			nodeStats.StaleConns = poolStats.StaleConns
		}

		stats.Nodes = append(stats.Nodes, nodeStats)

		if node.IsAlive {
			stats.HealthyNodes++
		} else {
			stats.UnhealthyNodes++
		}

		if node.Role == "master" {
			stats.MasterNodes++
		} else {
			stats.SlaveNodes++
		}
	}

	return stats, nil
}

// ClusterStats 集群统计信息
type ClusterStats struct {
	NodeCount       int          `json:"node_count"`
	HealthyNodes    int          `json:"healthy_nodes"`
	UnhealthyNodes  int          `json:"unhealthy_nodes"`
	MasterNodes     int          `json:"master_nodes"`
	SlaveNodes      int          `json:"slave_nodes"`
	Nodes           []*NodeStats `json:"nodes"`
	LastUpdate      time.Time    `json:"last_update"`
}

// NodeStats 节点统计信息
type NodeStats struct {
	Address        string `json:"address"`
	Role           string `json:"role"`
	Weight         int    `json:"weight"`
	IsAlive        bool   `json:"is_alive"`
	TotalConns     uint32 `json:"total_conns"`
	IdleConns      uint32 `json:"idle_conns"`
	StaleConns     uint32 `json:"stale_conns"`
}

// Start 启动Redis集群
func (rc *RedisCluster) Start(ctx context.Context) error {
	// 测试连接
	if err := rc.currentCluster.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis cluster: %w", err)
	}

	rc.logger.Info("Redis cluster started successfully")
	return nil
}

// Stop 停止Redis集群
func (rc *RedisCluster) Stop() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// 停止健康检查
	rc.healthChecker.Stop()

	// 关闭所有客户端连接
	if rc.currentCluster != nil {
		if err := rc.currentCluster.Close(); err != nil {
			rc.logger.Error("Error closing cluster client", zap.Error(err))
		}
	}

	for _, client := range rc.clients {
		if err := client.Close(); err != nil {
			rc.logger.Error("Error closing client", zap.Error(err))
		}
	}

	rc.logger.Info("Redis cluster stopped")
	return nil
}

// IsHealthy 检查集群是否健康
func (rc *RedisCluster) IsHealthy() bool {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	return rc.isHealthy
}

// Start 启动健康检查器
func (hc *ClusterHealthChecker) Start() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if hc.isRunning {
		return
	}

	hc.isRunning = true
	go hc.runHealthCheck()
	hc.logger.Info("Cluster health checker started")
}

// Stop 停止健康检查器
func (hc *ClusterHealthChecker) Stop() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if hc.isRunning {
		hc.isRunning = false
		close(hc.stopCh)
		hc.stopCh = make(chan struct{})
		hc.logger.Info("Cluster health checker stopped")
	}
}

// runHealthCheck 运行健康检查
func (hc *ClusterHealthChecker) runHealthCheck() {
	ticker := time.NewTicker(hc.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.stopCh:
			return
		case <-ticker.C:
			hc.checkAllNodes()
		}
	}
}

// checkAllNodes 检查所有节点健康状态
func (hc *ClusterHealthChecker) checkAllNodes() {
	for addr, node := range hc.nodes {
		go func(address string, n *RedisNode) {
			ctx, cancel := context.WithTimeout(context.Background(), hc.config.HealthCheckTimeout)
			defer cancel()

			client := redis.NewClient(&redis.Options{
				Addr:        address,
				Password:    hc.config.Password,
				Username:    hc.config.Username,
				DialTimeout: hc.config.HealthCheckTimeout,
			})

			// 执行ping检查
			err := client.Ping(ctx).Err()
			client.Close()

			hc.mu.Lock()
			n.IsAlive = (err == nil)
			hc.mu.Unlock()

			if err != nil {
				hc.logger.Warn("Node health check failed",
					zap.String("address", address),
					zap.Error(err))
			}
		}(addr, node)
	}
}

// RecordFailure 记录故障
func (fd *FailureDetector) RecordFailure(nodeAddr string) {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	fd.failures[nodeAddr]++
	fd.lastFailure[nodeAddr] = time.Now()
}

// RecordSuccess 记录成功
func (fd *FailureDetector) RecordSuccess(nodeAddr string) {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	// 重置失败计数
	if failures, exists := fd.failures[nodeAddr]; exists && failures > 0 {
		fd.failures[nodeAddr] = 0
	}
}

// SelectNode 选择节点
func (lb *LoadBalancer) SelectNode(key string) *RedisNode {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// 过滤健康的节点
	healthyNodes := make([]*RedisNode, 0)
	for _, node := range lb.nodes {
		if node.IsAlive {
			healthyNodes = append(healthyNodes, node)
		}
	}

	if len(healthyNodes) == 0 {
		return nil
	}

	// 根据策略选择节点
	switch lb.config.LoadBalanceStrategy {
	case "round_robin":
		node := healthyNodes[lb.currentIndex%len(healthyNodes)]
		lb.currentIndex++
		return node
	case "weighted":
		return lb.selectWeightedNode(healthyNodes, key)
	case "least_connections":
		return lb.selectLeastConnectionsNode(healthyNodes)
	default:
		// 默认随机选择
		return healthyNodes[rand.Intn(len(healthyNodes))]
	}
}

// selectWeightedNode 根据权重选择节点
func (lb *LoadBalancer) selectWeightedNode(nodes []*RedisNode, key string) *RedisNode {
	totalWeight := 0
	for _, node := range nodes {
		totalWeight += node.Weight
	}

	if totalWeight == 0 {
		return nodes[0]
	}

	// 使用key的哈希值作为随机种子
	hash := simpleHash(key)
	random := hash % totalWeight

	currentWeight := 0
	for _, node := range nodes {
		currentWeight += node.Weight
		if random < currentWeight {
			return node
		}
	}

	return nodes[0]
}

// selectLeastConnectionsNode 选择连接数最少的节点
func (lb *LoadBalancer) selectLeastConnectionsNode(nodes []*RedisNode) *RedisNode {
	// 简化实现，返回第一个健康节点
	return nodes[0]
}

// simpleHash 简单哈希函数
func simpleHash(key string) int {
	hash := 0
	for _, c := range key {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}