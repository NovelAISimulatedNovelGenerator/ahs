import { useState } from 'react'
import { 
  Layout, 
  List, 
  Button, 
  Typography, 
  Popconfirm, 
  Input, 
  Space,
  Tooltip
} from 'antd'
import { 
  PlusOutlined, 
  DeleteOutlined, 
  EditOutlined, 
  MessageOutlined,
  ClearOutlined
} from '@ant-design/icons'
import { useChatStore } from '../stores/chat'
import type { ChatSession } from '../types/api'

const { Sider } = Layout
const { Text } = Typography

interface ChatSidebarProps {
  collapsed: boolean // 侧边栏是否折叠
}

/**
 * 聊天会话侧边栏组件
 * 功能：
 * 1. 显示所有聊天会话列表
 * 2. 创建新会话
 * 3. 切换当前会话
 * 4. 编辑会话标题
 * 5. 删除单个会话
 * 6. 清空所有会话
 * 7. 支持折叠/展开模式
 */
export function ChatSidebar({ collapsed }: ChatSidebarProps) {
  // 从聊天状态管理器获取状态和操作函数
  const {
    sessions,           // 所有会话列表
    currentSessionId,   // 当前激活的会话ID
    createSession,      // 创建新会话
    deleteSession,      // 删除会话
    setCurrentSession,  // 设置当前会话
    updateSessionTitle, // 更新会话标题
    clearAllSessions,   // 清空所有会话
  } = useChatStore()

  // 本地状态：正在编辑标题的会话ID
  const [editingSessionId, setEditingSessionId] = useState<string | null>(null)
  // 本地状态：编辑中的标题文本
  const [editingTitle, setEditingTitle] = useState('')

  /**
   * 创建新的聊天会话
   * 会自动设置为当前激活会话
   */
  const handleNewSession = () => {
    createSession()
  }

  /**
   * 点击会话项，切换到该会话
   * @param sessionId 会话ID
   */
  const handleSessionClick = (sessionId: string) => {
    setCurrentSession(sessionId)
  }

  /**
   * 删除指定会话
   * @param sessionId 要删除的会话ID
   */
  const handleDeleteSession = (sessionId: string) => {
    deleteSession(sessionId)
  }

  /**
   * 开始编辑会话标题
   * 设置编辑状态并预填充当前标题
   * @param session 要编辑的会话对象
   */
  const handleEditTitle = (session: ChatSession) => {
    setEditingSessionId(session.id)
    setEditingTitle(session.title)
  }

  /**
   * 保存编辑的会话标题
   * 只有在标题非空时才保存
   */
  const handleSaveTitle = () => {
    if (editingSessionId && editingTitle.trim()) {
      updateSessionTitle(editingSessionId, editingTitle.trim())
    }
    // 重置编辑状态
    setEditingSessionId(null)
    setEditingTitle('')
  }

  /**
   * 取消编辑会话标题
   * 重置编辑状态，不保存更改
   */
  const handleCancelEdit = () => {
    setEditingSessionId(null)
    setEditingTitle('')
  }

  /**
   * 清空所有会话
   * 删除所有会话记录和当前会话状态
   */
  const handleClearAllSessions = () => {
    clearAllSessions()
  }

  /**
   * 格式化时间显示
   * 根据时间差显示相对时间（刚刚、X分钟前、X小时前、X天前）
   * @param timestamp 时间戳
   * @returns 格式化后的时间字符串
   */
  const formatTime = (timestamp: number) => {
    const now = Date.now()
    const diff = now - timestamp
    const minutes = Math.floor(diff / (1000 * 60))
    const hours = Math.floor(diff / (1000 * 60 * 60))
    const days = Math.floor(diff / (1000 * 60 * 60 * 24))

    if (days > 0) return `${days}天前`
    if (hours > 0) return `${hours}小时前`
    if (minutes > 0) return `${minutes}分钟前`
    return '刚刚'
  }

  return (
    <Sider 
      width={280} 
      collapsed={collapsed}
      theme="light" 
      style={{ 
        borderRight: '1px solid #e8e8e8',
        height: '100%',
        overflow: 'hidden',
        display: 'flex',
        flexDirection: 'column'
      }}
    >
      {/* 侧边栏头部：新建对话和清空按钮 */}
      <div style={{ padding: collapsed ? '12px 8px' : '16px', borderBottom: '1px solid #e8e8e8' }}>
        {collapsed ? (
          // 折叠模式：只显示新建按钮图标，使用Tooltip显示说明
          <Tooltip title="新建对话" placement="right">
            <Button 
              type="primary" 
              icon={<PlusOutlined />}
              onClick={handleNewSession}
              style={{ width: '100%' }}
            />
          </Tooltip>
        ) : (
          // 展开模式：显示完整按钮和清空所有对话按钮
          <Space direction="vertical" style={{ width: '100%' }}>
            <Button 
              type="primary" 
              icon={<PlusOutlined />}
              onClick={handleNewSession}
              style={{ width: '100%' }}
            >
              新建对话
            </Button>
            {/* 只有存在会话时才显示清空按钮 */}
            {sessions.length > 0 && (
              <Popconfirm
                title="确认清空所有对话?"
                description="这将删除所有对话记录，此操作无法撤销。"
                onConfirm={handleClearAllSessions}
                okText="确认"
                cancelText="取消"
              >
                <Button 
                  icon={<ClearOutlined />}
                  size="small"
                  style={{ width: '100%' }}
                  danger
                >
                  清空所有对话
                </Button>
              </Popconfirm>
            )}
          </Space>
        )}
      </div>

      {/* 会话列表区域 */}
      <div style={{ flex: 1, overflow: 'auto' }}>
        {sessions.length === 0 ? (
          // 空状态：没有会话时的提示（仅在展开模式显示）
          !collapsed && (
            <div style={{ padding: '24px 16px', textAlign: 'center' }}>
              <MessageOutlined style={{ fontSize: 24, color: '#bfbfbf', marginBottom: 8 }} />
              <Text type="secondary" style={{ display: 'block' }}>
                暂无对话
              </Text>
            </div>
          )
        ) : (
          // 会话列表：使用Ant Design的List组件渲染
          <List
            dataSource={sessions}
            renderItem={(session) => (
              <List.Item
                key={session.id}
                style={{
                  padding: collapsed ? '8px' : '12px 16px',
                  cursor: 'pointer',
                  // 当前激活会话的高亮样式
                  backgroundColor: currentSessionId === session.id ? '#e6f7ff' : 'transparent',
                  borderLeft: currentSessionId === session.id ? '3px solid #1890ff' : '3px solid transparent',
                }}
                onClick={() => handleSessionClick(session.id)}
              >
                {collapsed ? (
                  // 折叠模式：只显示消息图标，使用Tooltip显示会话标题
                  <Tooltip title={session.title} placement="right">
                    <MessageOutlined style={{ fontSize: 16, color: '#1890ff' }} />
                  </Tooltip>
                ) : (
                  // 展开模式：显示完整的会话信息
                  <div style={{ width: '100%', minHeight: '40px' }}>
                    {editingSessionId === session.id ? (
                      // 编辑模式：显示输入框
                      <Input
                        value={editingTitle}
                        onChange={(e) => setEditingTitle(e.target.value)}
                        onPressEnter={handleSaveTitle}    // Enter键保存
                        onBlur={handleSaveTitle}          // 失去焦点时保存
                        onKeyDown={(e) => {
                          if (e.key === 'Escape') {       // Escape键取消编辑
                            handleCancelEdit()
                          }
                        }}
                        autoFocus
                        style={{ marginBottom: 4 }}
                      />
                    ) : (
                      // 显示模式：显示会话信息和操作按钮
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                        {/* 会话基本信息 */}
                        <div style={{ flex: 1, minWidth: 0 }}>
                          {/* 会话标题 */}
                          <Text 
                            style={{ 
                              fontWeight: currentSessionId === session.id ? 600 : 400,
                              fontSize: '14px',
                              display: 'block',
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                              whiteSpace: 'nowrap',
                            }}
                          >
                            {session.title}
                          </Text>
                          {/* 会话元信息：更新时间和消息数量 */}
                          <Text 
                            type="secondary" 
                            style={{ fontSize: '12px' }}
                          >
                            {formatTime(session.updatedAt)} · {session.messages.length}条消息
                          </Text>
                        </div>
                        
                        {/* 操作按钮：编辑和删除 */}
                        <Space size={4} style={{ opacity: 0.6, marginLeft: 8 }}>
                          {/* 编辑标题按钮 */}
                          <Button
                            type="text"
                            size="small"
                            icon={<EditOutlined />}
                            onClick={(e) => {
                              e.stopPropagation() // 阻止事件冒泡，避免触发会话切换
                              handleEditTitle(session)
                            }}
                            style={{ padding: '0 4px', height: 20 }}
                          />
                          {/* 删除会话按钮 */}
                          <Popconfirm
                            title="确认删除此对话?"
                            onConfirm={(e) => {
                              e?.stopPropagation() // 阻止事件冒泡
                              handleDeleteSession(session.id)
                            }}
                            okText="确认"
                            cancelText="取消"
                          >
                            <Button
                              type="text"
                              size="small"
                              icon={<DeleteOutlined />}
                              onClick={(e) => e.stopPropagation()} // 阻止事件冒泡
                              style={{ padding: '0 4px', height: 20 }}
                            />
                          </Popconfirm>
                        </Space>
                      </div>
                    )}
                  </div>
                )}
              </List.Item>
            )}
          />
        )}
      </div>
    </Sider>
  )
}