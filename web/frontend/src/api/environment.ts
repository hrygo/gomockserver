import client from './client'
import type { Environment, CreateEnvironmentInput, UpdateEnvironmentInput } from '@/types/environment'

// 环境相关 API
export const environmentApi = {
  // 创建环境
  create: (data: CreateEnvironmentInput) => {
    return client.post<Environment>('/environments', data)
  },

  // 获取环境详情
  get: (id: string) => {
    return client.get<Environment>(`/environments/${id}`)
  },

  // 更新环境
  update: (id: string, data: UpdateEnvironmentInput) => {
    return client.put<Environment>(`/environments/${id}`, data)
  },

  // 删除环境
  delete: (id: string) => {
    return client.delete(`/environments/${id}`)
  },

  // 获取环境列表（按项目）
  listByProject: (projectId: string) => {
    return client.get<{ data: Environment[] }>(`/environments?project_id=${projectId}`)
  },
}
