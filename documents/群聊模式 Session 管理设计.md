## 核心设计原则

**Agent 无状态 + 最小上下文**：
- 每个 Agent 只收到单一任务指令 + 必要变量
- Orchestrator 负责保存完整群聊历史（仅用于 UI）
- 避免上下文爆炸

---

## 改动内容

### 1. 修改 Session 结构 (`pkg/session/store.go`)

```go
type GroupMessage struct {
    ID          string           `json:"id"`
    From        string           `json:"from"`
    To          string           `json:"to"`
    Content     string           `json:"content"`
    ToolCalls   []ToolCallInfo   `json:"tool_calls,omitempty"`
    ToolResults []ToolResultInfo `json:"tool_results,omitempty"`
    Timestamp   time.Time        `json:"timestamp"`
}

type Session struct {
    ID            string         `json:"id"`
    GroupMessages []GroupMessage `json:"group_messages"` // 群聊消息（用于UI）
    Participants  []string       `json:"participants"`   // 参与者列表
    Context       map[string]interface{} `json:"context"`
    CreatedAt     time.Time      `json:"created_at"`
    UpdatedAt     time.Time      `json:"updated_at"`
}
```

### 2. 修改 A2A 请求 (`pkg/a2a/types.go`)

```go
type CreateTaskRequest struct {
    SessionID     string                 `json:"session_id,omitempty"`
    TraceID       string                 `json:"trace_id,omitempty"`
    From          string                 `json:"from,omitempty"`  // 新增：谁在@它
    Message       string                 `json:"message"`
    SharedContext map[string]interface{} `json:"shared_context,omitempty"`
}
```

### 3. 修改 Orchestrator (`examples/orchestrator/main.go`)

- 调用 Agent 时传递 `From` 字段
- 保存 `GroupMessage` 到 Session
- 更新 `Participants` 列表

### 4. 修改前端 (`examples/orchestrator/static/index.html`)

- Session 列表显示参与者
- 加载历史时渲染 `GroupMessages`

---

## Agent 收到的上下文示例

**user-service 收到**：
```json
{
  "from": "@orchestrator",
  "message": "查询用户名为张三的用户信息",
  "shared_context": {}
}
```

**order-service 收到**：
```json
{
  "from": "@orchestrator",
  "message": "查询用户ID为123的订单",
  "shared_context": {"user_id": 123}
}
```

---

## 文件改动清单

| 文件 | 改动 |
|------|------|
| `pkg/session/store.go` | 新增 GroupMessage、Participants，移除旧 Messages |
| `pkg/a2a/types.go` | CreateTaskRequest 新增 From 字段 |
| `examples/orchestrator/main.go` | 保存群聊消息、传递 From、更新参与者 |
| `examples/orchestrator/static/index.html` | Session 列表和详情展示群聊格式 |
