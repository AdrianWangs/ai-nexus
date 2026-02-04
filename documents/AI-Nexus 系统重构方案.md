## 重构目标

将混乱的架构重新整理，建立清晰的分层结构，让前端能够稳定、灵活地调用后端服务。

---

## 第一阶段：统一协议层 (pkg/protocol)

**问题**：`a2a/types.go` 和 `session/store.go` 类型重复，定义分散

**方案**：
1. 新建 `pkg/protocol/` 目录，统一所有通信类型定义
2. 定义清晰的消息结构：
   - `Message`: 基础消息
   - `GroupMessage`: 群聊消息（from, to, content）
   - `StreamEvent`: 流式事件
   - `TaskRequest/Response`: 任务请求响应
3. 删除 `a2a/types.go` 中的重复定义，改为引用 protocol

---

## 第二阶段：重构 Agent 基础框架 (pkg/agent)

**问题**：`agent.go` 700+ 行，职责过重

**方案**：
1. 拆分为多个文件：
   - `agent.go`: 核心结构和生命周期管理
   - `handler.go`: HTTP 请求处理
   - `processor.go`: 消息处理逻辑
   - `config.go`: 配置（保持不变）
2. 简化 Agent 接口，只保留核心能力：
   - `RegisterCapability`: 注册能力
   - `RegisterService`: 注册服务
   - `ProcessMessage`: 处理消息
   - `Run`: 启动服务

---

## 第三阶段：Orchestrator 模块化 (pkg/orchestrator)

**问题**：`examples/orchestrator/main.go` 1500 行，所有逻辑堆在一起

**方案**：
1. 将 Orchestrator 核心逻辑移到 `pkg/orchestrator/`
2. 拆分为独立模块：
   - `orchestrator.go`: 核心协调器
   - `planner.go`: 任务规划（LLM 生成执行计划）
   - `executor.go`: 任务执行（并行/串行）
   - `router.go`: 消息路由（判断发给谁）
   - `summarizer.go`: 结果汇总
3. `examples/orchestrator/main.go` 只保留启动代码和 HTTP 路由

---

## 第四阶段：会话管理优化 (pkg/session)

**问题**：Session 结构混乱，GroupMessage 和 Message 并存

**方案**：
1. 统一使用 `GroupMessage` 作为唯一消息格式
2. 简化 Session 结构：
   ```go
   Session {
       ID           string
       Messages     []GroupMessage  // 统一用群聊消息
       Participants []string
       Metadata     map[string]any
       CreatedAt    time.Time
       UpdatedAt    time.Time
   }
   ```
3. 删除冗余的 `Message` 类型

---

## 第五阶段：前端 API 契约定义

**问题**：前后端没有清晰的 API 契约

**方案**：
1. 定义清晰的 API 接口：
   - `GET /api/sessions`: 会话列表
   - `GET /api/sessions/:id`: 会话详情
   - `DELETE /api/sessions/:id`: 删除会话
   - `POST /api/chat`: 发送消息（SSE 流式）
   - `GET /api/agents`: Agent 列表
   - `POST /api/agents/refresh`: 刷新 Agent
2. 统一响应格式：
   ```json
   {
     "success": true,
     "data": {...},
     "error": null
   }
   ```
3. 统一流式事件格式：
   ```json
   {"type": "message", "data": {...}}
   {"type": "thinking", "content": "..."}
   {"type": "done"}
   ```

---

## 第六阶段：清理和测试

1. 删除未使用的代码（如 `a2a/server.go` 大部分）
2. 移除硬编码（`shouldCallAgents` 的关键词判断）
3. 添加单元测试
4. 更新 examples 下的服务

---

## 文件变更预览

```
pkg/
├── protocol/           # 新增：统一协议定义
│   ├── types.go        # 所有通信类型
│   ├── events.go       # 流式事件定义
│   └── errors.go       # 错误定义
│
├── agent/              # 重构：拆分职责
│   ├── agent.go        # 精简后的核心
│   ├── handler.go      # HTTP 处理
│   ├── processor.go    # 消息处理
│   ├── config.go       # 配置（保持）
│   ├── card.go         # Agent Card（保持）
│   └── options.go      # 选项（保持）
│
├── orchestrator/       # 新增：协调器模块
│   ├── orchestrator.go # 核心协调器
│   ├── planner.go      # 任务规划
│   ├── executor.go     # 任务执行
│   ├── router.go       # 消息路由
│   └── summarizer.go   # 结果汇总
│
├── session/            # 优化：简化结构
│   ├── store.go        # 接口定义（简化）
│   ├── memory.go       # 内存实现
│   ├── redis.go        # Redis 实现
│   └── context.go      # 共享上下文（重命名）
│
├── a2a/                # 简化：只保留客户端
│   ├── client.go       # A2A 客户端
│   └── types.go        # 引用 protocol
│
├── tools/              # 保持不变
├── llm/                # 保持不变
└── registry/           # 保持不变

examples/
├── orchestrator/
│   ├── main.go         # 精简后的启动代码
│   └── static/         # 前端文件
└── *-service/          # 保持不变
```

---

## 执行顺序

1. **先建后拆**：先创建新的 protocol 模块，再逐步迁移
2. **保持可运行**：每个阶段完成后确保系统可运行
3. **逐步替换**：旧代码标记 deprecated，新代码验证后再删除

是否开始执行这个重构方案？
