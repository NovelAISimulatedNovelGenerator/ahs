import { useState, useEffect, useRef } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import {
  Card,
  Form,
  Input,
  Button,
  Typography,
  Space,
  Select,
  notification,
  Divider,
  Progress,
  Tag,
} from 'antd'
import {
  ArrowLeftOutlined,
  PlayCircleOutlined,
  StopOutlined,
  PauseCircleOutlined,
  ClearOutlined,
} from '@ant-design/icons'
import { useQuery } from '@tanstack/react-query'
import { apiClient } from '../api/client'
import { useSettingsStore } from '../stores/settings'
import type { WorkflowRequest } from '../types/api'

const { Title, Text } = Typography
const { TextArea } = Input

interface StreamMessage {
  id: string
  timestamp: number
  type: 'data' | 'done' | 'error' | 'start'
  content: string
  raw?: unknown
}

export function WorkflowStream() {
  const location = useLocation()
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [isStreaming, setIsStreaming] = useState(false)
  const [messages, setMessages] = useState<StreamMessage[]>([])
  const [progress, setProgress] = useState(0)
  const abortControllerRef = useRef<AbortController | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const { addDebugLog } = useSettingsStore()

  // 从URL参数获取预选工作流
  const searchParams = new URLSearchParams(location.search)
  const preSelectedWorkflow = searchParams.get('workflow')

  // 获取工作流列表
  const { data: workflowsData, isLoading: workflowsLoading } = useQuery({
    queryKey: ['workflows'],
    queryFn: () => apiClient.getWorkflows(),
  })

  // 初始化表单
  useEffect(() => {
    if (preSelectedWorkflow) {
      form.setFieldValue('workflow', preSelectedWorkflow)
    }
  }, [preSelectedWorkflow, form])

  // 自动滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const addMessage = (type: StreamMessage['type'], content: string, raw?: unknown) => {
    const message: StreamMessage = {
      id: Math.random().toString(36).substr(2, 9),
      timestamp: Date.now(),
      type,
      content,
      raw,
    }
    setMessages(prev => [...prev, message])
    return message
  }

  const handleSubmit = async (values: { workflow: string; input: string; timeout?: number }) => {
    const request: WorkflowRequest = {
      workflow: values.workflow,
      input: values.input,
      timeout: values.timeout || 300, // 流式执行默认5分钟超时
    }

    // 清空之前的消息
    setMessages([])
    setProgress(0)
    setIsStreaming(true)

    // 创建新的AbortController
    abortControllerRef.current = new AbortController()

    addMessage('start', `开始执行工作流: ${request.workflow}`)

    try {
      // 使用异步生成器处理SSE流
      const stream = apiClient.streamWorkflow(request, abortControllerRef.current)

      let messageCount = 0
      for await (const event of stream) {
        messageCount++

        // 记录调试日志
        addDebugLog({
          id: '',
          timestamp: Date.now(),
          direction: 'event',
          path: '/api/stream',
          eventType: event.type,
          summary: `SSE事件: ${event.type}`,
          details: event.payload,
        })

        // 处理不同类型的事件
        if (event.type === 'error') {
          const errorPayload =
            typeof event.payload === 'object'
              ? (event.payload as { error: string }).error
              : (event.payload as string)

          addMessage('error', `错误: ${errorPayload}`, event.payload)
          notification.error({
            message: '执行出错',
            description: errorPayload,
          })
          break
        } else if (event.type === 'done') {
          addMessage('done', '执行完成', event.payload)
          setProgress(100)
          notification.success({
            message: '执行完成',
            description: '工作流已成功执行完成',
          })
          break
        } else {
          // 数据事件
          const content =
            typeof event.payload === 'string'
              ? event.payload
              : JSON.stringify(event.payload, null, 2)

          addMessage('data', content, event.payload)

          // 根据消息数量更新进度（这里是估算）
          const estimatedProgress = Math.min(messageCount * 10, 90)
          setProgress(estimatedProgress)
        }
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '未知错误'
      addMessage('error', `连接错误: ${errorMessage}`, error)

      notification.error({
        message: '流式执行失败',
        description: errorMessage,
      })
    } finally {
      setIsStreaming(false)
      abortControllerRef.current = null
    }
  }

  const handleStop = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
      addMessage('error', '用户主动停止执行')
      setIsStreaming(false)
      notification.info({
        message: '已停止',
        description: '工作流执行已停止',
      })
    }
  }

  const handleClear = () => {
    setMessages([])
    setProgress(0)
  }

  const handleBack = () => {
    navigate('/workflows')
  }

  const handleSwitchToNormal = () => {
    const currentValues = form.getFieldsValue()
    if (currentValues.workflow) {
      navigate(`/execute?workflow=${encodeURIComponent(currentValues.workflow)}`)
    } else {
      navigate('/execute')
    }
  }

  const getMessageTypeColor = (type: StreamMessage['type']) => {
    switch (type) {
      case 'start':
        return 'blue'
      case 'data':
        return 'green'
      case 'done':
        return 'cyan'
      case 'error':
        return 'red'
      default:
        return 'default'
    }
  }

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <div>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={handleBack}
            style={{ marginBottom: '16px' }}
          >
            返回工作流列表
          </Button>

          <Title level={2}>流式执行工作流</Title>
          <Text type="secondary">实时查看工作流执行过程和中间结果</Text>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px' }}>
          {/* 左侧：配置面板 */}
          <Card title="执行配置" size="small">
            <Form form={form} layout="vertical" onFinish={handleSubmit} disabled={isStreaming}>
              <Form.Item
                label="选择工作流"
                name="workflow"
                rules={[{ required: true, message: '请选择要执行的工作流' }]}
              >
                <Select
                  placeholder="选择工作流..."
                  loading={workflowsLoading}
                  showSearch
                  filterOption={(input, option) =>
                    (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                  }
                  options={
                    workflowsData?.workflows.map(name => ({
                      value: name,
                      label: name,
                    })) || []
                  }
                />
              </Form.Item>

              <Form.Item
                label="输入参数"
                name="input"
                rules={[{ required: true, message: '请输入工作流参数' }]}
              >
                <TextArea rows={4} placeholder="输入工作流所需的参数或数据..." showCount />
              </Form.Item>

              <Form.Item
                label="超时时间（秒）"
                name="timeout"
                initialValue={300}
                extra="流式执行的最大等待时间，默认5分钟"
              >
                <Input type="number" min={1} max={3600} addonAfter="秒" />
              </Form.Item>

              <Form.Item>
                <Space>
                  {!isStreaming ? (
                    <Button
                      type="primary"
                      htmlType="submit"
                      icon={<PlayCircleOutlined />}
                      size="large"
                    >
                      开始流式执行
                    </Button>
                  ) : (
                    <Button danger onClick={handleStop} icon={<StopOutlined />} size="large">
                      停止执行
                    </Button>
                  )}

                  <Button
                    onClick={handleSwitchToNormal}
                    icon={<PauseCircleOutlined />}
                    disabled={isStreaming}
                  >
                    切换到标准执行
                  </Button>
                </Space>
              </Form.Item>
            </Form>

            {/* 进度指示器 */}
            {isStreaming && (
              <div style={{ marginTop: '16px' }}>
                <Text strong>执行进度:</Text>
                <Progress
                  percent={progress}
                  status={isStreaming ? 'active' : 'normal'}
                  strokeColor={progress < 100 ? '#1890ff' : '#52c41a'}
                />
              </div>
            )}
          </Card>

          {/* 右侧：实时输出 */}
          <Card
            title={
              <Space>
                <span>实时输出</span>
                {isStreaming && <Tag color="green">执行中</Tag>}
                <Button
                  size="small"
                  icon={<ClearOutlined />}
                  onClick={handleClear}
                  disabled={isStreaming}
                >
                  清空
                </Button>
              </Space>
            }
            size="small"
          >
            <div
              style={{
                height: '400px',
                overflowY: 'auto',
                backgroundColor: '#f6f8fa',
                padding: '12px',
                borderRadius: '6px',
                fontFamily: 'monospace',
                fontSize: '13px',
              }}
            >
              {messages.length === 0 ? (
                <Text type="secondary">等待执行开始...</Text>
              ) : (
                messages.map(message => (
                  <div key={message.id} style={{ marginBottom: '8px' }}>
                    <div
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '8px',
                        marginBottom: '4px',
                      }}
                    >
                      <Tag color={getMessageTypeColor(message.type)}>{message.type}</Tag>
                      <Text type="secondary" style={{ fontSize: '12px' }}>
                        {new Date(message.timestamp).toLocaleTimeString()}
                      </Text>
                    </div>
                    <div
                      style={{
                        whiteSpace: 'pre-wrap',
                        wordBreak: 'break-word',
                        marginLeft: '8px',
                        color: message.type === 'error' ? '#ff4d4f' : '#000',
                      }}
                    >
                      {message.content}
                    </div>
                  </div>
                ))
              )}
              <div ref={messagesEndRef} />
            </div>
          </Card>
        </div>

        {/* 使用说明 */}
        <Card title="流式执行说明" size="small">
          <Space direction="vertical" style={{ width: '100%' }}>
            <div>
              <Text strong>流式执行特点:</Text>
              <ul style={{ marginTop: '8px', paddingLeft: '20px' }}>
                <li>
                  <Text>实时显示工作流执行过程</Text>
                </li>
                <li>
                  <Text>支持中途停止执行</Text>
                </li>
                <li>
                  <Text>可以查看中间结果和调试信息</Text>
                </li>
                <li>
                  <Text>适合长时间运行的工作流</Text>
                </li>
              </ul>
            </div>

            <Divider />

            <div>
              <Text strong>消息类型说明:</Text>
              <Space wrap style={{ marginTop: '8px' }}>
                <Tag color="blue">start - 开始执行</Tag>
                <Tag color="green">data - 数据输出</Tag>
                <Tag color="cyan">done - 执行完成</Tag>
                <Tag color="red">error - 执行错误</Tag>
              </Space>
            </div>
          </Space>
        </Card>
      </Space>
    </div>
  )
}
