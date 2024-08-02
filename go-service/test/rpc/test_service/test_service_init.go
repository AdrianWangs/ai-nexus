package test_service

import (
	"sync"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/transport"
)

var (
	// todo edit custom config
	defaultClient     RPCClient
	defaultDstService = "test_service"
	defaultClientOpts = []client.Option{
		client.WithHostPorts("127.0.0.1:8848"),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
		client.WithTransportProtocol(transport.TTHeader),
	}
	once sync.Once
)

func init() {
	DefaultClient()
}

func DefaultClient() RPCClient {
	once.Do(func() {
		defaultClient = newClient(defaultDstService, defaultClientOpts...)
	})
	return defaultClient
}

func newClient(dstService string, opts ...client.Option) RPCClient {
	c, err := NewRPCClient(dstService, opts...)
	if err != nil {
		panic("failed to init client: " + err.Error())
	}
	return c
}

func InitClient(dstService string, opts ...client.Option) {
	defaultClient = newClient(dstService, opts...)
}
