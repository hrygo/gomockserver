import React, { useState } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Input,
  Select,
  Tag,
  Switch,
  Popconfirm,
  Badge,
  message,
  Tooltip,
} from 'antd'
import {
  PlusOutlined,
  SearchOutlined,
  DeleteOutlined,
  CopyOutlined,
  EditOutlined,
  FilterOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import type { TableRowSelection as AntdTableRowSelection } from 'antd/lib/table/interface'
import {
  useRules,
  useDeleteRule,
  useUpdateRule,
  useBatchToggleRules,
  useBatchDeleteRules,
  useCopyRule,
  useCreateRule,
} from '@/hooks/useRules'
import type { Rule, Protocol, CreateRuleInput } from '@/types/rule'
import RuleForm from '@/components/RuleForm'

const { Search } = Input
const { Option } = Select

const Rules: React.FC = () => {
  // 状态管理
  const [keyword, setKeyword] = useState('')
  const [protocolFilter, setProtocolFilter] = useState<Protocol | undefined>()
  const [enabledFilter, setEnabledFilter] = useState<boolean | undefined>()
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])
  const [ruleFormVisible, setRuleFormVisible] = useState(false)
  const [currentRule, setCurrentRule] = useState<Rule | null>(null)

  // API Hooks
  const { data: rules = [], isLoading } = useRules({
    keyword,
    protocol: protocolFilter,
    enabled: enabledFilter,
  })
  const createRuleMutation = useCreateRule()
  const deleteRuleMutation = useDeleteRule()
  const updateRuleMutation = useUpdateRule()
  const batchToggleMutation = useBatchToggleRules()
  const batchDeleteMutation = useBatchDeleteRules()
  const copyRuleMutation = useCopyRule()

  // 打开创建规则弹窗
  const handleCreateRule = () => {
    setCurrentRule(null)
    setRuleFormVisible(true)
  }

  // 打开编辑规则弹窗
  const handleEditRule = (rule: Rule) => {
    setCurrentRule(rule)
    setRuleFormVisible(true)
  }

  // 提交规则表单
  const handleRuleSubmit = async (values: CreateRuleInput) => {
    if (currentRule) {
      // 更新规则
      await updateRuleMutation.mutateAsync({
        id: currentRule.id,
        data: values,
      })
    } else {
      // 创建规则
      await createRuleMutation.mutateAsync(values)
    }
    setRuleFormVisible(false)
  }

  // 搜索处理
  const handleSearch = (value: string) => {
    setKeyword(value)
  }

  // 重置过滤器
  const handleResetFilters = () => {
    setKeyword('')
    setProtocolFilter(undefined)
    setEnabledFilter(undefined)
  }

  // 删除规则
  const handleDelete = async (id: string) => {
    await deleteRuleMutation.mutateAsync(id)
  }

  // 切换启用状态
  const handleToggleEnabled = async (id: string, enabled: boolean) => {
    await updateRuleMutation.mutateAsync({
      id,
      data: { enabled },
    })
  }

  // 复制规则
  const handleCopy = async (id: string) => {
    await copyRuleMutation.mutateAsync(id)
  }

  // 批量启用
  const handleBatchEnable = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要启用的规则')
      return
    }
    await batchToggleMutation.mutateAsync({
      ids: selectedRowKeys as string[],
      enabled: true,
    })
    setSelectedRowKeys([])
  }

  // 批量禁用
  const handleBatchDisable = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要禁用的规则')
      return
    }
    await batchToggleMutation.mutateAsync({
      ids: selectedRowKeys as string[],
      enabled: false,
    })
    setSelectedRowKeys([])
  }

  // 批量删除
  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要删除的规则')
      return
    }
    await batchDeleteMutation.mutateAsync(selectedRowKeys as string[])
    setSelectedRowKeys([])
  }

  // 表格列定义
  const columns: ColumnsType<Rule> = [
    {
      title: '规则名称',
      dataIndex: 'name',
      key: 'name',
      width: '20%',
      ellipsis: true,
      render: (text: string, record) => (
        <div>
          <div style={{ fontWeight: 500 }}>
            {text}
            {!record.enabled && (
              <Tag color="default" style={{ marginLeft: 8 }}>
                已禁用
              </Tag>
            )}
          </div>
          {record.description && (
            <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>
              {record.description}
            </div>
          )}
        </div>
      ),
    },
    {
      title: '协议',
      dataIndex: 'protocol',
      key: 'protocol',
      width: '8%',
      render: (protocol: Protocol) => (
        <Tag color={protocol === 'HTTPS' ? 'green' : 'blue'}>{protocol}</Tag>
      ),
    },
    {
      title: '匹配类型',
      dataIndex: 'match_type',
      key: 'match_type',
      width: '10%',
      render: (type: string) => {
        const colorMap: Record<string, string> = {
          Simple: 'default',
          Regex: 'orange',
          Script: 'purple',
        }
        return <Tag color={colorMap[type] || 'default'}>{type}</Tag>
      },
    },
    {
      title: '匹配条件',
      dataIndex: 'match_condition',
      key: 'match_condition',
      width: '18%',
      ellipsis: true,
      render: (condition) => {
        const method = Array.isArray(condition.method)
          ? condition.method.join(', ')
          : condition.method || 'ANY'
        const path = condition.path || '/'
        return (
          <Tooltip title={`${method} ${path}`}>
            <span>
              {method} {path}
            </span>
          </Tooltip>
        )
      },
    },
    {
      title: '响应状态',
      dataIndex: 'response',
      key: 'response_status',
      width: '10%',
      render: (response) => {
        const statusCode = response?.content?.status_code || 200
        const color = statusCode >= 200 && statusCode < 300 ? 'success' : 'error'
        return <Badge status={color} text={statusCode} />
      },
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: '8%',
      sorter: (a, b) => a.priority - b.priority,
      render: (priority: number) => <Tag color="cyan">{priority}</Tag>,
    },
    {
      title: '标签',
      dataIndex: 'tags',
      key: 'tags',
      width: '12%',
      render: (tags: string[]) =>
        tags && tags.length > 0 ? (
          <div>
            {tags.slice(0, 2).map((tag) => (
              <Tag key={tag} style={{ marginBottom: 4 }}>
                {tag}
              </Tag>
            ))}
            {tags.length > 2 && (
              <Tooltip title={tags.slice(2).join(', ')}>
                <Tag>+{tags.length - 2}</Tag>
              </Tooltip>
            )}
          </div>
        ) : (
          '-'
        ),
    },
    {
      title: '启用',
      dataIndex: 'enabled',
      key: 'enabled',
      width: '8%',
      render: (enabled: boolean, record) => (
        <Switch
          checked={enabled}
          onChange={(checked) => handleToggleEnabled(record.id, checked)}
          loading={updateRuleMutation.isPending}
        />
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: '12%',
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="编辑">
            <Button
              type="link"
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEditRule(record)}
            >
              编辑
            </Button>
          </Tooltip>
          <Tooltip title="复制">
            <Button
              type="link"
              size="small"
              icon={<CopyOutlined />}
              onClick={() => handleCopy(record.id)}
              loading={copyRuleMutation.isPending}
            >
              复制
            </Button>
          </Tooltip>
          <Popconfirm
            title="确认删除"
            description="删除后不可恢复，确认删除吗?"
            onConfirm={() => handleDelete(record.id)}
            okText="确认"
            cancelText="取消"
          >
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
              loading={deleteRuleMutation.isPending}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  // 行选择配置
  const rowSelection: AntdTableRowSelection<Rule> = {
    selectedRowKeys,
    onChange: (keys: React.Key[]) => setSelectedRowKeys(keys),
  }

  return (
    <div>
      <h1 style={{ marginBottom: 24 }}>规则管理</h1>

      {/* 搜索和过滤区域 */}
      <Card style={{ marginBottom: 16 }}>
        <Space size="middle" wrap>
          <Search
            placeholder="搜索规则名称、描述、标签"
            allowClear
            style={{ width: 300 }}
            onSearch={handleSearch}
            enterButton={<SearchOutlined />}
          />

          <Select
            placeholder="协议类型"
            allowClear
            style={{ width: 120 }}
            value={protocolFilter}
            onChange={setProtocolFilter}
          >
            <Option value="HTTP">HTTP</Option>
            <Option value="HTTPS">HTTPS</Option>
          </Select>

          <Select
            placeholder="启用状态"
            allowClear
            style={{ width: 120 }}
            value={enabledFilter}
            onChange={setEnabledFilter}
          >
            <Option value={true}>已启用</Option>
            <Option value={false}>已禁用</Option>
          </Select>

          <Button icon={<FilterOutlined />} onClick={handleResetFilters}>
            重置
          </Button>
        </Space>
      </Card>

      {/* 批量操作和创建按钮 */}
      <Card>
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
          <Space>
            <Button
              type="default"
              onClick={handleBatchEnable}
              disabled={selectedRowKeys.length === 0}
              loading={batchToggleMutation.isPending}
            >
              批量启用
            </Button>
            <Button
              type="default"
              onClick={handleBatchDisable}
              disabled={selectedRowKeys.length === 0}
              loading={batchToggleMutation.isPending}
            >
              批量禁用
            </Button>
            <Popconfirm
              title="批量删除确认"
              description={`确认删除选中的 ${selectedRowKeys.length} 条规则吗?`}
              onConfirm={handleBatchDelete}
              disabled={selectedRowKeys.length === 0}
              okText="确认"
              cancelText="取消"
            >
              <Button
                danger
                disabled={selectedRowKeys.length === 0}
                loading={batchDeleteMutation.isPending}
                icon={<DeleteOutlined />}
              >
                批量删除
              </Button>
            </Popconfirm>
            {selectedRowKeys.length > 0 && (
              <span style={{ marginLeft: 8, color: '#999' }}>
                已选择 {selectedRowKeys.length} 项
              </span>
            )}
          </Space>

          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreateRule}>
            创建规则
          </Button>
        </div>

        {/* 规则列表表格 */}
        <Table
          columns={columns}
          dataSource={rules}
          rowKey="id"
          loading={isLoading}
          rowSelection={rowSelection}
          pagination={{
            pageSize: 10,
            showTotal: (total) => `共 ${total} 条`,
            showSizeChanger: true,
            showQuickJumper: true,
          }}
          scroll={{ x: 1400 }}
        />
      </Card>

      {/* 规则表单弹窗 */}
      <RuleForm
        visible={ruleFormVisible}
        rule={currentRule}
        onCancel={() => setRuleFormVisible(false)}
        onSubmit={handleRuleSubmit}
      />
    </div>
  )
}

export default Rules
