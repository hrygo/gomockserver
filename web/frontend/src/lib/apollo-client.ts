import { ApolloClient, InMemoryCache, createHttpLink, from } from '@apollo/client'
import { setContext } from '@apollo/client/link/context'
import { onError } from '@apollo/client/link/error'

// HTTP连接到服务端GraphQL端点
const httpLink = createHttpLink({
  uri: '/api/graphql',
  credentials: 'include',
})

// 认证链 - 添加认证头（如果需要）
const authLink = setContext((_, { headers }) => {
  // 获取token，如果有的话
  const token = localStorage.getItem('auth-token')

  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : '',
    },
  }
})

// 错误处理链
const errorLink = onError(({ graphQLErrors, networkError }) => {
  if (graphQLErrors) {
    graphQLErrors.forEach((error: any) => {
      console.error(
        `[GraphQL error]: Message: ${error.message}, Location: ${error.locations}, Path: ${error.path}`
      )
    })
  }

  if (networkError) {
    console.error(`[Network error]: ${networkError}`)

    // 简单检查401错误
    if ((networkError as any)?.statusCode === 401 || (networkError as Error)?.message?.includes('401')) {
      // 未授权，可能需要重新登录
      localStorage.removeItem('auth-token')
      window.location.href = '/login'
    }
  }
})

// Apollo Client实例
export const apolloClient = new ApolloClient({
  link: from([errorLink, authLink, httpLink]),
  cache: new InMemoryCache({
    typePolicies: {
      Query: {
        fields: {
          // 配置缓存策略
          projects: {
            merge(_: any, incoming: any) {
              return incoming
            },
          },
          rules: {
            merge(_: any, incoming: any) {
              return incoming
            },
          },
        },
      },
    },
  }),
  defaultOptions: {
    watchQuery: {
      errorPolicy: 'all',
      notifyOnNetworkStatusChange: true,
    },
    query: {
      errorPolicy: 'all',
    },
    mutate: {
      errorPolicy: 'all',
    },
  },
})

export default apolloClient