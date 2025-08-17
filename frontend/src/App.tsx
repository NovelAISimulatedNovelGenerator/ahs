import { BrowserRouter, Routes, Route, Navigate, useNavigate, useLocation } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider, Layout, Menu, Typography } from 'antd'
import { useState, useEffect } from 'react'
import { SettingsPage } from './components/SettingsPage'
import { WorkflowList } from './components/WorkflowList'
import { WorkflowDetail } from './components/WorkflowDetail'
import { WorkflowExecute } from './components/WorkflowExecute'
import { WorkflowStream } from './components/WorkflowStream'
import { DebugLogs } from './components/DebugLogs'
import { ChatLayout } from './components/ChatLayout'
import { useSettingsStore } from './stores/settings'

const { Header, Content, Sider } = Layout
const { Title } = Typography

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 2,
      staleTime: 5 * 60 * 1000, // 5分钟
    },
  },
})

function AppLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const [selectedKeys, setSelectedKeys] = useState<string[]>([])
  const navigate = useNavigate()
  const location = useLocation()
  const { initialize } = useSettingsStore()

  // 检查是否为聊天页面
  const isChatPage = location.pathname === '/chat'

  // 初始化设置
  useEffect(() => {
    initialize()
  }, [initialize])

  // 更新选中的菜单项
  useEffect(() => {
    setSelectedKeys([location.pathname])
  }, [location.pathname])

  // 如果是聊天页面，直接返回聊天布局，不显示主应用导航
  if (isChatPage) {
    return <ChatLayout />
  }

  const menuItems = [
    {
      key: 'chat',
      label: '聊天',
      children: [
        { key: '/chat', label: 'AI对话' },
      ],
    },
    {
      key: 'workflows',
      label: '工作流',
      children: [
        { key: '/workflows', label: '工作流列表' },
        { key: '/execute', label: '执行工作流' },
        { key: '/stream', label: '流式执行' },
      ],
    },
    {
      key: 'debug',
      label: '调试',
      children: [{ key: '/debug', label: '调试日志' }],
    },
    {
      key: 'settings',
      label: '设置',
      children: [{ key: '/settings', label: '系统设置' }],
    },
  ]

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed} theme="light" width={250}>
        <div style={{ padding: '16px', textAlign: 'center' }}>
          <Title level={4} style={{ margin: 0, fontSize: collapsed ? '14px' : '16px' }}>
            {collapsed ? 'AHS' : 'Agent HTTP Server'}
          </Title>
        </div>

        <Menu
          mode="inline"
          selectedKeys={selectedKeys}
          defaultOpenKeys={['chat', 'workflows', 'debug', 'settings']}
          items={menuItems}
          onClick={({ key }) => {
            navigate(key)
          }}
        />
      </Sider>

      <Layout>
        <Header style={{ background: '#fff', padding: '0 24px' }}>
          <Title level={3} style={{ margin: '16px 0' }}>
            Agent HTTP Server - 内部工作台
          </Title>
        </Header>

        <Content style={{ background: '#f0f2f5' }}>
          <Routes>
            <Route path="/" element={<Navigate to="/chat" replace />} />
            <Route path="/workflows" element={<WorkflowList />} />
            <Route path="/workflows/:name" element={<WorkflowDetail />} />
            <Route path="/execute" element={<WorkflowExecute />} />
            <Route path="/stream" element={<WorkflowStream />} />
            <Route path="/debug" element={<DebugLogs />} />
            <Route path="/settings" element={<SettingsPage />} />
          </Routes>
        </Content>
      </Layout>
    </Layout>
  )
}

function App() {
  return (
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: '#1890ff',
        },
      }}
    >
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <AppLayout />
        </BrowserRouter>
      </QueryClientProvider>
    </ConfigProvider>
  )
}

export default App
