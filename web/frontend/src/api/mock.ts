import client from './client'
import type { MockResponse, SendMockRequestInput, MockTestHistory } from '@/types/mock'

/**
 * Mock 测试 API 接口
 */
export const mockApi = {
  /**
   * 发送 Mock 测试请求
   */
  sendRequest: (data: SendMockRequestInput) => {
    return client.post<MockResponse>('/mock/test', data)
  },

  /**
   * 获取测试历史记录
   */
  getHistory: (projectId: string, environmentId?: string) => {
    return client.get<MockTestHistory[]>('/mock/history', {
      params: { project_id: projectId, environment_id: environmentId },
    })
  },

  /**
   * 清空测试历史
   */
  clearHistory: (projectId: string, environmentId?: string) => {
    return client.delete('/mock/history', {
      params: { project_id: projectId, environment_id: environmentId },
    })
  },

  /**
   * 删除单条历史记录
   */
  deleteHistoryItem: (id: string) => {
    return client.delete(`/mock/history/${id}`)
  },
}
