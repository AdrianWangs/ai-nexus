## AI-Nexus 会话管理与架构增强

### 1. 框架层 Session 接口 (pkg/session)

```go
// pkg/session/store.go
type Store interface {
    Get(ctx context.Context, sessionID string) (*Session, error)
    Set(ctx context.Context, session *Session) error
    Delete(ctx context.Context, sessionID string) error
    AddMessage(ctx context.Context, sessionID string, msg Message) error
    List(ctx context.Context, prefix string) ([]*Session, error)  // 列出会话
}

type Session struct {
    ID        string                 `json:"id"`
    Messages  []Message              `json:"messages"`
    Context   map[string]interface{} `json:"context"`
    CreatedAt time.Time              `json:"created_at"`
    UpdatedAt time.Time              `json:"updated_at"`
}
```

### 2. 存储实现

| 文件 | 说明 |
|------|------|
| `pkg/session/store.go` | 接口定义 + Session 结构体 |
| `pkg/session/memory.go` | go-cache 实现（滑动过期 24h） |
| `pkg/session/redis.go` | Redis 实现（滑动过期 24h） |

### 3. Orchestrator Session HTTP 接口

```
GET    /sessions              # 列出所有会话
GET    /sessions/:id          # 获取会话详情（含历史消息）
DELETE /sessions/:id          # 删除会话
POST   /sessions              # 创建新会话（可选，chat 时自动创建）
```

### 4. Chat 接口增强

```
POST /chat
{
  "session_id": "可选，不传则创建新会话",
  "message": "用户消息",
  "stream": false
}

Response:
{
  "session_id": "sess_xxx",
  "success": true,
  "result": "查询完成，共找到5个用户...",
  "data": [...]
}
```

### 5. 完整改动清单

#### 新增文件
| 文件 | 说明 |
|------|------|
| `pkg/session/store.go` | Store 接口 + Session 结构体 |
| `pkg/session/memory.go` | go-cache 内存实现 |
| `pkg/session/redis.go` | Redis 实现 |

#### 修改文件
| 文件 | 改动 |
|------|------|
| `pkg/llm/interface.go` | 添加 ChatStream() 流式接口 |
| `pkg/llm/openai.go` | 流式响应 + 重试机制（指数退避） |
| `pkg/a2a/types.go` | 添加 session_id、trace_id |
| `pkg/agent/config.go` | 添加 session 配置 |
| `pkg/agent/agent.go` | 集成 Session、流式、结果总结 |
| `orchestrator/main.go` | Session 接口、多 Agent 协作、结果汇总 |
| `orchestrator/static/index.html` | 流式显示、会话管理 UI |

### 6. 配置

```yaml
session:
  store: "memory"     # memory 或 redis
  ttl: 86400          # 24小时
  sliding: true       # 滑动过期
  redis:              # store=redis 时生效
    addr: "localhost:6379"
    password: ""
    db: 0
```

### 7. 核心功能清单

- [x] Session 存储接口（memory/redis）
- [x] 滑动过期（24h，每次对话刷新）
- [x] Session HTTP 管理接口
- [x] 多轮对话支持
- [x] 流式响应（SSE）
- [x] 多 Agent 协作
- [x] LLM 结果总结（自然语言回复）
- [x] 重试机制（指数退避）
- [x] 链路追踪（trace_id）

### 8. 依赖

```
github.com/patrickmn/go-cache v2.1.0
github.com/redis/go-redis/v9 v9.0.0
```
