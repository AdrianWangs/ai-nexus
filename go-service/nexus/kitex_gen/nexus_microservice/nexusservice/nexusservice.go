// Code generated by Kitex v0.10.3. DO NOT EDIT.

package nexusservice

import (
	"context"
	"errors"
	"fmt"
	nexus_microservice "github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice"
	client "github.com/cloudwego/kitex/client"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
	streaming "github.com/cloudwego/kitex/pkg/streaming"
)

var errInvalidMessageType = errors.New("invalid message type for service method handler")

var serviceMethods = map[string]kitex.MethodInfo{
	"AskServer": kitex.NewMethodInfo(
		askServerHandler,
		newNexusServiceAskServerArgs,
		newNexusServiceAskServerResult,
		false,
		kitex.WithStreamingMode(kitex.StreamingServer),
	),
}

var (
	nexusServiceServiceInfo                = NewServiceInfo()
	nexusServiceServiceInfoForClient       = NewServiceInfoForClient()
	nexusServiceServiceInfoForStreamClient = NewServiceInfoForStreamClient()
)

// for server
func serviceInfo() *kitex.ServiceInfo {
	return nexusServiceServiceInfo
}

// for stream client
func serviceInfoForStreamClient() *kitex.ServiceInfo {
	return nexusServiceServiceInfoForStreamClient
}

// for client
func serviceInfoForClient() *kitex.ServiceInfo {
	return nexusServiceServiceInfoForClient
}

// NewServiceInfo creates a new ServiceInfo containing all methods
func NewServiceInfo() *kitex.ServiceInfo {
	return newServiceInfo(true, true, true)
}

// NewServiceInfo creates a new ServiceInfo containing non-streaming methods
func NewServiceInfoForClient() *kitex.ServiceInfo {
	return newServiceInfo(false, false, true)
}
func NewServiceInfoForStreamClient() *kitex.ServiceInfo {
	return newServiceInfo(true, true, false)
}

func newServiceInfo(hasStreaming bool, keepStreamingMethods bool, keepNonStreamingMethods bool) *kitex.ServiceInfo {
	serviceName := "NexusService"
	handlerType := (*nexus_microservice.NexusService)(nil)
	methods := map[string]kitex.MethodInfo{}
	for name, m := range serviceMethods {
		if m.IsStreaming() && !keepStreamingMethods {
			continue
		}
		if !m.IsStreaming() && !keepNonStreamingMethods {
			continue
		}
		methods[name] = m
	}
	extra := map[string]interface{}{
		"PackageName": "nexus_microservice",
	}
	if hasStreaming {
		extra["streaming"] = hasStreaming
	}
	svcInfo := &kitex.ServiceInfo{
		ServiceName:     serviceName,
		HandlerType:     handlerType,
		Methods:         methods,
		PayloadCodec:    kitex.Thrift,
		KiteXGenVersion: "v0.10.3",
		Extra:           extra,
	}
	return svcInfo
}

func askServerHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	st, ok := arg.(*streaming.Args)
	if !ok {
		return errors.New("NexusService.AskServer is a thrift streaming method, please call with Kitex StreamClient")
	}
	stream := &nexusServiceAskServerServer{st.Stream}
	req := new(nexus_microservice.AskRequest)
	if err := st.Stream.RecvMsg(req); err != nil {
		return err
	}
	return handler.(nexus_microservice.NexusService).AskServer(req, stream)
}

type nexusServiceAskServerClient struct {
	streaming.Stream
}

func (x *nexusServiceAskServerClient) DoFinish(err error) {
	if finisher, ok := x.Stream.(streaming.WithDoFinish); ok {
		finisher.DoFinish(err)
	} else {
		panic(fmt.Sprintf("streaming.WithDoFinish is not implemented by %T", x.Stream))
	}
}
func (x *nexusServiceAskServerClient) Recv() (*nexus_microservice.AskResponse, error) {
	m := new(nexus_microservice.AskResponse)
	return m, x.Stream.RecvMsg(m)
}

type nexusServiceAskServerServer struct {
	streaming.Stream
}

func (x *nexusServiceAskServerServer) Send(m *nexus_microservice.AskResponse) error {
	return x.Stream.SendMsg(m)
}

func newNexusServiceAskServerArgs() interface{} {
	return nexus_microservice.NewNexusServiceAskServerArgs()
}

func newNexusServiceAskServerResult() interface{} {
	return nexus_microservice.NewNexusServiceAskServerResult()
}

type kClient struct {
	c client.Client
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c: c,
	}
}

func (p *kClient) AskServer(ctx context.Context, req *nexus_microservice.AskRequest) (NexusService_AskServerClient, error) {
	streamClient, ok := p.c.(client.Streaming)
	if !ok {
		return nil, fmt.Errorf("client not support streaming")
	}
	res := new(streaming.Result)
	err := streamClient.Stream(ctx, "AskServer", nil, res)
	if err != nil {
		return nil, err
	}
	stream := &nexusServiceAskServerClient{res.Stream}

	if err := stream.Stream.SendMsg(req); err != nil {
		return nil, err
	}
	if err := stream.Stream.Close(); err != nil {
		return nil, err
	}
	return stream, nil
}
