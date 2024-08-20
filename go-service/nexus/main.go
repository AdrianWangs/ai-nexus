package main

import (
	nexus_microservice "github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice/nexusservice"
	"log"
)

func main() {
	svr := nexus_microservice.NewServer(new(NexusServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
