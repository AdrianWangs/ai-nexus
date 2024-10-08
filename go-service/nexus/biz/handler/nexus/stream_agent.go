// Package nexus @Author Adrian.Wang 2024/8/26 下午8:09:00
package nexus

import (
	"errors"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/biz/handler/nexus/models"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice"
	"github.com/cloudwego/hertz/pkg/common/json"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/niemeyer/pretty"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/ssestream"
)

type ToolParam struct {
	Id                string
	Type              string
	FunctionName      string
	FunctionArguments string
}

// StreamAgent 是一个流代理，用于处理流的请求
// 并且在流处理过程中要进行函数调用和其他中间处理过程
type StreamAgent struct {
	toolMap      map[int64]ToolParam
	currentTool  ToolParam
	content      string
	messages     []openai.ChatCompletionMessageParamUnion
	isStop       bool                                            //多轮对话控制结束对话
	outputStream nexus_microservice.NexusService_AskServerServer //用于输出的流对象
}

// NewStreamAgent 用于生成一个新的流代理对象
func NewStreamAgent() *StreamAgent {
	streamAgent := &StreamAgent{
		isStop: false,
	}
	streamAgent.Init()
	return streamAgent
}

// Init 流代理初始化，也就是将消息设置为空
func (sa *StreamAgent) Init() {
	sa.messages = []openai.ChatCompletionMessageParamUnion{}
	sa.toolMap = make(map[int64]ToolParam)
}

// ForwardResponse  转发响应请求并进行中间处理
// 相当于中间件+流转发器
func (sa *StreamAgent) ForwardResponse(source *ssestream.Stream[openai.ChatCompletionChunk], target nexus_microservice.NexusService_AskServerServer, req *nexus_microservice.AskRequest) {

	sa.outputStream = target

	// 开始对话,使用代理模式进行对话
	for source.Next() {

		event := source.Current()

		// 如果本轮对话没有任何回复就不需要进行其他额外的操作了
		if len(event.Choices) <= 0 {
			klog.Error("好像没对话内容...")
			pretty.Println(event)
			continue
		}

		// 将 openai 传过来的数据转化成我们网站对应的 response 格式
		askResponse := Event2response(event)

		// 监控流，在监控过程中函数生成成功的那一刻进行函数调用
		// Monitor中会执行消息插入操作
		sa.Monitor(event, target, req)

		// 不输出函数相关的内容，等函数生成完毕，才开始调用
		if len(askResponse.Choices[0].Message[0].ToolCalls) > 0 {
			continue
		}

		// 监控完以后该转发刚刚的对话了
		err := target.Send(askResponse)
		if err != nil {
			klog.Error("ForwardResponse--> 发送给用户的响应 :    执行错误: ", err)
		}
	}

	if err := source.Err(); err != nil {
		klog.Error("StreamAgent ForwardResponse error:", err)
		klog.Error("最终暂停处主模型 messages:")
		pretty.Println(models.DefaultQwenInstance.Messages())
		pretty.Println(sa.messages)
		sa.isStop = true
	}

	// 创建一个消息添加到消息列表中

}

// Monitor 监控流的请求,并执行相关函数调用
func (sa *StreamAgent) Monitor(event openai.ChatCompletionChunk, target nexus_microservice.NexusService_AskServerServer, req *nexus_microservice.AskRequest) {

	// 结束对话
	if event.Choices[0].FinishReason == openai.ChatCompletionChunkChoicesFinishReasonStop {

		klog.Debug("=====================结束本轮对话=====================")
		// 结束本轮对话
		sa.EndConversation()

		return
	}

	// 当函数调用相关的参数生成完毕后，进行函数调用
	if event.Choices[0].FinishReason == openai.ChatCompletionChunkChoicesFinishReasonFunctionCall ||
		event.Choices[0].FinishReason == openai.ChatCompletionChunkChoicesFinishReasonToolCalls {

		finishReason := string(event.Choices[0].FinishReason)

		klog.Debug("结束函数调用：", finishReason)

		for _, toolParam := range sa.toolMap {
			sa.currentTool = toolParam
			// 生成响应，告诉前端当前正在调用函数
			functionCallResponse := sa.GenerateToolMessageResponse(finishReason)
			// 监控完以后该转发刚刚的对话了
			err := target.Send(functionCallResponse)
			if err != nil {
				klog.Error("Monitor--> 发送给用户的响应 :    执行错误: ", err)
			}

			// 调用服务，可能涉及子 ai 调用，所以要把流对象和相关请求一起传入
			sa.CallService(target, req)

			// 清空上下文，防止前面流影响后面的操作
			sa.ClearContext()

		}

		return
	}

	delta := event.Choices[0].Delta

	if delta.Content != "" {

		// 打印对话内容
		//klog.Debug("主 ai 的对话： ", delta.Content)
		sa.content += delta.Content

	}

	// 没有调用,直接返回
	if len(delta.ToolCalls) <= 0 {
		return
	}

	for _, toolCall := range delta.ToolCalls {
		toolCallChunk := toolCall

		// 完善函数调用相关的信息，也就是切片组合成完整信息
		sa.MergeFunctionCallChunks(toolCallChunk)
	}

}

// CallService 调用服务
func (sa *StreamAgent) CallService(target nexus_microservice.NexusService_AskServerServer, req *nexus_microservice.AskRequest) {

	if sa.currentTool.Type == "" {
		sa.currentTool.Type = "function"
	}

	// 返回机器人的消息，插入到消息队列中,一般是指明一个函数调用操作
	assistantMessages := sa.GenerateAssistantMessage()

	// 将消息添加到消息列表中
	sa.messages = append(sa.messages, assistantMessages)

	// 调用次级 ai
	// 次级ai 会进行额外的信息插入的操作
	res, err := sa.DoService(target, req)

	if err != nil {
		klog.Error("服务调用失败:", err)
	}

	klog.Debug("服务调用结果:", res)

	// 返回结果，要告知主 ai已经调用完毕
	sa.messages = append(sa.messages, sa.GenerateToolMessage(res))

}

// DoService 执行相关服务，调用服务相关的流，交由下一级ai 进行函数调用
func (sa *StreamAgent) DoService(target nexus_microservice.NexusService_AskServerServer, req *nexus_microservice.AskRequest) (res string, er error) {

	// 需要调用的服务名称
	serviceName := sa.currentTool.FunctionName
	// ai 生成给次级 ai 的提示词
	arguments := sa.currentTool.FunctionArguments

	klog.Debug("调用服务:", serviceName)
	klog.Debug("请求的提示词：", arguments)
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
	return AskService(serviceName, nexusPrompt, req, target, sa)
}

// MergeFunctionCallChunks 拼接各个函数调用相关的流切片以
// 完善函数调用相关的信息，主要负责拼接流分片中的函数调用相关的信息
func (sa *StreamAgent) MergeFunctionCallChunks(toolCallChunk openai.ChatCompletionChunkChoicesDeltaToolCall) {

	// 判断是否已经有这个 id 的 functionCall 在哈希表中
	if _, ok := sa.toolMap[toolCallChunk.Index]; !ok {
		sa.toolMap[toolCallChunk.Index] = ToolParam{}
	}

	toolParam := sa.toolMap[toolCallChunk.Index]

	// 函数调用类型（不知道有啥用）
	toolParam.Type = string(toolCallChunk.Type)

	// 函数调用 id（不知道有啥用）
	if toolCallChunk.ID != "" {
		klog.Debug("合并函数调用 ID:", toolCallChunk.ID)
		toolParam.Id = toolCallChunk.ID
	}

	// 函数调用名称
	function := toolCallChunk.Function
	if function.Name != "" {
		klog.Debug("合并函数调用名称:", function.Name)
		toolParam.FunctionName += function.Name
	}

	// 函数调用参数
	if function.Arguments != "" {
		klog.Debug("合并函数调用参数:", function.Arguments)
		toolParam.FunctionArguments += function.Arguments
	}

	sa.toolMap[toolCallChunk.Index] = toolParam
}

// GenerateAssistantMessage 根据本次（不是本轮，一轮有多次对话）对话的生成机器人的消息格式
func (sa *StreamAgent) GenerateAssistantMessage() openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{
		Content:      sa.content,
		Role:         openai.ChatCompletionMessageRoleAssistant,
		FunctionCall: openai.ChatCompletionMessageFunctionCall{},
		ToolCalls: []openai.ChatCompletionMessageToolCall{
			{
				ID:   sa.currentTool.Id,
				Type: openai.ChatCompletionMessageToolCallType(sa.currentTool.Type),
				Function: openai.ChatCompletionMessageToolCallFunction{
					Arguments: sa.currentTool.FunctionArguments,
					Name:      sa.currentTool.FunctionName,
				},
			},
		},
	}
}

// GenerateToolMessage 生成工具类型的消息
func (sa *StreamAgent) GenerateToolMessage(res string) openai.ChatCompletionToolMessageParam {
	toolMessage := openai.ToolMessage(sa.currentTool.Id, res)
	return toolMessage
}

// ClearContext 清空本次对话的上下文
func (sa *StreamAgent) ClearContext() {
	sa.toolMap = make(map[int64]ToolParam)
	sa.content = ""
}

// EndConversation 结束对话
func (sa *StreamAgent) EndConversation() {
	// 返回机器人的消息，插入到消息队列中
	assistantMessages := sa.GenerateAssistantMessage()

	klog.Info("==========")
	klog.Info("结束对话,最终本轮对话：\n", sa.content)
	klog.Info("==========")

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

// GenerateToolMessageResponse 生成用于响应流的 FunctionCall 消息
func (sa *StreamAgent) GenerateToolMessageResponse(reason string) *nexus_microservice.AskResponse {
	return &nexus_microservice.AskResponse{
		Id:    "",
		Model: "",
		Choices: []*nexus_microservice.Choice{
			{
				FinishReason: &reason,
				Message: []*nexus_microservice.Message{
					{
						Role:    "assistant",
						Content: "正在调用函数...",
						ToolCalls: []*nexus_microservice.ToolCall{
							{
								Id:   sa.currentTool.Id,
								Type: sa.currentTool.Type,
								FunctionCall: &nexus_microservice.FunctionCall{
									Name:      sa.currentTool.FunctionName,
									Arguments: &sa.currentTool.FunctionArguments,
								},
							},
						},
					},
				},
			},
		},
	}
}
