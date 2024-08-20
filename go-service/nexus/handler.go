package main

import (
	nexus_microservice "github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice"
)

// NexusServiceImpl implements the last service interface defined in the IDL.
type NexusServiceImpl struct{}

func (s *NexusServiceImpl) EchoServer(req *nexus_microservice.AskRequest, stream nexus_microservice.NexusService_EchoServerServer) (err error) {
	println("EchoServer called")
	return
}
