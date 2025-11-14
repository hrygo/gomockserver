import client from './client'
import type {
  DashboardStatistics,
  ProjectStatistics,
  RuleStatistics,
  RequestTrend,
  ResponseTimeDistribution,
} from '@/types/statistics'

/**
 * 统计数据 API 接口
 */
export const statisticsApi = {
  /**
   * 获取仪表盘统计数据
   */
  getDashboard: () => {
    return client.get<DashboardStatistics>('/statistics/dashboard')
  },

  /**
   * 获取项目统计
   */
  getProjects: () => {
    return client.get<ProjectStatistics[]>('/statistics/projects')
  },

  /**
   * 获取规则统计
   */
  getRules: (projectId?: string) => {
    return client.get<RuleStatistics[]>('/statistics/rules', {
      params: { project_id: projectId },
    })
  },

  /**
   * 获取请求趋势（最近7天）
   */
  getRequestTrend: () => {
    return client.get<RequestTrend[]>('/statistics/request-trend')
  },

  /**
   * 获取响应时间分布
   */
  getResponseTimeDistribution: () => {
    return client.get<ResponseTimeDistribution[]>('/statistics/response-time-distribution')
  },
}
