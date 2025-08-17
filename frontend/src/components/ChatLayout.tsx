import { useState } from 'react'
import { Layout, Typography, Button, Space, Switch, Tooltip } from 'antd'
import { ArrowLeftOutlined, SettingOutlined, ThunderboltOutlined, ClockCircleOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { ChatSidebar } from './ChatSidebar'
import { ChatPage } from './ChatPage'
import { ChatProvider, useChatContext } from '../contexts/ChatContext'

const { Header } = Layout
const { Title } = Typography

/**
 * 聊天应用布局组件
 * 功能：
 * 1. 组合聊天侧边栏和聊天页面
 * 2. 管理侧边栏的折叠/展开状态
 * 3. 提供完整的聊天应用界面
 * 4. 响应式布局支持
 * 
 * 布局结构：
 * ┌─────────────┬─────────────────────┐
 * │             │                     │
 * │   会话列表   │      聊天页面        │
 * │   侧边栏    │     (消息+输入)      │
 * │             │                     │
 * └─────────────┴─────────────────────┘
 */
/**
 * 聊天布局内部组件
 * 包含实际的布局逻辑，需要在ChatProvider内部使用
 */
function ChatLayoutInner() {
  // 侧边栏折叠状态：false=展开，true=折叠
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const navigate = useNavigate()
  const { settings, toggleStreaming } = useChatContext()

  /**
   * 返回主应用
   */
  const handleBackToMain = () => {
    navigate('/workflows')
  }

  /**
   * 打开设置页面
   */
  const handleOpenSettings = () => {
    navigate('/settings')
  }

  return (
    <Layout className="chat-layout" style={{ height: '100vh' }}>
      {/* 顶部导航栏：提供返回主应用的方式和聊天设置 */}
      <Header style={{ background: '#fff', padding: '0 24px', borderBottom: '1px solid #e8e8e8' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', height: '100%' }}>
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <Button 
              type="text" 
              icon={<ArrowLeftOutlined />}
              onClick={handleBackToMain}
              style={{ marginRight: 16 }}
            >
              返回主应用
            </Button>
            <Title level={4} style={{ margin: 0 }}>
              AI智能对话助手
            </Title>
          </div>
          
          <Space>
            {/* 流式模式切换开关 */}
            <Tooltip 
              title={settings.useStreaming ? "切换到同步模式（一次性返回完整回复）" : "切换到流式模式（实时显示回复过程）"}
            >
              <Space size={8}>
                {settings.useStreaming ? (
                  <ThunderboltOutlined style={{ color: '#1890ff' }} />
                ) : (
                  <ClockCircleOutlined style={{ color: '#666' }} />
                )}
                <Switch 
                  checked={settings.useStreaming}
                  onChange={toggleStreaming}
                  size="small"
                />
                <span style={{ fontSize: 12, color: '#666' }}>
                  {settings.useStreaming ? '流式' : '同步'}
                </span>
              </Space>
            </Tooltip>
            
            <Button 
              type="text" 
              icon={<SettingOutlined />}
              onClick={handleOpenSettings}
            >
              设置
            </Button>
          </Space>
        </div>
      </Header>

      {/* 聊天主体区域 */}
      <Layout style={{ height: 'calc(100vh - 64px)' }}>
        {/* 聊天会话侧边栏 */}
        <ChatSidebar collapsed={sidebarCollapsed} />
        
        {/* 主聊天区域 */}
        <Layout style={{ marginLeft: 0 }}>
          <ChatPage />
        </Layout>
      </Layout>
    </Layout>
  )
}

/**
 * 聊天布局主组件
 * 使用ChatProvider包装，为子组件提供聊天设置Context
 */
export function ChatLayout() {
  return (
    <ChatProvider>
      <ChatLayoutInner />
    </ChatProvider>
  )
}