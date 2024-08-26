// Package nexus @Author Adrian.Wang 2024/8/26 下午8:09:00
package nexus

import (
	"fmt"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/ssestream"
)

// StreamAgent 是一个流代理，用于处理流的请求
// 并且在流处理过程中要进行函数调用和其他中间处理过程
type StreamAgent struct {
	functionName      string
	functionArguments string
	_type             string
	id                string
	content           string
	messages          []openai.ChatCompletionMessageParamUnion
}

func NewStreamAgent() *StreamAgent {
	return &StreamAgent{}
}

func (sa *StreamAgent) Init() {
	sa.messages = []openai.ChatCompletionMessageParamUnion{}
}

// CallFunction 调用函数
func (sa *StreamAgent) CallFunction() string {
	return "金山寺"
}

// Monitor 监控流的请求,并执行相关函数调用
func (sa *StreamAgent) Monitor(event openai.ChatCompletionChunk) {

	// 当函数调用相关的参数生成完毕后，进行函数调用
	if event.Choices[0].FinishReason == openai.ChatCompletionChunkChoicesFinishReasonFunctionCall ||
		event.Choices[0].FinishReason == openai.ChatCompletionChunkChoicesFinishReasonToolCalls {

		// 调用函数
		res := sa.CallFunction()

		if sa._type == "" {
			sa._type = "tool"
		}

		// 返回工具调用结果作为工具调用消息，插入到消息队列中
		tool_message := openai.ChatCompletionMessage{
			Content:      res,
			Role:         "tool",
			FunctionCall: openai.ChatCompletionMessageFunctionCall{},
			ToolCalls:    []openai.ChatCompletionMessageToolCall{},
		}

		// 返回机器人的消息，插入到消息队列中
		assisant_messages := openai.ChatCompletionMessage{
			Content:      sa.content,
			Role:         openai.ChatCompletionMessageRoleAssistant,
			FunctionCall: openai.ChatCompletionMessageFunctionCall{},
			ToolCalls: []openai.ChatCompletionMessageToolCall{
				{
					ID:   sa.id,
					Type: openai.ChatCompletionMessageToolCallType(sa._type),
					Function: openai.ChatCompletionMessageToolCallFunction{
						Arguments: sa.functionArguments,
						Name:      sa.functionName,
					},
				},
			},
		}

		// 添加消息到消息队列中
		sa.messages = append(sa.messages, assisant_messages, tool_message)

		sa.id = ""
		sa.functionArguments = ""
		sa.functionName = ""
		sa._type = ""
		sa.content = ""

		return

	}

	delta := event.Choices[0].Delta

	if delta.Content != "" {

		fmt.Print(delta.Content)

		sa.content += delta.Content
	}

	// 没有调用,直接返回
	if len(delta.ToolCalls) <= 0 {
		return
	}

	toolCall := delta.ToolCalls[0]

	if toolCall.Type != openai.ChatCompletionChunkChoicesDeltaToolCallsTypeFunction {
		return
	}

	sa._type = string(toolCall.Type)

	if toolCall.ID != "" {
		sa.id = toolCall.ID
	}

	function := toolCall.Function

	if function.Name != "" {
		sa.functionName += function.Name
	}

	if function.Arguments != "" {
		sa.functionArguments += function.Arguments
	}

}

// Messages 获取本次对话的消息
func (sa *StreamAgent) Messages() []openai.ChatCompletionMessageParamUnion {
	return sa.messages
}

// ForwardResponse  转发响应请求并进行中间处理
func (sa *StreamAgent) ForwardResponse(source *ssestream.Stream[openai.ChatCompletionChunk], target nexus_microservice.NexusService_AskServerServer) {
	// 开始对话,使用代理模式进行对话
	for source.Next() {
		event := source.Current()
		if len(event.Choices) <= 0 {
			continue
		}

		askResponse := Event2response(event)

		// 监控流，在监控过程中函数生成成功的那一刻进行函数调用
		sa.Monitor(event)

		fmt.Println("resp:", askResponse)
		err := target.Send(askResponse)
		if err != nil {
			fmt.Println("EchoServer failed: %v", err)
		}
	}

	//TODO 想办法解决，函数调用+消息发送的混合消息流中，如何判断对话结束

	if err := source.Err(); err != nil {
		klog.Info("StreamAgent ForwardResponse error:", err)
	}

}
