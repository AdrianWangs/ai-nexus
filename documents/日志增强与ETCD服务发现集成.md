## 改进计划

### 1. 日志增强 - 添加 TraceID 和 SessionID

**修改文件：** `pkg/agent/agent.go`, `examples/orchestrator/main.go`

日志格式统一为：
```
[agent-name] [trace:xxx] [session:xxx] 日志内容
```

---

### 2. ETCD 服务发现（Watch 监听模式）

#### 2.1 修改 `pkg/agent/agent.go`

```go
type Agent struct {
    // 现有字段...
    registry    registry.Registry      // ETCD 注册中心
    a2aClient   *a2a.Client            // A2A 客户端
    agentCache  map[string]*AgentCard  // 已发现的 Agent 缓存
    cacheMu     sync.RWMutex           // 缓存锁
}

// 初始化注册中心
func (a *Agent) initRegistry()

// 注册自己到 ETCD
func (a *Agent) registerToEtcd()

// 启动 Watch 监听其他 Agent
func (a *Agent) watchAgents()

// 获取指定 Agent
func (a *Agent) GetAgent(name string) *AgentCard

// 列出所有其他 Agent
func (a *Agent) ListAgents() []*AgentCard

// 调用其他 Agent
func (a *Agent) CallAgent(ctx, agentName, message, sessionID, traceID string) (*a2a.CreateTaskResponse, error)
```

#### 2.2 启动流程

```
Agent.Run()
  ├── 1. 连接 ETCD
  ├── 2. 注册自己（带 30s TTL + KeepAlive）
  ├── 3. 加载所有已注册的 Agent 到缓存
  ├── 4. 启动 Watch 监听（goroutine）
  │       ├── add/update → 更新缓存
  │       └── delete → 从缓存移除
  └── 5. 启动 HTTP/A2A 服务
```

#### 2.3 修改 `pkg/agent/config.go`

```go
type RegistryConfig struct {
    Enabled       bool     `yaml:"enabled"`        // 是否启用
    EtcdEndpoints []string `yaml:"etcd_endpoints"`
}
```

---

### 3. Orchestrator 简化

`examples/orchestrator/main.go` 改为使用 Agent 的发现能力：
- 移除 `downstream_agents` 配置依赖
- 直接调用 `ListAgents()` 获取可用 Agent
- 保留配置文件作为 fallback（ETCD 不可用时）

---

### 4. 配置文件更新

各服务 `config.yaml`：
```yaml
registry:
  enabled: true
  etcd_endpoints:
    - "localhost:2379"
```

---

### 5. 改动文件清单

| 文件 | 改动 |
|------|------|
| `pkg/agent/agent.go` | 日志增强 + 添加 registry/cache/发现/调用方法 |
| `pkg/agent/config.go` | 添加 registry.enabled |
| `examples/orchestrator/main.go` | 日志增强 + 使用 Agent 发现能力 |
| `examples/user-service/config.yaml` | 添加 registry.enabled: true |
| `examples/order-service/config.yaml` | 添加 registry.enabled: true |
| `examples/product-service/config.yaml` | 添加 registry.enabled: true |
| `examples/orchestrator/config.yaml` | 添加 registry.enabled: true |

---

### 6. 预期效果

**日志示例：**
```
[user-service] [trace:abc123] [session:sess_001] Processing request
[user-service] [trace:abc123] [session:sess_001] Agent registered to ETCD
[user-service] [trace:abc123] [session:sess_001] Discovered agent: order-service
[user-service] [trace:abc123] [session:sess_001] Calling agent order-service
```

**Agent 互相发现：**
```
user-service 启动 → 注册 → Watch → 发现 order-service, product-service
order-service 启动 → 注册 → Watch → 发现 user-service, product-service
新 Agent 上线 → 所有 Agent 通过 Watch 自动感知
```