export interface HealthStatus {
  status: string;
  timestamp: string;
  services: Record<string, string>;
  uptime: number;
}

export interface SystemMetrics {
  timestamp: string;
  runtime: {
    uptime: number;
    go_version: string;
    num_cpu: number;
    goos: string;
    goarch: string;
  };
  memory: {
    alloc: number;
    total_alloc: number;
    sys: number;
    num_gc: number;
    heap_alloc: number;
    heap_sys: number;
    heap_inuse: number;
    stack_inuse: number;
    gc_cpu_fraction: number;
  };
  goroutines: number;
  database: {
    connected: boolean;
    status: string;
    latency: number;
  };
}

export interface RealtimeStats {
  timestamp: string;
  requests_per_min: number;
  active_connections: number;
  avg_response_time: number;
  error_rate: number;
  protocol_stats: Record<string, {
    count: number;
    success_rate: number;
    avg_duration: number;
  }>;
}

export interface TrendPoint {
  timestamp: string;
  request_count: number;
  success_count: number;
  error_count: number;
  avg_duration: number;
}

export interface TrendData {
  period: string;
  data_points: TrendPoint[];
}
