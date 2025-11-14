import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { message } from 'antd'
import { projectApi } from '@/api/project'
import type { CreateProjectInput, UpdateProjectInput } from '@/types/project'

// Query Keys
export const projectKeys = {
  all: ['projects'] as const,
  detail: (id: string) => ['projects', id] as const,
}

// 获取项目列表
export const useProjects = () => {
  return useQuery({
    queryKey: projectKeys.all,
    queryFn: async () => {
      const response = await projectApi.list()
      return response.data
    },
  })
}

// 获取项目详情
export const useProject = (id: string) => {
  return useQuery({
    queryKey: projectKeys.detail(id),
    queryFn: async () => {
      const response = await projectApi.get(id)
      return response.data
    },
    enabled: !!id,
  })
}

// 创建项目
export const useCreateProject = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateProjectInput) => {
      const response = await projectApi.create(data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: projectKeys.all })
      message.success('项目创建成功')
    },
    onError: () => {
      message.error('项目创建失败')
    },
  })
}

// 更新项目
export const useUpdateProject = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateProjectInput }) => {
      const response = await projectApi.update(id, data)
      return response.data
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: projectKeys.all })
      queryClient.invalidateQueries({ queryKey: projectKeys.detail(data.id) })
      message.success('项目更新成功')
    },
    onError: () => {
      message.error('项目更新失败')
    },
  })
}

// 删除项目
export const useDeleteProject = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: string) => {
      await projectApi.delete(id)
      return id
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: projectKeys.all })
      message.success('项目删除成功')
    },
    onError: () => {
      message.error('项目删除失败')
    },
  })
}
