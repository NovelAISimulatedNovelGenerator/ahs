import { createContext, useContext, useState, ReactNode } from 'react'

/**
 * 聊天设置接口
 * 用于配置聊天行为的各种选项
 */
interface ChatSettings {
  useStreaming: boolean     // 是否使用流式传输
  timeout: number          // 请求超时时间（秒）
  autoSaveMemory: boolean  // 是否自动保存记忆（未来扩展）
}

/**
 * 聊天Context接口
 * 提供聊天设置的状态管理和操作方法
 */
interface ChatContextType {
  settings: ChatSettings
  updateSettings: (newSettings: Partial<ChatSettings>) => void
  toggleStreaming: () => void
}

// 默认聊天设置
const defaultSettings: ChatSettings = {
  useStreaming: true,      // 默认使用流式传输
  timeout: 180,           // 默认3分钟超时
  autoSaveMemory: true,   // 默认自动保存记忆
}

// 创建Context
const ChatContext = createContext<ChatContextType | undefined>(undefined)

/**
 * 聊天Context提供者组件
 * 为所有子组件提供聊天设置的状态管理
 */
interface ChatProviderProps {
  children: ReactNode
}

export function ChatProvider({ children }: ChatProviderProps) {
  const [settings, setSettings] = useState<ChatSettings>(defaultSettings)

  /**
   * 更新聊天设置
   * @param newSettings 新的设置项（部分更新）
   */
  const updateSettings = (newSettings: Partial<ChatSettings>) => {
    setSettings(prev => ({ ...prev, ...newSettings }))
  }

  /**
   * 切换流式传输模式
   * 快捷方法，用于切换最常用的设置
   */
  const toggleStreaming = () => {
    setSettings(prev => ({ ...prev, useStreaming: !prev.useStreaming }))
  }

  const contextValue: ChatContextType = {
    settings,
    updateSettings,
    toggleStreaming,
  }

  return (
    <ChatContext.Provider value={contextValue}>
      {children}
    </ChatContext.Provider>
  )
}

/**
 * 使用聊天Context的Hook
 * 提供类型安全的方式访问聊天设置
 * @returns 聊天Context的值
 * @throws 如果在ChatProvider之外使用则抛出错误
 */
export function useChatContext(): ChatContextType {
  const context = useContext(ChatContext)
  if (context === undefined) {
    throw new Error('useChatContext must be used within a ChatProvider')
  }
  return context
}