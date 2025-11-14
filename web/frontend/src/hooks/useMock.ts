import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { message } from 'antd'
import { mockApi } from '@/api/mock'
import type { SendMockRequestInput } from '@/types/mock'

// Query Keys
export const mockKeys = {
  all: ['mock'] as const,
  history: (projectId: string, environmentId?: string) =>
    [...mockKeys.all, 'history', projectId, environmentId] as const,
}

// 获取测试历史
export const useMockHistory = (projectId: string, environmentId?: string) => {
  return useQuery({
    queryKey: mockKeys.history(projectId, environmentId),
    queryFn: async () => {
      const response = await mockApi.getHistory(projectId, environmentId)
      return response.data
    },
    enabled: !!projectId,
  })
}

// 发送 Mock 测试请求
export const useSendMockRequest = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: SendMockRequestInput) => {
      const response = await mockApi.sendRequest(data)
      return response.data
    },
    onSuccess: (_, variables) => {
      // 刷新历史记录
      queryClient.invalidateQueries({
        queryKey: mockKeys.history(variables.project_id, variables.environment_id),
      })
    },
    onError: () => {
      message.error('请求发送失败')
    },
  })
}

// 清空测试历史
export const useClearMockHistory = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      projectId,
      environmentId,
    }: {
      projectId: string
      environmentId?: string
    }) => {
      await mockApi.clearHistory(projectId, environmentId)
      return { projectId, environmentId }
    },
    onSuccess: ({ projectId, environmentId }) => {
      queryClient.invalidateQueries({ queryKey: mockKeys.history(projectId, environmentId) })
      message.success('历史记录已清空')
    },
    onError: () => {
      message.error('清空失败')
    },
  })
}

// 删除单条历史记录
export const useDeleteMockHistoryItem = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, projectId }: { id: string; projectId: string }) => {
      await mockApi.deleteHistoryItem(id)
      return { id, projectId }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mockKeys.all })
      message.success('记录已删除')
    },
    onError: () => {
      message.error('删除失败')
    },
  })
}
