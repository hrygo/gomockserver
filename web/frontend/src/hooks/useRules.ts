import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { message } from 'antd'
import { ruleApi } from '@/api/rule'
import type { CreateRuleInput, UpdateRuleInput, RuleFilter } from '@/types/rule'

// Query Keys
export const ruleKeys = {
  all: ['rules'] as const,
  lists: () => [...ruleKeys.all, 'list'] as const,
  list: (filter?: RuleFilter) => [...ruleKeys.lists(), { filter }] as const,
  details: () => [...ruleKeys.all, 'detail'] as const,
  detail: (id: string) => [...ruleKeys.details(), id] as const,
}

// 获取规则列表
export const useRules = (filter?: RuleFilter) => {
  return useQuery({
    queryKey: ruleKeys.list(filter),
    queryFn: async () => {
      const response = await ruleApi.list(filter)
      return response.data.data
    },
  })
}

// 获取规则详情
export const useRule = (id: string) => {
  return useQuery({
    queryKey: ruleKeys.detail(id),
    queryFn: async () => {
      const response = await ruleApi.get(id)
      return response.data
    },
    enabled: !!id,
  })
}

// 创建规则
export const useCreateRule = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateRuleInput) => {
      const response = await ruleApi.create(data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ruleKeys.lists() })
      message.success('规则创建成功')
    },
    onError: () => {
      message.error('规则创建失败')
    },
  })
}

// 更新规则
export const useUpdateRule = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateRuleInput }) => {
      const response = await ruleApi.update(id, data)
      return response.data
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ruleKeys.lists() })
      queryClient.invalidateQueries({ queryKey: ruleKeys.detail(data.id) })
      message.success('规则更新成功')
    },
    onError: () => {
      message.error('规则更新失败')
    },
  })
}

// 删除规则
export const useDeleteRule = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: string) => {
      await ruleApi.delete(id)
      return id
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ruleKeys.lists() })
      message.success('规则删除成功')
    },
    onError: () => {
      message.error('规则删除失败')
    },
  })
}

// 批量启用/禁用规则
export const useBatchToggleRules = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ ids, enabled }: { ids: string[]; enabled: boolean }) => {
      await ruleApi.batchToggle(ids, enabled)
      return { ids, enabled }
    },
    onSuccess: ({ enabled }) => {
      queryClient.invalidateQueries({ queryKey: ruleKeys.lists() })
      message.success(enabled ? '批量启用成功' : '批量禁用成功')
    },
    onError: () => {
      message.error('批量操作失败')
    },
  })
}

// 批量删除规则
export const useBatchDeleteRules = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (ids: string[]) => {
      await ruleApi.batchDelete(ids)
      return ids
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ruleKeys.lists() })
      message.success('批量删除成功')
    },
    onError: () => {
      message.error('批量删除失败')
    },
  })
}

// 复制规则
export const useCopyRule = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: string) => {
      const response = await ruleApi.copy(id)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ruleKeys.lists() })
      message.success('规则复制成功')
    },
    onError: () => {
      message.error('规则复制失败')
    },
  })
}
