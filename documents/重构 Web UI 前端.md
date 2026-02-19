## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                    Web UI (消息路由中台)                          │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐        │
│  │  群聊会话存储   │  │  Agent 发现    │  │  消息路由器    │        │
│  │  GroupStore   │  │  Discovery    │  │   Router      │        │
│  └───────────────┘  └───────────────┘  └───────────────┘        │
│           │                 │                  │                 │
│           └─────────────────┼──────────────────┘                 │
│                             │                                    │
│  ┌──────────────────────────┴──────────────────────────┐        │
│  │                    HTTP API                          │        │
│  │  /agents, /groups, /chat/stream, /a2a/callback      │        │
│  └──────────────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────────┘
         │                    │                    │
         ▼                    ▼                    ▼
    ┌─────────┐         ┌─────────┐         ┌─────────┐
    │Orchestrator│       │user-svc │         │order-svc│
    │  :9091   │         │  :8081  │         │  :8083  │
    └─────────┘         └─────────┘         └─────────┘
```

**核心理念：**
- Web UI 是所有消息的路由中心
- Agent 间调用通过 Web UI 转发
- 所有消息写入群聊，保持完整消息流
- 前端通过 SSE 实时接收消息更新

---

## 重构范围

### 1. 删除的文件
```
examples/web-ui/static/     # 全部删除重写
examples/web-ui/handlers.go # 删除重写
examples/web-ui/group_session.go # 删除重写
examples/web-ui/main.go     # 删除重写
```

### 2. 修改的模块
```
pkg/a2a/                    # 添加 callback 支持
pkg/orchestrator/           # 使用 callback 模式
pkg/agent/                  # 使用 callback 模式
examples/user-service/      # 适配新协议
examples/order-service/     # 适配新协议
examples/product-service/   # 适配新协议
examples/orchestrator/      # 适配新协议
```

---

## 阶段一：重构 Web UI 后端

### 文件结构
```
examples/web-ui/
├── config.yaml
├── go.mod
├── main.go              # 入口
├── server.go            # HTTP 服务器
├── discovery.go         # Agent 发现 (ETCD)
├── session.go           # 群聊会话管理
├── router.go            # 消息路由器
├── handlers/
│   ├── agents.go        # Agent API
│   ├── groups.go        # 群聊 API
│   ├── chat.go          # 聊天 API
│   └── callback.go      # A2A 回调 API
└── static/              # 前端文件
```

### API 设计
| 端点 | 方法 | 功能 |
|------|------|------|
| `/agents` | GET | 获取在线 Agent 列表 |
| `/groups` | GET | 获取群聊列表 |
| `/groups/:id` | GET | 获取群聊详情 |
| `/groups/:id` | DELETE | 删除群聊 |
| `/chat/stream` | POST | 发起聊天（流式） |
| `/a2a/callback` | POST | Agent 回调（写入消息） |
| `/a2a/route` | POST | Agent 间路由请求 |

### 消息路由流程
```
1. 用户发送 @orchestrator 你好
   │
   ▼
2. Web UI 创建群聊，添加 user→orchestrator 消息
   │
   ▼
3. Web UI 调用 Orchestrator，传入 callback_url
   │
   ▼
4. Orchestrator 处理，需要调用 user-service
   │
   ▼
5. Orchestrator POST 到 /a2a/route，请求路由到 user-service
   │
   ▼
6. Web UI 添加 orchestrator→user-service 消息，转发请求
   │
   ▼
7. user-service 返回结果，POST 到 /a2a/callback
   │
   ▼
8. Web UI 添加 user-service→orchestrator 消息，通知 Orchestrator
   │
   ▼
9. Orchestrator 完成，POST 到 /a2a/callback
   │
   ▼
10. Web UI 添加 orchestrator→user 消息，流式返回前端
```

---

## 阶段二：修改 A2A 协议

### 新增字段
```go
type TaskRequest struct {
    SessionID   string `json:"session_id"`
    TraceID     string `json:"trace_id"`
    GroupID     string `json:"group_id"`      // 新增：群聊ID
    CallbackURL string `json:"callback_url"`  // 新增：回调地址
    Message     string `json:"message"`
    From        string `json:"from"`
    Stream      bool   `json:"stream"`
}
```

### 回调消息格式
```go
type CallbackMessage struct {
    GroupID   string `json:"group_id"`
    TraceID   string `json:"trace_id"`
    From      string `json:"from"`
    To        string `json:"to"`
    Content   string `json:"content"`
    Type      string `json:"type"`  // user/reply/internal
    ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}
```

### 路由请求格式
```go
type RouteRequest struct {
    GroupID     string `json:"group_id"`
    TraceID     string `json:"trace_id"`
    From        string `json:"from"`
    To          string `json:"to"`      // 目标 Agent
    Message     string `json:"message"`
    CallbackURL string `json:"callback_url"`
}
```

---

## 阶段三：修改 Orchestrator

### 变更点
1. 收到请求时，使用 `callback_url` 而非直接 A2A
2. 调用 SubAgent 时，POST 到 `/a2a/route`
3. 所有消息通过 callback 写入群聊

### 代码修改
```
pkg/orchestrator/
├── orchestrator.go      # 添加 callback 模式
├── executor.go          # 使用路由调用 Agent
└── router.go            # 修改路由逻辑
```

---

## 阶段四：修改 Agent 服务

### 变更点
1. 收到请求时，保存 `callback_url`
2. 处理完成后，POST 结果到 `callback_url`
3. 流式响应同时发送到 callback

### 需要修改的服务
- `examples/user-service/`
- `examples/order-service/`
- `examples/product-service/`
- `examples/orchestrator/`

---

## 阶段五：重写前端

### 文件结构
```
static/
├── index.html
├── css/
│   └── main.css
└── js/
    ├── main.js          # 入口
    ├── api.js           # API 封装
    ├── state.js         # 状态管理
    └── ui/
        ├── sidebar.js   # 会话列表
        ├── chat.js      # 聊天区域
        ├── agents.js    # Agent 面板
        └── composer.js  # 输入框
```

### 功能
1. **会话管理** - 新建/切换/删除群聊
2. **消息显示** - 用户(蓝)/回复(灰)/内部(折叠)
3. **Agent 选择** - @ 提及任意 Agent
4. **实时更新** - SSE 接收群消息更新
5. **状态显示** - 处理进度卡片

---

## 执行顺序

1. **阶段一** - 重构 Web UI 后端（路由中台）
2. **阶段二** - 修改 A2A 协议（添加 callback）
3. **阶段三** - 修改 Orchestrator（使用 callback）
4. **阶段四** - 修改各 Agent 服务
5. **阶段五** - 重写前端
6. **测试** - 完整流程测试