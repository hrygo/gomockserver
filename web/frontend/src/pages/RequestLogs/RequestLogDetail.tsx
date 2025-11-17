import React from 'react';
import { Descriptions, Card, Typography, Tag, Tabs } from 'antd';
import { Light as SyntaxHighlighter } from 'react-syntax-highlighter';
import json from 'react-syntax-highlighter/dist/esm/languages/hljs/json';
import { docco } from 'react-syntax-highlighter/dist/esm/styles/hljs';
import dayjs from 'dayjs';
import type { RequestLog } from '../../types/log';

SyntaxHighlighter.registerLanguage('json', json);

const { Text } = Typography;
const { TabPane } = Tabs;

interface RequestLogDetailProps {
  log: RequestLog;
}

const RequestLogDetail: React.FC<RequestLogDetailProps> = ({ log }) => {
  const formatJSON = (obj: any) => {
    try {
      return JSON.stringify(obj, null, 2);
    } catch {
      return String(obj);
    }
  };

  return (
    <div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Descriptions column={2} size="small">
          <Descriptions.Item label="请求ID">
            <Text copyable>{log.request_id}</Text>
          </Descriptions.Item>
          <Descriptions.Item label="时间">
            {dayjs(log.timestamp).format('YYYY-MM-DD HH:mm:ss.SSS')}
          </Descriptions.Item>
          <Descriptions.Item label="协议">
            <Tag color={log.protocol === 'http' ? 'blue' : 'green'}>
              {log.protocol.toUpperCase()}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="方法">
            {log.method || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="路径" span={2}>
            <Text copyable>{log.path || '-'}</Text>
          </Descriptions.Item>
          <Descriptions.Item label="状态码">
            {log.status_code ? (
              <Tag color={
                log.status_code >= 200 && log.status_code < 300 ? 'success' :
                log.status_code >= 400 ? 'error' : 'warning'
              }>
                {log.status_code}
              </Tag>
            ) : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="耗时">
            <Text strong>{log.duration}ms</Text>
          </Descriptions.Item>
          <Descriptions.Item label="来源IP">
            <Text copyable>{log.source_ip}</Text>
          </Descriptions.Item>
          <Descriptions.Item label="项目ID">
            <Text copyable>{log.project_id}</Text>
          </Descriptions.Item>
          <Descriptions.Item label="环境ID">
            <Text copyable>{log.environment_id}</Text>
          </Descriptions.Item>
          {log.rule_id && (
            <Descriptions.Item label="规则ID">
              <Text copyable>{log.rule_id}</Text>
            </Descriptions.Item>
          )}
        </Descriptions>
      </Card>

      <Tabs defaultActiveKey="request">
        <TabPane tab="请求详情" key="request">
          <Card size="small">
            <SyntaxHighlighter 
              language="json" 
              style={docco}
              customStyle={{ margin: 0, maxHeight: 400, overflow: 'auto' }}
            >
              {formatJSON(log.request)}
            </SyntaxHighlighter>
          </Card>
        </TabPane>
        <TabPane tab="响应详情" key="response">
          <Card size="small">
            <SyntaxHighlighter 
              language="json" 
              style={docco}
              customStyle={{ margin: 0, maxHeight: 400, overflow: 'auto' }}
            >
              {formatJSON(log.response)}
            </SyntaxHighlighter>
          </Card>
        </TabPane>
      </Tabs>
    </div>
  );
};

export default RequestLogDetail;
