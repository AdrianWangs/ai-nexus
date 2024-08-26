// Package nexus @Author Adrian.Wang 2024/8/26 下午8:09:00
package nexus

import (
	"fmt"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/ssestream"
)

// StreamAgent 是一个流代理，用于处理流的请求
// 并且在流处理过程中要进行函数调用和其他中间处理过程
type StreamAgent struct {
}

// ForwardResponse  转发响应请求并进行中间处理
func (sa *StreamAgent) ForwardResponse(source *ssestream.Stream[openai.ChatCompletionChunk], target nexus_microservice.NexusService_AskServerServer) {
	// 开始对话,使用代理模式进行对话
	for source.Next() {
		event := source.Current()
		if len(event.Choices) <= 0 {
			continue
		}

		//TODO 在获得到 event 后，判断是否是函数调用，如果是则需要进行函数调用
		askResponse := Event2response(event)
		fmt.Println("resp:", askResponse)
		err := target.Send(askResponse)
		if err != nil {
			fmt.Println("EchoServer failed: %v", err)
		}

	}
}
