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

// 聊天相关类型
export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: number
  isStreaming?: boolean
}

export interface ChatSession {
  id: string
  title: string
  messages: ChatMessage[]
  createdAt: number
  updatedAt: number
}

export interface ChatStore {
  sessions: ChatSession[]
  currentSessionId: string | null
  isStreaming: boolean
}
