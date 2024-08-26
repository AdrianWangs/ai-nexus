package main

import (
	"fmt"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/biz/nexus"
	nexus_microservice "github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice"
	"os"
)

// NexusServiceImpl implements the last service interface defined in the IDL.
type NexusServiceImpl struct {
}

// 通义大模型
var baseUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1/"
var apiKey = "" // 自行去官网申请 apiKey
var model = "qwen-max"
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

	apiKey = os.Getenv("API_KEY")

	// 初始化通义千问大模型
	nexus.QwenInstance.Init(baseUrl, apiKey)
	nexus.QwenInstance.SetModel(model)
	nexus.QwenInstance.SetPrompt(prompt)
	nexus.QwenInstance.SetMessages(nexus.Request2openai(req.Messages))
	nexus.QwenInstance.SetTools(nexus.GetParamsFromThrift())

	// 注册流代理，用于转发流
	streamAgent := nexus.NewStreamAgent()

	// 使用代理转发流，并在转发过程中自动执行函数调用
	for !streamAgent.IsStop() {

		// 初始化流
		chatStream := nexus.QwenInstance.NewStream()
		streamAgent.ForwardResponse(chatStream, stream)
		fmt.Println(streamAgent.Messages())
		// 将消息添加到消息列表中
		nexus.QwenInstance.AddMessages(streamAgent.Messages())
		streamAgent.ClearMessages()
	}

	return
}
