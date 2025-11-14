// 项目类型定义
export interface Project {
  id: string
  name: string
  workspace_id: string
  description?: string
  created_at: string
  updated_at: string
}

export interface CreateProjectInput {
  name: string
  workspace_id: string
  description?: string
}

export interface UpdateProjectInput {
  name?: string
  description?: string
}
