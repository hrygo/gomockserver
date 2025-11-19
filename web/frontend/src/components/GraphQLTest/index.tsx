import React from 'react'
import { Card, Spin, Alert, Descriptions, Tag } from 'antd'
import { useDashboardStatistics, useServerStatus } from '@/hooks/useGraphQL'

const GraphQLTest: React.FC = () => {
  const { data: dashboardData, loading: dashboardLoading, error: dashboardError } = useDashboardStatistics()
  const { data: statusData, loading: statusLoading, error: statusError } = useServerStatus()

  const isLoading = dashboardLoading || statusLoading
  const hasError = dashboardError || statusError

  return (
    <div style={{ padding: '24px' }}>
      <Card title="GraphQL 连接测试" style={{ marginBottom: '16px' }}>
        {isLoading && (
          <div style={{ textAlign: 'center', padding: '20px' }}>
            <Spin size="large" />
            <p style={{ marginTop: '16px' }}>正在连接到GraphQL服务...</p>
          </div>
        )}

        {hasError && (
          <Alert
            message="GraphQL连接错误"
            description={
              <div>
                {dashboardError && (
                  <div style={{ marginBottom: '8px' }}>
                    <strong>Dashboard查询错误:</strong> {dashboardError.message}
                  </div>
                )}
                {statusError && (
                  <div>
                    <strong>Status查询错误:</strong> {statusError.message}
                  </div>
                )}
              </div>
            }
            type="error"
            showIcon
            style={{ marginBottom: '16px' }}
          />
        )}

        {!isLoading && !hasError && (
          <div>
            <Tag color="success" style={{ marginBottom: '16px' }}>
              GraphQL连接成功
            </Tag>

            <Descriptions title="查询结果" bordered column={1}>
              <Descriptions.Item label="Hello查询结果">
                {dashboardData?.hello || '无数据'}
              </Descriptions.Item>
              <Descriptions.Item label="Status查询结果">
                {dashboardData?.status || statusData?.status || '无数据'}
              </Descriptions.Item>
              <Descriptions.Item label="完整Dashboard数据">
                <pre style={{ fontSize: '12px', maxHeight: '200px', overflow: 'auto' }}>
                  {JSON.stringify(dashboardData, null, 2)}
                </pre>
              </Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Card>
    </div>
  )
}

export default GraphQLTest