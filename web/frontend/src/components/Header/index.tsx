import React from 'react'
import { Layout, Breadcrumb, Space, Avatar, Dropdown } from 'antd'
import { UserOutlined, SettingOutlined, LogoutOutlined } from '@ant-design/icons'
import type { MenuProps } from 'antd'

const { Header: AntHeader } = Layout

const Header: React.FC = () => {
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出',
      danger: true,
    },
  ]

  return (
    <AntHeader style={{ background: '#fff', padding: '0 24px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
      <Breadcrumb
        items={[
          { title: '首页' },
          { title: '仪表盘' },
        ]}
      />
      
      <Space size="middle">
        <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
          <Avatar icon={<UserOutlined />} style={{ cursor: 'pointer' }} />
        </Dropdown>
      </Space>
    </AntHeader>
  )
}

export default Header
