export interface RequestLog {
  id: string;
  request_id: string;
  project_id: string;
  environment_id: string;
  rule_id?: string;
  protocol: 'http' | 'websocket';
  method?: string;
  path?: string;
  request: Record<string, any>;
  response: Record<string, any>;
  status_code?: number;
  duration: number;
  source_ip: string;
  timestamp: string;
}

export interface RequestLogFilter {
  project_id?: string;
  environment_id?: string;
  rule_id?: string;
  protocol?: string;
  method?: string;
  path?: string;
  status_code?: number;
  source_ip?: string;
  start_time?: string;
  end_time?: string;
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface RequestLogListResponse {
  data: RequestLog[];
  total: number;
  page: number;
  size: number;
}

export interface RequestLogStatistics {
  total_requests: number;
  success_requests: number;
  error_requests: number;
  avg_duration: number;
  max_duration: number;
  min_duration: number;
  protocol_stats: Record<string, number>;
  status_code_stats: Record<string, number>;
}
