import React, { useState } from 'react'
import {
  Card,
  Form,
  Input,
  Select,
  Button,
  Space,
  Tabs,
  Descriptions,
  Table,
  Tag,
  Empty,
  Popconfirm,
  message,
  Badge,
} from 'antd'
import {
  SendOutlined,
  DeleteOutlined,
  ClearOutlined,
  HistoryOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { useProjects } from '@/hooks/useProjects'
import { useEnvironments } from '@/hooks/useEnvironments'
import {
  useSendMockRequest,
  useMockHistory,
  useClearMockHistory,
  useDeleteMockHistoryItem,
} from '@/hooks/useMock'
import type { MockRequest, MockTestHistory } from '@/types/mock'

const { Option } = Select
const { TextArea } = Input
const { TabPane } = Tabs

const MockTest: React.FC = () => {
  const [form] = Form.useForm()
  const [selectedProjectId, setSelectedProjectId] = useState<string>()
  const [selectedEnvironmentId, setSelectedEnvironmentId] = useState<string>()
  const [responseData, setResponseData] = useState<any>(null)
  const [activeTab, setActiveTab] = useState('request')

  // 数据查询
  const { data: projects = [] } = useProjects()
  const { data: environments = [] } = useEnvironments(selectedProjectId || '')
  const { data: history = [], isLoading: historyLoading } = useMockHistory(
    selectedProjectId || '',
    selectedEnvironmentId
  )

  // 数据操作
  const sendRequestMutation = useSendMockRequest()
  const clearHistoryMutation = useClearMockHistory()
  const deleteHistoryMutation = useDeleteMockHistoryItem()

  // 发送测试请求
  const handleSendRequest = async () => {
    if (!selectedProjectId || !selectedEnvironmentId) {
      message.warning('请选择项目和环境')
      return
    }

    try {
      const values = await form.validateFields()

      // 解析 JSON 字符串
      let headers: any = {}
      let query: any = {}
      let body: any = null

      if (values.headers) {
        try {
          headers = JSON.parse(values.headers)
        } catch {
          message.error('请求头格式错误,请输入有效的JSON')
          return
        }
      }

      if (values.query) {
        try {
          query = JSON.parse(values.query)
        } catch {
          message.error('查询参数格式错误,请输入有效的JSON')
          return
        }
      }

      if (values.body) {
        try {
          body = JSON.parse(values.body)
        } catch {
          message.error('请求体格式错误,请输入有效的JSON')
          return
        }
      }

      const request: MockRequest = {
        method: values.method,
        url: values.url,
        headers: Object.keys(headers).length > 0 ? headers : undefined,
        query: Object.keys(query).length > 0 ? query : undefined,
        body,
      }

      const response = await sendRequestMutation.mutateAsync({
        project_id: selectedProjectId,
        environment_id: selectedEnvironmentId,
        request,
      })

      setResponseData(response)
      setActiveTab('response')
      message.success('请求发送成功')
    } catch (error) {
      console.error('发送请求失败:', error)
    }
  }

  // 清空历史
  const handleClearHistory = async () => {
    if (!selectedProjectId) return
    await clearHistoryMutation.mutateAsync({
      projectId: selectedProjectId,
      environmentId: selectedEnvironmentId,
    })
  }

  // 删除历史记录
  const handleDeleteHistory = async (id: string) => {
    if (!selectedProjectId) return
    await deleteHistoryMutation.mutateAsync({ id, projectId: selectedProjectId })
  }

  // 从历史记录加载请求
  const handleLoadFromHistory = (record: MockTestHistory) => {
    form.setFieldsValue({
      method: record.request.method,
      url: record.request.url,
      headers: record.request.headers ? JSON.stringify(record.request.headers, null, 2) : '',
      query: record.request.query ? JSON.stringify(record.request.query, null, 2) : '',
      body: record.request.body ? JSON.stringify(record.request.body, null, 2) : '',
    })
    setResponseData(record.response)
    setActiveTab('request')
    message.success('已加载历史请求')
  }

  // 历史记录表格列
  const historyColumns: ColumnsType<MockTestHistory> = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: '15%',
      render: (text: string) => dayjs(text).format('MM-DD HH:mm:ss'),
    },
    {
      title: '请求',
      key: 'request',
      width: '30%',
      render: (_, record) => (
        <div>
          <Tag color="blue">{record.request.method}</Tag>
          <span style={{ fontSize: 12 }}>{record.request.url}</span>
        </div>
      ),
    },
    {
      title: '响应状态',
      dataIndex: ['response', 'status_code'],
      key: 'status_code',
      width: '10%',
      render: (code: number) => {
        const color = code >= 200 && code < 300 ? 'success' : 'error'
        return <Badge status={color} text={code} />
      },
    },
    {
      title: '匹配规则',
      dataIndex: ['response', 'matched_rule_name'],
      key: 'matched_rule',
      width: '20%',
      render: (name: string) => name || '-',
    },
    {
      title: '耗时',
      dataIndex: ['response', 'duration'],
      key: 'duration',
      width: '10%',
      render: (ms: number) => (
        <span>
          <ClockCircleOutlined style={{ marginRight: 4 }} />
          {ms}ms
        </span>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: '15%',
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => handleLoadFromHistory(record)}>
            加载
          </Button>
          <Popconfirm
            title="确认删除"
            onConfirm={() => handleDeleteHistory(record.id)}
            okText="确认"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <h1 style={{ marginBottom: 24 }}>Mock 测试</h1>

      {/* 项目和环境选择 */}
      <Card style={{ marginBottom: 16 }}>
        <Space size="large">
          <div>
            <span style={{ marginRight: 8 }}>项目:</span>
            <Select
              placeholder="选择项目"
              style={{ width: 200 }}
              value={selectedProjectId}
              onChange={(value) => {
                setSelectedProjectId(value)
                setSelectedEnvironmentId(undefined)
              }}
            >
              {projects?.map((p) => (
                <Option key={p.id} value={p.id}>
                  {p.name}
                </Option>
              ))}
            </Select>
          </div>

          <div>
            <span style={{ marginRight: 8 }}>环境:</span>
            <Select
              placeholder="选择环境"
              style={{ width: 200 }}
              value={selectedEnvironmentId}
              onChange={setSelectedEnvironmentId}
              disabled={!selectedProjectId || environments.length === 0}
            >
              {environments?.map((e) => (
                <Option key={e.id} value={e.id}>
                  {e.name}
                </Option>
              ))}
            </Select>
          </div>
        </Space>
      </Card>

      {/* 请求和响应面板 */}
      <Card>
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          {/* 请求配置 */}
          <TabPane tab="请求配置" key="request">
            <Form form={form} layout="vertical" initialValues={{ method: 'GET' }}>
              <Space style={{ width: '100%', marginBottom: 16 }}>
                <Form.Item
                  name="method"
                  rules={[{ required: true, message: '请选择请求方法' }]}
                  style={{ marginBottom: 0 }}
                >
                  <Select style={{ width: 120 }}>
                    <Option value="GET">GET</Option>
                    <Option value="POST">POST</Option>
                    <Option value="PUT">PUT</Option>
                    <Option value="DELETE">DELETE</Option>
                    <Option value="PATCH">PATCH</Option>
                  </Select>
                </Form.Item>

                <Form.Item
                  name="url"
                  rules={[{ required: true, message: '请输入请求URL' }]}
                  style={{ marginBottom: 0, flex: 1 }}
                >
                  <Input placeholder="请输入请求路径，例如: /api/users" />
                </Form.Item>

                <Button
                  type="primary"
                  icon={<SendOutlined />}
                  onClick={handleSendRequest}
                  loading={sendRequestMutation.isPending}
                  disabled={!selectedProjectId || !selectedEnvironmentId}
                >
                  发送请求
                </Button>
              </Space>

              <Form.Item label="请求头 (JSON)" name="headers">
                <TextArea
                  rows={4}
                  placeholder='{"Content-Type": "application/json"}'
                  style={{ fontFamily: 'monospace' }}
                />
              </Form.Item>

              <Form.Item label="查询参数 (JSON)" name="query">
                <TextArea
                  rows={3}
                  placeholder='{"page": "1", "size": "10"}'
                  style={{ fontFamily: 'monospace' }}
                />
              </Form.Item>

              <Form.Item label="请求体 (JSON)" name="body">
                <TextArea
                  rows={6}
                  placeholder='{"name": "test", "value": 123}'
                  style={{ fontFamily: 'monospace' }}
                />
              </Form.Item>
            </Form>
          </TabPane>

          {/* 响应结果 */}
          <TabPane tab="响应结果" key="response">
            {responseData ? (
              <div>
                <Descriptions column={2} bordered style={{ marginBottom: 16 }}>
                  <Descriptions.Item label="状态码">
                    <Badge
                      status={
                        responseData.status_code >= 200 && responseData.status_code < 300
                          ? 'success'
                          : 'error'
                      }
                      text={responseData.status_code}
                    />
                  </Descriptions.Item>
                  <Descriptions.Item label="响应时间">
                    <ClockCircleOutlined style={{ marginRight: 4 }} />
                    {responseData.duration}ms
                  </Descriptions.Item>
                  <Descriptions.Item label="匹配规则" span={2}>
                    {responseData.matched_rule_name ? (
                      <Tag color="green">{responseData.matched_rule_name}</Tag>
                    ) : (
                      <Tag color="default">未匹配</Tag>
                    )}
                  </Descriptions.Item>
                </Descriptions>

                <Card title="响应头" size="small" style={{ marginBottom: 16 }}>
                  <pre style={{ margin: 0, fontSize: 12 }}>
                    {JSON.stringify(responseData.headers || {}, null, 2)}
                  </pre>
                </Card>

                <Card title="响应体" size="small">
                  <pre style={{ margin: 0, fontSize: 12 }}>
                    {typeof responseData.body === 'string'
                      ? responseData.body
                      : JSON.stringify(responseData.body, null, 2)}
                  </pre>
                </Card>
              </div>
            ) : (
              <Empty description="暂无响应数据，请先发送请求" />
            )}
          </TabPane>

          {/* 测试历史 */}
          <TabPane
            tab={
              <span>
                <HistoryOutlined /> 测试历史
              </span>
            }
            key="history"
          >
            <div style={{ marginBottom: 16 }}>
              <Popconfirm
                title="确认清空所有历史记录？"
                onConfirm={handleClearHistory}
                okText="确认"
                cancelText="取消"
              >
                <Button
                  icon={<ClearOutlined />}
                  disabled={history.length === 0}
                  loading={clearHistoryMutation.isPending}
                >
                  清空历史
                </Button>
              </Popconfirm>
            </div>

            <Table
              columns={historyColumns}
              dataSource={history}
              rowKey="id"
              loading={historyLoading}
              pagination={{ pageSize: 10, showTotal: (total) => `共 ${total} 条` }}
            />
          </TabPane>
        </Tabs>
      </Card>
    </div>
  )
}

export default MockTest
