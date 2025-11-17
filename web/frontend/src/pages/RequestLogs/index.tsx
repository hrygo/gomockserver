import React, { useState } from 'react';
import { 
  Table, Card, Form, Input, Select, DatePicker, Button, Space, Tag, 
  Tooltip, Typography, message, Modal 
} from 'antd';
import { useQuery } from '@tanstack/react-query';
import { SearchOutlined, ReloadOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { requestLogAPI } from '../../api/requestLog';
import type { RequestLog, RequestLogFilter } from '../../types/log';
import RequestLogDetail from './RequestLogDetail';

const { RangePicker } = DatePicker;
const { Option } = Select;
const { Text } = Typography;

const RequestLogsPage: React.FC = () => {
  const [form] = Form.useForm();
  const [filter, setFilter] = useState<RequestLogFilter>({
    page: 1,
    page_size: 20,
    sort_by: 'timestamp',
    sort_order: 'desc',
  });
  const [selectedLog, setSelectedLog] = useState<RequestLog | null>(null);
  const [detailVisible, setDetailVisible] = useState(false);

  // 查询日志列表
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['requestLogs', filter],
    queryFn: () => requestLogAPI.list(filter),
  });

  // 处理搜索
  const handleSearch = (values: any) => {
    const newFilter: RequestLogFilter = {
      ...filter,
      page: 1,
      project_id: values.project_id,
      environment_id: values.environment_id,
      protocol: values.protocol,
      method: values.method,
      path: values.path,
      status_code: values.status_code,
      source_ip: values.source_ip,
    };

    if (values.time_range) {
      newFilter.start_time = values.time_range[0].toISOString();
      newFilter.end_time = values.time_range[1].toISOString();
    }

    setFilter(newFilter);
  };

  // 重置搜索
  const handleReset = () => {
    form.resetFields();
    setFilter({
      page: 1,
      page_size: 20,
      sort_by: 'timestamp',
      sort_order: 'desc',
    });
  };

  // 查看详情
  const handleViewDetail = (log: RequestLog) => {
    setSelectedLog(log);
    setDetailVisible(true);
  };

  // 清理日志
  const handleCleanup = () => {
    Modal.confirm({
      title: '清理日志',
      content: '确定要清理7天前的日志吗？',
      onOk: async () => {
        try {
          const result = await requestLogAPI.cleanup(7);
          message.success(`成功清理 ${result.deleted_count} 条日志`);
          refetch();
        } catch (error) {
          message.error('清理失败');
        }
      },
    });
  };

  // 表格列定义
  const columns = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      render: (timestamp: string) => (
        <Text>{dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')}</Text>
      ),
    },
    {
      title: '协议',
      dataIndex: 'protocol',
      key: 'protocol',
      width: 100,
      render: (protocol: string) => (
        <Tag color={protocol === 'http' ? 'blue' : 'green'}>
          {protocol.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: '方法',
      dataIndex: 'method',
      key: 'method',
      width: 80,
      render: (method?: string) => method || '-',
    },
    {
      title: '路径',
      dataIndex: 'path',
      key: 'path',
      ellipsis: true,
      render: (path?: string) => (
        <Tooltip title={path}>
          <Text>{path || '-'}</Text>
        </Tooltip>
      ),
    },
    {
      title: '状态码',
      dataIndex: 'status_code',
      key: 'status_code',
      width: 100,
      render: (statusCode?: number) => {
        if (!statusCode) return '-';
        const color = statusCode >= 200 && statusCode < 300 ? 'success' : 
                     statusCode >= 400 ? 'error' : 'warning';
        return <Tag color={color}>{statusCode}</Tag>;
      },
    },
    {
      title: '耗时',
      dataIndex: 'duration',
      key: 'duration',
      width: 100,
      render: (duration: number) => (
        <Text>{duration}ms</Text>
      ),
    },
    {
      title: '来源IP',
      dataIndex: 'source_ip',
      key: 'source_ip',
      width: 140,
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      fixed: 'right' as const,
      render: (_: any, record: RequestLog) => (
        <Button 
          type="link" 
          icon={<EyeOutlined />}
          onClick={() => handleViewDetail(record)}
        >
          详情
        </Button>
      ),
    },
  ];

  // 分页配置
  const pagination = {
    current: filter.page || 1,
    pageSize: filter.page_size || 20,
    total: data?.total || 0,
    showSizeChanger: true,
    showQuickJumper: true,
    showTotal: (total: number) => `共 ${total} 条`,
    onChange: (page: number, pageSize: number) => {
      setFilter({ ...filter, page, page_size: pageSize });
    },
  };

  return (
    <div style={{ padding: '24px' }}>
      <Card title="请求日志" style={{ marginBottom: 16 }}>
        <Form
          form={form}
          layout="inline"
          onFinish={handleSearch}
          style={{ marginBottom: 16 }}
        >
          <Form.Item name="protocol" label="协议">
            <Select style={{ width: 120 }} allowClear placeholder="全部">
              <Option value="http">HTTP</Option>
              <Option value="websocket">WebSocket</Option>
            </Select>
          </Form.Item>
          <Form.Item name="method" label="方法">
            <Select style={{ width: 120 }} allowClear placeholder="全部">
              <Option value="GET">GET</Option>
              <Option value="POST">POST</Option>
              <Option value="PUT">PUT</Option>
              <Option value="DELETE">DELETE</Option>
              <Option value="PATCH">PATCH</Option>
            </Select>
          </Form.Item>
          <Form.Item name="path" label="路径">
            <Input placeholder="支持正则表达式" style={{ width: 200 }} />
          </Form.Item>
          <Form.Item name="status_code" label="状态码">
            <Input type="number" placeholder="如: 200" style={{ width: 120 }} />
          </Form.Item>
          <Form.Item name="source_ip" label="来源IP">
            <Input placeholder="IP地址" style={{ width: 150 }} />
          </Form.Item>
          <Form.Item name="time_range" label="时间范围">
            <RangePicker showTime />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>
                搜索
              </Button>
              <Button onClick={handleReset}>重置</Button>
              <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
                刷新
              </Button>
              <Button 
                danger 
                icon={<DeleteOutlined />} 
                onClick={handleCleanup}
              >
                清理旧日志
              </Button>
            </Space>
          </Form.Item>
        </Form>

        <Table
          columns={columns}
          dataSource={data?.data || []}
          loading={isLoading}
          pagination={pagination}
          rowKey="id"
          scroll={{ x: 1200 }}
        />
      </Card>

      <Modal
        title="请求日志详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={1000}
      >
        {selectedLog && <RequestLogDetail log={selectedLog} />}
      </Modal>
    </div>
  );
};

export default RequestLogsPage;
