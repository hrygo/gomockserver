import client from './client'
import type { ImportResult, ExportFormat } from '@/types/export'

/**
 * 导入导出 API 接口
 */
export const exportApi = {
  /**
   * 导出规则
   */
  exportRules: (projectId: string, environmentId?: string, format: ExportFormat = 'json') => {
    return client.get('/export/rules', {
      params: { project_id: projectId, environment_id: environmentId, format },
      responseType: 'blob',
    })
  },

  /**
   * 导出项目（包含环境和规则）
   */
  exportProject: (projectId: string, format: ExportFormat = 'json') => {
    return client.get(`/export/projects/${projectId}`, {
      params: { format },
      responseType: 'blob',
    })
  },

  /**
   * 导入规则
   */
  importRules: (projectId: string, file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('project_id', projectId)

    return client.post<ImportResult>('/import/rules', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },

  /**
   * 导入项目
   */
  importProject: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)

    return client.post<ImportResult>('/import/project', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },
}
