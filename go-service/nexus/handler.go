package main

import (
	"fmt"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/biz/nexus"
	nexus_microservice "github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/kr/pretty"
	"github.com/openai/openai-go"
)

// NexusServiceImpl implements the last service interface defined in the IDL.
type NexusServiceImpl struct {
}

// 通义大模型
var baseUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1/"
var apiKey = "sk-8285fe317edc44ef95f029be9b7cfe94" // 自行去官网申请 apiKey
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

// request2openai 将通用的消息格式转换为openai的消息格式
func request2openai(messages []*nexus_microservice.Message) (openaiMessages []openai.ChatCompletionMessageParamUnion) {

	klog.Info("Received request2openai request:")
	pretty.Println(messages)
	fmt.Println("=======================")

	for _, message := range messages {

		if message.Role == "system" {
			openaiMessages = append(openaiMessages, openai.SystemMessage(message.Content))
			continue
		}

		if message.Role == "user" {
			openaiMessages = append(openaiMessages, openai.UserMessage(message.Content))
			continue
		}

		if message.Role == "assistant" {

			// 解析工具调用列表
			toolCalls := []openai.ChatCompletionMessageToolCall{}

			for _, tool := range message.ToolCalls {
				toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCall{
					Type: openai.ChatCompletionMessageToolCallType(tool.Type),
					Function: openai.ChatCompletionMessageToolCallFunction{
						Arguments: *tool.FunctionCall.Arguments,
						Name:      tool.FunctionCall.Name,
					},
				})
			}

			// 生成最终机器人回复的消息类型
			assistantMessage := openai.ChatCompletionMessage{
				Content:      message.Content,
				Role:         openai.ChatCompletionMessageRoleAssistant,
				FunctionCall: openai.ChatCompletionMessageFunctionCall{},
				ToolCalls:    toolCalls,
			}

			openaiMessages = append(openaiMessages, assistantMessage)
			continue
		}

		if message.Role == "tool" {
			tool_message := openai.ChatCompletionMessage{
				Content:      message.Content,
				Role:         "tool",
				FunctionCall: openai.ChatCompletionMessageFunctionCall{},
				ToolCalls:    []openai.ChatCompletionMessageToolCall{},
			}
			openaiMessages = append(openaiMessages, tool_message)
			continue
		}

	}

	return
}

func getParamsFromThrift() []openai.ChatCompletionToolParam {
	return []openai.ChatCompletionToolParam{
		{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(openai.FunctionDefinitionParam{
				Name:        openai.String("get_travel_location"),
				Description: openai.String("用于获取值得推荐的旅游景点"),
				Parameters: openai.F(openai.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]string{
							"type":        "string",
							"description": "城市名字：比如浙江、昆山、杭州、北京",
						},
					},
					"required": []string{"location"},
				}),
			}),
		},
		{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(openai.FunctionDefinitionParam{
				Name:        openai.String("make_plan"),
				Description: openai.String("用于安排计划清单"),
				Parameters: openai.F(openai.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"todos": map[string]interface{}{
							"type": "array",
							"items": map[string]string{
								"type": "string",
							},
							"description": "任务清单：比如'买菜'、'逛街等'",
						},
					},
					"required": []string{"location"},
				}),
			}),
		},
	}
}

// CallByJson 将json格式的参数解析一下并且调用工具
func CallByJson(functionName string, params string) string {
	fmt.Println(functionName, params)
	return "金山寺"
}

// event2response 将openai的事件转换为通用的消息格式
func event2response(event openai.ChatCompletionChunk) (response *nexus_microservice.AskResponse) {

	// 构建响应
	response = &nexus_microservice.AskResponse{}

	response.Id = event.ID
	response.Model = event.Model

	// 构建函数调用相关参数
	toolCalls := make([]*nexus_microservice.ToolCall, 0)

	for _, toolCall := range event.Choices[0].Delta.ToolCalls {
		toolCalls = append(toolCalls, &nexus_microservice.ToolCall{
			Type: string(toolCall.Type),
			FunctionCall: &nexus_microservice.FunctionCall{
				Name:      toolCall.Function.Name,
				Arguments: &toolCall.Function.Arguments,
			},
		})
	}

	messages := make([]*nexus_microservice.Message, 0)

	delta := event.Choices[0].Delta

	messages = append(messages, &nexus_microservice.Message{
		Role:    string(delta.Role),
		Content: delta.Content,
		FunctionCall: &nexus_microservice.FunctionCall{
			Name:      delta.FunctionCall.Name,
			Arguments: &delta.FunctionCall.Arguments,
		},
		ToolCalls: toolCalls,
	})

	choices := make([]*nexus_microservice.Choice, 0)

	for _, choice := range event.Choices {
		finishReason := string(choice.FinishReason)
		choices = append(choices, &nexus_microservice.Choice{
			FinishReason: &finishReason,
			Message:      messages,
			Index:        int32(choice.Index),
		})
	}

	response.Choices = choices
	return

}

func (s *NexusServiceImpl) AskServer(req *nexus_microservice.AskRequest, stream nexus_microservice.NexusService_AskServerServer) (err error) {
	klog.Info("Received AskServer request:", req)

	// 初始化通义千问大模型
	nexus.QwenInstance.Init(baseUrl, apiKey)
	nexus.QwenInstance.SetModel(model)
	nexus.QwenInstance.SetPrompt(prompt)
	nexus.QwenInstance.SetMessages(request2openai(req.Messages))
	nexus.QwenInstance.SetTools(getParamsFromThrift())

	// 初始化流
	chatStream := nexus.QwenInstance.NewStream()

	pretty.Print(request2openai(req.Messages))
	//pretty.Print(getParamsFromThrift())

	for chatStream.Next() {
		event := chatStream.Current()
		if len(event.Choices) <= 0 {
			continue
		}

		askResponse := event2response(event)
		fmt.Println("resp:", askResponse)
		err = stream.Send(askResponse)
		if err != nil {
			fmt.Println("EchoServer failed: %v", err)
		}

	}

	return
}
