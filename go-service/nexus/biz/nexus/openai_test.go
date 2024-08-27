package nexus

import (
	"fmt"
	"github.com/openai/openai-go"
	"testing"
)

var baseUrl string
var model string
var apiKey string

func init() {

	//// 本地 ollama
	//baseUrl = "http://localhost:11434/v1/"
	//apiKey = "ollama"
	//model = "llama3.1:8b"

	// 通义大模型
	baseUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1/"
	apiKey = "sk-8285fe317edc44ef95f029be9b7cfe94" // 自行去官网申请 apiKey
	model = "qwen-max"

	QwenInstance.Init(baseUrl, apiKey)
	QwenInstance.SetModel(model)

}

// 将json格式的参数解析一下并且调用工具
func CallByJson(functionName string, params string) string {
	fmt.Println(functionName, params)
	return "金山寺"
}

func callGpt() (bool, []openai.ChatCompletionMessageParamUnion) {

	chatStream := QwenInstance.NewStream()

	var function_name string
	var function_arguments string
	var _type string
	var id string
	var content string

	for chatStream.Next() {
		event := chatStream.Current()
		if len(event.Choices) <= 0 {
			continue
		}

		if event.Choices[0].FinishReason == openai.ChatCompletionChunkChoicesFinishReasonFunctionCall ||
			event.Choices[0].FinishReason == openai.ChatCompletionChunkChoicesFinishReasonToolCalls {
			res := CallByJson(function_name, function_arguments)

			tool_message := openai.ChatCompletionMessage{
				Content:      res,
				Role:         "tool",
				FunctionCall: openai.ChatCompletionMessageFunctionCall{},
				ToolCalls:    []openai.ChatCompletionMessageToolCall{},
			}

			assisant_messages := openai.ChatCompletionMessage{
				Content:      content,
				Role:         openai.ChatCompletionMessageRoleAssistant,
				FunctionCall: openai.ChatCompletionMessageFunctionCall{},
				ToolCalls: []openai.ChatCompletionMessageToolCall{
					{
						ID:   id,
						Type: openai.ChatCompletionMessageToolCallType(_type),
						Function: openai.ChatCompletionMessageToolCallFunction{
							Arguments: function_arguments,
							Name:      function_name,
						},
					},
				},
			}

			QwenInstance.AddMessage(assisant_messages)
			QwenInstance.AddMessage(tool_message)

			fmt.Println("函数调用结果：", res)

			function_name = ""
			function_arguments = ""

			return false, QwenInstance.Messages()
		}

		delta := event.Choices[0].Delta

		if delta.Content != "" {

			fmt.Print(delta.Content)

			content += delta.Content
		}

		// 没有调用
		if len(delta.ToolCalls) <= 0 {
			continue
		}

		toolCall := delta.ToolCalls[0]

		if toolCall.Type != openai.ChatCompletionChunkChoicesDeltaToolCallsTypeFunction {
			continue
		}

		_type = string(toolCall.Type)

		if toolCall.ID != "" {
			id = toolCall.ID
		}

		function := toolCall.Function

		if function.Name != "" {
			function_name += function.Name
		}

		if function.Arguments != "" {
			function_arguments += function.Arguments
		}

	}

	if err := chatStream.Err(); err != nil {

		println(err.Error())

	}

	println()

	assisant_messages := openai.ChatCompletionMessage{
		Content:      content,
		Role:         openai.ChatCompletionMessageRoleAssistant,
		FunctionCall: openai.ChatCompletionMessageFunctionCall{},
		ToolCalls: []openai.ChatCompletionMessageToolCall{
			{
				ID:   id,
				Type: openai.ChatCompletionMessageToolCallType(_type),
				Function: openai.ChatCompletionMessageToolCallFunction{
					Arguments: function_arguments,
					Name:      function_name,
				},
			},
		},
	}

	QwenInstance.AddMessage(assisant_messages)

	return true, QwenInstance.Messages()

}

// TestFunctionCall
func TestFunctionCall(t *testing.T) {

	QwenInstance.SetPrompt(`
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
	`)

	question := "我周末想要去苏州玩，你有什么意见？"

	messages := QwenInstance.Messages()

	messages = append(messages, openai.UserMessage(question))

	QwenInstance.SetMessages(messages)

	QwenInstance.SetTools([]openai.ChatCompletionToolParam{
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
	})

	isEnd := false

	for !isEnd {
		isEnd, messages = callGpt()
		if len(messages) > 0 {
			fmt.Println("==================Start=====================")

			for _, message := range messages {

				fmt.Println(message)
			}

			fmt.Println("==================End=====================")
		}

	}

}
