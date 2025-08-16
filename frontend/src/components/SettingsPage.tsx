import { useState, useEffect } from 'react'
import { Card, Form, Input, Button, Typography, Space, notification } from 'antd'
import { useSettingsStore } from '../stores/settings'
import type { TenantInfo, ApiSettings } from '../types/api'

const { Title, Text } = Typography

export function SettingsPage() {
  const { apiSettings, tenantInfo, setApiSettings, setTenantInfo, initialize } = useSettingsStore()

  const [apiForm] = Form.useForm()
  const [tenantForm] = Form.useForm()
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    initialize()

    // 初始化表单值
    apiForm.setFieldsValue(apiSettings)
    if (tenantInfo) {
      tenantForm.setFieldsValue(tenantInfo)
    }
  }, [apiForm, tenantForm, apiSettings, tenantInfo, initialize])

  const handleApiSettingsSubmit = async (values: ApiSettings) => {
    setLoading(true)
    try {
      setApiSettings(values)
      notification.success({
        message: '保存成功',
        description: 'API设置已更新',
      })
    } catch (error) {
      notification.error({
        message: '保存失败',
        description: error instanceof Error ? error.message : '未知错误',
      })
    } finally {
      setLoading(false)
    }
  }

  const handleTenantSubmit = async (values: TenantInfo) => {
    setLoading(true)
    try {
      setTenantInfo(values)
      notification.success({
        message: '保存成功',
        description: '租户信息已更新',
      })
    } catch (error) {
      notification.error({
        message: '保存失败',
        description: error instanceof Error ? error.message : '未知错误',
      })
    } finally {
      setLoading(false)
    }
  }

  const handleClearTenant = () => {
    setTenantInfo(null)
    tenantForm.resetFields()
    notification.info({
      message: '已清除',
      description: '租户信息已清除',
    })
  }

  return (
    <div style={{ padding: '24px', maxWidth: '800px', margin: '0 auto' }}>
      <Title level={2}>系统设置</Title>

      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        {/* API设置 */}
        <Card title="API设置" size="small">
          <Form
            form={apiForm}
            layout="vertical"
            onFinish={handleApiSettingsSubmit}
            initialValues={apiSettings}
          >
            <Form.Item
              label="API基础地址"
              name="apiBase"
              rules={[
                { required: true, message: '请输入API基础地址' },
                { type: 'url', message: '请输入有效的URL' },
              ]}
              extra="后端服务的基础URL，如: http://localhost:8081"
            >
              <Input placeholder="http://localhost:8081" />
            </Form.Item>

            <Form.Item>
              <Button type="primary" htmlType="submit" loading={loading}>
                保存API设置
              </Button>
            </Form.Item>
          </Form>
        </Card>

        {/* 租户设置 */}
        <Card title="租户设置" size="small">
          <Text type="secondary" style={{ display: 'block', marginBottom: '16px' }}>
            配置用户ID和归档ID，用于RAG系统的多租户隔离。留空则使用默认租户。
          </Text>

          <Form
            form={tenantForm}
            layout="vertical"
            onFinish={handleTenantSubmit}
            initialValues={tenantInfo || {}}
          >
            <Form.Item
              label="用户ID"
              name="userId"
              rules={[{ pattern: /^[a-zA-Z0-9_-]+$/, message: '只允许字母、数字、下划线和连字符' }]}
              extra="用于标识用户身份的唯一ID"
            >
              <Input placeholder="user_001" />
            </Form.Item>

            <Form.Item
              label="归档ID"
              name="archiveId"
              rules={[{ pattern: /^[a-zA-Z0-9_-]+$/, message: '只允许字母、数字、下划线和连字符' }]}
              extra="用于数据归档和隔离的ID"
            >
              <Input placeholder="archive_001" />
            </Form.Item>

            <Form.Item>
              <Space>
                <Button type="primary" htmlType="submit" loading={loading}>
                  保存租户设置
                </Button>
                <Button onClick={handleClearTenant}>清除设置</Button>
              </Space>
            </Form.Item>
          </Form>
        </Card>

        {/* 当前状态显示 */}
        <Card title="当前状态" size="small">
          <Space direction="vertical" style={{ width: '100%' }}>
            <div>
              <Text strong>API地址: </Text>
              <Text code>{apiSettings.apiBase}</Text>
            </div>

            {tenantInfo ? (
              <>
                <div>
                  <Text strong>用户ID: </Text>
                  <Text code>{tenantInfo.userId}</Text>
                </div>
                <div>
                  <Text strong>归档ID: </Text>
                  <Text code>{tenantInfo.archiveId}</Text>
                </div>
              </>
            ) : (
              <div>
                <Text strong>租户状态: </Text>
                <Text type="secondary">未配置（使用默认租户）</Text>
              </div>
            )}
          </Space>
        </Card>
      </Space>
    </div>
  )
}
