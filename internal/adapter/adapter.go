package adapter

import (
	"time"

	"github.com/gomockserver/mockserver/internal/models"
)

// Request 统一请求模型
type Request struct {
	ID         string                 // 唯一追踪ID
	Protocol   models.ProtocolType    // 协议类型
	Metadata   map[string]interface{} // 协议特定元数据
	Path       string                 // 请求路径/方法标识
	Headers    map[string]string      // 头信息
	Body       []byte                 // 载荷数据（原始）
	ParsedBody interface{}            // 解析后的数据
	SourceIP   string                 // 来源IP
	SourcePort int                    // 来源端口
	ReceivedAt time.Time              // 接收时间
}

// Response 统一响应模型
type Response struct {
	StatusCode int                    // 状态码
	Headers    map[string]string      // 响应头
	Body       []byte                 // 响应体
	Metadata   map[string]interface{} // 协议特定元数据
}

// ProtocolAdapter 协议适配器接口
type ProtocolAdapter interface {
	// Parse 将协议特定请求转换为统一请求模型
	Parse(rawRequest interface{}) (*Request, error)

	// Build 将统一响应模型转换为协议特定响应
	Build(response *Response) (interface{}, error)
}
