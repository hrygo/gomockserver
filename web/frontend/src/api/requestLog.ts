import axios from 'axios';
import type { RequestLog, RequestLogFilter, RequestLogListResponse, RequestLogStatistics } from '../types/log';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export const requestLogAPI = {
  /**
   * 获取请求日志列表
   */
  async list(filter: RequestLogFilter): Promise<RequestLogListResponse> {
    const params = new URLSearchParams();
    
    if (filter.project_id) params.append('project_id', filter.project_id);
    if (filter.environment_id) params.append('environment_id', filter.environment_id);
    if (filter.rule_id) params.append('rule_id', filter.rule_id);
    if (filter.protocol) params.append('protocol', filter.protocol);
    if (filter.method) params.append('method', filter.method);
    if (filter.path) params.append('path', filter.path);
    if (filter.status_code) params.append('status_code', filter.status_code.toString());
    if (filter.source_ip) params.append('source_ip', filter.source_ip);
    if (filter.start_time) params.append('start_time', filter.start_time);
    if (filter.end_time) params.append('end_time', filter.end_time);
    if (filter.page) params.append('page', filter.page.toString());
    if (filter.page_size) params.append('page_size', filter.page_size.toString());
    if (filter.sort_by) params.append('sort_by', filter.sort_by);
    if (filter.sort_order) params.append('sort_order', filter.sort_order);
    
    const { data } = await axios.get<RequestLogListResponse>(
      `${API_BASE_URL}/api/v1/request-logs?${params.toString()}`
    );
    return data;
  },

  /**
   * 获取单个请求日志详情
   */
  async getById(id: string): Promise<RequestLog> {
    const { data } = await axios.get<RequestLog>(`${API_BASE_URL}/api/v1/request-logs/${id}`);
    return data;
  },

  /**
   * 清理旧日志
   */
  async cleanup(beforeDays: number): Promise<{ deleted_count: number }> {
    const { data } = await axios.delete(
      `${API_BASE_URL}/api/v1/request-logs/cleanup?before_days=${beforeDays}`
    );
    return data;
  },

  /**
   * 获取统计信息
   */
  async getStatistics(
    projectId?: string,
    environmentId?: string,
    period: '24h' | '7d' | '30d' = '24h'
  ): Promise<RequestLogStatistics> {
    const params = new URLSearchParams();
    if (projectId) params.append('project_id', projectId);
    if (environmentId) params.append('environment_id', environmentId);
    params.append('period', period);
    
    const { data } = await axios.get<RequestLogStatistics>(
      `${API_BASE_URL}/api/v1/request-logs/statistics?${params.toString()}`
    );
    return data;
  },
};
