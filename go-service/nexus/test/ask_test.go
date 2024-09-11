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

	messages := []*nexus_microservice.Message{}

	messages = append(messages, &nexus_microservice.Message{
		Role:    "user",
		Content: "我想去苏州玩，帮我找点好玩的经典，然后帮我安排一下行程计划",
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

		if resp == nil {
			continue
		}

		t.Log("resp:", resp.Choices[0].Message[0].Content)
		toolCalls := resp.Choices[0].Message[0].ToolCalls

		for _, toolCall := range toolCalls {
			functionCall := toolCall.FunctionCall
			t.Log("type:", toolCall.Type)
			t.Log("toolCall:", functionCall.Name, "(", *functionCall.Arguments, ")")
		}

		t.Log("====================================")

	}

}
