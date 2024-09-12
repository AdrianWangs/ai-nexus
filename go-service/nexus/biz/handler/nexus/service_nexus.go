// Package nexus @Author Adrian.Wang 2024/9/1 16:21:00
// 主要用于次级 ai 的函数调用等功能
package nexus

import (
	"fmt"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/biz/handler/nexus/models"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/kr/pretty"
	"github.com/openai/openai-go"
	"os"
)

// NexusServiceImpl implements the last service interface defined in the IDL.
type NexusServiceImpl struct {
}

// 通义大模型
// var baseUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1/"
var baseUrl = "https://4.0.wokaai.com/v1/"
var apiKey = "" // 自行去官网申请 apiKey
var model = "gpt-4o"

// 提示词
var prompt = `
# 角色
你是一个高效的次级 AI 大脑，精准接收用户请求并调用相应函数，确保参数填写准确无误。只负责调用一个函数，不进行任何额外回复。

## 技能
### 技能 1：调用函数
1. 仔细分析用户请求，确定需要调用的函数。
2. 准确填写函数所需参数，确保参数与用户请求匹配。

## 限制
- 只进行函数调用，不做任何其他回复。
- 确保参数填写正确，符合用户请求。
- 函数调用失败的话重试次数不超过三次
`

// AskService 是一个流式接口，接收主 ai 的需求并指导次级 ai 进行函数调用
// [mainStreamAgent] 是主 ai 的流代理对象，毕竟我们当前调用还是在主 ai 中，所以需要将消息加入到主 ai 的消息列表中
func AskService(service string, nexusPrompt string, req *nexus_microservice.AskRequest, stream nexus_microservice.NexusService_AskServerServer, mainStreamAgent *StreamAgent) (res string, err error) {

	klog.Info("======================================================")
	klog.Info("调用服务:", service)
	klog.Info("请求的提示词：", nexusPrompt)
	klog.Info("调用结果:")
	klog.Info("======================================================")
	// 从环境变量中获取 API_KEY
	apiKey = os.Getenv("API_KEY")

	qwenInstance := models.NewQwen()

	// 如果模型设置不为空，就设置用户指定的模型
	if req.Model != nil {
		qwenInstance.SetModel(*req.Model)
	}

	idlPath := fmt.Sprintf("./resources/idl/%s.thrift", service)

	qwenInstance.SetTools(GetParamsFromThrift(service, idlPath))

	qwenInstance.Init(baseUrl, apiKey)
	qwenInstance.SetModel(model)

	// 初始化通义千问大模型
	qwenInstance.SetPrompt(prompt)

	// 将传过来的参数中的对话添加到设置为真正的对话列表
	qwenInstance.SetMessages([]openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(nexusPrompt),
	})

	// 注册流代理，用于转发流，也就是将 openai 返回的流消息转发给 kitex 的流对象
	streamAgent := NewStreamAgent()

	// 使用代理转发流，并在转发过程中自动执行函数调用
	for !streamAgent.IsStop() {

		// 初始化流
		chatStream := qwenInstance.NewStream()

		// 使用代理跟踪发送流，并且在一段流对话后把消息加入到原始消息中
		// 使用代理可以在转发流的过程中进行额外操作，比如进行函数调用
		streamAgent.ForwardResponseForSubNexus(chatStream, stream, mainStreamAgent)

		// 将消息添加到消息列表中
		qwenInstance.AddMessages(streamAgent.Messages())

		// 将当前流代理的消息清除，否则会导致本轮的对话堆积起来，streamAgent 只负责一次对话
		// 但是函数调用需要多次对话直到 ai 自己认为可以结束了，才会真的暂停
		// 假设本轮对话消息是A，下一轮对话B，下下轮对话是 C,
		// 如果不清除对话，下一轮对话就会添加[A,B]到总对话列表[A]中，变成[A,A,B]
		// 下下轮就会是[A,A,B,A,B,C]
		// 但是正常来说最终对话列表应该是[A,B,C]
		streamAgent.ClearMessages()

	}

	klog.Info("************************************************")
	klog.Info("次级对话结果:")
	pretty.Println(qwenInstance.Messages())
	klog.Info("************************************************")

	res = ""

	// 将次级 ai 对话的所有工具的结果作为结果返回
	for _, message := range qwenInstance.Messages() {
		if val, ok := message.(openai.ChatCompletionMessage); ok {
			res += fmt.Sprintf("%s:%s\n", val.Role, val.Content)
			if len(val.ToolCalls) <= 0 {
				continue
			}

			for _, tool := range val.ToolCalls {
				res += fmt.Sprintf("调用函数：%s\n", tool.Function.Name)
				res += fmt.Sprintf("参数：%s\n", tool.Function.Arguments)
			}

		} else if val, ok := message.(openai.ChatCompletionToolMessageParam); ok {
			res += fmt.Sprintf("%s:%s\n", val.Role, val.Content)
		}
	}

	return
}
