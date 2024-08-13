# 项目架构图

## 总架构图


### **模块构成与工作流**

1. **用户界面与主大模型微服务**
    - **用户交互层**：用户通过API接口提交请求，包含自然语言表述的需求。
    - **核心入口——主大模型**：这是系统的第一道处理环节，负责理解与解析用户需求，确定最佳的服务路径。主大模型具备高级的语义理解和决策能力，决定哪些模块微服务最适合响应特定需求。

2. **模块微服务与次级大模型**
    - 每个微服务作为一个“能力模块”，入口配置了另一层大模型，此模型专注识别服务请求类型并映射到合适的内部服务或函数。
    - **需求细化**：次级大模型进一步分析细化用户的具体要求，明确应当调用哪个功能模块及其接口。
    - **调用参数构造**：确定调用需求后，该模型自动生成调用相应接口所需的参数集，确保准确传达用户的意图。

3. **函数接口大模型与服务执行**
    - 最深层次的模型专注于实现调用的细节，包括参数化和错误处理。
    - **判断与反馈**：基于前序分析，此模型验证需求是否符合其管理的接口功能，并作出响应。符合条件则动态生成调用参数；如遇无法处理的情况，则返回精确的错误信息。

4. **主大模型微服务的执行层**
    - 配备了双重调用机制：既能直接通过反射机制执行本地方法，又能通过RPC（远程过程调用）框架与其他微服务通信。
    - **服务调度**：基于分析结果及生成的参数，核心微服务高效地发起实际函数调用，完成用户请求的处理并返回结果。

```mermaid
graph TD
    subgraph 用户交互层
        UI[用户界面]
    end
    subgraph 核心服务层
        MainModel["主大模型微服务"]
        MainModel -->|多类型请求解析| ManyModels["次级大模型"]
        ManyModels -->|调用参数生成| FuncInterfaces["函数接口大模型"]
    end
    subgraph 执行&响应层
        ExecutionLayer["服务执行与反馈"]
        MainModel -->|直接调用或RPC| ExecutionLayer
    end
    subgraph 支持服务与基础设施
        Nacos[Nacos服务发现与配置中心]
        ThirdPartyMS["第三方插件微服务"]
        TaskMS["任务管理微服务"]
        EmailMS["邮件微服务"]
        KMMS["知识管理微服务"]
        AMMS["文章管理微服务"]
        CommunityMS["社区微服务"]
        PayMS["支付微服务"]
    end
    subgraph 鉴权层
        AuthMS["用户认证微服务"]
    end


    FuncInterfaces -- 参数构造 --> ExecutionLayer
    UI --> MainModel & ExecutionLayer
    UI -->|授权| AuthMS
    AuthMS --> UI
    ManyModels -- 能力细化 --> FuncInterfaces
    ExecutionLayer -.-> PayMS & ThirdPartyMS & TaskMS & EmailMS & KMMS & AMMS & CommunityMS
    Nacos -.-> AuthMS & MainModel & FuncInterfaces & ExecutionLayer & PayMS & ThirdPartyMS & TaskMS & EmailMS & KMMS & AMMS & CommunityMS
    ThirdPartyMS -.-> DemoPlugins
    ExecutionLayer -->|处理结果| UI

    style MainModel fill:#6FA6DE,color:#fff
    style ManyModels fill:#AED6F1,color:#000,border-color:#333
    style FuncInterfaces fill:#AED6F1,color:#000,border-color:#333
    style ExecutionLayer fill:#6FA6DE,color:#fff
```

## 功能架构

### ai 微服务

#### 用户视角

```mermaid
flowchart LR
    A[用户]
    subgraph 任务1
        direction LR
        BA[分析需求]
        BC[列出计划]
        BC[调用下个AI]
    end

    subgraph 任务2
        direction LR
        CB[总结文案]
        CC[调用函数]
        CD[调用下个AI]
    end
    subgraph 任务N
        direction LR
        DA[总结任务]
        
    end
    
    A -->|提交请求| ai -->|执行|任务1 --> |执行|任务2  -->|...| 任务N -->|反馈结果，等待用户确认|A
    
```

#### 开发视角逻辑架构

```mermaid

graph TD
    A[用户请求]
    B[AI]
    subgraph C[微服务1]
        direction TB
        CA[函数选择AI]
        subgraph C1[函数1]
            direction TB
           C1A[参数生成AI中间件]
           C1B[函数1]
           C1A -->|生成参数并调用函数| C1B
        end

       subgraph C2[函数1]
          direction TB
          C2A[参数生成AI中间件]
          C2B[函数1]
          C2A -->|生成参数并调用函数| C2B
       end

       subgraph C3[函数1]
          direction TB
          C3A[参数生成AI中间件]
          C3B[函数1]
          C3A -->|生成参数并调用函数| C3B
       end
        
        CA -->|选择函数| C1
        CA -->|选择函数| C2
        CA -->|选择函数| C3
    end
   subgraph D[微服务2]
      direction TB
      DA[函数选择AI]
      subgraph D1[函数1]
         direction TB
         D1A[参数生成AI中间件]
         D1B[函数1]
         D1A -->|生成参数并调用函数| D1B
      end

      subgraph D2[函数1]
         direction TB
         D2A[参数生成AI中间件]
         D2B[函数1]
         D2A -->|生成参数并调用函数| D2B
      end

      subgraph D3[函数1]
         direction TB
         D3A[参数生成AI中间件]
         D3B[函数1]
         D3A -->|生成参数并调用函数| D3B
      end

      DA -->|选择函数| D1
      DA -->|选择函数| D2
      DA -->|选择函数| D3
   end
    
    A -->|发送请求| B 
   B --> |选择微服务| C
   B --> |选择微服务| D

```

#### 开发视角物理架构

```mermaid
graph TD
A[用户请求]
subgraph B[AI微服务]
    direction LR
    BA[主AI--选择微服务]
    BB[二级AI--选择函数]
    BC[三级AI--生成参数]
end

subgraph C[微服务idl-记录 api 相关数据]
    direction LR
   CA[微服务的相关功能介绍清单]
    CB[微服务 1 的 idl]
    CC[微服务 2 的 idl]
end

subgraph D[微服务1]
    direction LR
    DA[函数1]
    DB[函数2]
end

subgraph E[微服务2]
    direction LR
    EA[函数1]
    EB[函数2]
end



BA -->|查询|CA
BA -->|传递所需微服务名称|BB
BB -->|查询|CB
BB -->|查询|CC
BB -->|传递所需函数|BC
BC -->|生成参数并调用|DA
BC -->|生成参数并调用|DB
BC -->|生成参数并调用|EA
BC -->|生成参数并调用|EB

   A -->|请求| B
   

```