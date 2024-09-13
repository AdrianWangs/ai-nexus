// Package nexus @Author Adrian.Wang 2024/8/26 下午8:01:00
package nexus

import (
	"fmt"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/biz/handler/nexus/parser"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/openai/openai-go"
	"os"
	"path/filepath"
)

// Request2openai 将通用的消息格式转换为openai的消息格式
func Request2openai(messages []*nexus_microservice.Message) (openaiMessages []openai.ChatCompletionMessageParamUnion) {

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

// GetServicesFromThrift 从thrift中获取服务
func GetServicesFromThrift() []openai.ChatCompletionToolParam {

	filenames, err := filepath.Glob("./resources/idl/*.thrift")

	if err != nil {
		klog.Error("解析 thrift 文件失败")
		os.Exit(1)
	}

	toolServices, err := parser.ParseThriftServiceFromPaths(filenames)

	return toolServices

}

// GetParamsFromThrift 从thrift中获取参数
func GetParamsFromThrift(serviceName string, idlPath string) []openai.ChatCompletionToolParam {
	toolParams, err := parser.ParseThriftIdlFromPath(idlPath)
	if err != nil {
		klog.Error("解析 thrift 文件失败")
		return []openai.ChatCompletionToolParam{}
	}

	for index, _ := range toolParams {
		// 服务名称
		toolParams[index].Function.Value.Name = openai.String(
			serviceName + "-" + toolParams[index].Function.Value.Name.String())
	}

	return toolParams
}

// CallByJson 将json格式的参数解析一下并且调用工具
func CallByJson(functionName string, params string) string {
	fmt.Println(functionName, params)
	return "金山寺"
}

// Event2response 将openai的事件转换为通用的消息格式
func Event2response(event openai.ChatCompletionChunk) (response *nexus_microservice.AskResponse) {

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
