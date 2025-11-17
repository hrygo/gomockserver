import React, { useEffect, useState } from 'react'
import {
  Modal,
  Form,
  Input,
  Select,
  Switch,
  InputNumber,
  Tabs,
  Space,
  Card,
  message,
} from 'antd'
import type { Rule, CreateRuleInput } from '@/types/rule'

const { Option } = Select
const { TextArea } = Input
const { TabPane } = Tabs

interface RuleFormProps {
  visible: boolean
  rule?: Rule | null
  projectId?: string
  environmentId?: string
  onCancel: () => void
  onSubmit: (values: CreateRuleInput) => void
}

const RuleForm: React.FC<RuleFormProps> = ({
  visible,
  rule,
  projectId,
  environmentId,
  onCancel,
  onSubmit,
}) => {
  const [form] = Form.useForm()
  const [activeTab, setActiveTab] = useState('basic')

  useEffect(() => {
    if (visible) {
      if (rule) {
        // 编辑模式 - 填充现有数据
        const fieldsValue: any = {
          name: rule.name,
          description: rule.description,
          protocol: rule.protocol,
          match_type: rule.match_type,
          priority: rule.priority,
          enabled: rule.enabled,
          tags: rule.tags,
          // 匹配条件
          method: rule.match_condition.method,
          path: rule.match_condition.path,
          query: rule.match_condition.query,
          headers: rule.match_condition.headers,
          ip_whitelist: rule.match_condition.ip_whitelist,
          match_script: rule.match_condition.script,
          // 响应配置
          response_type: rule.response.type,
          status_code: rule.response.content.status_code,
          content_type: rule.response.content.content_type,
          response_headers: rule.response.content.headers,
          response_body: rule.response.content.body,
          response_script: rule.response.content.script,
        }
        
        // 延迟配置
        if (rule.delay) {
          fieldsValue.delay_type = rule.delay.type
          switch (rule.delay.type) {
            case 'fixed':
              fieldsValue.delay_fixed = rule.delay.fixed
              break
            case 'random':
              fieldsValue.delay_random_min = rule.delay.min
              fieldsValue.delay_random_max = rule.delay.max
              break
            case 'normal':
              fieldsValue.delay_normal_mean = rule.delay.mean
              fieldsValue.delay_normal_stddev = rule.delay.std_dev
              break
            case 'step':
              fieldsValue.delay_step_fixed = rule.delay.fixed
              fieldsValue.delay_step_step = rule.delay.step
              fieldsValue.delay_step_limit = rule.delay.limit
              break
          }
        }

        form.setFieldsValue(fieldsValue)
      } else {
        // 创建模式 - 设置默认值
        form.resetFields()
        form.setFieldsValue({
          protocol: 'HTTP',
          match_type: 'Simple',
          priority: 100,
          enabled: true,
          response_type: 'Static',
          status_code: 200,
          content_type: 'JSON',
        })
      }
    }
  }, [visible, rule, form])

  const handleOk = async () => {
    try {
      const values = await form.validateFields()

      // 构造响应配置
      const response: any = {
        type: values.response_type,
        content: {
          status_code: values.status_code,
          content_type: values.content_type,
          headers: values.response_headers,
          body: values.response_body,
        },
      }

      // 如果是脚本响应类型，添加脚本内容
      if (values.response_type === 'Script') {
        response.content.script = values.response_script
      }

      // 构造匹配条件
      const match_condition: any = {}
      if (values.method) match_condition.method = values.method
      if (values.path) match_condition.path = values.path
      if (values.query) match_condition.query = values.query
      if (values.headers) match_condition.headers = values.headers
      if (values.ip_whitelist) match_condition.ip_whitelist = values.ip_whitelist
      
      // 如果是脚本匹配类型，添加脚本内容
      if (values.match_type === 'Script') {
        match_condition.script = values.match_script
      }

      // 构造延迟配置
      const delay: any = {}
      const delayType = values.delay_type
      if (delayType === 'fixed' && values.delay_fixed) {
        delay.type = 'fixed'
        delay.fixed = values.delay_fixed
      } else if (delayType === 'random' && values.delay_random_min && values.delay_random_max) {
        delay.type = 'random'
        delay.min = values.delay_random_min
        delay.max = values.delay_random_max
      } else if (delayType === 'normal' && values.delay_normal_mean && values.delay_normal_stddev) {
        delay.type = 'normal'
        delay.mean = values.delay_normal_mean
        delay.std_dev = values.delay_normal_stddev
      } else if (delayType === 'step') {
        delay.type = 'step'
        if (values.delay_step_fixed) delay.fixed = values.delay_step_fixed
        if (values.delay_step_step) delay.step = values.delay_step_step
        if (values.delay_step_limit) delay.limit = values.delay_step_limit
      }

      const ruleData: CreateRuleInput = {
        name: values.name,
        project_id: projectId || rule?.project_id || '',
        environment_id: environmentId || rule?.environment_id || '',
        protocol: values.protocol,
        match_type: values.match_type,
        priority: values.priority,
        enabled: values.enabled,
        tags: values.tags,
        description: values.description,
        match_condition,
        response,
        delay: Object.keys(delay).length > 0 ? delay : undefined,
      }

      onSubmit(ruleData)
    } catch (error) {
      message.error('请检查表单输入')
    }
  }

  return (
    <Modal
      title={rule ? '编辑规则' : '创建规则'}
      open={visible}
      onOk={handleOk}
      onCancel={onCancel}
      width={800}
      destroyOnClose
    >
      <Form form={form} layout="vertical" autoComplete="off">
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          {/* 基础信息 */}
          <TabPane tab="基础信息" key="basic">
            <Form.Item
              label="规则名称"
              name="name"
              rules={[
                { required: true, message: '请输入规则名称' },
                { min: 2, max: 100, message: '规则名称长度为2-100个字符' },
              ]}
            >
              <Input placeholder="请输入规则名称" />
            </Form.Item>

            <Form.Item label="描述" name="description" rules={[{ max: 500 }]}>
              <TextArea rows={2} placeholder="请输入规则描述（可选）" maxLength={500} showCount />
            </Form.Item>

            <Space size="large" style={{ width: '100%' }}>
              <Form.Item
                label="协议"
                name="protocol"
                rules={[{ required: true, message: '请选择协议' }]}
              >
                <Select style={{ width: 120 }}>
                  <Option value="HTTP">HTTP</Option>
                  <Option value="HTTPS">HTTPS</Option>
                  <Option value="WebSocket">WebSocket</Option>
                </Select>
              </Form.Item>

              <Form.Item
                label="匹配类型"
                name="match_type"
                rules={[{ required: true, message: '请选择匹配类型' }]}
              >
                <Select style={{ width: 120 }}>
                  <Option value="Simple">简单匹配</Option>
                  <Option value="Regex">正则匹配</Option>
                  <Option value="Script">脚本匹配</Option>
                </Select>
              </Form.Item>

              <Form.Item
                label="优先级"
                name="priority"
                rules={[{ required: true, message: '请输入优先级' }]}
              >
                <InputNumber min={1} max={999} style={{ width: 120 }} />
              </Form.Item>

              <Form.Item label="启用" name="enabled" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Space>

            <Form.Item label="标签" name="tags">
              <Select mode="tags" placeholder="输入标签并回车" style={{ width: '100%' }} />
            </Form.Item>
          </TabPane>

          {/* 匹配条件 */}
          <TabPane tab="匹配条件" key="match">
            <Form.Item 
              noStyle 
              shouldUpdate={(prevValues, currentValues) => 
                prevValues.match_type !== currentValues.match_type
              }
            >
              {({ getFieldValue }) => 
                getFieldValue('match_type') === 'Script' ? (
                  <Form.Item 
                    label="匹配脚本" 
                    name="match_script"
                    rules={[{ required: true, message: '请输入匹配脚本' }]}
                  >
                    <TextArea
                      rows={6}
                      placeholder={`输入 JavaScript 匹配脚本，例如:
// 可访问的变量:
// request: { id, protocol, path, headers, body, sourceIP, metadata }
// rule: { id, name, project_id, priority }

// 示例: 匹配包含特定头的请求
return request.headers['x-api-key'] === 'secret-key';

// 示例: 匹配特定路径模式
return request.path.startsWith('/api/v2/');

// 必须返回布尔值`}
                      />
                    </Form.Item>
                  ) : null
                }
              </Form.Item>
              
              {form.getFieldValue('match_type') !== 'Script' && (
                <>
                  <Form.Item label="HTTP 方法" name="method">
                    <Select mode="multiple" placeholder="选择HTTP方法，不选则匹配所有">
                      <Option value="GET">GET</Option>
                      <Option value="POST">POST</Option>
                      <Option value="PUT">PUT</Option>
                      <Option value="DELETE">DELETE</Option>
                      <Option value="PATCH">PATCH</Option>
                      <Option value="HEAD">HEAD</Option>
                      <Option value="OPTIONS">OPTIONS</Option>
                    </Select>
                  </Form.Item>

                  <Form.Item label="路径" name="path">
                    <Input placeholder="例如: /api/users 或 /api/users/.*（正则）" />
                  </Form.Item>

                  <Form.Item label="查询参数" name="query">
                    <TextArea
                      rows={3}
                      placeholder='JSON 格式，例如: {"page": "1", "size": "10"}'
                    />
                  </Form.Item>

                  <Form.Item label="请求头" name="headers">
                    <TextArea
                      rows={3}
                      placeholder='JSON 格式，例如: {"Content-Type": "application/json"}'
                    />
                  </Form.Item>

                  <Form.Item label="IP 白名单" name="ip_whitelist">
                    <Select mode="tags" placeholder="输入IP地址并回车，例如: 192.168.1.1" />
                  </Form.Item>
                </>
              )}
            </TabPane>

          {/* 响应配置 */}
          <TabPane tab="响应配置" key="response">
            <Form.Item
              label="响应类型"
              name="response_type"
              rules={[{ required: true, message: '请选择响应类型' }]}
            >
              <Select>
                <Option value="Static">静态响应</Option>
                <Option value="Dynamic">动态响应</Option>
                <Option value="Proxy">代理转发</Option>
                <Option value="Script">脚本响应</Option>
              </Select>
            </Form.Item>

            <Space size="large" style={{ width: '100%' }}>
              <Form.Item
                label="状态码"
                name="status_code"
                rules={[{ required: true, message: '请输入状态码' }]}
              >
                <InputNumber min={100} max={599} style={{ width: 120 }} />
              </Form.Item>

              <Form.Item
                label="内容类型"
                name="content_type"
                rules={[{ required: true, message: '请选择内容类型' }]}
              >
                <Select style={{ width: 150 }}>
                  <Option value="JSON">JSON</Option>
                  <Option value="XML">XML</Option>
                  <Option value="HTML">HTML</Option>
                  <Option value="Text">Text</Option>
                  <Option value="Binary">Binary</Option>
                </Select>
              </Form.Item>
            </Space>

            <Form.Item label="响应头" name="response_headers">
              <TextArea
                rows={3}
                placeholder='JSON 格式，例如: {"X-Custom-Header": "value"}'
              />
            </Form.Item>

            <Form.Item 
              noStyle 
              shouldUpdate={(prevValues, currentValues) => 
                prevValues.content_type !== currentValues.content_type
              }
            >
              {({ getFieldValue }) => 
                getFieldValue('content_type') === 'Binary' ? (
                  <Form.Item 
                    label="Base64 编码的二进制数据" 
                    name="response_body"
                    rules={[{ required: true, message: '请输入 Base64 编码的二进制数据' }]}
                  >
                    <TextArea
                      rows={6}
                      placeholder='输入 Base64 编码的二进制数据'
                    />
                  </Form.Item>
                ) : null
              }
            </Form.Item>
            
            <Form.Item 
              noStyle 
              shouldUpdate={(prevValues, currentValues) => 
                prevValues.content_type !== currentValues.content_type
              }
            >
              {({ getFieldValue }) => 
                getFieldValue('content_type') !== 'Binary' ? (
                  <Form.Item label="响应体" name="response_body">
                    <TextArea
                      rows={6}
                      placeholder='根据内容类型输入响应内容，例如: {"message": "success"}'
                    />
                  </Form.Item>
                ) : null
              }
            </Form.Item>
            
            {/* 脚本响应编辑器 */}
            <Form.Item 
              noStyle 
              shouldUpdate={(prevValues, currentValues) => 
                prevValues.response_type !== currentValues.response_type
              }
            >
              {({ getFieldValue }) => 
                getFieldValue('response_type') === 'Script' ? (
                  <Form.Item 
                    label="脚本代码" 
                    name="response_script"
                    rules={[{ required: true, message: '请输入脚本代码' }]}
                  >
                    <TextArea
                      rows={8}
                      placeholder='输入 JavaScript 脚本代码，例如: return { message: "Hello from script" };'
                    />
                  </Form.Item>
                ) : null
              }
            </Form.Item>
          </TabPane>

          {/* 延迟配置 */}
          <TabPane tab="延迟配置" key="delay">
            <Form.Item
              label="延迟类型"
              name="delay_type"
            >
              <Select>
                <Option value="">无延迟</Option>
                <Option value="fixed">固定延迟</Option>
                <Option value="random">随机延迟</Option>
                <Option value="normal">正态分布延迟</Option>
                <Option value="step">阶梯延迟</Option>
              </Select>
            </Form.Item>
            
            <Form.Item 
              noStyle 
              shouldUpdate={(prevValues, currentValues) => 
                prevValues.delay_type !== currentValues.delay_type
              }
            >
              {({ getFieldValue }) => {
                const delayType = getFieldValue('delay_type')
                
                return (
                  <>
                    {/* 固定延迟 */}
                    {delayType === 'fixed' && (
                      <Card title="固定延迟" size="small" style={{ marginBottom: 16 }}>
                        <Form.Item label="延迟时间（毫秒）" name="delay_fixed">
                          <InputNumber min={0} max={60000} style={{ width: '100%' }} placeholder="0" />
                        </Form.Item>
                      </Card>
                    )}
                    
                    {/* 随机延迟 */}
                    {delayType === 'random' && (
                      <Card title="随机延迟" size="small" style={{ marginBottom: 16 }}>
                        <Space size="large">
                          <Form.Item label="最小值（毫秒）" name="delay_random_min">
                            <InputNumber min={0} max={60000} style={{ width: 150 }} placeholder="0" />
                          </Form.Item>
                          
                          <Form.Item label="最大值（毫秒）" name="delay_random_max">
                            <InputNumber min={0} max={60000} style={{ width: 150 }} placeholder="0" />
                          </Form.Item>
                        </Space>
                      </Card>
                    )}
                    
                    {/* 正态分布延迟 */}
                    {delayType === 'normal' && (
                      <Card title="正态分布延迟" size="small" style={{ marginBottom: 16 }}>
                        <Space size="large">
                          <Form.Item label="均值（毫秒）" name="delay_normal_mean">
                            <InputNumber min={0} max={60000} style={{ width: 150 }} placeholder="1000" />
                          </Form.Item>
                          
                          <Form.Item label="标准差（毫秒）" name="delay_normal_stddev">
                            <InputNumber min={0} max={60000} style={{ width: 150 }} placeholder="500" />
                          </Form.Item>
                        </Space>
                      </Card>
                    )}
                    
                    {/* 阶梯延迟 */}
                    {delayType === 'step' && (
                      <Card title="阶梯延迟" size="small" style={{ marginBottom: 16 }}>
                        <Space size="large" wrap>
                          <Form.Item label="基础延迟（毫秒）" name="delay_step_fixed">
                            <InputNumber min={0} max={60000} style={{ width: 150 }} placeholder="0" />
                          </Form.Item>
                          
                          <Form.Item label="步长（毫秒）" name="delay_step_step">
                            <InputNumber min={0} max={60000} style={{ width: 150 }} placeholder="100" />
                          </Form.Item>
                          
                          <Form.Item label="上限（毫秒）" name="delay_step_limit">
                            <InputNumber min={0} max={60000} style={{ width: 150 }} placeholder="5000" />
                          </Form.Item>
                        </Space>
                      </Card>
                    )}
                  </>
                )
              }}
            </Form.Item>
          </TabPane>
        </Tabs>
      </Form>
    </Modal>
  )
}

export default RuleForm
