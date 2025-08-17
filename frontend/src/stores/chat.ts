import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { ChatMessage, ChatSession, ChatStore } from '../types/api'

interface ChatStoreActions {
  // 会话管理
  createSession: (title?: string) => string
  deleteSession: (sessionId: string) => void
  setCurrentSession: (sessionId: string | null) => void
  getCurrentSession: () => ChatSession | null
  updateSessionTitle: (sessionId: string, title: string) => void
  
  // 消息管理
  addMessage: (sessionId: string, message: Omit<ChatMessage, 'id' | 'timestamp'>) => string
  updateMessage: (sessionId: string, messageId: string, content: string) => void
  setMessageStreaming: (sessionId: string, messageId: string, isStreaming: boolean) => void
  
  // 流式传输状态
  setStreaming: (isStreaming: boolean) => void
  
  // 清理
  clearAllSessions: () => void
}

export const useChatStore = create<ChatStore & ChatStoreActions>()(
  persist(
    (set, get) => ({
      // 状态
      sessions: [],
      currentSessionId: null,
      isStreaming: false,

      // 会话管理
      createSession: (title = '新对话') => {
        const id = `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
        const newSession: ChatSession = {
          id,
          title,
          messages: [],
          createdAt: Date.now(),
          updatedAt: Date.now(),
        }
        
        set((state) => ({
          sessions: [newSession, ...state.sessions],
          currentSessionId: id,
        }))
        
        return id
      },

      deleteSession: (sessionId) => {
        set((state) => ({
          sessions: state.sessions.filter(s => s.id !== sessionId),
          currentSessionId: state.currentSessionId === sessionId ? null : state.currentSessionId,
        }))
      },

      setCurrentSession: (sessionId) => {
        set({ currentSessionId: sessionId })
      },

      getCurrentSession: () => {
        const { sessions, currentSessionId } = get()
        return sessions.find(s => s.id === currentSessionId) || null
      },

      updateSessionTitle: (sessionId, title) => {
        set((state) => ({
          sessions: state.sessions.map(session =>
            session.id === sessionId
              ? { ...session, title, updatedAt: Date.now() }
              : session
          ),
        }))
      },

      // 消息管理
      addMessage: (sessionId, messageData) => {
        const messageId = `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
        const message: ChatMessage = {
          ...messageData,
          id: messageId,
          timestamp: Date.now(),
        }

        set((state) => ({
          sessions: state.sessions.map(session =>
            session.id === sessionId
              ? {
                  ...session,
                  messages: [...session.messages, message],
                  updatedAt: Date.now(),
                }
              : session
          ),
        }))

        return messageId
      },

      updateMessage: (sessionId, messageId, content) => {
        set((state) => ({
          sessions: state.sessions.map(session =>
            session.id === sessionId
              ? {
                  ...session,
                  messages: session.messages.map(msg =>
                    msg.id === messageId ? { ...msg, content } : msg
                  ),
                  updatedAt: Date.now(),
                }
              : session
          ),
        }))
      },

      setMessageStreaming: (sessionId, messageId, isStreaming) => {
        set((state) => ({
          sessions: state.sessions.map(session =>
            session.id === sessionId
              ? {
                  ...session,
                  messages: session.messages.map(msg =>
                    msg.id === messageId ? { ...msg, isStreaming } : msg
                  ),
                  updatedAt: Date.now(),
                }
              : session
          ),
        }))
      },

      // 流式传输状态
      setStreaming: (isStreaming) => {
        set({ isStreaming })
      },

      // 清理
      clearAllSessions: () => {
        set({
          sessions: [],
          currentSessionId: null,
          isStreaming: false,
        })
      },
    }),
    {
      name: 'chat-store',
      partialize: (state) => ({
        sessions: state.sessions,
        currentSessionId: state.currentSessionId,
      }),
    }
  )
)