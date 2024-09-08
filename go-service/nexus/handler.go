package main

import (
	"fmt"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/biz/handler/nexus"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/biz/handler/nexus/models"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/biz/handler/nexus/printer"
	nexus_microservice "github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/kr/pretty"
	"os"
)

// NexusServiceImpl implements the last service interface defined in the IDL.
type NexusServiceImpl struct {
}

// 通义大模型
var baseUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1/"
var apiKey = "" // 自行去官网申请 apiKey
var model = "qwen-max"

// 提示词
var prompt = `
# 角色
你是一个高效的智核，能够精准分析用户需求，转化为任务序列，逐一调用微服务提供服务。每个任务执行小操作，单个服务可多次调用。

## 技能
### 技能 1：分析用户需求
1. 仔细理解用户输入内容，确定核心需求点。
2. 将需求拆分为具体的任务序列。

### 技能 2：调用微服务
1. 按照任务序列依次调用对应的微服务。
2. 每个微服务执行一个小操作。
3. 记录每个微服务的调用结果。

### 技能 3：归纳总结回应
1. 整合所有微服务调用结果。
2. 对结果进行归纳总结，给出清晰回应。

## 限制
- 严格按照任务序列执行，不重复执行失败任务。
- 只执行与用户需求相关的任务和调用相关微服务。
- 回应内容要准确、简洁，符合用户需求。
`

// AskServer 是一个流式接口，接收用户的请求并调用函数和工具进行处理
func (s *NexusServiceImpl) AskServer(req *nexus_microservice.AskRequest, stream nexus_microservice.NexusService_AskServerServer) (err error) {

	// 从环境变量中获取 API_KEY
	apiKey = os.Getenv("API_KEY")

	qwenInstance := models.NewQwen()

	qwenInstance.Init(baseUrl, apiKey)
	qwenInstance.SetModel(model)

	// 初始化通义千问大模型
	qwenInstance.SetPrompt(prompt)

	// 将传过来的参数中的对话添加到设置为真正的对话列表
	qwenInstance.SetMessages(nexus.Request2openai(req.Messages))

	// 如果模型设置不为空，就设置用户指定的模型
	if req.Model != nil {
		qwenInstance.SetModel(*req.Model)
	}

	// 最顶级的应该是先将微服务列表传入，然后让ai选择使用哪一个微服务
	qwenInstance.SetTools(nexus.GetServicesFromThrift())

	// 注册流代理，用于转发流，也就是将 openai 返回的流消息转发给 kitex 的流对象
	streamAgent := nexus.NewStreamAgent()

	// 使用代理转发流，并在转发过程中自动执行函数调用
	for !streamAgent.IsStop() {

		// 初始化流
		chatStream := qwenInstance.NewStream()

		// 使用代理跟踪发送流，并且在一段流对话后把消息加入到原始消息中
		// 使用代理可以在转发流的过程中进行额外操作，比如进行函数调用
		streamAgent.ForwardResponse(chatStream, stream, req)

		// TODO 主函数的调用结果没能返回，所以出现了一些问题，这时候需要把主函数的调用结果传到消息列表才能解决一些问题
		// 将消息添加到消息列表中
		qwenInstance.AddMessages(streamAgent.Messages())

		// 将当前流代理的消息清除，否则会导致本轮的对话堆积起来，streamAgent 只负责一次对话
		// 但是函数调用需要多次对话直到 ai 自己认为可以结束了，才会真的暂停
		// 假设本轮对话消息是A，下一轮对话B，下下轮对话是 C,
		// 如果不清除对话，下一轮对话就会添加[A,B]到总对话列表[A]中，变成[A,A,B]
		// 下下轮就会是[A,A,B,A,B,C]
		// 但是正常来说最终对话列表应该是[A,B,C]
		streamAgent.ClearMessages()

		fmt.Println("000000000000000000000000000000000000000000")
		klog.Info("本轮对话结果:")
		printer.PrintMessages(qwenInstance.Messages())
		fmt.Println("000000000000000000000000000000000000000000")

	}

	fmt.Println("111111111111111111111111111111111111111111111111111111111111")
	pretty.Println(qwenInstance.Messages())
	fmt.Println("111111111111111111111111111111111111111111111111111111111111")

	return
}
