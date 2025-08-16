# React 前端规划与实施指南（本地部署版）

本指南面向对本仓库零上下文的新同事，严格对齐后端协议与当前约束，保证无信息差。目标是在本地完成前后端联调与使用，不需要任何公网穿透。

- 应用目标：内部工作台（不追求 SEO）
- 部署范围：仅本地
- 后端端口：`8081`（见 [config.example.yaml](cci:7://file:///home/kiz/Code/agent_http_server/config.example.yaml:0:0-0:0) → `server.port: 8081`）
- 关键接口与协议见 [internal/handler/handler.go](cci:7://file:///home/kiz/Code/agent_http_server/internal/handler/handler.go:0:0-0:0) 与 [internal/service/workflow.go](cci:7://file:///home/kiz/Code/agent_http_server/internal/service/workflow.go:0:0-0:0)

---

## 一、后端接口与数据契约

- 列表：`GET /api/workflows`
  - 响应：
    - `workflows: string[]`
    - `count: number`
- 详情：`GET /api/workflows/{name}`
  - 响应（[WorkflowInfo](cci:2://file:///home/kiz/Code/agent_http_server/internal/service/workflow.go:29:0-34:1)，来源 [internal/service/workflow.go](cci:7://file:///home/kiz/Code/agent_http_server/internal/service/workflow.go:0:0-0:0)）：
    - `name: string`
    - `description: string`
    - `version: string`
    - `status: string`
- 非流式执行：`POST /api/execute`
  - 请求（[WorkflowRequest](cci:2://file:///home/kiz/Code/agent_http_server/internal/service/workflow.go:13:0-19:1)）：
    - `workflow: string` 必填
    - `input: string` 必填
    - `timeout?: number` 可选（单位秒）
    - 兼容期（如需）：`user_id?: string`, `archive_id?: string`
  - 响应（[WorkflowResponse](cci:2://file:///home/kiz/Code/agent_http_server/internal/service/workflow.go:22:0-26:1)）：
    - 成功：`{ status: "success", result: string }`
    - 失败：`{ status: "error", error: string }`
- 流式执行（SSE）：`POST /api/stream`
  - 请求体同 [WorkflowRequest](cci:2://file:///home/kiz/Code/agent_http_server/internal/service/workflow.go:13:0-19:1)
  - 返回为 Server-Sent Events 流，事件格式：
    - 多次：`event: data` + `data: <string>`
    - 结束：`event: done` + `data: <string>`
    - 错误：`event: error` + `data: {"error":"<message>"}`（注意是 JSON 字符串）

多租户请求头（前端需注入）：

- `X-User-ID: <string>`
- `X-Archive-ID: <string>`

说明：

- 浏览器原生 `EventSource` 不支持 POST 且不能自定义 Header，不适合此后端。应使用 `fetch + ReadableStream` 手动解析 SSE。

---

## 二、技术选型（✅ 已实现）

- ✅ 构建：Vite + React 19 + TypeScript（CSR）
- ✅ UI 组件库：Ant Design 5.x
- ✅ 路由：React Router DOM 6.x
- ✅ 状态：TanStack Query（服务端状态）+ Zustand（本地轻状态）
- ✅ 流式：`fetch + ReadableStream` 手动解析 SSE（POST + 自定义 Header）
- ✅ 语言：中文
- ✅ 质量：ESLint + Prettier + Husky（pre-commit）
- ✅ 部署形态：前后端分离，均在本地运行（后端 8081）

---

## 三、目录结构（✅ 已实现）

```
frontend/
├── src/
│   ├── api/client.ts          # ✅ API客户端（支持SSE、多租户、调试日志）
│   ├── types/api.ts           # ✅ TypeScript类型定义
│   ├── stores/settings.ts     # ✅ Zustand设置存储
│   ├── components/            # ✅ React组件
│   │   ├── SettingsPage.tsx   # ✅ API & 租户配置页
│   │   ├── WorkflowList.tsx   # ✅ 工作流列表页
│   │   ├── WorkflowDetail.tsx # ✅ 工作流详情页
│   │   ├── WorkflowExecute.tsx # ✅ 同步执行页
│   │   ├── WorkflowStream.tsx  # ✅ SSE流式执行页
│   │   └── DebugLogs.tsx      # ✅ API调试日志页
│   ├── App.tsx               # ✅ 主应用（布局+路由）
│   └── main.tsx              # ✅ React入口
├── public/                   # ✅ 静态资源
├── .prettierrc              # ✅ 代码格式化配置
├── .prettierignore          # ✅ 格式化忽略规则
├── vite.config.ts           # ✅ Vite配置（含代理）
├── package.json             # ✅ 依赖管理
└── tsconfig.json            # ✅ TypeScript配置
```

---

## 四、环境变量与本地代理（✅ 已实现）

- ✅ 环境变量：
  - `VITE_API_BASE=http://localhost:8081`（默认值，可选配置）
- ✅ 本地开发代理（`vite.config.ts`）：
  - `/api/*` → `http://localhost:8081`
  - 自动处理 `text/event-stream`（SSE流式传输）

---

## 五、数据与状态边界（✅ 已实现）

✅ **TypeScript类型**（`src/types/api.ts`）：
- `WorkflowInfo`: `{ name, description, version, status }`
- `WorkflowListResp`: `{ workflows: string[], count: number }`
- `WorkflowRequest`: `{ workflow: string; input: string; timeout?: number; user_id?, archive_id? }`
- `WorkflowResponse`: `{ status: "success" | "error"; result?: string; error?: string }`
- `SSEEvent`: `{ type: "data" | "done" | "error"; payload: string | { error: string } }`
- `TenantInfo`: `{ userId: string; archiveId: string }`
- `DebugLogEntry`: 调试日志条目

✅ **状态管理**：
- ✅ 服务器状态：TanStack Query（API缓存、重试、错误处理）
- ✅ 本地状态：Zustand（设置、租户信息持久化）

---

## 六、请求层设计（✅ 已实现）

✅ **API客户端**（`src/api/client.ts`）：
- ✅ 自动读取 `VITE_API_BASE` 或默认 `localhost:8081`
- ✅ 自动注入租户头：`X-User-ID`、`X-Archive-ID`
- ✅ 统一错误处理：HTTP状态码 + 业务错误
- ✅ 调试日志：自动记录所有请求/响应/SSE事件

✅ **API方法**：
- ✅ `getWorkflows()`: `GET /api/workflows`
- ✅ `getWorkflowInfo(name)`: `GET /api/workflows/{name}`
- ✅ `executeWorkflow(request)`: `POST /api/execute`
- ✅ `streamWorkflow(request)`: `POST /api/stream`（SSE异步生成器）

✅ **兼容性支持**：
- ✅ 同时在Header和Body中发送租户信息（过渡期兼容）

---

## 七、SSE 解析实现（✅ 已完成）

✅ **核心实现**（`src/api/client.ts:streamWorkflow`）：
1. ✅ 使用 `fetch` + `AbortController` 支持中断
2. ✅ `TextDecoder` + `ReadableStream` 逐块解码
3. ✅ 按行解析SSE格式：`event:` / `data:` 
4. ✅ 异步生成器模式：`async *streamWorkflow()`

✅ **事件处理**：
- ✅ `data`：实时文本流
- ✅ `done`：标记完成
- ✅ `error`：JSON错误解析
- ✅ 自动调试日志记录

✅ **UI集成**：
- ✅ "停止"按钮（`AbortController.abort()`）
- ✅ 实时输出渲染
- ✅ 错误状态显示

---

## 八、多租户策略（✅ 已实现）

✅ **当前实现**：
- ✅ Header注入：`X-User-ID` + `X-Archive-ID`
- ✅ Body兼容：同时写入 `user_id` + `archive_id`（过渡期）
- ✅ 设置持久化：localStorage存储租户信息

✅ **前端功能**：
- ✅ 设置页：租户配置界面
- ✅ 自动注入：所有API请求自动带上租户信息
- ✅ 状态管理：Zustand管理租户状态

🔄 **迁移路径**：
- [未完成] 后端Header中间件完善后，移除Body中的租户字段

---

## 九、MVP 页面规格（✅ 全部完成）

- ✅ **A. 工作流列表页**（`WorkflowList.tsx`）
  - ✅ `GET /api/workflows` + 表格展示
  - ✅ 搜索过滤 + 刷新按钮
  - ✅ 点击跳转详情页
  - ✅ 错误重试 + 调试日志

- ✅ **B. 工作流详情页**（`WorkflowDetail.tsx`）
  - ✅ `GET /api/workflows/{name}` 显示详细信息
  - ✅ 名称/描述/版本/状态展示
  - ✅ 执行按钮快速跳转

- ✅ **C. 非流式执行页**（`WorkflowExecute.tsx`）
  - ✅ 表单：工作流选择、输入内容、超时设置
  - ✅ `POST /api/execute` 执行
  - ✅ 结果/错误展示（Ant Design组件）
  - ✅ 完整调试日志

- ✅ **D. 流式执行页**（`WorkflowStream.tsx`）
  - ✅ `POST /api/stream` SSE流式执行
  - ✅ 实时文本输出渲染
  - ✅ "停止"按钮（AbortController）
  - ✅ 错误事件处理 + 状态显示

- ✅ **E. 设置页**（`SettingsPage.tsx`）
  - ✅ API Base URL配置
  - ✅ UserID + ArchiveID租户设置
  - ✅ localStorage持久化
  - ✅ 表单验证 + 保存提示

- ✅ **F. 调试日志面板**（`DebugLogs.tsx`）
  - ✅ 时间、方向、路径、状态码/事件类型
  - ✅ 请求/响应/SSE事件详情展示
  - ✅ 过滤器 + 清空 + JSON导出

---

## 十、本地开发与联调（✅ 就绪）

✅ **快速启动**：
```bash
# 1. 启动后端
./agent_http_server -config config.yaml

# 2. 启动前端
cd frontend && npm run dev
# 访问: http://localhost:5173
```

✅ **验证步骤**：
1. ✅ 后端健康检查：`GET localhost:8081/health`
2. ✅ 前端访问：`http://localhost:5173`
3. ✅ 设置页配置：API Base + UserID + ArchiveID
4. ✅ 功能验证：列表 → 详情 → 执行 → 流式 → 调试日志

✅ **一键部署脚本**：
```bash
# 安装依赖 + 构建
./scripts/deploy_frontend.sh

# 构建 + 启动开发服务器
./scripts/deploy_frontend.sh --run

# 仅启动开发服务器
./scripts/deploy_frontend.sh --dev
```

✅ **故障排除**：
- ✅ Vite代理：自动处理 `/api` → `:8081`
- ✅ SSE支持：原生处理 `text/event-stream`
- ✅ 调试工具：内置调试日志面板

---

## 十一、部署方式（✅ 多种选择）

✅ **方式一：开发模式**（推荐）
```bash
./scripts/deploy_frontend.sh --dev
# 访问: http://localhost:5173
```

✅ **方式二：生产预览**
```bash
./scripts/deploy_frontend.sh      # 构建
cd frontend && npm run preview    # 预览
# 访问: http://localhost:4173
```

✅ **方式三：静态服务器**
```bash
./scripts/deploy_frontend.sh     # 构建
cd frontend/dist
python -m http.server 8080       # 或任意静态服务器
# 需要后端CORS支持
```

✅ **特点**：
- ✅ 完全本地部署，无需公网
- ✅ 前后端分离，独立端口
- ✅ 支持SPA路由（History API）

---

## 十二、里程碑进度（✅ 全部完成）

- ✅ **M1**：脚手架 + Vite代理 + 设置页 + 租户store + API封装
- ✅ **M2**：工作流列表页 + 详情页
- ✅ **M3**：非流式执行页
- ✅ **M4**：流式执行页（SSE + 停止 + 错误处理）
- ✅ **M5**：调试日志面板
- ✅ **M6**：文档完善 + 部署脚本 + 本地部署确认

---

## 十三、部署脚本使用指南

✅ **一键部署脚本**：`./scripts/deploy_frontend.sh`

### 🚀 使用方法

**基础构建**：
```bash
./scripts/deploy_frontend.sh
```
- 自动检测Node.js/npm环境
- 智能选择 `npm ci` 或 `npm install`
- 构建生产版本到 `frontend/dist/`

**完整开发环境**：
```bash
./scripts/deploy_frontend.sh --run
```
- 执行构建 + 启动开发服务器
- 适合全新环境快速上手

**仅启动开发**：
```bash
./scripts/deploy_frontend.sh --dev
```
- 跳过构建，直接启动 `npm run dev`
- 适合已构建过的开发调试

**查看帮助**：
```bash
./scripts/deploy_frontend.sh --help
```

### ✨ 脚本特性

- ✅ **环境检测**：自动验证Node.js和npm版本
- ✅ **智能安装**：检测package-lock.json自动选择安装方式
- ✅ **错误处理**：详细的错误提示和解决建议
- ✅ **彩色输出**：友好的进度显示和状态提示
- ✅ **无脑操作**：零配置，开箱即用

---

## 十四、当前状态总结

✅ **完成功能**：
- ✅ 完整的React前端应用（6个页面全部实现）
- ✅ SSE流式处理 + 多租户支持
- ✅ 完善的调试工具和错误处理
- ✅ 一键部署脚本和文档
- ✅ 本地开发环境就绪

🔄 **待完善**：
- [未完成] 后端Header中间件完善后，移除API Body中的租户字段
- [未完成] 生产环境部署配置（当前仅支持本地）

📋 **技术债务**：
- 无重大技术债务，代码质量良好
- 遵循最佳实践，具备良好的可维护性

---
