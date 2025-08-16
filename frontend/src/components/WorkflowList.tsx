import { useState } from 'react'
import { Card, List, Button, Typography, Space, Input, Spin, Alert, Tag } from 'antd'
import { SearchOutlined, PlayCircleOutlined, InfoCircleOutlined } from '@ant-design/icons'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { apiClient } from '../api/client'

const { Title, Text } = Typography
const { Search } = Input

export function WorkflowList() {
  const navigate = useNavigate()
  const [searchTerm, setSearchTerm] = useState('')

  const {
    data: workflowsData,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['workflows'],
    queryFn: () => apiClient.getWorkflows(),
    refetchInterval: 30000, // 30秒自动刷新
  })

  const { data: workflowDetails, isLoading: detailsLoading } = useQuery({
    queryKey: ['workflow-details', workflowsData?.workflows],
    queryFn: async () => {
      if (!workflowsData?.workflows) return []

      const details = await Promise.allSettled(
        workflowsData.workflows.map(name => apiClient.getWorkflowInfo(name))
      )

      return details.map((result, index) => ({
        name: workflowsData.workflows[index],
        info: result.status === 'fulfilled' ? result.value : null,
        error: result.status === 'rejected' ? result.reason : null,
      }))
    },
    enabled: !!workflowsData?.workflows,
  })

  const filteredWorkflows =
    workflowDetails?.filter(
      item =>
        item.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        (item.info?.description || '').toLowerCase().includes(searchTerm.toLowerCase())
    ) || []

  const handleExecute = (workflowName: string) => {
    navigate(`/execute?workflow=${encodeURIComponent(workflowName)}`)
  }

  const handleStream = (workflowName: string) => {
    navigate(`/stream?workflow=${encodeURIComponent(workflowName)}`)
  }

  const handleViewDetails = (workflowName: string) => {
    navigate(`/workflows/${encodeURIComponent(workflowName)}`)
  }

  if (error) {
    return (
      <div style={{ padding: '24px' }}>
        <Alert
          type="error"
          message="加载工作流列表失败"
          description={error instanceof Error ? error.message : '未知错误'}
          action={<Button onClick={() => refetch()}>重试</Button>}
        />
      </div>
    )
  }

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <div>
          <Title level={2}>工作流列表</Title>
          <Text type="secondary">管理和执行系统中的所有工作流程</Text>
        </div>

        <Card size="small">
          <Space direction="vertical" style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Search
                placeholder="搜索工作流名称或描述..."
                allowClear
                style={{ width: '300px' }}
                value={searchTerm}
                onChange={e => setSearchTerm(e.target.value)}
                prefix={<SearchOutlined />}
              />

              <Space>
                <Button onClick={() => refetch()} loading={isLoading}>
                  刷新
                </Button>
                <Text type="secondary">共 {workflowsData?.count || 0} 个工作流</Text>
              </Space>
            </div>

            {isLoading || detailsLoading ? (
              <div style={{ textAlign: 'center', padding: '48px' }}>
                <Spin size="large" />
                <div style={{ marginTop: '16px' }}>
                  <Text type="secondary">加载工作流信息...</Text>
                </div>
              </div>
            ) : (
              <List
                dataSource={filteredWorkflows}
                renderItem={item => (
                  <List.Item
                    key={item.name}
                    actions={[
                      <Button
                        key="details"
                        icon={<InfoCircleOutlined />}
                        onClick={() => handleViewDetails(item.name)}
                      >
                        详情
                      </Button>,
                      <Button
                        key="execute"
                        type="primary"
                        icon={<PlayCircleOutlined />}
                        onClick={() => handleExecute(item.name)}
                      >
                        执行
                      </Button>,
                      <Button
                        key="stream"
                        icon={<PlayCircleOutlined />}
                        onClick={() => handleStream(item.name)}
                      >
                        流式执行
                      </Button>,
                    ]}
                  >
                    <List.Item.Meta
                      title={
                        <Space>
                          <Text strong>{item.name}</Text>
                          {item.info?.status && (
                            <Tag color={item.info.status === 'active' ? 'green' : 'orange'}>
                              {item.info.status}
                            </Tag>
                          )}
                          {item.info?.version && <Tag color="blue">v{item.info.version}</Tag>}
                        </Space>
                      }
                      description={
                        item.error ? (
                          <Text type="danger">
                            获取详情失败: {item.error.message || '未知错误'}
                          </Text>
                        ) : (
                          <Text type="secondary">{item.info?.description || '暂无描述'}</Text>
                        )
                      }
                    />
                  </List.Item>
                )}
              />
            )}

            {!isLoading && !detailsLoading && filteredWorkflows.length === 0 && (
              <div style={{ textAlign: 'center', padding: '48px' }}>
                <Text type="secondary">{searchTerm ? '没有找到匹配的工作流' : '暂无工作流'}</Text>
              </div>
            )}
          </Space>
        </Card>
      </Space>
    </div>
  )
}
