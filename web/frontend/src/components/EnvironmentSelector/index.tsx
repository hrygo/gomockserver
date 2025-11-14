import React from 'react'
import { Select } from 'antd'
import { GlobalOutlined } from '@ant-design/icons'
import type { Environment } from '@/types/environment'

interface EnvironmentSelectorProps {
  environments: Environment[]
  currentEnvironmentId?: string
  onEnvironmentChange: (environmentId: string) => void
  loading?: boolean
  disabled?: boolean
  style?: React.CSSProperties
}

const EnvironmentSelector: React.FC<EnvironmentSelectorProps> = ({
  environments,
  currentEnvironmentId,
  onEnvironmentChange,
  loading = false,
  disabled = false,
  style,
}) => {
  return (
    <Select
      style={{ minWidth: 200, ...style }}
      placeholder="选择环境"
      value={currentEnvironmentId}
      onChange={onEnvironmentChange}
      loading={loading}
      disabled={disabled || environments.length === 0}
      suffixIcon={<GlobalOutlined />}
      options={environments.map((env) => ({
        label: (
          <div>
            <div style={{ fontWeight: 500 }}>{env.name}</div>
            <div style={{ fontSize: 12, color: '#999' }}>{env.base_url}</div>
          </div>
        ),
        value: env.id,
      }))}
      optionLabelProp="label"
      notFoundContent="暂无环境"
    />
  )
}

export default EnvironmentSelector
