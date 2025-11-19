import React from 'react'
import { useServerStatus } from '@/hooks/useGraphQL'
import { Card, Spin, Alert, Typography, Tag } from 'antd'

const { Title, Text } = Typography

const GraphQLTestComponent: React.FC = () => {
  const { loading, error, data } = useServerStatus()

  if (loading) {
    return (
      <Card title="GraphQL连接测试" style={{ margin: '20px' }}>
        <div style={{ textAlign: 'center', padding: '20px' }}>
          <Spin size="large" />
          <div style={{ marginTop: '10px' }}>正在连接GraphQL服务器...</div>
        </div>
      </Card>
    )
  }

  if (error) {
    return (
      <Card title="GraphQL连接测试" style={{ margin: '20px' }}>
        <Alert
          message="GraphQL连接错误"
          description={error.message}
          type="error"
          showIcon
        />
      </Card>
    )
  }

  return (
    <Card title="GraphQL连接测试" style={{ margin: '20px' }}>
      <Title level={4}>连接状态: <Tag color="green">成功</Tag></Title>

      <div style={{ marginBottom: '16px' }}>
        <Text strong>服务器状态: </Text>
        <Tag color={data?.status?.status === 'healthy' ? 'green' : 'red'}>
          {data?.status?.status || '未知'}
        </Tag>
      </div>

      <div style={{ marginBottom: '16px' }}>
        <Text strong>版本信息: </Text>
        <Text code>{data?.status?.version || '未知'}</Text>
      </div>

      <div style={{ marginBottom: '16px' }}>
        <Text strong>时间戳: </Text>
        <Text>{data?.status?.timestamp || '未知'}</Text>
      </div>

      {data?.hello && (
        <div style={{ marginBottom: '16px' }}>
          <Text strong>Hello消息: </Text>
          <Text>{data.hello.message || 'Hello from MockServer!'}</Text>
        </div>
      )}

      <div style={{ marginTop: '20px', padding: '10px', backgroundColor: '#f5f5f5', borderRadius: '4px' }}>
        <Text type="secondary">
          <strong>原始数据:</strong>
        </Text>
        <pre style={{ marginTop: '8px', fontSize: '12px', whiteSpace: 'pre-wrap' }}>
          {JSON.stringify(data, null, 2)}
        </pre>
      </div>
    </Card>
  )
}

export default GraphQLTestComponent