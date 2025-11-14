// 环境类型定义
export interface Environment {
  id: string
  name: string
  project_id: string
  base_url: string
  description?: string
  created_at: string
  updated_at: string
}

export interface CreateEnvironmentInput {
  name: string
  project_id: string
  base_url: string
  description?: string
}

export interface UpdateEnvironmentInput {
  name?: string
  base_url?: string
  description?: string
}
