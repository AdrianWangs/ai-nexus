// Package nexus @Author Adrian.Wang 2024/8/26 下午8:09:00
package nexus

import (
	"errors"
	"fmt"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice"
	"github.com/cloudwego/hertz/pkg/common/json"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/niemeyer/pretty"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/ssestream"
)

// StreamAgent 是一个流代理，用于处理流的请求
// 并且在流处理过程中要进行函数调用和其他中间处理过程
type StreamAgent struct {
	functionName      string //函数名称
	functionArguments string // 函数调用参数
	_type             string
	id                string
	content           string
	messages          []openai.ChatCompletionMessageParamUnion
	isStop            bool //多轮对话控制结束对话
}

// NewStreamAgent 用于生成一个新的流代理对象
func NewStreamAgent() *StreamAgent {
	return &StreamAgent{
		isStop: false,
	}
}

// Init 流代理初始化，也就是将消息设置为空
func (sa *StreamAgent) Init() {
	sa.messages = []openai.ChatCompletionMessageParamUnion{}
}

// ForwardResponse  转发响应请求并进行中间处理
func (sa *StreamAgent) ForwardResponse(source *ssestream.Stream[openai.ChatCompletionChunk], target nexus_microservice.NexusService_AskServerServer, req *nexus_microservice.AskRequest) {

	// 开始对话,使用代理模式进行对话
	for source.Next() {

		event := source.Current()

		// 如果本轮对话没有任何回复就不需要进行其他额外的操作了
		if len(event.Choices) <= 0 {
			klog.Info("好像没对话内容...")
			pretty.Println(event)
			continue
		}

		// 将 openai 传过来的数据转化成我们网站对应的 response 格式
		askResponse := Event2response(event)

		// 监控流，在监控过程中函数生成成功的那一刻进行函数调用
		sa.Monitor(event, target, req)

		// 监控完以后不出意外就该转发刚刚的对话了
		err := target.Send(askResponse)
		if err != nil {
			fmt.Println("EchoServer failed: ", err)
		}
	}

	if err := source.Err(); err != nil {
		klog.Error("StreamAgent ForwardResponse error:", err)
		sa.isStop = true
	}

}

// Monitor 监控流的请求,并执行相关函数调用
func (sa *StreamAgent) Monitor(event openai.ChatCompletionChunk, target nexus_microservice.NexusService_AskServerServer, req *nexus_microservice.AskRequest) {

	// 结束对话
	if event.Choices[0].FinishReason == openai.ChatCompletionChunkChoicesFinishReasonStop {

		// 结束本轮对话
		sa.EndConversation()

		return
	}

	// 当函数调用相关的参数生成完毕后，进行函数调用
	if event.Choices[0].FinishReason == openai.ChatCompletionChunkChoicesFinishReasonFunctionCall ||
		event.Choices[0].FinishReason == openai.ChatCompletionChunkChoicesFinishReasonToolCalls {

		// 调用服务，可能涉及子 ai 调用，所以要把流对象和相关请求一起传入
		sa.CallService(target, req)

		return
	}

	delta := event.Choices[0].Delta

	if delta.Content != "" {

		// 打印对话内容
		fmt.Print(delta.Content)
		sa.content += delta.Content

	}

	// 没有调用,直接返回
	if len(delta.ToolCalls) <= 0 {
		return
	}

	toolCall := delta.ToolCalls[0]

	// 判断是否是函数调用
	if toolCall.Type != openai.ChatCompletionChunkChoicesDeltaToolCallsTypeFunction {
		return
	}

	// 完善函数调用相关的信息，也就是切片组合成完整信息
	sa.CompleteFunctionCall(toolCall)

}

// CallService 调用服务
func (sa *StreamAgent) CallService(target nexus_microservice.NexusService_AskServerServer, req *nexus_microservice.AskRequest) {

	// 调用次级 ai
	res, err := sa.DoService(target, req)

	if err != nil {
		klog.Error("服务调用失败:", err)
		// 清空上下文，防止前面流影响后面的操作
		sa.ClearContext()
		return
	}

	// 这里应该是固定的 openai 格式（目前）
	if sa._type == "" {
		sa._type = "tool"
	}

	// TODO: 应该是知道服务名称，然后将消息转发给新的拥有
	//  对应服务的函数清单的 ai 服务去选择并且调用函数，这样可以确保函数调用的准确性
	//  TODO: 想一下这里的函数还需不需要调用流？
	fmt.Println("==========")
	fmt.Println("调用服务:", sa.functionName)
	fmt.Println("请求的提示词：", sa.functionArguments)
	fmt.Println("调用结果:", res)
	fmt.Println("==========")

	// 主 ai 不负责 ai 调用方面的逻辑，只负责将消息转发给 ai 服务，真正的调用应该交付给次级 ai
	// 因此下面的代码不再需要了
	// TODO 确认到底是主 ai 处理函数调用逻辑还是次级 ai 处理函数调用逻辑
	if false {
		// 返回微服务调用结果作为工具调用消息，插入到消息队列中
		toolMessage := sa.GenerateToolMessage(res)

		// 返回机器人的消息，插入到消息队列中
		assistantMessages := sa.GenerateAssistantMessage()

		// 添加消息到消息队列中
		sa.messages = append(sa.messages, assistantMessages, toolMessage)
	}

	// 清空上下文，防止前面流影响后面的操作
	sa.ClearContext()

}

// DoService 执行相关服务，调用服务相关的流，交由下一级ai 进行函数调用
func (sa *StreamAgent) DoService(target nexus_microservice.NexusService_AskServerServer, req *nexus_microservice.AskRequest) (string, error) {

	// 需要调用的服务名称
	serviceName := sa.functionName
	// ai 生成给次级 ai 的提示词
	arguments := sa.functionArguments

	// 将 argument 尝试从 json 转化成 map
	var argumentMap map[string]interface{}
	err := json.Unmarshal([]byte(arguments), &argumentMap)
	if err != nil {
		return "", err
	}

	var nexusPrompt string

	// 判断 prompt 是否存在
	prompt, exist := argumentMap["prompt"]
	if !exist {
		err = errors.New("不存在 prompt 字段，调用失败")
		return "", err
	}

	// 判断 prompt 是否是字符串类型
	if promptValue, ok := prompt.(string); ok {
		nexusPrompt = promptValue
	} else {
		err = errors.New("prompt 字段不是字符串类型，调用失败")
		return "", err
	}

	// 将方法转化给次级 ai 进行调用
	return CallService(serviceName, nexusPrompt, req, target, sa)
}

// CompleteFunctionCall 完善函数调用相关的信息，主要负责拼接流分片中的函数调用相关的信息
func (sa *StreamAgent) CompleteFunctionCall(toolCall openai.ChatCompletionChunkChoicesDeltaToolCall) {
	// 函数调用类型（不知道有啥用）
	sa._type = string(toolCall.Type)

	// 函数调用 id（不知道有啥用）
	if toolCall.ID != "" {
		sa.id = toolCall.ID
	}

	// 函数调用名称
	function := toolCall.Function
	if function.Name != "" {
		sa.functionName += function.Name
	}

	// 函数调用参数
	if function.Arguments != "" {
		sa.functionArguments += function.Arguments
	}
}

// GenerateAssistantMessage 根据本次（不是本轮，一轮有多次对话）对话的生成机器人的消息格式
func (sa *StreamAgent) GenerateAssistantMessage() openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{
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
}

// GenerateToolMessage 生成工具类型的消息
func (sa *StreamAgent) GenerateToolMessage(res string) openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{
		Content:      res,
		Role:         "tool",
		FunctionCall: openai.ChatCompletionMessageFunctionCall{},
		ToolCalls:    []openai.ChatCompletionMessageToolCall{},
	}
}

// ClearContext 清空本次对话的上下文
func (sa *StreamAgent) ClearContext() {
	sa.id = ""
	sa.functionArguments = ""
	sa.functionName = ""
	sa._type = ""
	sa.content = ""
}

// EndConversation 结束对话
func (sa *StreamAgent) EndConversation() {
	// 返回机器人的消息，插入到消息队列中
	assistantMessages := sa.GenerateAssistantMessage()

	fmt.Println("==========")
	fmt.Println("结束对话,最终本轮对话：\n", sa.content)
	fmt.Println("==========")

	// 添加消息到消息队列中
	sa.messages = append(sa.messages, assistantMessages)

	// 结束对话的时候可以设置结束对话，意味着本轮（不是本次）对话结束
	sa.isStop = true
	sa.ClearContext()
}

// IsStop 判断当前对话轮次是否结束
func (sa *StreamAgent) IsStop() bool {
	return sa.isStop
}

// SetIsStop 设置当前对话轮次结束
func (sa *StreamAgent) SetIsStop(isStop bool) {
	sa.isStop = isStop
}

// Messages 获取本次对话的消息
func (sa *StreamAgent) Messages() []openai.ChatCompletionMessageParamUnion {
	return sa.messages
}

// AddMessages 添加多个消息到消息队列中
func (sa *StreamAgent) AddMessages(messages []openai.ChatCompletionMessageParamUnion) {
	sa.messages = append(sa.messages, messages...)
}

// AddMessage 添加消息到消息队列中
func (sa *StreamAgent) AddMessage(message openai.ChatCompletionMessageParamUnion) {
	sa.messages = append(sa.messages, message)

}

// ClearMessages 获取本次对话的消息
func (sa *StreamAgent) ClearMessages() {
	sa.messages = []openai.ChatCompletionMessageParamUnion{}
}
