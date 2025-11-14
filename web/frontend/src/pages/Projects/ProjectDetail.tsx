import React, { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Descriptions, Button, Space, Spin, Empty, Statistic, Row, Col, Table, Tag, Popconfirm } from 'antd'
import { ArrowLeftOutlined, EditOutlined, PlusOutlined, DeleteOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { useProject } from '@/hooks/useProjects'
import {
  useEnvironments,
  useCreateEnvironment,
  useUpdateEnvironment,
  useDeleteEnvironment,
} from '@/hooks/useEnvironments'
import type { Environment, CreateEnvironmentInput } from '@/types/environment'
import EnvironmentForm from '@/components/EnvironmentForm'

const ProjectDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: project, isLoading } = useProject(id!)

  // 环境管理状态
  const [envFormVisible, setEnvFormVisible] = useState(false)
  const [currentEnvironment, setCurrentEnvironment] = useState<Environment | null>(null)

  // 环境管理 Hooks
  const { data: environments = [], isLoading: envLoading } = useEnvironments(id || '')
  const createEnvMutation = useCreateEnvironment()
  const updateEnvMutation = useUpdateEnvironment()
  const deleteEnvMutation = useDeleteEnvironment()

  // 打开创建环境弹窗
  const handleCreateEnvironment = () => {
    setCurrentEnvironment(null)
    setEnvFormVisible(true)
  }

  // 打开编辑环境弹窗
  const handleEditEnvironment = (env: Environment) => {
    setCurrentEnvironment(env)
    setEnvFormVisible(true)
  }

  // 删除环境
  const handleDeleteEnvironment = async (envId: string) => {
    await deleteEnvMutation.mutateAsync({ id: envId, projectId: id! })
  }

  // 提交环境表单
  const handleEnvironmentSubmit = async (values: CreateEnvironmentInput) => {
    if (currentEnvironment) {
      // 更新环境
      await updateEnvMutation.mutateAsync({
        id: currentEnvironment.id,
        data: {
          name: values.name,
          base_url: values.base_url,
          description: values.description,
        },
      })
    } else {
      // 创建环境
      await createEnvMutation.mutateAsync(values)
    }
    setEnvFormVisible(false)
  }

  // 环境表格列定义
  const environmentColumns: ColumnsType<Environment> = [
    {
      title: '环境名称',
      dataIndex: 'name',
      key: 'name',
      width: '20%',
      render: (text: string) => <strong>{text}</strong>,
    },
    {
      title: 'Base URL',
      dataIndex: 'base_url',
      key: 'base_url',
      width: '30%',
      render: (text: string) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      width: '25%',
      render: (text: string) => text || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: '15%',
      render: (text: string) => dayjs(text).format('YYYY-MM-DD'),
    },
    {
      title: '操作',
      key: 'action',
      width: '10%',
      render: (_, record) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            onClick={() => handleEditEnvironment(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除"
            description="删除环境后，该环境下的所有规则将不可用，确认删除吗？"
            onConfirm={() => handleDeleteEnvironment(record.id)}
            okText="确认"
            cancelText="取消"
          >
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px 0' }}>
        <Spin size="large" />
      </div>
    )
  }

  if (!project) {
    return (
      <Empty
        description="项目不存在"
        style={{ marginTop: 100 }}
      >
        <Button type="primary" onClick={() => navigate('/projects')}>
          返回项目列表
        </Button>
      </Empty>
    )
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/projects')}>
          返回
        </Button>
      </Space>

      <Card
        title={project.name}
        extra={
          <Button icon={<EditOutlined />} onClick={() => navigate(`/projects/${id}/edit`)}>
            编辑项目
          </Button>
        }
      >
        <Descriptions column={2}>
          <Descriptions.Item label="项目ID">{project.id}</Descriptions.Item>
          <Descriptions.Item label="工作空间">{project.workspace_id}</Descriptions.Item>
          <Descriptions.Item label="创建时间">
            {dayjs(project.created_at).format('YYYY-MM-DD HH:mm:ss')}
          </Descriptions.Item>
          <Descriptions.Item label="更新时间">
            {dayjs(project.updated_at).format('YYYY-MM-DD HH:mm:ss')}
          </Descriptions.Item>
          <Descriptions.Item label="描述" span={2}>
            {project.description || '暂无描述'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={8}>
          <Card>
            <Statistic title="环境总数" value={environments.length} suffix="个" />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic title="规则总数" value={0} suffix="个" />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic title="启用规则" value={0} suffix="个" valueStyle={{ color: '#3f8600' }} />
          </Card>
        </Col>
      </Row>

      <Card
        title="环境列表"
        style={{ marginTop: 16 }}
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreateEnvironment}
          >
            创建环境
          </Button>
        }
      >
        {envLoading ? (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <Spin />
          </div>
        ) : environments.length > 0 ? (
          <Table
            columns={environmentColumns}
            dataSource={environments}
            rowKey="id"
            pagination={false}
          />
        ) : (
          <Empty description="暂无环境，点击右上角按钮创建" />
        )}
      </Card>

      <EnvironmentForm
        visible={envFormVisible}
        environment={currentEnvironment}
        projectId={id!}
        onCancel={() => setEnvFormVisible(false)}
        onSubmit={handleEnvironmentSubmit}
      />

      <Card title="最近规则" style={{ marginTop: 16 }} extra={<Button type="primary" icon={<PlusOutlined />}>创建规则</Button>}>
        <Empty description="暂无规则" />
      </Card>
    </div>
  )
}

export default ProjectDetail
