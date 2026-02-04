## 目标
- 删除现有的 [index.html](file:///Users/bytedance/GolandProjects/ai-nexus/examples/orchestrator/static/index.html)，完全重写一个更美观、交互更完整、BUG 更少的单页前端。
- 不使用浏览器自带的 alert/confirm，自己实现一套页面风格一致的模态框（确认/提示/错误）。
- 前端把后端现有接口都用上（不使用非流式对话接口 `POST /chat` 的非 stream 分支），重点用 `POST /chat/stream`。

## 我已确认的后端接口与流事件
- 页面托管：`GET /` -> `./static/index.html`，静态目录 `GET /static/*`（见 [main.go:L1299-L1303](file:///Users/bytedance/GolandProjects/ai-nexus/examples/orchestrator/main.go#L1299-L1303)）。
- Agents：`GET /agents`，`POST /agents/refresh`（见 [main.go:L1305-L1318](file:///Users/bytedance/GolandProjects/ai-nexus/examples/orchestrator/main.go#L1305-L1318)）。
- Sessions：`GET /sessions`，`GET /sessions/:id`，`DELETE /sessions/:id`（见 [main.go:L1320-L1346](file:///Users/bytedance/GolandProjects/ai-nexus/examples/orchestrator/main.go#L1320-L1346)）。
- Chat（流式）：`POST /chat/stream`（见 [main.go:L1382-L1403](file:///Users/bytedance/GolandProjects/ai-nexus/examples/orchestrator/main.go#L1382-L1403)）。
- SSE 事件：`thinking/content/agent_call/agent_result/summary/error/done`，并且结尾还会额外发 `data: [DONE]`（见 [types.go:L97-L116](file:///Users/bytedance/GolandProjects/ai-nexus/pkg/a2a/types.go#L97-L116)）。

## 新页面的功能与布局（不参考旧实现）
- **整体布局**：左侧会话列表（可折叠/移动端抽屉）+ 中间聊天区 + 顶部状态栏（连接状态/在线 agents 数量/刷新）。
- **会话管理**：
  - 加载会话列表（按更新时间排序）。
  - 点击进入会话（拉取并渲染消息）。
  - 新建对话（本地清空 UI，首次发送时由后端创建 session）。
  - 删除会话（自定义模态框确认 -> 调 `DELETE /sessions/:id`）。
- **聊天体验**：
  - 发送框自适应高度、Enter 发送 / Shift+Enter 换行。
  - 流式输出：正确处理 `thinking/agent_call/agent_result/content/summary/error/done/[DONE]`，展示“状态行 + 时间线/调用了哪些 agent + traceId”。
  - 支持“停止生成”（AbortController 断开 SSE 读取，UI 回到可输入状态）。
  - Markdown 渲染 + 代码高亮（沿用现有 CDN 依赖即可，也可以改成更轻的方案；实现时会选更稳的）。
- **Agents 面板**：
  - 展示在线 agents、capabilities、url、description。
  - “刷新服务”按钮：先 `POST /agents/refresh` 再 `GET /agents` 更新列表。
- **提示与错误**：
  - 统一 Toast（成功/失败/网络断开）。
  - 模态框用于危险操作确认（删除）、以及需要用户注意的错误详情。

## 自定义模态框（替代系统弹窗）
- 一个通用 Modal 组件：
  - `openConfirm({title, body, confirmText, cancelText, danger}) -> Promise<boolean>`
  - `openAlert({title, body})`
- 行为要求：遮罩点击关闭（可配置）、ESC 关闭、自动聚焦、简单的焦点限制（避免键盘操作跑到页面后面）。

## 关键实现点（为避免 BUG 会特别处理）
- SSE 解析更健壮：按行解析 `data:`，忽略空行；遇到 `[DONE]` 直接结束；遇到 `type:"done"` 也结束。
- 事件字段都是可选：所有 `session_id/trace_id/content/agent/status/error/data` 都要做空值保护。
- 并行 agent 返回顺序不固定：按 `event.agent` 聚合展示，不依赖顺序。

## 实施步骤（你确认后我会开始改代码）
1. 删除旧的 `examples/orchestrator/static/index.html`，新增全新的单文件实现（HTML+CSS+JS），不复用旧结构。
2. 实现 UI（布局、会话列表、聊天区、agents 面板、移动端适配）。
3. 实现自定义 Modal + Toast，并把“删除会话/错误提示”等全部切换过去。
4. 接入所有接口：`/sessions`、`/sessions/:id`、`DELETE /sessions/:id`、`/agents`、`POST /agents/refresh`、`POST /chat/stream`。
5. 本地验证：启动 orchestrator，逐项验证（新建/加载/删除会话、流式对话、agents 刷新、断网/服务端报错、停止生成）。

## 说明
- 项目里未找到 `openspec/AGENTS.md`（已搜索 `**/AGENTS.md`），所以这次按代码现状直接规划和实现。