## AI-Nexus 2.0 MVP 方案

### 项目愿景

一个 Go 语言的 Agent 开发框架，让开发者用最少的代码把普通服务变成 AI Agent，支持 Agent 之间的发现与协作。

---

## 核心架构

```mermaid
graph TB
    subgraph Cluster["AI-Nexus 2.0"]
        subgraph Agents["Agent 层"]
            A1["Agent A - 用户服务"]
            A2["Agent B - 订单服务"]
            A3["Agent C - 支付服务"]
        end
        
        subgraph Infra["基础设施"]
            etcd[("etcd 注册中心")]
            sqlite[("SQLite 本地存储")]
        end
        
        A1 <-->|"A2A Protocol"| A2
        A2 <-->|"A2A Protocol"| A3
        
        A1 -->|"注册/发现"| etcd
        A2 -->|"注册/发现"| etcd
        A3 -->|"注册/发现"| etcd
    end
```

### 两种调用路径（HTTP vs AI）

```mermaid
graph TB
    subgraph External["外部调用（传统 HTTP）"]
        Client["HTTP Client"]
        GIN["Gin Router"]
        Handler["Handler"]
    end
    
    subgraph AICall["AI 调用（内部反射）"]
        Agent["Agent / LLM"]
        Tools["Tool Registry"]
        Executor["反射执行器"]
    end
    
    subgraph Business["业务代码（共享）"]
        Service["UserService"]
    end
    
    Client -->|"GET /api/users/1"| GIN
    GIN --> Handler
    Handler --> Service
    
    Agent -->|"调用 GetUser"| Tools
    Tools --> Executor
    Executor -->|"反射调用"| Service
```

**关键点**：
- **Gin 是 Gin** - 对外的 HTTP 接口，正常的 REST API
- **AI 调用是 AI 调用** - 内部通过反射直接调用 Service 方法
- **Service 是共享的** - 两边都调用同一个业务代码

### 单个 Agent 内部结构

```mermaid
graph TB
    subgraph AgentInternal["Agent 内部"]
        subgraph A2AEndpoints["A2A 端点（独立端口）"]
            CARD["/.well-known/agent.json"]
            TASK["/a2a/tasks"]
        end
        
        subgraph AICore["AI 核心"]
            LLM["LLM - OpenAI 兼容接口"]
        end
        
        subgraph ToolsLayer["Tools 层 - 两级结构"]
            CAP["Capability 清单 - 能力分类"]
            TOOLS["Tool Registry - 具体工具"]
            EXEC["反射执行器"]
        end
        
        subgraph BizCode["业务代码（共享）"]
            SVC["开发者的 Service"]
        end
        
        subgraph HTTPLayer["HTTP 层（Gin，独立端口）"]
            GIN["Gin Router"]
            HANDLER["Handler"]
        end
        
        TASK --> LLM
        LLM -->|"1.选择能力"| CAP
        CAP -->|"2.加载工具"| TOOLS
        LLM -->|"3.调用工具"| TOOLS
        TOOLS --> EXEC
        EXEC --> SVC
        
        GIN --> HANDLER
        HANDLER --> SVC
    end
```

---

## 懒加载 Tools 机制

AI 一开始只知道"能力清单"，按需加载具体 Tools，节省 Token 并提高灵活性。

```mermaid
sequenceDiagram
    participant User as 用户
    participant Agent as Agent
    participant LLM as LLM
    participant Registry as Tool Registry

    Note over Agent,Registry: 启动时只加载能力清单
    Agent->>Registry: 获取能力清单
    Registry-->>Agent: 用户管理/订单处理/文件操作

    User->>Agent: 帮我查一下用户张三的信息
    Agent->>LLM: 用户请求 + 能力清单
    LLM-->>Agent: 需要 用户管理 能力
    
    Note over Agent,Registry: 按需加载具体 Tools
    Agent->>Registry: 获取 用户管理 的 Tools
    Registry-->>Agent: GetUser/CreateUser/UpdateUser
    
    Agent->>LLM: 用户请求 + 具体 Tools
    LLM-->>Agent: 调用 GetUser name=张三
    Agent->>Agent: 反射执行 Tool
    Agent-->>User: 返回结果
```

### 两级结构设计

```mermaid
graph TB
    subgraph Level1["第一级：能力清单 Capability"]
        C1["用户管理 - 处理用户相关操作"]
        C2["订单处理 - 处理订单相关操作"]
        C3["文件操作 - 文件读写操作"]
    end
    
    subgraph Level2["第二级：具体 Tools"]
        subgraph UserTools["用户管理 Tools"]
            T1["GetUser"]
            T2["CreateUser"]
            T3["UpdateUser"]
        end
        
        subgraph OrderTools["订单处理 Tools"]
            T4["CreateOrder"]
            T5["GetOrder"]
        end
    end
    
    C1 -->|"按需加载"| T1
    C1 -->|"按需加载"| T2
    C1 -->|"按需加载"| T3
    C2 -->|"按需加载"| T4
    C2 -->|"按需加载"| T5
```

---

## 服务发现流程

```mermaid
sequenceDiagram
    participant Agent as Agent 实例
    participant etcd as etcd
    participant Other as 其他 Agent

    Note over Agent: 启动
    Agent->>Agent: 反射扫描生成 Tools
    Agent->>Agent: 构建 AgentCard
    Agent->>etcd: 注册 AgentCard

    Note over Other: 需要调用
    Other->>etcd: 查询可用 Agent
    etcd-->>Other: 返回 AgentCard 列表
    Other->>Agent: POST /a2a/tasks
    Agent-->>Other: 返回结果
```

---

## Tool 描述方式（混合模式）

```go
// 用结构体 + Tag 定义输入参数
type GetUserInput struct {
    UserID int64  `json:"user_id" desc:"用户的唯一标识"`
    Name   string `json:"name" desc:"用户姓名，可选"`
}

type UserService struct{}

func (s *UserService) GetUser(input GetUserInput) (*User, error) {
    return &User{ID: input.UserID}, nil
}

// 实现接口提供方法描述
func (s *UserService) ToolDescriptions() map[string]string {
    return map[string]string{
        "GetUser":    "根据用户ID或姓名获取用户详细信息",
        "CreateUser": "创建新用户账号",
    }
}
```

---

## YAML 配置文件

元信息通过 YAML 配置，无需重新编译即可修改：

```yaml
# config.yaml
agent:
  name: "user-service"
  description: "用户管理服务，提供用户增删改查功能"
  version: "1.0.0"
  
server:
  http_port: 8080    # Gin HTTP 端口
  a2a_port: 8081     # A2A 协议端口
  
llm:
  base_url: "https://api.openai.com/v1"
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4o-mini"
  
registry:
  etcd_endpoints:
    - "localhost:2379"
    
capabilities:
  - name: "user-management"
    description: "用户管理，包括用户的增删改查操作"
  - name: "order-management"
    description: "订单管理，包括订单的创建和查询"
```

---

## MVP 目录结构

```
ai-nexus/
├── cmd/
│   └── example/              # 示例入口
├── pkg/
│   ├── agent/                # Agent 核心
│   │   ├── agent.go          # Agent 结构
│   │   ├── options.go        # 配置选项
│   │   ├── config.go         # YAML 配置加载
│   │   └── card.go           # AgentCard
│   ├── tools/                # 反射工具层
│   │   ├── capability.go     # 能力定义（第一级）
│   │   ├── scanner.go        # 反射扫描
│   │   ├── schema.go         # JSON Schema 生成
│   │   └── executor.go       # 执行器
│   ├── a2a/                  # A2A 协议
│   │   ├── server.go         # HTTP 端点
│   │   ├── client.go         # 调用其他 Agent
│   │   └── types.go          # 类型定义
│   ├── llm/                  # LLM 层
│   │   └── openai.go         # OpenAI 兼容接口
│   └── registry/             # 注册中心
│       └── etcd.go           # etcd 实现
└── examples/                 # 示例（Gin + Mock 数据）
    ├── user-service/         # 用户服务示例
    ├── order-service/        # 订单服务示例
    └── product-service/      # 商品服务示例
```

---

## 示例服务（Gin + Mock 数据）

### 1. 用户服务 (user-service)

```go
// examples/user-service/service.go
type UserService struct{}

type GetUserInput struct {
    UserID int64 `json:"user_id" desc:"用户ID"`
}

func (s *UserService) GetUser(input GetUserInput) (*User, error) {
    return &User{
        ID:    input.UserID,
        Name:  "张三",
        Email: "zhangsan@example.com",
        Phone: "13800138000",
    }, nil
}

func (s *UserService) ListUsers(input ListUsersInput) ([]*User, error) {
    return []*User{
        {ID: 1, Name: "张三"},
        {ID: 2, Name: "李四"},
        {ID: 3, Name: "王五"},
    }, nil
}

func (s *UserService) ToolDescriptions() map[string]string {
    return map[string]string{
        "GetUser":   "根据用户ID获取用户详细信息",
        "ListUsers": "获取用户列表",
    }
}
```

### 2. 订单服务 (order-service)

```go
// examples/order-service/service.go
type OrderService struct{}

type CreateOrderInput struct {
    UserID    int64 `json:"user_id" desc:"用户ID"`
    ProductID int64 `json:"product_id" desc:"商品ID"`
    Quantity  int   `json:"quantity" desc:"购买数量"`
}

func (s *OrderService) CreateOrder(input CreateOrderInput) (*Order, error) {
    return &Order{
        ID:        10001,
        UserID:    input.UserID,
        ProductID: input.ProductID,
        Quantity:  input.Quantity,
        Amount:    99.99,
        Status:    "pending",
        CreatedAt: time.Now(),
    }, nil
}

func (s *OrderService) GetOrder(input GetOrderInput) (*Order, error) {
    return &Order{
        ID:     input.OrderID,
        Status: "completed",
        Amount: 199.99,
    }, nil
}

func (s *OrderService) ToolDescriptions() map[string]string {
    return map[string]string{
        "CreateOrder": "创建新订单",
        "GetOrder":    "查询订单详情",
    }
}
```

### 3. 商品服务 (product-service)

```go
// examples/product-service/service.go
type ProductService struct{}

func (s *ProductService) GetProduct(input GetProductInput) (*Product, error) {
    return &Product{
        ID:          input.ProductID,
        Name:        "iPhone 15 Pro",
        Price:       7999.00,
        Stock:       100,
        Description: "Apple 最新款智能手机",
    }, nil
}

func (s *ProductService) SearchProducts(input SearchInput) ([]*Product, error) {
    return []*Product{
        {ID: 1, Name: "iPhone 15 Pro", Price: 7999.00},
        {ID: 2, Name: "MacBook Pro", Price: 14999.00},
        {ID: 3, Name: "AirPods Pro", Price: 1899.00},
    }, nil
}

func (s *ProductService) ToolDescriptions() map[string]string {
    return map[string]string{
        "GetProduct":     "获取商品详情",
        "SearchProducts": "搜索商品",
    }
}
```

### 示例启动方式

```go
// examples/user-service/main.go
func main() {
    // 共享的业务服务
    userService := &UserService{}
    
    // 1. Gin 路由（对外 HTTP，端口 8080）
    r := gin.Default()
    r.GET("/api/users/:id", func(c *gin.Context) {
        id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
        user, _ := userService.GetUser(GetUserInput{UserID: id})
        c.JSON(200, user)
    })
    r.GET("/api/users", func(c *gin.Context) {
        users, _ := userService.ListUsers(ListUsersInput{})
        c.JSON(200, users)
    })
    
    // 2. Agent（AI 调用，端口 8081）
    ag := agent.NewFromConfig("./config.yaml")
    ag.RegisterCapability("user-management", "用户管理")
    ag.RegisterService("user-management", userService)
    
    // 并行启动两个服务
    go r.Run(":8080")  // HTTP API
    ag.Run(":8081")    // A2A 端点
}
```

---

## MVP 实施步骤

### 第一步：反射工具层
- `pkg/tools/capability.go` - 能力定义
- `pkg/tools/scanner.go` - 扫描 struct 方法 + Tag
- `pkg/tools/schema.go` - 生成 JSON Schema
- `pkg/tools/executor.go` - 反射调用方法

### 第二步：LLM 集成
- `pkg/llm/openai.go` - OpenAI 兼容接口
- 两阶段调用：选能力 → 选工具

### 第三步：Agent 核心
- `pkg/agent/agent.go` - Agent 生命周期
- `pkg/agent/config.go` - YAML 配置加载
- `pkg/agent/card.go` - AgentCard 生成

### 第四步：A2A 协议
- `pkg/a2a/server.go` - HTTP 端点
- `pkg/a2a/client.go` - 调用其他 Agent
- `pkg/a2a/types.go` - Task/Message 类型

### 第五步：注册中心
- `pkg/registry/etcd.go` - etcd 注册/发现

### 第六步：示例服务
- `examples/user-service/` - 用户服务（Gin + Mock）
- `examples/order-service/` - 订单服务（Gin + Mock）
- `examples/product-service/` - 商品服务（Gin + Mock）

---

## 技术选型

| 组件 | 选择 |
|------|------|
| 语言 | Go |
| HTTP 框架 | Gin（业务 API） |
| 注册中心 | etcd |
| 数据库 | SQLite (MVP) |
| LLM | OpenAI 兼容接口 |
| 配置 | YAML |

---

## 核心价值

1. **零配置** - 写普通 Go 代码 + Tag 即可
2. **懒加载** - 按需加载 Tools，节省 Token
3. **热配置** - YAML 配置，无需重新编译
4. **双端口分离** - HTTP API 和 A2A 端点独立运行
5. **标准协议** - A2A 协议，可与其他生态互通
6. **简单部署** - etcd + SQLite，开箱即用