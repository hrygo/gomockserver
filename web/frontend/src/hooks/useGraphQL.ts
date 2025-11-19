import { useQuery, gql } from '@apollo/client'

// GraphQL查询定义
export const GET_DASHBOARD_STATISTICS = gql`
  query {
    hello
    status
  }
`

export const GET_SERVER_STATUS = gql`
  query {
    status
  }
`

export const useDashboardStatistics = () => {
  return useQuery(GET_DASHBOARD_STATISTICS, {
    errorPolicy: 'all',
    notifyOnNetworkStatusChange: true,
    fetchPolicy: 'network-only',
  })
}

export const useServerStatus = () => {
  return useQuery(GET_SERVER_STATUS, {
    errorPolicy: 'all',
    notifyOnNetworkStatusChange: true,
    fetchPolicy: 'network-only',
  })
}