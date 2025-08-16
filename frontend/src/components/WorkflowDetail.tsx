import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Descriptions, Button, Typography, Space, Spin, Alert, Tag, Divider } from 'antd'
import { ArrowLeftOutlined, PlayCircleOutlined } from '@ant-design/icons'
import { useQuery } from '@tanstack/react-query'
import { apiClient } from '../api/client'

const { Title, Text, Paragraph } = Typography

export function WorkflowDetail() {
  const { name } = useParams<{ name: string }>()
  const navigate = useNavigate()
  const [retryCount, setRetryCount] = useState(0)

  const {
    data: workflowInfo,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['workflow-info', name],
    queryFn: () => {
      if (!name) throw new Error('工作流名称不能为空')
      return apiClient.getWorkflowInfo(decodeURIComponent(name))
    },
    enabled: !!name,
    retry: failureCount => {
      // 最多重试3次
      if (failureCount < 3) {
        setTimeout(() => setRetryCount(prev => prev + 1), 1000)
        return true
      }
      return false
    },
  })

  const handleBack = () => {
    navigate('/workflows')
  }

  const handleExecute = () => {
    if (name) {
      navigate(`/execute?workflow=${encodeURIComponent(name)}`)
    }
  }

  const handleStream = () => {
    if (name) {
      navigate(`/stream?workflow=${encodeURIComponent(name)}`)
    }
  }

  const handleRetry = () => {
    setRetryCount(0)
    refetch()
  }

  if (!name) {
    return (
      <div style={{ padding: '24px' }}>
        <Alert
          type="error"
          message="参数错误"
          description="未指定工作流名称"
          action={<Button onClick={handleBack}>返回列表</Button>}
        />
      </div>
    )
  }

  if (error) {
    return (
      <div style={{ padding: '24px', maxWidth: '800px', margin: '0 auto' }}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <Button icon={<ArrowLeftOutlined />} onClick={handleBack}>
            返回工作流列表
          </Button>

          <Alert
            type="error"
            message="加载工作流详情失败"
            description={
              <div>
                <Paragraph>{error instanceof Error ? error.message : '未知错误'}</Paragraph>
                {retryCount > 0 && <Text type="secondary">已重试 {retryCount} 次</Text>}
              </div>
            }
            action={<Button onClick={handleRetry}>重试</Button>}
          />
        </Space>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div style={{ padding: '24px', maxWidth: '800px', margin: '0 auto' }}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <Button icon={<ArrowLeftOutlined />} onClick={handleBack}>
            返回工作流列表
          </Button>

          <div style={{ textAlign: 'center', padding: '48px' }}>
            <Spin size="large" />
            <div style={{ marginTop: '16px' }}>
              <Text type="secondary">加载工作流详情...</Text>
              {retryCount > 0 && (
                <div style={{ marginTop: '8px' }}>
                  <Text type="secondary">重试中... ({retryCount}/3)</Text>
                </div>
              )}
            </div>
          </div>
        </Space>
      </div>
    )
  }

  return (
    <div style={{ padding: '24px', maxWidth: '800px', margin: '0 auto' }}>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div style={{ flex: 1 }}>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={handleBack}
              style={{ marginBottom: '16px' }}
            >
              返回工作流列表
            </Button>

            <Title level={2} style={{ margin: 0 }}>
              {decodeURIComponent(name)}
            </Title>

            <div style={{ marginTop: '8px' }}>
              <Space>
                {workflowInfo?.status && (
                  <Tag color={workflowInfo.status === 'active' ? 'green' : 'orange'}>
                    {workflowInfo.status}
                  </Tag>
                )}
                {workflowInfo?.version && <Tag color="blue">版本 {workflowInfo.version}</Tag>}
              </Space>
            </div>
          </div>

          <Space>
            <Button
              type="primary"
              icon={<PlayCircleOutlined />}
              onClick={handleExecute}
              size="large"
            >
              执行工作流
            </Button>
            <Button icon={<PlayCircleOutlined />} onClick={handleStream} size="large">
              流式执行
            </Button>
          </Space>
        </div>

        <Card title="工作流信息" size="small">
          <Descriptions
            bordered
            column={1}
            size="small"
            labelStyle={{ width: '120px', fontWeight: 'bold' }}
          >
            <Descriptions.Item label="名称">
              <Text code>{decodeURIComponent(name)}</Text>
            </Descriptions.Item>

            <Descriptions.Item label="描述">
              {workflowInfo?.description ? (
                <Paragraph style={{ margin: 0 }}>{workflowInfo.description}</Paragraph>
              ) : (
                <Text type="secondary">暂无描述</Text>
              )}
            </Descriptions.Item>

            <Descriptions.Item label="版本">
              {workflowInfo?.version ? (
                <Tag color="blue">{workflowInfo.version}</Tag>
              ) : (
                <Text type="secondary">未知</Text>
              )}
            </Descriptions.Item>

            <Descriptions.Item label="状态">
              {workflowInfo?.status ? (
                <Tag color={workflowInfo.status === 'active' ? 'green' : 'orange'}>
                  {workflowInfo.status}
                </Tag>
              ) : (
                <Text type="secondary">未知</Text>
              )}
            </Descriptions.Item>
          </Descriptions>
        </Card>

        <Card title="操作" size="small">
          <Space direction="vertical" style={{ width: '100%' }}>
            <Text type="secondary">选择执行方式来运行此工作流：</Text>

            <Space wrap>
              <Button type="primary" icon={<PlayCircleOutlined />} onClick={handleExecute}>
                标准执行
              </Button>
              <Button icon={<PlayCircleOutlined />} onClick={handleStream}>
                流式执行
              </Button>
            </Space>

            <Divider />

            <div>
              <Text strong>执行说明：</Text>
              <ul style={{ marginTop: '8px', paddingLeft: '20px' }}>
                <li>
                  <Text>
                    <strong>标准执行</strong>: 等待工作流完成后返回完整结果
                  </Text>
                </li>
                <li>
                  <Text>
                    <strong>流式执行</strong>: 实时显示工作流执行过程和中间结果
                  </Text>
                </li>
              </ul>
            </div>
          </Space>
        </Card>
      </Space>
    </div>
  )
}
