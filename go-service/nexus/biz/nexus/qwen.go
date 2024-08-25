// @Author Adrian.Wang 2024/8/25 下午11:50:00
package nexus

import (
	"context"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
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

func (nexus *Qwen) SetModel(model string) {
	nexus.model = model
}

func (nexus *Qwen) SetPrompt(prompt string) {
	nexus.prompt = prompt
	nexus.messages = []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(nexus.prompt),
	}
}

func (nexus *Qwen) Messages() []openai.ChatCompletionMessageParamUnion {
	return nexus.messages
}

func (nexus *Qwen) SetMessages(messages []openai.ChatCompletionMessageParamUnion) {
	nexus.messages = messages
}

func (nexus *Qwen) AddMessages(message openai.ChatCompletionMessageParamUnion) {
	nexus.messages = append(nexus.messages, message)
}

func (nexus *Qwen) Tools() []openai.ChatCompletionToolParam {
	return nexus.tools
}

func (nexus *Qwen) SetTools(tools []openai.ChatCompletionToolParam) {
	nexus.tools = tools
}

func (nexus *Qwen) NewStream() *ssestream.Stream[openai.ChatCompletionChunk] {

	ctx := context.Background()

	nexus.params = openai.ChatCompletionNewParams{
		Model:    openai.F(nexus.model),
		Messages: openai.F(nexus.messages),
		Tools:    openai.F(nexus.tools),
	}
	return nexus.client.Chat.Completions.NewStreaming(ctx, nexus.params)
}
