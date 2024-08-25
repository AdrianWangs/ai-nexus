package main

import (
	"fmt"
	nexus_microservice "github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice"
)

// NexusServiceImpl implements the last service interface defined in the IDL.
type NexusServiceImpl struct {
}

func (s *NexusServiceImpl) AskServer(req *nexus_microservice.AskRequest, stream nexus_microservice.NexusService_AskServerServer) (err error) {
	fmt.Print("EchoServer")

	fmt.Println(req)

	choices := make([]*nexus_microservice.Choice, 0)

	message := &nexus_microservice.Message{
		Role:    "System",
		Content: "你好，我是AI助手，有什么可以帮助你的吗？",
	}

	choice := &nexus_microservice.Choice{
		FinishReason: nil,
		Message:      []*nexus_microservice.Message{message},
		Index:        0,
	}

	choices = append(choices, choice)

	askResponse := &nexus_microservice.AskResponse{
		Id:      "",
		Model:   "",
		Choices: choices,
	}

	err = stream.Send(askResponse)

	defer stream.Close()

	if err != nil {
		fmt.Println("EchoServer failed: %v", err)
	}

	return
}
