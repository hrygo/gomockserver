// Mock 测试相关类型定义

export interface MockRequest {
  method: string
  url: string
  headers?: Record<string, string>
  query?: Record<string, string>
  body?: any
}

export interface MockResponse {
  status_code: number
  headers?: Record<string, string>
  body?: any
  duration: number // 响应时间(ms)
  matched_rule_id?: string
  matched_rule_name?: string
}

export interface MockTestHistory {
  id: string
  request: MockRequest
  response: MockResponse
  timestamp: string
  project_id: string
  environment_id: string
}

export interface SendMockRequestInput {
  project_id: string
  environment_id: string
  request: MockRequest
}
