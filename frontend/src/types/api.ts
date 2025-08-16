// 基于 README.md 中定义的后端接口类型

export interface WorkflowInfo {
  name: string
  description: string
  version: string
  status: string
}

export interface WorkflowListResp {
  workflows: string[]
  count: number
}

export interface WorkflowRequest {
  workflow: string
  input: string
  timeout?: number
  // 兼容期字段
  user_id?: string
  archive_id?: string
}

export interface WorkflowResponse {
  status: 'success' | 'error'
  result?: string
  error?: string
}

export interface SSEEvent {
  type: 'data' | 'done' | 'error'
  payload: string | { error: string }
}

export interface TenantInfo {
  userId: string
  archiveId: string
}

export interface ApiSettings {
  apiBase: string
}

export interface DebugLogEntry {
  id: string
  timestamp: number
  direction: 'request' | 'response' | 'event'
  path: string
  method?: string
  statusCode?: number
  eventType?: string
  summary: string
  details?: unknown
}
