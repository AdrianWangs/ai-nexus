# 函数调用图

```mermaid
sequenceDiagram
    participant User
    participant NexusServiceImpl as 主 ai 服务
    participant QwenInstance as ai 服务使用的Gpt
    participant StreamAgent as 主 ai 的流代理
    participant SubNexusServiceImpl as 次级 ai 服务
    participant QwenInstanceSub as 次级 ai 服务使用的Gpt
    participant ToolFunction as 工具函数

    User->>NexusServiceImpl: 请求服务
    NexusServiceImpl->>QwenInstance: 初始化&设置参数
    QwenInstance->>StreamAgent: 初始化&流代理注册
    loop 对话与处理
        StreamAgent->>QwenInstance: 接收&转发事件
        QwenInstance-->>StreamAgent: 回复内容
        alt 函数调用检测
            StreamAgent->>SubNexusServiceImpl: 准备调用函数
            SubNexusServiceImpl->>QwenInstanceSub: 初始化&参数设置
            QwenInstanceSub->>StreamAgent: 初始化&流代理注册
            loop 次级对话处理
                StreamAgent->>QwenInstanceSub: 转发事件
                QwenInstanceSub-->>StreamAgent: 函数调用准备完成
                StreamAgent->>Function: 调用函数
                Function-->>StreamAgent: 返回结果
                StreamAgent->>SubNexusServiceImpl: 添加调用结果到消息
                SubNexusServiceImpl->>QwenInstanceSub: 继续对话
            end
            SubNexusServiceImpl->>NexusServiceImpl: 返回最终结果
        else 正常对话处理
            StreamAgent->>NexusServiceImpl: 添加对话内容
        end
    end
    NexusServiceImpl-->>User: 返回最终结果
```


# 使用方法

## thrift 文件命名规范
在 `resources/idl/`目录下存放微服务描述相关的 thrift文件,命名必须以”微服务名称.thrift“,其中微服务名称必须和服务注册的微服务名称保持一致，否则无法进行调用

thrift 文件以`// 注释`开头可以向 ai 指明当前微服务的作用，帮助 ai 更好地进行选择函数调用





# Project

## introduce

- Use the [Kitex](https://github.com/cloudwego/kitex/) framework
- Generating the base code for unit tests.
- Provides basic config functions
- Provides the most basic MVC code hierarchy.

## Directory structure

|  catalog   | introduce  |
|  ----  | ----  |
| conf  | Configuration files |
| main.go  | Startup file |
| handler.go  | Used for request processing return of response. |
| kitex_gen  | kitex generated code |
| biz/service  | The actual business logic. |
| biz/dal  | Logic for operating the storage layer |

## How to run

```shell
sh build.sh
sh output/bootstrap.sh
```