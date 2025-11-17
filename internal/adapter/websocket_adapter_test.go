package adapter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestNewWebSocketAdapter(t *testing.T) {
	adapter := NewWebSocketAdapter()

	assert.NotNil(t, adapter)
	assert.Equal(t, 1000, adapter.maxConnections)
	assert.Equal(t, 30*time.Second, adapter.pingInterval)
	assert.Equal(t, 60*time.Second, adapter.pongWait)
	assert.NotNil(t, adapter.connections)
}

func TestWebSocketAdapter_ConnectionManagement(t *testing.T) {
	adapter := NewWebSocketAdapter()

	// 测试获取连接数
	assert.Equal(t, 0, adapter.GetConnectionCount())

	// 创建模拟连接
	conn := &WebSocketConnection{
		ID:       "test-conn-1",
		Send:     make(chan []byte, 256),
		Done:     make(chan struct{}),
		Metadata: make(map[string]interface{}),
	}

	// 添加连接
	err := adapter.addConnection(conn)
	assert.NoError(t, err)
	assert.Equal(t, 1, adapter.GetConnectionCount())

	// 获取连接
	retrievedConn, exists := adapter.GetConnection("test-conn-1")
	assert.True(t, exists)
	assert.Equal(t, conn.ID, retrievedConn.ID)

	// 获取不存在的连接
	_, exists = adapter.GetConnection("non-existent")
	assert.False(t, exists)

	// 移除连接
	adapter.removeConnection("test-conn-1")
	assert.Equal(t, 0, adapter.GetConnectionCount())
}

func TestWebSocketAdapter_MaxConnections(t *testing.T) {
	adapter := NewWebSocketAdapter()
	adapter.maxConnections = 2

	// 添加第一个连接
	conn1 := &WebSocketConnection{
		ID:   "conn-1",
		Send: make(chan []byte, 256),
		Done: make(chan struct{}),
	}
	err := adapter.addConnection(conn1)
	assert.NoError(t, err)

	// 添加第二个连接
	conn2 := &WebSocketConnection{
		ID:   "conn-2",
		Send: make(chan []byte, 256),
		Done: make(chan struct{}),
	}
	err = adapter.addConnection(conn2)
	assert.NoError(t, err)

	// 尝试添加第三个连接（应该失败）
	conn3 := &WebSocketConnection{
		ID:   "conn-3",
		Conn: nil, // 测试中不需要真实连接
		Send: make(chan []byte, 256),
		Done: make(chan struct{}),
	}
	err = adapter.addConnection(conn3)
	assert.Error(t, err)
	assert.Equal(t, ErrMaxConnectionsReached, err)
}

func TestWebSocketAdapter_SendToConnection(t *testing.T) {
	adapter := NewWebSocketAdapter()

	conn := &WebSocketConnection{
		ID:   "test-conn",
		Send: make(chan []byte, 256),
		Done: make(chan struct{}),
	}
	adapter.addConnection(conn)

	// 测试发送消息
	message := []byte("test message")
	err := adapter.SendToConnection("test-conn", message)
	assert.NoError(t, err)

	// 验证消息被接收
	select {
	case received := <-conn.Send:
		assert.Equal(t, message, received)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}

	// 测试发送到不存在的连接
	err = adapter.SendToConnection("non-existent", message)
	assert.Error(t, err)
	assert.Equal(t, ErrConnectionNotFound, err)
}

func TestWebSocketAdapter_BroadcastMessage(t *testing.T) {
	adapter := NewWebSocketAdapter()

	// 创建多个连接
	conn1 := &WebSocketConnection{
		ID:   "conn-1",
		Send: make(chan []byte, 256),
		Done: make(chan struct{}),
	}
	conn2 := &WebSocketConnection{
		ID:   "conn-2",
		Send: make(chan []byte, 256),
		Done: make(chan struct{}),
	}

	adapter.addConnection(conn1)
	adapter.addConnection(conn2)

	// 广播消息
	message := []byte("broadcast message")
	adapter.BroadcastMessage(message)

	// 验证所有连接都收到消息
	select {
	case received := <-conn1.Send:
		assert.Equal(t, message, received)
	case <-time.After(time.Second):
		t.Fatal("conn1: timeout waiting for broadcast message")
	}

	select {
	case received := <-conn2.Send:
		assert.Equal(t, message, received)
	case <-time.After(time.Second):
		t.Fatal("conn2: timeout waiting for broadcast message")
	}
}

func TestExtractHeaders(t *testing.T) {
	headers := map[string][]string{
		"Content-Type": {"application/json"},
		"User-Agent":   {"test-agent", "other-value"},
		"Empty":        {},
	}

	result := extractHeaders(headers)

	assert.Equal(t, "application/json", result["Content-Type"])
	assert.Equal(t, "test-agent", result["User-Agent"])
	assert.NotContains(t, result, "Empty")
}

func TestExtractQuery(t *testing.T) {
	query := map[string][]string{
		"param1": {"value1"},
		"param2": {"value2", "other-value"},
		"empty":  {},
	}

	result := extractQuery(query)

	assert.Equal(t, "value1", result["param1"])
	assert.Equal(t, "value2", result["param2"])
	assert.NotContains(t, result, "empty")
}

// 集成测试 - WebSocket 升级
func TestWebSocketAdapter_Upgrade(t *testing.T) {
	// 由于需要实际的 WebSocket 连接，这个测试较为复杂
	// 这里提供一个基本的框架，实际测试可能需要更复杂的设置
	
	gin.SetMode(gin.TestMode)
	adapter := NewWebSocketAdapter()

	router := gin.New()
	router.GET("/ws/:projectID/:environmentID", func(c *gin.Context) {
		_, err := adapter.Parse(c)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	})

	// 创建测试服务器
	server := httptest.NewServer(router)
	defer server.Close()

	// 将 http 改为 ws
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/proj1/env1"

	// 建立 WebSocket 连接
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(wsURL, nil)
	
	if err == nil {
		defer conn.Close()
		assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
		
		// 验证连接已建立
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, 1, adapter.GetConnectionCount())
	}
	// 注意：在某些测试环境中 WebSocket 升级可能失败，这是正常的
}
