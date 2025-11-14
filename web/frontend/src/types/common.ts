// 通用类型定义
export interface ApiResponse<T = any> {
  code?: number
  message?: string
  data?: T
  request_id?: string
}

export interface PaginationParams {
  page?: number
  page_size?: number
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export interface ErrorResponse {
  code: number
  message: string
  details?: string
  request_id?: string
}

export type LoadingState = 'idle' | 'loading' | 'success' | 'error'
