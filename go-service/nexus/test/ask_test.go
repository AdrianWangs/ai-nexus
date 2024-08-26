// @Author Adrian.Wang 2024/8/23 下午5:03:00
package ask_test

import (
	"context"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice/nexusservice"
	"github.com/cloudwego/kitex/client/streamclient"
	"io"
	"testing"
)

func TestAsk(t *testing.T) {
	var streamClient = nexusservice.MustNewStreamClient(
		"nexus-service", // Service Name
		streamclient.WithHostPorts("127.0.0.1:8888"), // Service Address
	)

	messages := []*nexus_microservice.Message{{
		Role:    "system",
		Content: "你好，我是AI助手，有什么可以帮助你的吗？",
	}}

	messages = append(messages, &nexus_microservice.Message{
		Role:    "user",
		Content: "今天晚上我有什么安排？",
	})

	ctx := context.Background()
	askRequest := &nexus_microservice.AskRequest{
		Model:           nil,
		TopP:            nil,
		Temperature:     nil,
		PresencePenalty: nil,
		MaxTokens:       nil,
		Seed:            nil,
		Stop:            nil,
		EnableSearch:    nil,
		Messages:        messages,
	}
	stream, err := streamClient.AskServer(ctx, askRequest)

	if err != nil {
		t.Errorf("EchoServer failed: %v", err)
		return
	}

	for {
		resp, err := stream.Recv()

		if err != nil && err == io.EOF {
			break
		}

		if err != nil && err.Error() != "EOF" {
			t.Errorf("EchoServer failed: %v", err)
			break
		}

		t.Log("resp:", resp)

	}

}
