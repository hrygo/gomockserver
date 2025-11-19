package adapter

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/pkg/logger"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// MessageHandler 消息处理函数类型
type MessageHandler func(request *Request, conn *WebSocketConnection)

// WebSocketAdapter WebSocket 协议适配器
type WebSocketAdapter struct {
	upgrader websocket.Upgrader
	// 连接管理
	connections     map[string]*WebSocketConnection
	connectionsLock sync.RWMutex
	// 配置
	maxConnections int
	pingInterval   time.Duration
	pongWait       time.Duration
	writeWait      time.Duration
	maxMessageSize int64
	// 消息处理器
	messageHandler MessageHandler
}

// WebSocketConnection WebSocket 连接
type WebSocketConnection struct {
	ID        string
	Conn      *websocket.Conn
	ProjectID string
	EnvID     string
	Send      chan []byte
	Done      chan struct{}
	LastPing  time.Time
	LastPong  time.Time
	Metadata  map[string]interface{}
	mu        sync.RWMutex
}

// WebSocketMessage WebSocket 消息
type WebSocketMessage struct {
	Type      string                 `json:"type"`      // message, ping, pong, close
	Data      interface{}            `json:"data"`      // 消息数据
	Timestamp time.Time              `json:"timestamp"` // 时间戳
	Metadata  map[string]interface{} `json:"metadata"`  // 元数据
}

// NewWebSocketAdapter 创建 WebSocket 适配器
func NewWebSocketAdapter() *WebSocketAdapter {
	return &WebSocketAdapter{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// 允许所有来源（生产环境需要根据需求限制）
				return true
			},
		},
		connections:    make(map[string]*WebSocketConnection),
		maxConnections: 1000,             // 默认最大连接数
		pingInterval:   30 * time.Second, // Ping 间隔
		pongWait:       60 * time.Second, // Pong 等待时间
		writeWait:      10 * time.Second, // 写超时
		maxMessageSize: 512 * 1024,       // 最大消息大小 512KB
	}
}

// Parse 解析 WebSocket 请求为统一模型
func (a *WebSocketAdapter) Parse(rawRequest interface{}) (*Request, error) {
	c, ok := rawRequest.(*gin.Context)
	if !ok {
		return nil, nil
	}

	// 升级 HTTP 连接为 WebSocket
	conn, err := a.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("failed to upgrade websocket", zap.Error(err))
		return nil, err
	}

	// 生成连接 ID
	connID := uuid.New().String()

	// 创建连接对象
	wsConn := &WebSocketConnection{
		ID:       connID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Done:     make(chan struct{}),
		LastPing: time.Now(),
		LastPong: time.Now(),
		Metadata: make(map[string]interface{}),
	}

	// 从 URL 参数获取项目和环境 ID
	projectID := c.Param("projectID")
	envID := c.Param("environmentID")
	wsConn.ProjectID = projectID
	wsConn.EnvID = envID

	// 保存连接
	a.addConnection(wsConn)

	// 启动读写协程
	go a.readPump(wsConn)
	go a.writePump(wsConn)
	go a.pingPump(wsConn)

	// 创建统一请求模型（用于首次连接）
	request := &Request{
		ID:         connID,
		Protocol:   models.ProtocolWebSocket,
		Path:       c.Request.URL.Path,
		Headers:    extractHeaders(c.Request.Header),
		SourceIP:   c.ClientIP(),
		ReceivedAt: time.Now(),
		Metadata: map[string]interface{}{
			"connection_id":  connID,
			"project_id":     projectID,
			"environment_id": envID,
			"query":          extractQuery(c.Request.URL.Query()),
			"event":          "connect",
		},
	}

	logger.Info("websocket connection established",
		zap.String("connection_id", connID),
		zap.String("project_id", projectID),
		zap.String("environment_id", envID),
	)

	return request, nil
}

// Build 构建 WebSocket 响应
func (a *WebSocketAdapter) Build(response *Response) (interface{}, error) {
	// WebSocket 响应通过 Send channel 发送，这里不需要处理
	return response, nil
}

// addConnection 添加连接
func (a *WebSocketAdapter) addConnection(conn *WebSocketConnection) error {
	a.connectionsLock.Lock()
	defer a.connectionsLock.Unlock()

	// 检查连接数限制
	if len(a.connections) >= a.maxConnections {
		logger.Warn("max connections reached", zap.Int("max", a.maxConnections))
		if conn.Conn != nil {
			conn.Conn.Close()
		}
		return ErrMaxConnectionsReached
	}

	a.connections[conn.ID] = conn
	return nil
}

// removeConnection 移除连接
func (a *WebSocketAdapter) removeConnection(connID string) {
	a.connectionsLock.Lock()
	defer a.connectionsLock.Unlock()

	if conn, exists := a.connections[connID]; exists {
		close(conn.Done)
		close(conn.Send)
		delete(a.connections, connID)
		logger.Info("websocket connection removed", zap.String("connection_id", connID))
	}
}

// GetConnection 获取连接
func (a *WebSocketAdapter) GetConnection(connID string) (*WebSocketConnection, bool) {
	a.connectionsLock.RLock()
	defer a.connectionsLock.RUnlock()
	conn, exists := a.connections[connID]
	return conn, exists
}

// GetConnectionCount 获取当前连接数
func (a *WebSocketAdapter) GetConnectionCount() int {
	a.connectionsLock.RLock()
	defer a.connectionsLock.RUnlock()
	return len(a.connections)
}

// BroadcastMessage 广播消息到所有连接
func (a *WebSocketAdapter) BroadcastMessage(message []byte) {
	a.connectionsLock.RLock()
	defer a.connectionsLock.RUnlock()

	for _, conn := range a.connections {
		select {
		case conn.Send <- message:
		default:
			logger.Warn("failed to send broadcast message, channel full",
				zap.String("connection_id", conn.ID))
		}
	}
}

// SendToConnection 发送消息到指定连接
func (a *WebSocketAdapter) SendToConnection(connID string, message []byte) error {
	conn, exists := a.GetConnection(connID)
	if !exists {
		return ErrConnectionNotFound
	}

	select {
	case conn.Send <- message:
		return nil
	case <-time.After(a.writeWait):
		return ErrSendTimeout
	}
}

// SetMessageHandler 设置消息处理器
func (a *WebSocketAdapter) SetMessageHandler(handler MessageHandler) {
	a.messageHandler = handler
}

// readPump 从 WebSocket 读取消息
func (a *WebSocketAdapter) readPump(conn *WebSocketConnection) {
	defer func() {
		a.removeConnection(conn.ID)
		conn.Conn.Close()
	}()

	conn.Conn.SetReadLimit(a.maxMessageSize)
	conn.Conn.SetReadDeadline(time.Now().Add(a.pongWait))
	conn.Conn.SetPongHandler(func(string) error {
		conn.mu.Lock()
		conn.LastPong = time.Now()
		conn.mu.Unlock()
		conn.Conn.SetReadDeadline(time.Now().Add(a.pongWait))
		return nil
	})

	for {
		messageType, message, err := conn.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("websocket read error", zap.Error(err))
			}
			break
		}

		// 处理接收到的消息
		logger.Debug("websocket message received",
			zap.String("connection_id", conn.ID),
			zap.Int("type", messageType),
			zap.Int("size", len(message)),
		)

		// 将消息传递给消息处理器
		if a.messageHandler != nil {
			request := a.createMessageRequest(conn, message)
			a.messageHandler(request, conn)
		}
	}
}

// createMessageRequest 创建消息请求
func (a *WebSocketAdapter) createMessageRequest(conn *WebSocketConnection, message []byte) *Request {
	return &Request{
		ID:         uuid.New().String(),
		Protocol:   models.ProtocolWebSocket,
		Path:       "", // WebSocket 没有路径概念
		Body:       message,
		SourceIP:   "", // 从连接获取
		ReceivedAt: time.Now(),
		Metadata: map[string]interface{}{
			"connection_id":  conn.ID,
			"project_id":     conn.ProjectID,
			"environment_id": conn.EnvID,
			"event":          "message",
		},
	}
}

// writePump 向 WebSocket 写入消息
func (a *WebSocketAdapter) writePump(conn *WebSocketConnection) {
	ticker := time.NewTicker(a.pingInterval)
	defer func() {
		ticker.Stop()
		conn.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-conn.Send:
			conn.Conn.SetWriteDeadline(time.Now().Add(a.writeWait))
			if !ok {
				// Channel 已关闭
				conn.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := conn.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				logger.Error("websocket write error", zap.Error(err))
				return
			}

		case <-conn.Done:
			return
		}
	}
}

// pingPump 发送心跳 Ping
func (a *WebSocketAdapter) pingPump(conn *WebSocketConnection) {
	ticker := time.NewTicker(a.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			conn.mu.Lock()
			conn.LastPing = time.Now()
			conn.mu.Unlock()

			conn.Conn.SetWriteDeadline(time.Now().Add(a.writeWait))
			if err := conn.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Error("failed to send ping", zap.Error(err))
				return
			}

		case <-conn.Done:
			return
		}
	}
}

// extractHeaders 提取 HTTP headers
func extractHeaders(headers map[string][]string) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// extractQuery 提取查询参数
func extractQuery(query map[string][]string) map[string]string {
	result := make(map[string]string)
	for key, values := range query {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// 错误定义
var (
	ErrMaxConnectionsReached = errors.New("maximum connections reached")
	ErrConnectionNotFound    = errors.New("connection not found")
	ErrSendTimeout           = errors.New("send message timeout")
)
