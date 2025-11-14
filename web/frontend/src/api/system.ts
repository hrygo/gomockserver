import client from './client'

export interface SystemInfo {
  version: string
  build_time: string
  go_version: string
  admin_api_url: string
  mock_service_url: string
}

export interface HealthStatus {
  status: 'healthy' | 'unhealthy'
  database: boolean
  cache: boolean
  uptime: number
}

/**
 * 系统信息 API 接口
 */
export const systemApi = {
  /**
   * 获取系统信息
   */
  getInfo: () => {
    return client.get<SystemInfo>('/system/info')
  },

  /**
   * 获取健康状态
   */
  getHealth: () => {
    return client.get<HealthStatus>('/system/health')
  },
}
