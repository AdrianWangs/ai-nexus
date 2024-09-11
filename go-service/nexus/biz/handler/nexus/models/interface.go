// @Author Adrian.Wang 2024/8/25 下午11:59:00
package models

import (
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/ssestream"
)

type Nexus interface {
	NewStream() *ssestream.Stream[openai.ChatCompletionChunk]
	Init(baseUrl string, apiKey string, model string)
	SetPrompt(prompt string)
	SetModel(model string)
	SetMessages(messages []openai.ChatCompletionMessageParamUnion)
	AddMessages(message openai.ChatCompletionMessageParamUnion)
	Messages()
	SetTools(tools []openai.ChatCompletionToolParam)
	AddTools(tools openai.ChatCompletionToolParam)
	Tools()
}
