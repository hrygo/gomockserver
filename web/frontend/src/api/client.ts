import axios, { AxiosInstance, AxiosResponse, AxiosError } from 'axios'
import { message } from 'antd'
import type { ErrorResponse } from '@/types/common'

// 创建 axios 实例
const client: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
client.interceptors.request.use(
  (config) => {
    // 添加 request_id 用于请求追踪
    const requestId = `req_${Date.now()}_${Math.random().toString(36).substring(7)}`
    config.headers['X-Request-ID'] = requestId

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
client.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  (error: AxiosError<ErrorResponse>) => {
    // 统一错误处理
    if (error.response) {
      const { status, data } = error.response

      switch (status) {
        case 400:
          message.error(data?.message || '请求参数错误')
          break
        case 401:
          message.error('未授权，请登录')
          // 可以在这里跳转到登录页
          break
        case 403:
          message.error('权限不足')
          break
        case 404:
          message.error(data?.message || '资源不存在')
          break
        case 500:
          message.error(data?.message || '服务器错误')
          break
        default:
          message.error(data?.message || '请求失败')
      }
    } else if (error.request) {
      message.error('网络连接失败，请检查网络')
    } else {
      message.error('请求配置错误')
    }

    return Promise.reject(error)
  }
)

export default client
