import client from './client'
import type { Project, CreateProjectInput, UpdateProjectInput } from '@/types/project'

// 项目相关 API
export const projectApi = {
  // 创建项目
  create: (data: CreateProjectInput) => {
    return client.post<Project>('/projects', data)
  },

  // 获取项目详情
  get: (id: string) => {
    return client.get<Project>(`/projects/${id}`)
  },

  // 更新项目
  update: (id: string, data: UpdateProjectInput) => {
    return client.put<Project>(`/projects/${id}`, data)
  },

  // 删除项目
  delete: (id: string) => {
    return client.delete(`/projects/${id}`)
  },

  // 获取项目列表
  list: () => {
    return client.get<Project[]>('/projects')
  },
}
