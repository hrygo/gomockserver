import axios from 'axios';
import type { HealthStatus, SystemMetrics, RealtimeStats, TrendData } from '../types/monitoring';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export const monitoringAPI = {
  /**
   * 获取健康状态
   */
  async getHealth(): Promise<HealthStatus> {
    const { data } = await axios.get<HealthStatus>(`${API_BASE_URL}/api/v1/health`);
    return data;
  },

  /**
   * 获取系统指标
   */
  async getMetrics(): Promise<SystemMetrics> {
    const { data } = await axios.get<SystemMetrics>(`${API_BASE_URL}/api/v1/metrics`);
    return data;
  },

  /**
   * 存活检查
   */
  async checkLive(): Promise<{ status: string; timestamp: string }> {
    const { data } = await axios.get(`${API_BASE_URL}/api/v1/live`);
    return data;
  },

  /**
   * 就绪检查
   */
  async checkReady(): Promise<{ status: string; timestamp: string }> {
    const { data } = await axios.get(`${API_BASE_URL}/api/v1/ready`);
    return data;
  },
};

export const statisticsAPI = {
  /**
   * 获取实时统计
   */
  async getRealtime(): Promise<RealtimeStats> {
    const { data } = await axios.get<RealtimeStats>(`${API_BASE_URL}/api/v1/statistics/realtime`);
    return data;
  },

  /**
   * 获取趋势数据
   */
  async getTrend(period: 'hour' | 'day' | 'week' | 'month' = 'day', duration?: number): Promise<TrendData> {
    const params = new URLSearchParams({ period });
    if (duration) params.append('duration', duration.toString());
    const { data } = await axios.get<TrendData>(`${API_BASE_URL}/api/v1/statistics/trend?${params.toString()}`);
    return data;
  },
};
