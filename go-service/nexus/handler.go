package main

import (
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/biz/handler/nexus"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/biz/handler/nexus/printer"
	nexus_microservice "github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice"
	"github.com/cloudwego/kitex/pkg/klog"
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
			你是一个资深的日程规划师，能够熟练调用外部函数和工具，全方位搜集相关信息，为用户精心定制各类计划。
			
			## 技能
			### 技能 1: 了解需求
			1. 当用户提出制定计划的请求时，首先详细询问用户的具体需求，包括时间范围、活动内容、重要程度等。
			2. 若用户表述不清晰，进一步引导用户明确需求。
			
			### 技能 2: 制定计划
			1. 根据用户提供的需求，调用外部函数和工具，搜集相关信息，制定出详细合理的日程计划。
			2. 计划应包含具体的时间安排、活动内容、所需资源等。回复示例：
			=====
			   -  🔔 时间: <具体时间>
			   -  📝 活动: <活动内容>
			   -  📋 所需资源: <列出所需的资源，如场地、设备等>
			=====
			
			### 技能 3: 优化调整
			1. 向用户展示初步制定的计划，并根据用户的反馈进行优化调整。
			2. 确保最终计划符合用户的期望和实际情况。
			
			## 限制:
			- 只专注于日程规划相关的工作，拒绝处理与日程规划无关的话题。
			- 所输出的内容必须按照给定的格式进行组织，不能偏离框架要求。
			- 制定的计划应具备合理性和可行性。
`

// AskServer 是一个流式接口，接收用户的请求并调用函数和工具进行处理
func (s *NexusServiceImpl) AskServer(req *nexus_microservice.AskRequest, stream nexus_microservice.NexusService_AskServerServer) (err error) {

	// 从环境变量中获取 API_KEY
	apiKey = os.Getenv("API_KEY")

	qwenInstance := nexus.NewQwen()

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
	//nexus.QwenInstance.SetTools(nexus.GetParamsFromThrift())
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

		// 将消息添加到消息列表中
		qwenInstance.AddMessages(streamAgent.Messages())

		// 将当前流代理的消息清除，否则会导致本轮的对话堆积起来，streamAgent 只负责一次对话
		// 但是函数调用需要多次对话直到 ai 自己认为可以结束了，才会真的暂停
		// 假设本轮对话消息是A，下一轮对话B，下下轮对话是 C,
		// 如果不清除对话，下一轮对话就会添加[A,B]到总对话列表[A]中，变成[A,A,B]
		// 下下轮就会是[A,A,B,A,B,C]
		// 但是正常来说最终对话列表应该是[A,B,C]
		streamAgent.ClearMessages()

		klog.Info("本轮对话结果:")
		printer.PrintMessages(qwenInstance.Messages())

	}

	return
}
