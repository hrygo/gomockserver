import { useQuery } from '@tanstack/react-query'
import { statisticsApi } from '@/api/statistics'

// Query Keys
export const statisticsKeys = {
  all: ['statistics'] as const,
  dashboard: () => [...statisticsKeys.all, 'dashboard'] as const,
  projects: () => [...statisticsKeys.all, 'projects'] as const,
  rules: (projectId?: string) => [...statisticsKeys.all, 'rules', projectId] as const,
  requestTrend: () => [...statisticsKeys.all, 'request-trend'] as const,
  responseTime: () => [...statisticsKeys.all, 'response-time'] as const,
}

// 获取仪表盘统计数据
export const useDashboardStatistics = () => {
  return useQuery({
    queryKey: statisticsKeys.dashboard(),
    queryFn: async () => {
      const response = await statisticsApi.getDashboard()
      return response.data
    },
    refetchInterval: 30000, // 每30秒刷新一次
  })
}

// 获取项目统计
export const useProjectStatistics = () => {
  return useQuery({
    queryKey: statisticsKeys.projects(),
    queryFn: async () => {
      const response = await statisticsApi.getProjects()
      return response.data
    },
  })
}

// 获取规则统计
export const useRuleStatistics = (projectId?: string) => {
  return useQuery({
    queryKey: statisticsKeys.rules(projectId),
    queryFn: async () => {
      const response = await statisticsApi.getRules(projectId)
      return response.data
    },
  })
}

// 获取请求趋势
export const useRequestTrend = () => {
  return useQuery({
    queryKey: statisticsKeys.requestTrend(),
    queryFn: async () => {
      const response = await statisticsApi.getRequestTrend()
      return response.data
    },
  })
}

// 获取响应时间分布
export const useResponseTimeDistribution = () => {
  return useQuery({
    queryKey: statisticsKeys.responseTime(),
    queryFn: async () => {
      const response = await statisticsApi.getResponseTimeDistribution()
      return response.data
    },
  })
}
