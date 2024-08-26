// @Author Adrian.Wang 2024/8/25 下午11:50:00
package nexus

import (
	"context"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
	"reflect"
)

var QwenInstance Qwen

type Qwen struct {
	baseUrl  string
	apiKey   string
	model    string
	prompt   string
	client   *openai.Client
	messages []openai.ChatCompletionMessageParamUnion
	tools    []openai.ChatCompletionToolParam
	params   openai.ChatCompletionNewParams
}

// Init 初始化
func (nexus *Qwen) Init(baseUrl string, apiKey string) {

	//// 本地 ollama
	//baseUrl = "http://localhost:11434/v1/"
	//apiKey = "ollama"
	//model = "llama3.1:8b"

	// 通义大模型
	nexus.baseUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1/"
	nexus.apiKey = "" // 自行去官网申请 apiKey
	nexus.SetModel("qwen-max")

	nexus.client = openai.NewClient(
		option.WithBaseURL(baseUrl),
		option.WithAPIKey(apiKey), // defaults to os.LookupEnv("OPENAI_API_KEY")
	)

}

// SetModel 设置模型
func (nexus *Qwen) SetModel(model string) {
	nexus.model = model
}

// SetPrompt 设置系统提示词
func (nexus *Qwen) SetPrompt(prompt string) {
	nexus.prompt = prompt
	nexus.messages = []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(nexus.prompt),
	}
}

// Messages 获取消息
func (nexus *Qwen) Messages() []openai.ChatCompletionMessageParamUnion {
	return nexus.messages
}

// SetMessages 设置消息
func (nexus *Qwen) SetMessages(messages []openai.ChatCompletionMessageParamUnion) {

	systemMessage := []openai.ChatCompletionMessageParamUnion{}

	// 判断第一条消息的类型是不是系统消息，注意 messages 切片是一个接口数组，要用反射判断
	if len(messages) > 0 {
		// 获取结构体的类型
		t := reflect.TypeOf(messages[0])
		// 判断是否是系统消息
		if t.Name() != "SystemMessage" {
			systemMessage = append(systemMessage, openai.SystemMessage(nexus.prompt))
		}
	}

	nexus.messages = append(systemMessage, messages...)
}

// AddMessages 添加消息
func (nexus *Qwen) AddMessages(message openai.ChatCompletionMessageParamUnion) {
	nexus.messages = append(nexus.messages, message)
}

// Tools 获取工具
func (nexus *Qwen) Tools() []openai.ChatCompletionToolParam {
	return nexus.tools
}

// SetTools 设置工具
func (nexus *Qwen) SetTools(tools []openai.ChatCompletionToolParam) {
	nexus.tools = tools
}

// AddTools 添加工具
func (nexus *Qwen) AddTools(tool openai.ChatCompletionToolParam) {
	nexus.tools = append(nexus.tools, tool)
}

// NewStream 创建流并返回
func (nexus *Qwen) NewStream() *ssestream.Stream[openai.ChatCompletionChunk] {

	ctx := context.Background()

	nexus.params = openai.ChatCompletionNewParams{
		Model:    openai.F(nexus.model),
		Messages: openai.F(nexus.messages),
		Tools:    openai.F(nexus.tools),
	}
	return nexus.client.Chat.Completions.NewStreaming(ctx, nexus.params)

}
