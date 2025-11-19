import React, { useRef, useEffect } from 'react'
import { Card, Row, Col, Statistic, Table, Spin, Empty } from 'antd'
import {
  ProjectOutlined,
  ApiOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  EnvironmentOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import ReactECharts from 'echarts-for-react'
import type { EChartsOption } from 'echarts'
import dayjs from 'dayjs'
import {
  useDashboardStatistics,
  useProjectStatistics,
  useRuleStatistics,
  useRequestTrend,
  useResponseTimeDistribution,
} from '@/hooks/useStatistics'
import type { ProjectStatistics, RuleStatistics } from '@/types/statistics'

const Dashboard: React.FC = () => {
  // ECharts 实例引用
  const trendChartRef = useRef<any>(null)
  const pieChartRef = useRef<any>(null)

  // 获取统计数据
  const { data: dashboardStats, isLoading: dashboardLoading } = useDashboardStatistics()
  const { data: projectStats } = useProjectStatistics()
  const { data: ruleStats } = useRuleStatistics()
  const { data: requestTrend } = useRequestTrend()
  const { data: responseTimeDist } = useResponseTimeDistribution()

  // 确保数据为数组，防止 null 或 undefined
  const safeProjectStats = projectStats || []
  const safeRuleStats = ruleStats || []
  const safeRequestTrend = requestTrend || []
  const safeResponseTimeDist = responseTimeDist || []

  // 组件卸载时清理 ECharts 实例
  useEffect(() => {
    return () => {
      try {
        if (trendChartRef.current) {
          const instance = trendChartRef.current.getEchartsInstance()
          if (instance && !instance.isDisposed()) {
            instance.dispose()
          }
        }
        if (pieChartRef.current) {
          const instance = pieChartRef.current.getEchartsInstance()
          if (instance && !instance.isDisposed()) {
            instance.dispose()
          }
        }
      } catch (error) {
        // 忽略清理错误
        console.debug('ECharts cleanup error:', error)
      }
    }
  }, [])

  // 请求趋势图表配置
  const requestTrendOption: EChartsOption = {
    title: {
      text: '请求趋势（最近7天）',
      left: 'center',
      textStyle: { fontSize: 14, fontWeight: 'normal' },
    },
    tooltip: {
      trigger: 'axis',
      formatter: '{b}<br/>请求数: {c}',
    },
    xAxis: {
      type: 'category',
      data: safeRequestTrend.map((item) => dayjs(item.date).format('MM-DD')),
      axisLabel: { fontSize: 12 },
    },
    yAxis: {
      type: 'value',
      axisLabel: { fontSize: 12 },
    },
    series: [
      {
        data: safeRequestTrend.map((item) => item.count),
        type: 'line',
        smooth: true,
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(24, 144, 255, 0.3)' },
              { offset: 1, color: 'rgba(24, 144, 255, 0.05)' },
            ],
          },
        },
        itemStyle: { color: '#1890ff' },
      },
    ],
    grid: { left: 50, right: 20, top: 50, bottom: 30 },
  }

  // 响应时间分布图表配置
  const responseTimeOption: EChartsOption = {
    title: {
      text: '响应时间分布',
      left: 'center',
      textStyle: { fontSize: 14, fontWeight: 'normal' },
    },
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)',
    },
    legend: {
      orient: 'vertical',
      right: 10,
      top: 'center',
      textStyle: { fontSize: 12 },
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 8,
          borderColor: '#fff',
          borderWidth: 2,
        },
        label: {
          show: false,
          position: 'center',
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 16,
            fontWeight: 'bold',
          },
        },
        labelLine: {
          show: false,
        },
        data: safeResponseTimeDist.map((item) => ({
          value: item.count,
          name: item.range,
        })),
      },
    ],
  }

  // 项目统计表格列
  const projectColumns: ColumnsType<ProjectStatistics> = [
    {
      title: '项目名称',
      dataIndex: 'project_name',
      key: 'project_name',
      width: '30%',
    },
    {
      title: '环境数',
      dataIndex: 'environment_count',
      key: 'environment_count',
      width: '20%',
      sorter: (a, b) => a.environment_count - b.environment_count,
    },
    {
      title: '规则数',
      dataIndex: 'rule_count',
      key: 'rule_count',
      width: '20%',
      sorter: (a, b) => a.rule_count - b.rule_count,
    },
    {
      title: '请求数',
      dataIndex: 'request_count',
      key: 'request_count',
      width: '30%',
      sorter: (a, b) => a.request_count - b.request_count,
      render: (count: number) => (
        <span style={{ fontWeight: 500, color: '#1890ff' }}>{count.toLocaleString()}</span>
      ),
    },
  ]

  // 规则统计表格列
  const ruleColumns: ColumnsType<RuleStatistics> = [
    {
      title: '规则名称',
      dataIndex: 'rule_name',
      key: 'rule_name',
      width: '35%',
      ellipsis: true,
    },
    {
      title: '匹配次数',
      dataIndex: 'match_count',
      key: 'match_count',
      width: '20%',
      sorter: (a, b) => a.match_count - b.match_count,
      defaultSortOrder: 'descend',
    },
    {
      title: '平均响应时间',
      dataIndex: 'avg_response_time',
      key: 'avg_response_time',
      width: '25%',
      render: (time: number) => `${time.toFixed(2)}ms`,
    },
    {
      title: '最后匹配',
      dataIndex: 'last_matched_at',
      key: 'last_matched_at',
      width: '20%',
      render: (time?: string) =>
        time ? dayjs(time).format('MM-DD HH:mm') : '-',
    },
  ]

  if (dashboardLoading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px 0' }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div>
      <h1 style={{ marginBottom: 24 }}>仪表盘</h1>

      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="项目总数"
              value={dashboardStats?.total_projects || 0}
              prefix={<ProjectOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="环境总数"
              value={dashboardStats?.total_environments || 0}
              prefix={<EnvironmentOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="规则总数"
              value={dashboardStats?.total_rules || 0}
              prefix={<ApiOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="总请求数"
              value={dashboardStats?.total_requests || 0}
              prefix={<ThunderboltOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={8}>
          <Card>
            <Statistic
              title="启用规则"
              value={dashboardStats?.enabled_rules || 0}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="禁用规则"
              value={dashboardStats?.disabled_rules || 0}
              prefix={<CloseCircleOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="今日请求"
              value={dashboardStats?.requests_today || 0}
              prefix={<ThunderboltOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 图表区域 */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={16}>
          <Card styles={{ body: { padding: '20px' } }}>
            {safeRequestTrend.length > 0 ? (
              <ReactECharts
                ref={trendChartRef}
                option={requestTrendOption}
                style={{ height: 300 }}
                opts={{ renderer: 'svg' }}
              />
            ) : (
              <Empty description="暂无请求数据" style={{ padding: '80px 0' }} />
            )}
          </Card>
        </Col>
        <Col span={8}>
          <Card styles={{ body: { padding: '20px' } }}>
            {safeResponseTimeDist.length > 0 ? (
              <ReactECharts
                ref={pieChartRef}
                option={responseTimeOption}
                style={{ height: 300 }}
                opts={{ renderer: 'svg' }}
              />
            ) : (
              <Empty description="暂无响应数据" style={{ padding: '80px 0' }} />
            )}
          </Card>
        </Col>
      </Row>

      {/* 项目统计 */}
      <Card title="项目统计" style={{ marginBottom: 16 }}>
        <Table
          columns={projectColumns}
          dataSource={safeProjectStats}
          rowKey="project_id"
          pagination={false}
          size="small"
        />
      </Card>

      {/* 热门规则 (Top 10) */}
      <Card title="热门规则 (Top 10)">
        <Table
          columns={ruleColumns}
          dataSource={safeRuleStats.slice(0, 10)}
          rowKey="rule_id"
          pagination={false}
          size="small"
        />
      </Card>
    </div>
  )
}

export default Dashboard
