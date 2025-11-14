import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Card,
  Button,
  Table,
  Space,
  Modal,
  Form,
  Input,
  Popconfirm,
  Tag,
  Empty,
  Spin,
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  FolderOutlined,
  SearchOutlined,
  EyeOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { useProjects, useCreateProject, useUpdateProject, useDeleteProject } from '@/hooks/useProjects'
import type { Project, CreateProjectInput } from '@/types/project'

const Projects: React.FC = () => {
  const navigate = useNavigate()
  const [searchText, setSearchText] = useState('')
  const [createModalVisible, setCreateModalVisible] = useState(false)
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [editingProject, setEditingProject] = useState<Project | null>(null)

  const [createForm] = Form.useForm()
  const [editForm] = Form.useForm()

  // 数据获取
  const { data: projects, isLoading } = useProjects()
  const createMutation = useCreateProject()
  const updateMutation = useUpdateProject()
  const deleteMutation = useDeleteProject()

  // 过滤项目列表
  const filteredProjects = projects?.filter((project) =>
    project.name.toLowerCase().includes(searchText.toLowerCase())
  )

  // 创建项目
  const handleCreate = async () => {
    try {
      const values = await createForm.validateFields()
      await createMutation.mutateAsync(values as CreateProjectInput)
      setCreateModalVisible(false)
      createForm.resetFields()
    } catch (error) {
      console.error('Create failed:', error)
    }
  }

  // 更新项目
  const handleUpdate = async () => {
    if (!editingProject) return

    try {
      const values = await editForm.validateFields()
      await updateMutation.mutateAsync({
        id: editingProject.id,
        data: values,
      })
      setEditModalVisible(false)
      setEditingProject(null)
      editForm.resetFields()
    } catch (error) {
      console.error('Update failed:', error)
    }
  }

  // 删除项目
  const handleDelete = async (id: string) => {
    await deleteMutation.mutateAsync(id)
  }

  // 打开编辑弹窗
  const openEditModal = (project: Project) => {
    setEditingProject(project)
    editForm.setFieldsValue({
      name: project.name,
      description: project.description,
    })
    setEditModalVisible(true)
  }

  // 表格列定义
  const columns: ColumnsType<Project> = [
    {
      title: '项目名称',
      dataIndex: 'name',
      key: 'name',
      render: (text) => (
        <Space>
          <FolderOutlined style={{ color: '#1890ff' }} />
          <span>{text}</span>
        </Space>
      ),
    },
    {
      title: '工作空间',
      dataIndex: 'workspace_id',
      key: 'workspace_id',
      render: (text) => <Tag>{text}</Tag>,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (text) => text || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (text) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_, record) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/projects/${record.id}`)}
          >
            查看
          </Button>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => openEditModal(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个项目吗？"
            description="删除后将无法恢复，且会删除该项目下的所有规则"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
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
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h1>项目管理</h1>
        <Space>
          <Input
            placeholder="搜索项目名称"
            prefix={<SearchOutlined />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: 200 }}
            allowClear
          />
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalVisible(true)}
          >
            创建项目
          </Button>
        </Space>
      </div>

      <Card>
        {isLoading ? (
          <div style={{ textAlign: 'center', padding: '50px 0' }}>
            <Spin size="large" />
          </div>
        ) : filteredProjects && filteredProjects.length > 0 ? (
          <Table
            columns={columns}
            dataSource={filteredProjects}
            rowKey="id"
            pagination={{
              pageSize: 10,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 个项目`,
            }}
          />
        ) : (
          <Empty description="暂无项目，点击上方按钮创建" />
        )}
      </Card>

      {/* 创建项目弹窗 */}
      <Modal
        title="创建项目"
        open={createModalVisible}
        onOk={handleCreate}
        onCancel={() => {
          setCreateModalVisible(false)
          createForm.resetFields()
        }}
        confirmLoading={createMutation.isPending}
      >
        <Form form={createForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="name"
            label="项目名称"
            rules={[
              { required: true, message: '请输入项目名称' },
              { max: 50, message: '项目名称不能超过50个字符' },
            ]}
          >
            <Input placeholder="请输入项目名称" />
          </Form.Item>
          <Form.Item
            name="workspace_id"
            label="工作空间ID"
            rules={[{ required: true, message: '请输入工作空间ID' }]}
            initialValue="default"
          >
            <Input placeholder="请输入工作空间ID" />
          </Form.Item>
          <Form.Item name="description" label="项目描述">
            <Input.TextArea
              rows={4}
              placeholder="请输入项目描述（可选）"
              maxLength={200}
              showCount
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* 编辑项目弹窗 */}
      <Modal
        title="编辑项目"
        open={editModalVisible}
        onOk={handleUpdate}
        onCancel={() => {
          setEditModalVisible(false)
          setEditingProject(null)
          editForm.resetFields()
        }}
        confirmLoading={updateMutation.isPending}
      >
        <Form form={editForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="name"
            label="项目名称"
            rules={[
              { required: true, message: '请输入项目名称' },
              { max: 50, message: '项目名称不能超过50个字符' },
            ]}
          >
            <Input placeholder="请输入项目名称" />
          </Form.Item>
          <Form.Item name="description" label="项目描述">
            <Input.TextArea
              rows={4}
              placeholder="请输入项目描述（可选）"
              maxLength={200}
              showCount
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Projects
