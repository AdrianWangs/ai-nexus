## 核心架构

```
                         ┌─────────────┐
                         │    ETCD     │
                         └──────┬──────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
 ┌──────────────┐       ┌──────────────┐       ┌──────────────┐
 │ user-service │       │ order-service│       │product-service│
 └──────────────┘       └──────────────┘       └──────────────┘
        ▲                       ▲                       ▲
        └───────────────────────┼───────────────────────┘
                                │
                         ┌──────┴──────┐
                         │ Orchestrator│
                         └─────────────┘
                                ▲
                         ┌──────┴──────┐
                         │   Web UI    │
                         └─────────────┘
```

***

## 第一阶段：创建独立的 Web UI 服务

**新建** **`examples/web-ui/`**

* `main.go` - 入口，连接 ETCD，发现 Agent

* `config.yaml` - 配置文件

* `handlers.go` - HTTP 处理器

* `group_session.go` - 群聊会话管理（含 Type 字段）

* `go.mod`

***

## 第二阶段：修改 Orchestrator

* 调用 Agent 时记录内部请求/响应消息

* 消息 Type 字段：`user`, `reply`, `internal`

* 移除 Web UI 相关代码

***

## 第三阶段：重构前端（组件化）

**`examples/web-ui/static/`**

* 状态卡片在顶部

* 内部协商消息可折叠

* 组件化：Layout, Sidebar, ChatPanel, StatusCard, MessageBubble, InternalMessages, Composer, AgentPanel 等

***

## 第四阶段：更新启动脚本

* 添加 web-ui 服务启动/停止

***

## 文件变更

### 新建

* `examples/web-ui/*` (\~8个文件)

* `examples/web-ui/static/*` (\~15个前端文件)

### 修改

* `pkg/protocol/types.go` - GroupMessage 增加 Type

* `pkg/orchestrator/orchestrator.go` - 记录内部消息

* `examples/orchestrator/main.go` - 移除 Web UI

* `scripts/dev-start.sh`, `scripts/dev-stop.sh`

### 删除

* `examples/orchestrator/static/` - 移到 web-ui

