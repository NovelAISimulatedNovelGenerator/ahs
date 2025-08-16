import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { TenantInfo, ApiSettings, DebugLogEntry } from '../types/api'
import { apiClient } from '../api/client'

interface SettingsState {
  // API设置
  apiSettings: ApiSettings
  setApiSettings: (settings: ApiSettings) => void

  // 租户信息
  tenantInfo: TenantInfo | null
  setTenantInfo: (info: TenantInfo | null) => void

  // 调试日志
  debugLogs: DebugLogEntry[]
  addDebugLog: (log: DebugLogEntry) => void
  clearDebugLogs: () => void

  // 初始化
  initialize: () => void
}

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set, get) => ({
      apiSettings: {
        apiBase: import.meta.env.VITE_API_BASE || 'http://localhost:8081',
      },

      tenantInfo: null,

      debugLogs: [],

      setApiSettings: settings => {
        set({ apiSettings: settings })
      },

      setTenantInfo: info => {
        set({ tenantInfo: info })
        // 同步更新API客户端的租户信息
        if (info) {
          apiClient.setTenant(info)
        }
      },

      addDebugLog: log => {
        set(state => ({
          debugLogs: [log, ...state.debugLogs].slice(0, 1000), // 保持最近1000条
        }))
      },

      clearDebugLogs: () => {
        set({ debugLogs: [] })
        apiClient.clearDebugLogs()
      },

      initialize: () => {
        const state = get()
        // 初始化API客户端配置
        if (state.tenantInfo) {
          apiClient.setTenant(state.tenantInfo)
        }
      },
    }),
    {
      name: 'agent-settings-storage',
      partialize: state => ({
        apiSettings: state.apiSettings,
        tenantInfo: state.tenantInfo,
      }),
    }
  )
)
