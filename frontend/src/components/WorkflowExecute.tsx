import { useState, useEffect } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import {
  Card,
  Form,
  Input,
  Button,
  Typography,
  Space,
  Alert,
  Spin,
  Select,
  notification,
  Divider,
} from 'antd'
import { ArrowLeftOutlined, PlayCircleOutlined, ClockCircleOutlined } from '@ant-design/icons'
import { useMutation, useQuery } from '@tanstack/react-query'
import { apiClient } from '../api/client'
import { useSettingsStore } from '../stores/settings'
import type { WorkflowRequest, WorkflowResponse } from '../types/api'

const { Title, Text, Paragraph } = Typography
const { TextArea } = Input

export function WorkflowExecute() {
  const location = useLocation()
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [result, setResult] = useState<WorkflowResponse | null>(null)
  const { addDebugLog } = useSettingsStore()

  // 从URL参数获取预选工作流
  const searchParams = new URLSearchParams(location.search)
  const preSelectedWorkflow = searchParams.get('workflow')

  // 获取工作流列表
  const { data: workflowsData, isLoading: workflowsLoading } = useQuery({
    queryKey: ['workflows'],
    queryFn: () => apiClient.getWorkflows(),
  })

  // 执行工作流的mutation
  const executeMutation = useMutation({
    mutationFn: (request: WorkflowRequest) => {
      const startTime = Date.now()
      addDebugLog({
        id: '',
        timestamp: startTime,
        direction: 'request',
        path: '/api/execute',
        method: 'POST',
        summary: `开始执行工作流: ${request.workflow}`,
        details: request,
      })

      return apiClient.executeWorkflow(request)
    },
    onSuccess: data => {
      setResult(data)
      notification.success({
        message: '执行成功',
        description: '工作流执行完成',
      })
    },
    onError: error => {
      notification.error({
        message: '执行失败',
        description: error instanceof Error ? error.message : '未知错误',
      })
    },
  })

  // 初始化表单
  useEffect(() => {
    if (preSelectedWorkflow) {
      form.setFieldValue('workflow', preSelectedWorkflow)
    }
  }, [preSelectedWorkflow, form])

  const handleSubmit = (values: { workflow: string; input: string; timeout?: number }) => {
    const request: WorkflowRequest = {
      workflow: values.workflow,
      input: values.input,
      timeout: values.timeout || 180,
    }

    setResult(null)
    executeMutation.mutate(request)
  }

  const handleBack = () => {
    navigate('/workflows')
  }

  const handleSwitchToStream = () => {
    const currentValues = form.getFieldsValue()
    if (currentValues.workflow) {
      navigate(`/stream?workflow=${encodeURIComponent(currentValues.workflow)}`)
    } else {
      navigate('/stream')
    }
  }

  return (
    <div style={{ padding: '24px', maxWidth: '1000px', margin: '0 auto' }}>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <div>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={handleBack}
            style={{ marginBottom: '16px' }}
          >
            返回工作流列表
          </Button>

          <Title level={2}>执行工作流</Title>
          <Text type="secondary">选择工作流并提供输入参数，等待完整执行结果</Text>
        </div>

        <Card title="执行配置" size="small">
          <Form
            form={form}
            layout="vertical"
            onFinish={handleSubmit}
            disabled={executeMutation.isPending}
          >
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
              <TextArea rows={6} placeholder="输入工作流所需的参数或数据..." showCount />
            </Form.Item>

            <Form.Item
              label="超时时间（秒）"
              name="timeout"
              initialValue={180}
              extra="工作流执行的最大等待时间，默认3分钟"
            >
              <Input type="number" min={1} max={3600} addonAfter="秒" />
            </Form.Item>

            <Form.Item>
              <Space>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={executeMutation.isPending}
                  icon={<PlayCircleOutlined />}
                  size="large"
                >
                  {executeMutation.isPending ? '执行中...' : '执行工作流'}
                </Button>

                <Button onClick={handleSwitchToStream} icon={<ClockCircleOutlined />}>
                  切换到流式执行
                </Button>
              </Space>
            </Form.Item>
          </Form>
        </Card>

        {/* 执行状态 */}
        {executeMutation.isPending && (
          <Card size="small">
            <div style={{ textAlign: 'center', padding: '24px' }}>
              <Spin size="large" />
              <div style={{ marginTop: '16px' }}>
                <Text>正在执行工作流，请稍候...</Text>
              </div>
            </div>
          </Card>
        )}

        {/* 执行结果 */}
        {result && (
          <Card
            title={
              <Space>
                <span>执行结果</span>
                {result.status === 'success' ? (
                  <span style={{ color: '#52c41a' }}>✓ 成功</span>
                ) : (
                  <span style={{ color: '#ff4d4f' }}>✗ 失败</span>
                )}
              </Space>
            }
            size="small"
          >
            {result.status === 'success' ? (
              <div>
                <Paragraph>
                  <Text strong>结果:</Text>
                </Paragraph>
                <div
                  style={{
                    background: '#f6f8fa',
                    padding: '16px',
                    borderRadius: '6px',
                    fontFamily: 'monospace',
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-word',
                  }}
                >
                  {result.result || '(无输出)'}
                </div>
              </div>
            ) : (
              <Alert
                type="error"
                message="执行失败"
                description={
                  <div
                    style={{
                      fontFamily: 'monospace',
                      whiteSpace: 'pre-wrap',
                      wordBreak: 'break-word',
                    }}
                  >
                    {result.error || '未知错误'}
                  </div>
                }
              />
            )}
          </Card>
        )}

        {/* 错误信息 */}
        {executeMutation.error && (
          <Alert
            type="error"
            message="请求失败"
            description={executeMutation.error.message}
            closable
            onClose={() => executeMutation.reset()}
          />
        )}

        {/* 使用说明 */}
        <Card title="使用说明" size="small">
          <Space direction="vertical" style={{ width: '100%' }}>
            <div>
              <Text strong>执行模式说明:</Text>
              <ul style={{ marginTop: '8px', paddingLeft: '20px' }}>
                <li>
                  <Text>
                    <strong>标准执行</strong>: 提交请求后等待工作流完全执行完成
                  </Text>
                </li>
                <li>
                  <Text>
                    <strong>流式执行</strong>: 实时显示执行过程和中间结果
                  </Text>
                </li>
              </ul>
            </div>

            <Divider />

            <div>
              <Text strong>参数格式:</Text>
              <ul style={{ marginTop: '8px', paddingLeft: '20px' }}>
                <li>
                  <Text>可以是纯文本、JSON字符串或其他格式</Text>
                </li>
                <li>
                  <Text>具体格式取决于所选工作流的要求</Text>
                </li>
                <li>
                  <Text>建议查看工作流详情了解所需参数格式</Text>
                </li>
              </ul>
            </div>
          </Space>
        </Card>
      </Space>
    </div>
  )
}
