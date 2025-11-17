import client from './client'
import type { Rule, CreateRuleInput, UpdateRuleInput, RuleFilter } from '@/types/rule'

/**
 * 规则 API 接口
 */
export const ruleApi = {
  /**
   * 创建规则
   */
  create: (data: CreateRuleInput) => {
    return client.post<Rule>('/rules', data)
  },

  /**
   * 获取规则详情
   */
  get: (id: string) => {
    return client.get<Rule>(`/rules/${id}`)
  },

  /**
   * 更新规则
   */
  update: (id: string, data: UpdateRuleInput) => {
    return client.put<Rule>(`/rules/${id}`, data)
  },

  /**
   * 删除规则
   */
  delete: (id: string) => {
    return client.delete(`/rules/${id}`)
  },

  /**
   * 获取规则列表（支持过滤）
   */
  list: (filter?: RuleFilter) => {
    return client.get<{
      data: Rule[]
      total: number
      page: number
      page_size: number
    }>('/rules', { params: filter })
  },

  /**
   * 批量启用/禁用规则
   */
  batchToggle: (ids: string[], enabled: boolean) => {
    return client.post('/rules/batch/toggle', { ids, enabled })
  },

  /**
   * 批量删除规则
   */
  batchDelete: (ids: string[]) => {
    return client.post('/rules/batch/delete', { ids })
  },

  /**
   * 复制规则
   */
  copy: (id: string) => {
    return client.post<Rule>(`/rules/${id}/copy`)
  },
}
