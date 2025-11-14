// 统计数据类型定义

export interface DashboardStatistics {
  total_projects: number
  total_environments: number
  total_rules: number
  enabled_rules: number
  disabled_rules: number
  total_requests: number
  requests_today: number
}

export interface ProjectStatistics {
  project_id: string
  project_name: string
  environment_count: number
  rule_count: number
  request_count: number
}

export interface RuleStatistics {
  rule_id: string
  rule_name: string
  match_count: number
  avg_response_time: number
  last_matched_at?: string
}

export interface RequestTrend {
  date: string
  count: number
}

export interface ResponseTimeDistribution {
  range: string
  count: number
}
