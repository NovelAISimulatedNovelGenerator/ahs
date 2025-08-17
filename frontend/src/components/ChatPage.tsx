import { useState, useEffect, useRef } from 'react'
import { Layout, Input, Button, List, Typography, Avatar, Space, Spin, Card } from 'antd'
import { SendOutlined, UserOutlined, RobotOutlined } from '@ant-design/icons'
import { useChatStore } from '../stores/chat'
import { apiClient } from '../api/client'
import { useChatContext } from '../contexts/ChatContext'
import { MarkdownRenderer } from './MarkdownRenderer'
import type { ChatMessage } from '../types/api'

const { Content } = Layout
const { TextArea } = Input
const { Text, Paragraph } = Typography

/**
 * 聊天页面主组件
 * 功能：
 * 1. 显示当前会话的消息列表
 * 2. 处理用户输入和消息发送
 * 3. 与后端agent工作流进行流式通信
 * 4. 实时显示AI回复（支持流式传输）
 * 5. 自动保存对话到本地存储
 * 6. 集成RAG记忆系统（通过后端agent工作流）
 */

export function ChatPage() {
  // 从聊天状态管理器获取状态和操作函数
  const {
    getCurrentSession,    // 获取当前会话
    currentSessionId,     // 当前会话ID
    createSession,        // 创建新会话
    addMessage,          // 添加消息到会话
    updateMessage,       // 更新消息内容
    setMessageStreaming, // 设置消息流式状态
    setStreaming,        // 设置全局流式状态
    isStreaming,         // 当前是否正在流式传输
  } = useChatStore()

  // 从聊天设置Context获取配置
  const { settings } = useChatContext()

  // 本地状态：用户输入的文本
  const [inputValue, setInputValue] = useState('')
  // 本地状态：当前正在流式传输的消息ID
  const [currentStreamingId, setCurrentStreamingId] = useState<string | null>(null)
  // 引用：消息列表底部元素，用于自动滚动
  const messagesEndRef = useRef<HTMLDivElement>(null)
  // 引用：用于中止流式传输的控制器
  const abortControllerRef = useRef<AbortController | null>(null)
  
  // 获取当前激活的会话
  const currentSession = getCurrentSession()

  /**
   * 自动滚动到消息列表底部
   * 确保用户总是能看到最新的消息
   */
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  // 当消息列表更新时，自动滚动到底部
  useEffect(() => {
    scrollToBottom()
  }, [currentSession?.messages])

  // 如果没有当前会话，自动创建一个新会话
  useEffect(() => {
    if (!currentSessionId) {
      createSession()
    }
  }, [currentSessionId, createSession])

  /**
   * 处理用户发送消息
   * 整个流程：
   * 1. 验证输入和状态
   * 2. 添加用户消息到会话
   * 3. 创建AI助手消息占位符
   * 4. 调用后端agent工作流进行流式处理
   * 5. 实时更新AI回复内容
   * 6. 处理错误和完成状态
   * 
   * 注意：后端agent工作流会自动调用RAG记忆系统进行：
   * - 查询相关历史记忆
   * - 保存新的对话内容到记忆系统
   */
  const handleSendMessage = async () => {
    // 输入验证：确保有内容、没有正在进行的流式传输、有当前会话
    if (!inputValue.trim() || isStreaming || !currentSession) return

    const userInput = inputValue.trim()
    setInputValue('') // 清空输入框

    // 1. 添加用户消息到当前会话
    addMessage(currentSession.id, {
      role: 'user',
      content: userInput,
    })

    // 2. 添加AI助手消息占位符，初始内容为空，标记为流式传输中
    const assistantMessageId = addMessage(currentSession.id, {
      role: 'assistant',
      content: '',
      isStreaming: true,
    })

    // 3. 设置流式传输状态
    setCurrentStreamingId(assistantMessageId)  // 记录当前流式传输的消息ID
    setStreaming(true)                         // 设置全局流式传输状态
    setMessageStreaming(currentSession.id, assistantMessageId, true) // 设置消息级别的流式状态

    try {
      // 4. 根据设置选择执行模式
      if (settings.useStreaming) {
        // === 流式模式 ===
        // 创建中止控制器，用于支持用户主动停止生成
        abortControllerRef.current = new AbortController()

        let accumulatedContent = '' // 累积的AI回复内容

        // 调用后端agent工作流进行流式处理
        // 这里会触发后端的RAG记忆查询和保存操作
        for await (const event of apiClient.streamWorkflow(
          {
            workflow: 'agent',             // 使用agent工作流（集成了RAG记忆系统）
            input: userInput,              // 用户输入
            timeout: settings.timeout,    // 使用设置中的超时时间
          },
          abortControllerRef.current // 传入中止控制器
        )) {
          // 处理流式事件
          if (event.type === 'data' && typeof event.payload === 'string') {
            // 数据事件：累积内容并实时更新UI
            accumulatedContent += event.payload
            updateMessage(currentSession.id, assistantMessageId, accumulatedContent)
          } else if (event.type === 'error') {
            // 错误事件：显示错误信息并停止流式传输
            const errorMessage = typeof event.payload === 'object' && event.payload?.error
              ? event.payload.error
              : '发生未知错误'
            updateMessage(currentSession.id, assistantMessageId, `错误: ${errorMessage}`)
            break
          }
          // 注意：'done'事件会自动结束for await循环
        }
      } else {
        // === 同步模式 ===
        // 直接调用同步API，一次性获得完整回复
        const response = await apiClient.executeWorkflow({
          workflow: 'agent',             // 使用agent工作流（集成了RAG记忆系统）
          input: userInput,              // 用户输入
          timeout: settings.timeout,    // 使用设置中的超时时间
        })

        // 检查响应状态
        if (response.status === 'success' && response.result) {
          // 成功：更新助手消息内容
          updateMessage(currentSession.id, assistantMessageId, response.result)
        } else {
          // 失败：显示错误信息
          const errorMessage = response.error || '未知错误'
          updateMessage(currentSession.id, assistantMessageId, `错误: ${errorMessage}`)
        }
      }
    } catch (error) {
      // 7. 处理异常情况（网络错误、用户中止等）
      console.error('聊天错误:', error)
      updateMessage(
        currentSession.id,
        assistantMessageId,
        `错误: ${error instanceof Error ? error.message : '未知错误'}`
      )
    } finally {
      // 8. 清理状态
      setMessageStreaming(currentSession.id, assistantMessageId, false) // 取消消息流式状态
      setStreaming(false)                                               // 取消全局流式状态
      setCurrentStreamingId(null)                                       // 清空当前流式消息ID
      abortControllerRef.current = null                                 // 清空中止控制器
    }
  }

  /**
   * 处理停止流式传输
   * 用户点击停止按钮时调用，中止当前的AI回复生成
   */
  const handleStopStreaming = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort() // 发送中止信号
    }
  }

  /**
   * 处理输入框键盘事件
   * Enter键发送消息，Shift+Enter换行
   * @param e 键盘事件
   */
  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault() // 阻止默认的换行行为
      handleSendMessage() // 发送消息
    }
    // Shift+Enter 允许正常换行（不阻止默认行为）
  }

  // 加载状态：当没有当前会话时显示加载动画
  if (!currentSession) {
    return (
      <Content style={{ padding: '24px', display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
        <Spin size="large" />
      </Content>
    )
  }

  return (
    <Layout style={{ height: '100%', background: '#f5f5f5' }}>
      <Content style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
        {/* 消息列表区域 */}
        <div style={{ flex: 1, overflow: 'auto', padding: '16px 24px' }}>
          {currentSession.messages.length === 0 ? (
            // 空状态：没有消息时显示欢迎卡片
            <Card style={{ textAlign: 'center', margin: '20% auto', maxWidth: 400 }}>
              <RobotOutlined style={{ fontSize: 48, color: '#1890ff', marginBottom: 16 }} />
              <Typography.Title level={4}>开始对话</Typography.Title>
              <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
                我是你的AI助手，具有记忆功能。你可以问我任何问题，我会记住我们的对话内容。
              </Typography.Text>
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                当前模式：{settings.useStreaming ? '流式传输（实时回复）' : '同步模式（完整回复）'}
              </Typography.Text>
            </Card>
          ) : (
            // 消息列表：渲染所有历史消息
            <List
              dataSource={currentSession.messages}
              renderItem={(message) => (
                <MessageItem 
                  key={message.id} 
                  message={message}
                  // 只有当前正在流式传输的消息才显示流式状态
                  isStreaming={message.isStreaming && currentStreamingId === message.id}
                />
              )}
            />
          )}
          {/* 自动滚动锚点：隐藏的div，用于scrollIntoView */}
          <div ref={messagesEndRef} />
        </div>

        {/* 输入区域：固定在底部的消息输入和发送区域 */}
        <div style={{ padding: '16px 24px', background: '#fff', borderTop: '1px solid #e8e8e8' }}>
          <Space.Compact style={{ display: 'flex', width: '100%' }}>
            {/* 多行文本输入框 */}
            <TextArea
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyPress={handleKeyPress}
              placeholder="输入消息... (Enter发送，Shift+Enter换行)"
              autoSize={{ minRows: 1, maxRows: 4 }} // 自动调整高度，最多4行
              disabled={isStreaming} // 流式传输时禁用输入
              style={{ flex: 1 }}
            />
            {/* 发送/停止按钮：根据流式传输状态和模式切换 */}
            {isStreaming ? (
              // 传输中：显示停止按钮（仅在流式模式下可用）
              settings.useStreaming ? (
                <Button 
                  type="primary" 
                  danger
                  onClick={handleStopStreaming}
                  style={{ alignSelf: 'flex-end' }}
                >
                  停止
                </Button>
              ) : (
                // 同步模式下显示加载状态的发送按钮
                <Button
                  type="primary"
                  loading
                  style={{ alignSelf: 'flex-end' }}
                >
                  处理中
                </Button>
              )
            ) : (
              // 正常状态：显示发送按钮
              <Button
                type="primary"
                icon={<SendOutlined />}
                onClick={handleSendMessage}
                disabled={!inputValue.trim()} // 输入为空时禁用
                style={{ alignSelf: 'flex-end' }}
              >
                发送
              </Button>
            )}
          </Space.Compact>
        </div>
      </Content>
    </Layout>
  )
}

// 消息项组件的属性接口
interface MessageItemProps {
  message: ChatMessage    // 消息数据
  isStreaming?: boolean   // 是否正在流式传输
}

/**
 * 单个消息项组件
 * 功能：
 * 1. 渲染用户或AI助手的消息
 * 2. 根据消息角色调整布局和样式
 * 3. 支持流式传输时的打字机效果
 * 4. 显示消息时间戳
 * 5. 保持消息内容的格式（换行、空格等）
 */
function MessageItem({ message, isStreaming }: MessageItemProps) {
  const isUser = message.role === 'user' // 判断是否为用户消息
  
  return (
    <List.Item style={{ border: 'none', padding: '12px 0' }}>
      <div
        style={{
          display: 'flex',
          width: '100%',
          // 用户消息右对齐，AI消息左对齐
          justifyContent: isUser ? 'flex-end' : 'flex-start',
          alignItems: 'flex-start',
          gap: 12,
        }}
      >
        {/* AI助手头像：只在AI消息时显示，位于左侧 */}
        {!isUser && (
          <Avatar 
            icon={<RobotOutlined />} 
            style={{ backgroundColor: '#1890ff', flexShrink: 0 }}
          />
        )}
        
        {/* 消息气泡 */}
        <div
          style={{
            maxWidth: '70%', // 限制最大宽度，避免消息过长
            padding: '12px 16px',
            borderRadius: 12,
            // 用户消息使用蓝色背景，AI消息使用白色背景
            backgroundColor: isUser ? '#1890ff' : '#fff',
            color: isUser ? '#fff' : '#333',
            boxShadow: '0 1px 2px rgba(0,0,0,0.1)', // 轻微阴影
            position: 'relative',
          }}
        >
          {/* 消息内容：根据角色选择渲染方式 */}
          {isUser ? (
            // 用户消息：纯文本显示
            <Paragraph
              style={{ 
                margin: 0, 
                color: 'inherit',
                whiteSpace: 'pre-wrap',  // 保持换行和空格
                wordBreak: 'break-word', // 长单词自动换行
              }}
            >
              {message.content}
            </Paragraph>
          ) : (
            // AI消息：Markdown渲染
            <MarkdownRenderer 
              content={message.content}
              isStreaming={isStreaming}
            />
          )}
          
          {/* 消息时间戳 */}
          <Text 
            style={{ 
              fontSize: 12, 
              opacity: 0.7, 
              color: 'inherit',
              display: 'block',
              marginTop: 4,
            }}
          >
            {new Date(message.timestamp).toLocaleTimeString()}
          </Text>
        </div>

        {/* 用户头像：只在用户消息时显示，位于右侧 */}
        {isUser && (
          <Avatar 
            icon={<UserOutlined />} 
            style={{ backgroundColor: '#52c41a', flexShrink: 0 }}
          />
        )}
      </div>
    </List.Item>
  )
}