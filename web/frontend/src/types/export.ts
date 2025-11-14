// 导入导出类型定义

export interface ExportData {
  version: string
  export_time: string
  project?: ExportProject
  rules?: ExportRule[]
}

export interface ExportProject {
  id: string
  name: string
  workspace_id: string
  description?: string
  environments: ExportEnvironment[]
  rules: ExportRule[]
}

export interface ExportEnvironment {
  id: string
  name: string
  base_url: string
  description?: string
}

export interface ExportRule {
  id: string
  name: string
  environment_id: string
  protocol: string
  match_type: string
  priority: number
  enabled: boolean
  tags?: string[]
  description?: string
  match_condition: any
  response: any
  delay?: any
}

export interface ImportResult {
  success: boolean
  message: string
  imported_count: number
  failed_count: number
  errors?: string[]
}

export type ExportFormat = 'json' | 'yaml'
