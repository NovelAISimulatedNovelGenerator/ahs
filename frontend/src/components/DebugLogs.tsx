import { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Button,
  Typography,
  Space,
  Tag,
  Modal,
  Input,
  Select,
  Statistic,
  Row,
  Col,
  notification,
} from 'antd'
import {
  DeleteOutlined,
  EyeOutlined,
  ReloadOutlined,
  ExportOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import { useSettingsStore } from '../stores/settings'
import { apiClient } from '../api/client'
import type { DebugLogEntry } from '../types/api'
import type { ColumnsType } from 'antd/es/table'

const { Title, Text } = Typography
const { Search } = Input

export function DebugLogs() {
  const { clearDebugLogs } = useSettingsStore()
  const [logs, setLogs] = useState<DebugLogEntry[]>([])
  const [filteredLogs, setFilteredLogs] = useState<DebugLogEntry[]>([])
  const [selectedLog, setSelectedLog] = useState<DebugLogEntry | null>(null)
  const [isModalVisible, setIsModalVisible] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')
  const [directionFilter, setDirectionFilter] = useState<string>('all')
  const [methodFilter, setMethodFilter] = useState<string>('all')

  // 定期刷新日志
  useEffect(() => {
    const refreshLogs = () => {
      const apiLogs = apiClient.getDebugLogs()
      setLogs(apiLogs.reverse()) // 最新的在前面
    }

    // 初始加载
    refreshLogs()

    // 每2秒刷新一次
    const interval = setInterval(refreshLogs, 2000)

    return () => clearInterval(interval)
  }, [])

  // 过滤日志
  useEffect(() => {
    let filtered = logs

    // 文本搜索
    if (searchTerm) {
      filtered = filtered.filter(
        log =>
          log.path.toLowerCase().includes(searchTerm.toLowerCase()) ||
          log.summary.toLowerCase().includes(searchTerm.toLowerCase()) ||
          (log.method && log.method.toLowerCase().includes(searchTerm.toLowerCase()))
      )
    }

    // 方向过滤
    if (directionFilter !== 'all') {
      filtered = filtered.filter(log => log.direction === directionFilter)
    }

    // 方法过滤
    if (methodFilter !== 'all') {
      filtered = filtered.filter(log => log.method === methodFilter)
    }

    setFilteredLogs(filtered)
  }, [logs, searchTerm, directionFilter, methodFilter])

  const handleViewDetails = (log: DebugLogEntry) => {
    setSelectedLog(log)
    setIsModalVisible(true)
  }

  const handleClearLogs = () => {
    clearDebugLogs()
    setLogs([])
    notification.success({
      message: '已清空',
      description: '调试日志已清空',
    })
  }

  const handleExportLogs = () => {
    const data = JSON.stringify(logs, null, 2)
    const blob = new Blob([data], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `debug-logs-${new Date().toISOString().slice(0, 10)}.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)

    notification.success({
      message: '导出成功',
      description: '调试日志已导出到本地文件',
    })
  }

  const getDirectionColor = (direction: string) => {
    switch (direction) {
      case 'request':
        return 'blue'
      case 'response':
        return 'green'
      case 'event':
        return 'purple'
      default:
        return 'default'
    }
  }

  const getStatusColor = (statusCode?: number) => {
    if (!statusCode) return 'default'
    if (statusCode >= 200 && statusCode < 300) return 'green'
    if (statusCode >= 400 && statusCode < 500) return 'orange'
    if (statusCode >= 500) return 'red'
    return 'default'
  }

  const columns: ColumnsType<DebugLogEntry> = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 100,
      render: (timestamp: number) => (
        <Text style={{ fontSize: '12px' }}>{new Date(timestamp).toLocaleTimeString()}</Text>
      ),
    },
    {
      title: '方向',
      dataIndex: 'direction',
      key: 'direction',
      width: 80,
      render: (direction: string) => <Tag color={getDirectionColor(direction)}>{direction}</Tag>,
    },
    {
      title: '方法',
      dataIndex: 'method',
      key: 'method',
      width: 70,
      render: (method?: string) => (method ? <Tag>{method}</Tag> : <Text type="secondary">-</Text>),
    },
    {
      title: '路径',
      dataIndex: 'path',
      key: 'path',
      width: 200,
      render: (path: string) => (
        <Text code style={{ fontSize: '12px' }}>
          {path}
        </Text>
      ),
    },
    {
      title: '状态',
      dataIndex: 'statusCode',
      key: 'statusCode',
      width: 80,
      render: (statusCode: number | undefined, record: DebugLogEntry) => {
        if (record.direction === 'request') return <Text type="secondary">-</Text>
        if (!statusCode) return <Text type="secondary">-</Text>
        return <Tag color={getStatusColor(statusCode)}>{statusCode}</Tag>
      },
    },
    {
      title: '摘要',
      dataIndex: 'summary',
      key: 'summary',
      ellipsis: true,
      render: (summary: string) => <Text style={{ fontSize: '13px' }}>{summary}</Text>,
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_, record) => (
        <Button size="small" icon={<EyeOutlined />} onClick={() => handleViewDetails(record)}>
          详情
        </Button>
      ),
    },
  ]

  // 统计信息
  const stats = {
    total: logs.length,
    requests: logs.filter(log => log.direction === 'request').length,
    responses: logs.filter(log => log.direction === 'response').length,
    events: logs.filter(log => log.direction === 'event').length,
    errors: logs.filter(
      log =>
        (log.statusCode && log.statusCode >= 400) ||
        log.summary.toLowerCase().includes('error') ||
        log.summary.toLowerCase().includes('错误')
    ).length,
  }

  const uniqueMethods = Array.from(new Set(logs.map(log => log.method).filter(Boolean)))

  return (
    <div style={{ padding: '24px', maxWidth: '1400px', margin: '0 auto' }}>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <div>
          <Title level={2}>调试日志</Title>
          <Text type="secondary">实时监控API请求、响应和事件流，便于调试和问题排查</Text>
        </div>

        {/* 统计卡片 */}
        <Row gutter={16}>
          <Col span={6}>
            <Card size="small">
              <Statistic title="总日志数" value={stats.total} />
            </Card>
          </Col>
          <Col span={6}>
            <Card size="small">
              <Statistic title="请求" value={stats.requests} valueStyle={{ color: '#1890ff' }} />
            </Card>
          </Col>
          <Col span={6}>
            <Card size="small">
              <Statistic title="响应" value={stats.responses} valueStyle={{ color: '#52c41a' }} />
            </Card>
          </Col>
          <Col span={6}>
            <Card size="small">
              <Statistic title="错误" value={stats.errors} valueStyle={{ color: '#ff4d4f' }} />
            </Card>
          </Col>
        </Row>

        {/* 过滤和操作区域 */}
        <Card size="small">
          <Space direction="vertical" style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Space>
                <Search
                  placeholder="搜索路径、摘要或方法..."
                  allowClear
                  style={{ width: '300px' }}
                  value={searchTerm}
                  onChange={e => setSearchTerm(e.target.value)}
                  prefix={<SearchOutlined />}
                />

                <Select
                  value={directionFilter}
                  onChange={setDirectionFilter}
                  style={{ width: '120px' }}
                >
                  <Select.Option value="all">全部方向</Select.Option>
                  <Select.Option value="request">请求</Select.Option>
                  <Select.Option value="response">响应</Select.Option>
                  <Select.Option value="event">事件</Select.Option>
                </Select>

                <Select value={methodFilter} onChange={setMethodFilter} style={{ width: '120px' }}>
                  <Select.Option value="all">全部方法</Select.Option>
                  {uniqueMethods.map(method => (
                    <Select.Option key={method} value={method}>
                      {method}
                    </Select.Option>
                  ))}
                </Select>
              </Space>

              <Space>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={() => {
                    const apiLogs = apiClient.getDebugLogs()
                    setLogs(apiLogs.reverse())
                  }}
                >
                  刷新
                </Button>

                <Button
                  icon={<ExportOutlined />}
                  onClick={handleExportLogs}
                  disabled={logs.length === 0}
                >
                  导出
                </Button>

                <Button
                  danger
                  icon={<DeleteOutlined />}
                  onClick={handleClearLogs}
                  disabled={logs.length === 0}
                >
                  清空日志
                </Button>
              </Space>
            </div>

            <Text type="secondary">
              显示 {filteredLogs.length} / {logs.length} 条日志
              {logs.length > 0 && <span style={{ marginLeft: '16px' }}>自动刷新中...</span>}
            </Text>
          </Space>
        </Card>

        {/* 日志表格 */}
        <Card size="small">
          <Table
            columns={columns}
            dataSource={filteredLogs}
            rowKey="id"
            size="small"
            pagination={{
              pageSize: 50,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
            }}
            scroll={{ x: 800 }}
          />
        </Card>

        {/* 详情模态框 */}
        <Modal
          title="日志详情"
          open={isModalVisible}
          onCancel={() => setIsModalVisible(false)}
          footer={[
            <Button key="close" onClick={() => setIsModalVisible(false)}>
              关闭
            </Button>,
          ]}
          width={800}
        >
          {selectedLog && (
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <Text strong>基本信息:</Text>
                <div style={{ marginTop: '8px', marginLeft: '16px' }}>
                  <p>
                    <Text strong>ID:</Text> {selectedLog.id}
                  </p>
                  <p>
                    <Text strong>时间:</Text> {new Date(selectedLog.timestamp).toLocaleString()}
                  </p>
                  <p>
                    <Text strong>方向:</Text>{' '}
                    <Tag color={getDirectionColor(selectedLog.direction)}>
                      {selectedLog.direction}
                    </Tag>
                  </p>
                  <p>
                    <Text strong>路径:</Text> <Text code>{selectedLog.path}</Text>
                  </p>
                  {selectedLog.method && (
                    <p>
                      <Text strong>方法:</Text> <Tag>{selectedLog.method}</Tag>
                    </p>
                  )}
                  {selectedLog.statusCode && (
                    <p>
                      <Text strong>状态码:</Text>{' '}
                      <Tag color={getStatusColor(selectedLog.statusCode)}>
                        {selectedLog.statusCode}
                      </Tag>
                    </p>
                  )}
                  {selectedLog.eventType && (
                    <p>
                      <Text strong>事件类型:</Text> <Tag>{selectedLog.eventType}</Tag>
                    </p>
                  )}
                  <p>
                    <Text strong>摘要:</Text> {selectedLog.summary}
                  </p>
                </div>
              </div>

              {selectedLog.details != null && (
                <div>
                  <Text strong>详细信息:</Text>
                  <div
                    style={
                      {
                        marginTop: '8px',
                        background: '#f6f8fa',
                        padding: '12px',
                        borderRadius: '6px',
                        fontFamily: 'monospace',
                        fontSize: '12px',
                        maxHeight: '300px',
                        overflowY: 'auto',
                      } as React.CSSProperties
                    }
                  >
                    <pre>{JSON.stringify(selectedLog.details, null, 2)}</pre>
                  </div>
                </div>
              )}
            </Space>
          )}
        </Modal>
      </Space>
    </div>
  )
}
