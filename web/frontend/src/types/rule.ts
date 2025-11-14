// 规则类型定义
export type Protocol = 'HTTP' | 'HTTPS'
export type MatchType = 'Simple' | 'Regex' | 'Script'
export type ContentType = 'JSON' | 'XML' | 'HTML' | 'Text'
export type ResponseType = 'Static' | 'Dynamic' | 'Proxy'

export interface MatchCondition {
  method?: string | string[]
  path?: string
  query?: Record<string, string>
  headers?: Record<string, string>
  ip_whitelist?: string[]
}

export interface ResponseContent {
  status_code: number
  content_type: ContentType
  headers?: Record<string, string>
  body?: any
}

export interface Response {
  type: ResponseType
  content: ResponseContent
}

export interface Delay {
  fixed?: number
  random?: {
    min: number
    max: number
  }
}

export interface Rule {
  id: string
  name: string
  project_id: string
  environment_id: string
  protocol: Protocol
  match_type: MatchType
  priority: number
  enabled: boolean
  tags?: string[]
  description?: string
  match_condition: MatchCondition
  response: Response
  delay?: Delay
  created_at: string
  updated_at: string
}

export interface CreateRuleInput {
  name: string
  project_id: string
  environment_id: string
  protocol: Protocol
  match_type: MatchType
  priority: number
  enabled: boolean
  tags?: string[]
  description?: string
  match_condition: MatchCondition
  response: Response
  delay?: Delay
}

export interface UpdateRuleInput {
  name?: string
  priority?: number
  enabled?: boolean
  tags?: string[]
  description?: string
  match_condition?: MatchCondition
  response?: Response
  delay?: Delay
}

export interface RuleFilter {
  project_id?: string
  environment_id?: string
  enabled?: boolean
  protocol?: Protocol
  keyword?: string
}
