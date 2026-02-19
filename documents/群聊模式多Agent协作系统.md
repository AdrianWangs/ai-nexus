## 目标
将现有的"幕后编排"模式改为**群聊模式**，让用户能看到 Agent 之间的对话过程。

---

## 核心改动

### 1. 新增消息类型 (`pkg/a2a/types.go`)
```go
// 群聊消息
type GroupMessage struct {
    ID        string    `json:"id"`
    From      string    `json:"from"`       // 发送者 (user/@agent-name)
    To        string    `json:"to"`         // 接收者 (@agent-name/@user)
    Content   string    `json:"content"`    // 消息内容
    Mention   string    `json:"mention"`    // @谁
    ToolCalls []ToolCallInfo `json:"tool_calls,omitempty"`  // 工具调用信息
    ToolResults []ToolResultInfo `json:"tool_results,omitempty"` // 工具执行结果
    Timestamp time.Time `json:"timestamp"`
}

// 工具调用信息（展示用）
type ToolCallInfo struct {
    Name   string `json:"name"`
    Args   string `json:"args"`
}

// 工具执行结果（展示用）
type ToolResultInfo struct {
    Name    string `json:"name"`
    Success bool   `json:"success"`
    Result  string `json:"result"`
}
```

### 2. 新增 StreamEvent 类型 (`pkg/a2a/types.go`)
```go
const (
    EventTypeGroupMessage = "group_message"  // 群聊消息
    EventTypeMention      = "mention"        // @提及
)
```

### 3. 修改 Orchestrator (`examples/orchestrator/main.go`)

**改动点**：
- `ProcessRequestStream` 方法：当调用其他 Agent 时，发送 `group_message` 事件，展示 @Agent 的指令
- Agent 返回结果时，发送 `group_message` 事件，展示 @主Agent 的回复（包含工具调用信息）
- 新增 `/chat/direct` 端点：支持用户直接 @某个 Agent

**新增方法**：
```go
// 发送群聊消息事件
func (o *Orchestrator) sendGroupMessage(callback StreamCallback, from, to, content string, toolCalls []ToolCallInfo, toolResults []ToolResultInfo)

// 处理用户直接@某个Agent的请求
func (o *Orchestrator) ProcessDirectMention(ctx context.Context, req *DirectMentionRequest) (*ChatResponse, error)
```

### 4. 修改前端 (`examples/orchestrator/static/index.html`)

**改动点**：
- 消息展示改为群聊风格，显示 `@agent-name` 的提及
- 显示工具调用和结果信息
- 新增输入框支持 `@agent-name` 语法，用户可直接 @某个 Agent

---

## 消息流程示例

**用户发送**：`帮我查张三的订单`

**群聊展示**：
```
[用户] @orchestrator: 帮我查张三的订单

[orchestrator] @user-service: 查询用户名为张三的用户信息

[user-service] @orchestrator: 
  📞 调用工具: GetUserByName({"name": "张三"})
  ✅ 结果: {"id": 123, "name": "张三"}

[orchestrator] @order-service: 查询用户ID为123的所有订单

[order-service] @orchestrator:
  📞 调用工具: GetOrdersByUserID({"user_id": 123})
  ✅ 结果: [{"id": 1, "product": "手机"}, ...]

[orchestrator] @用户: 张三有3个订单：1. 手机...
```

**用户直接@Agent**：`@product-service 查询所有商品`
```
[用户] @product-service: 查询所有商品

[product-service] @用户:
  📞 调用工具: ListProducts({})
  ✅ 结果: 共有5个商品...
```

---

## 文件改动清单

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `pkg/a2a/types.go` | 修改 | 新增 GroupMessage 等类型 |
| `examples/orchestrator/main.go` | 修改 | 支持群聊消息流、直接@Agent |
| `examples/orchestrator/static/index.html` | 修改 | 群聊风格UI、@语法支持 |

---

## 不改动的部分

- `pkg/agent/agent.go` - Agent 核心逻辑不变
- `pkg/a2a/client.go` / `server.go` - A2A 通信协议不变
- 各 service 示例（user-service, order-service, product-service）- 不需要改动

---

## 实现顺序

1. **先改 types.go**：定义新的消息类型
2. **再改 orchestrator/main.go**：实现群聊消息流
3. **最后改前端**：展示群聊效果
