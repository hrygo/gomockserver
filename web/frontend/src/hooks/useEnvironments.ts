import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { message } from 'antd'
import { environmentApi } from '@/api/environment'
import type { CreateEnvironmentInput, UpdateEnvironmentInput } from '@/types/environment'

// Query Keys
export const environmentKeys = {
  all: ['environments'] as const,
  byProject: (projectId: string) => ['environments', 'project', projectId] as const,
  detail: (id: string) => ['environments', id] as const,
}

// 获取项目的环境列表
export const useEnvironments = (projectId: string) => {
  return useQuery({
    queryKey: environmentKeys.byProject(projectId),
    queryFn: async () => {
      const response = await environmentApi.listByProject(projectId)
      return response.data.data || []
    },
    enabled: !!projectId,
  })
}

// 获取环境详情
export const useEnvironment = (id: string) => {
  return useQuery({
    queryKey: environmentKeys.detail(id),
    queryFn: async () => {
      const response = await environmentApi.get(id)
      return response.data
    },
    enabled: !!id,
  })
}

// 创建环境
export const useCreateEnvironment = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateEnvironmentInput) => {
      const response = await environmentApi.create(data)
      return response.data
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: environmentKeys.byProject(data.project_id) })
      message.success('环境创建成功')
    },
    onError: () => {
      message.error('环境创建失败')
    },
  })
}

// 更新环境
export const useUpdateEnvironment = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateEnvironmentInput }) => {
      const response = await environmentApi.update(id, data)
      return response.data
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: environmentKeys.byProject(data.project_id) })
      queryClient.invalidateQueries({ queryKey: environmentKeys.detail(data.id) })
      message.success('环境更新成功')
    },
    onError: () => {
      message.error('环境更新失败')
    },
  })
}

// 删除环境
export const useDeleteEnvironment = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, projectId }: { id: string; projectId: string }) => {
      await environmentApi.delete(id)
      return { id, projectId }
    },
    onSuccess: ({ projectId }) => {
      queryClient.invalidateQueries({ queryKey: environmentKeys.byProject(projectId) })
      message.success('环境删除成功')
    },
    onError: () => {
      message.error('环境删除失败')
    },
  })
}
