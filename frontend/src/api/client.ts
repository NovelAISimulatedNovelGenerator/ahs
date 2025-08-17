import type {
  WorkflowInfo,
  WorkflowListResp,
  WorkflowRequest,
  WorkflowResponse,
  SSEEvent,
  TenantInfo,
  DebugLogEntry,
} from '../types/api'

class ApiClient {
  private baseURL: string
  private tenant: TenantInfo | null = null
  private debugLogs: DebugLogEntry[] = []

  constructor() {
    this.baseURL = import.meta.env.VITE_API_BASE || 'http://localhost:8081'
  }

  setTenant(tenant: TenantInfo) {
    this.tenant = tenant
  }

  private getHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }

    // 始终添加租户头部，使用配置的值或默认值
    if (this.tenant) {
      headers['X-User-ID'] = this.tenant.userId
      headers['X-Archive-ID'] = this.tenant.archiveId
    } else {
      // 使用默认租户信息，避免后端拒绝请求
      headers['X-User-ID'] = 'default_user'
      headers['X-Archive-ID'] = 'default_archive'
    }

    return headers
  }

  private addDebugLog(entry: Omit<DebugLogEntry, 'id' | 'timestamp'>) {
    const log: DebugLogEntry = {
      ...entry,
      id: Math.random().toString(36).substr(2, 9),
      timestamp: Date.now(),
    }
    this.debugLogs.push(log)
    // 保持最近1000条日志
    if (this.debugLogs.length > 1000) {
      this.debugLogs = this.debugLogs.slice(-1000)
    }
  }

  getDebugLogs(): DebugLogEntry[] {
    return [...this.debugLogs]
  }

  clearDebugLogs() {
    this.debugLogs = []
  }

  private async handleResponse<T>(response: Response, path: string): Promise<T> {
    const statusCode = response.status

    if (!response.ok) {
      const errorText = await response.text()
      this.addDebugLog({
        direction: 'response',
        path,
        statusCode,
        summary: `HTTP ${statusCode}: ${errorText}`,
        details: { error: errorText },
      })
      throw new Error(`HTTP ${statusCode}: ${errorText}`)
    }

    const data = await response.json()

    this.addDebugLog({
      direction: 'response',
      path,
      statusCode,
      summary: `成功响应 (${statusCode})`,
      details: data,
    })

    // 检查业务层错误
    if (data.status === 'error') {
      throw new Error(data.error || '未知错误')
    }

    return data
  }

  async getWorkflows(): Promise<WorkflowListResp> {
    const path = '/api/workflows'
    const url = `${this.baseURL}${path}`

    this.addDebugLog({
      direction: 'request',
      path,
      method: 'GET',
      summary: '获取工作流列表',
    })

    const response = await fetch(url, {
      method: 'GET',
      headers: this.getHeaders(),
    })

    return this.handleResponse<WorkflowListResp>(response, path)
  }

  async getWorkflowInfo(name: string): Promise<WorkflowInfo> {
    const path = `/api/workflows/${encodeURIComponent(name)}`
    const url = `${this.baseURL}${path}`

    this.addDebugLog({
      direction: 'request',
      path,
      method: 'GET',
      summary: `获取工作流详情: ${name}`,
    })

    const response = await fetch(url, {
      method: 'GET',
      headers: this.getHeaders(),
    })

    return this.handleResponse<WorkflowInfo>(response, path)
  }

  async executeWorkflow(request: WorkflowRequest): Promise<WorkflowResponse> {
    const path = '/api/execute'
    const url = `${this.baseURL}${path}`

    // 兼容期：同时在Header和Body中添加租户信息
    const body = { ...request }
    if (this.tenant) {
      body.user_id = this.tenant.userId
      body.archive_id = this.tenant.archiveId
    }

    this.addDebugLog({
      direction: 'request',
      path,
      method: 'POST',
      summary: `执行工作流: ${request.workflow}`,
      details: body,
    })

    const response = await fetch(url, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(body),
    })

    return this.handleResponse<WorkflowResponse>(response, path)
  }

  async *streamWorkflow(
    request: WorkflowRequest,
    abortController?: AbortController
  ): AsyncGenerator<SSEEvent, void, unknown> {
    const path = '/api/stream'
    const url = `${this.baseURL}${path}`

    // 兼容期：同时在Header和Body中添加租户信息
    const body = { ...request }
    if (this.tenant) {
      body.user_id = this.tenant.userId
      body.archive_id = this.tenant.archiveId
    }

    this.addDebugLog({
      direction: 'request',
      path,
      method: 'POST',
      summary: `流式执行工作流: ${request.workflow}`,
      details: body,
    })

    const response = await fetch(url, {
      method: 'POST',
      headers: {
        ...this.getHeaders(),
        Accept: 'text/event-stream',
      },
      body: JSON.stringify(body),
      signal: abortController?.signal,
    })

    if (!response.ok) {
      const errorText = await response.text()
      this.addDebugLog({
        direction: 'response',
        path,
        statusCode: response.status,
        summary: `HTTP ${response.status}: ${errorText}`,
        details: { error: errorText },
      })
      throw new Error(`HTTP ${response.status}: ${errorText}`)
    }

    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error('无法获取响应流')
    }

    const decoder = new TextDecoder()
    let buffer = ''

    try {
      while (true) {
        const { done, value } = await reader.read()

        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || '' // 保留最后一个不完整的行

        for (const line of lines) {
          if (line.trim() === '') continue

          if (line.startsWith('event: ')) {
            // const eventType = line.slice(7).trim() as SSEEvent['type']
            continue
          }

          if (line.startsWith('data: ')) {
            const data = line.slice(6)
            let payload: string | { error: string }

            try {
              // 尝试解析为JSON（error事件）
              payload = JSON.parse(data)
            } catch {
              // 普通字符串数据
              payload = data
            }

            const event: SSEEvent = {
              type: 'data', // 默认类型，实际应该从前面的event行获取
              payload,
            }

            this.addDebugLog({
              direction: 'event',
              path,
              eventType: event.type,
              summary: `SSE事件: ${event.type}`,
              details: event.payload,
            })

            yield event
          }
        }
      }
    } finally {
      reader.releaseLock()
    }
  }
}

export const apiClient = new ApiClient()
export default apiClient
