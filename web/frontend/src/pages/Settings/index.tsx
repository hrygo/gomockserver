import React, { useState } from 'react'
import {
  Card,
  Descriptions,
  Button,
  Space,
  Upload,
  Select,
  message,
  Modal,
  Alert,
  Badge,
  Statistic,
} from 'antd'
import {
  DownloadOutlined,
  UploadOutlined,
  InfoCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons'
import type { UploadFile } from 'antd'
import { useQuery } from '@tanstack/react-query'
import { useProjects } from '@/hooks/useProjects'
import { exportApi } from '@/api/export'
import { systemApi } from '@/api/system'
import type { ExportFormat } from '@/types/export'

const { Option } = Select

const Settings: React.FC = () => {
  const [selectedProjectId, setSelectedProjectId] = useState<string>()
  const [exportFormat, setExportFormat] = useState<ExportFormat>('json')
  const [importModalVisible, setImportModalVisible] = useState(false)
  const [importType, setImportType] = useState<'rules' | 'project'>('rules')
  const [fileList, setFileList] = useState<UploadFile[]>([])

  // 获取项目列表
  const { data: projects = [] } = useProjects()

  // 获取系统信息
  const { data: systemInfo } = useQuery({
    queryKey: ['system', 'info'],
    queryFn: async () => {
      const response = await systemApi.getInfo()
      return response.data
    },
  })

  // 获取健康状态
  const { data: healthStatus } = useQuery({
    queryKey: ['system', 'health'],
    queryFn: async () => {
      const response = await systemApi.getHealth()
      return response.data
    },
    refetchInterval: 10000, // 每10秒刷新
  })

  // 导出规则
  const handleExportRules = async () => {
    if (!selectedProjectId) {
      message.warning('请选择项目')
      return
    }

    try {
      const response = await exportApi.exportRules(selectedProjectId, undefined, exportFormat)
      const blob = new Blob([response.data])
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `rules_${selectedProjectId}_${Date.now()}.${exportFormat}`
      link.click()
      window.URL.revokeObjectURL(url)
      message.success('规则导出成功')
    } catch (error) {
      message.error('导出失败')
    }
  }

  // 导出项目
  const handleExportProject = async () => {
    if (!selectedProjectId) {
      message.warning('请选择项目')
      return
    }

    try {
      const response = await exportApi.exportProject(selectedProjectId, exportFormat)
      const blob = new Blob([response.data])
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `project_${selectedProjectId}_${Date.now()}.${exportFormat}`
      link.click()
      window.URL.revokeObjectURL(url)
      message.success('项目导出成功')
    } catch (error) {
      message.error('导出失败')
    }
  }

  // 处理导入
  const handleImport = async () => {
    if (fileList.length === 0) {
      message.warning('请选择文件')
      return
    }

    if (importType === 'rules' && !selectedProjectId) {
      message.warning('请选择项目')
      return
    }

    const file = fileList[0].originFileObj as File

    try {
      let response
      if (importType === 'rules') {
        response = await exportApi.importRules(selectedProjectId!, file)
      } else {
        response = await exportApi.importProject(file)
      }

      const result = response.data
      if (result.success) {
        message.success(`导入成功：${result.imported_count} 条`)
        setImportModalVisible(false)
        setFileList([])
      } else {
        Modal.error({
          title: '导入失败',
          content: (
            <div>
              <p>{result.message}</p>
              {result.errors && result.errors.length > 0 && (
                <ul>
                  {result.errors.map((err, idx) => (
                    <li key={idx}>{err}</li>
                  ))}
                </ul>
              )}
            </div>
          ),
        })
      }
    } catch (error) {
      message.error('导入失败')
    }
  }

  // 格式化运行时间
  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    return `${days}天 ${hours}小时 ${minutes}分钟`
  }

  return (
    <div>
      <h1 style={{ marginBottom: 24 }}>系统设置</h1>

      {/* 系统信息 */}
      <Card title="系统信息" style={{ marginBottom: 16 }}>
        <Descriptions column={2}>
          <Descriptions.Item label="系统版本">
            {systemInfo?.version || 'v0.6.0'}
          </Descriptions.Item>
          <Descriptions.Item label="构建时间">
            {systemInfo?.build_time || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="Go 版本">
            {systemInfo?.go_version || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="运行时间">
            {healthStatus ? formatUptime(healthStatus.uptime) : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="管理 API 地址">
            {systemInfo?.admin_api_url || 'http://localhost:8080/api/v1'}
          </Descriptions.Item>
          <Descriptions.Item label="Mock 服务地址">
            {systemInfo?.mock_service_url || 'http://localhost:9090'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* 健康状态 */}
      <Card title="健康状态" style={{ marginBottom: 16 }}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div>
            <Badge
              status={healthStatus?.status === 'healthy' ? 'success' : 'error'}
              text={
                <span style={{ fontSize: 16, fontWeight: 500 }}>
                  系统状态: {healthStatus?.status === 'healthy' ? '健康' : '异常'}
                </span>
              }
            />
          </div>

          <Space size="large">
            <Statistic
              title="数据库"
              value={healthStatus?.database ? '正常' : '异常'}
              prefix={
                healthStatus?.database ? (
                  <CheckCircleOutlined style={{ color: '#52c41a' }} />
                ) : (
                  <CloseCircleOutlined style={{ color: '#f5222d' }} />
                )
              }
            />
            <Statistic
              title="缓存"
              value={healthStatus?.cache ? '正常' : '异常'}
              prefix={
                healthStatus?.cache ? (
                  <CheckCircleOutlined style={{ color: '#52c41a' }} />
                ) : (
                  <CloseCircleOutlined style={{ color: '#f5222d' }} />
                )
              }
            />
          </Space>
        </Space>
      </Card>

      {/* 导入导出 */}
      <Card title="导入导出" style={{ marginBottom: 16 }}>
        <Alert
          message="说明"
          description="支持 JSON 和 YAML 格式的规则和项目导入导出。导出的文件包含全部配置信息。"
          type="info"
          showIcon
          icon={<InfoCircleOutlined />}
          style={{ marginBottom: 16 }}
        />

        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <div>
            <h3>导出</h3>
            <Space wrap>
              <Select
                placeholder="选择项目"
                style={{ width: 200 }}
                value={selectedProjectId}
                onChange={setSelectedProjectId}
              >
                {projects?.map((p) => (
                  <Option key={p.id} value={p.id}>
                    {p.name}
                  </Option>
                ))}
              </Select>

              <Select
                value={exportFormat}
                onChange={setExportFormat}
                style={{ width: 120 }}
              >
                <Option value="json">JSON</Option>
                <Option value="yaml">YAML</Option>
              </Select>

              <Button
                icon={<DownloadOutlined />}
                onClick={handleExportRules}
                disabled={!selectedProjectId}
              >
                导出规则
              </Button>

              <Button
                icon={<DownloadOutlined />}
                onClick={handleExportProject}
                disabled={!selectedProjectId}
              >
                导出项目
              </Button>
            </Space>
          </div>

          <div>
            <h3>导入</h3>
            <Button
              type="primary"
              icon={<UploadOutlined />}
              onClick={() => setImportModalVisible(true)}
            >
              导入文件
            </Button>
          </div>
        </Space>
      </Card>

      {/* 导入弹窗 */}
      <Modal
        title="导入文件"
        open={importModalVisible}
        onOk={handleImport}
        onCancel={() => {
          setImportModalVisible(false)
          setFileList([])
        }}
        okText="开始导入"
        cancelText="取消"
      >
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <div>
            <span style={{ marginRight: 8 }}>导入类型:</span>
            <Select
              value={importType}
              onChange={setImportType}
              style={{ width: 200 }}
            >
              <Option value="rules">规则</Option>
              <Option value="project">项目</Option>
            </Select>
          </div>

          {importType === 'rules' && (
            <div>
              <span style={{ marginRight: 8 }}>目标项目:</span>
              <Select
                placeholder="选择项目"
                style={{ width: 200 }}
                value={selectedProjectId}
                onChange={setSelectedProjectId}
              >
                {projects?.map((p) => (
                  <Option key={p.id} value={p.id}>
                    {p.name}
                  </Option>
                ))}
              </Select>
            </div>
          )}

          <Upload
            fileList={fileList}
            onChange={({ fileList: newFileList }) => setFileList(newFileList)}
            beforeUpload={() => false}
            maxCount={1}
            accept=".json,.yaml,.yml"
          >
            <Button icon={<UploadOutlined />}>选择文件</Button>
          </Upload>

          <Alert
            message="导入将会创建新的规则或项目，不会覆盖现有数据。"
            type="warning"
            showIcon
          />
        </Space>
      </Modal>
    </div>
  )
}

export default Settings
