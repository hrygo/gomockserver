import React, { useEffect } from 'react'
import { Modal, Form, Input, message } from 'antd'
import type { Environment, CreateEnvironmentInput } from '@/types/environment'

interface EnvironmentFormProps {
  visible: boolean
  environment?: Environment | null
  projectId: string
  onCancel: () => void
  onSubmit: (values: CreateEnvironmentInput) => void
}

const EnvironmentForm: React.FC<EnvironmentFormProps> = ({
  visible,
  environment,
  projectId,
  onCancel,
  onSubmit,
}) => {
  const [form] = Form.useForm()

  useEffect(() => {
    if (visible) {
      if (environment) {
        // 编辑模式 - 填充现有数据
        form.setFieldsValue({
          name: environment.name,
          base_url: environment.base_url,
          description: environment.description,
        })
      } else {
        // 创建模式 - 重置表单
        form.resetFields()
      }
    }
  }, [visible, environment, form])

  const handleOk = async () => {
    try {
      const values = await form.validateFields()
      onSubmit({
        ...values,
        project_id: projectId,
      })
    } catch (error) {
      message.error('请检查表单输入')
    }
  }

  return (
    <Modal
      title={environment ? '编辑环境' : '创建环境'}
      open={visible}
      onOk={handleOk}
      onCancel={onCancel}
      width={600}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        autoComplete="off"
      >
        <Form.Item
          label="环境名称"
          name="name"
          rules={[
            { required: true, message: '请输入环境名称' },
            { min: 2, max: 50, message: '环境名称长度为2-50个字符' },
          ]}
        >
          <Input placeholder="例如: 开发环境、测试环境、生产环境" />
        </Form.Item>

        <Form.Item
          label="Base URL"
          name="base_url"
          rules={[
            { required: true, message: '请输入Base URL' },
            {
              pattern: /^https?:\/\/.+/,
              message: '请输入有效的HTTP/HTTPS URL',
            },
          ]}
        >
          <Input placeholder="例如: http://localhost:8080" />
        </Form.Item>

        <Form.Item
          label="描述"
          name="description"
          rules={[{ max: 200, message: '描述不能超过200个字符' }]}
        >
          <Input.TextArea
            rows={3}
            placeholder="请输入环境描述（可选）"
            showCount
            maxLength={200}
          />
        </Form.Item>
      </Form>
    </Modal>
  )
}

export default EnvironmentForm
