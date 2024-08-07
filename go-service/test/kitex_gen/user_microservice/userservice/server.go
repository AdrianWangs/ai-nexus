// Code generated by Kitex v0.9.1. DO NOT EDIT.
package userservice

import (
	user_microservice "github.com/AdrianWangs/ai-nexus/go-service/test/kitex_gen/user_microservice"
	klog "github.com/cloudwego/kitex/pkg/klog"
	rpcinfo "github.com/cloudwego/kitex/pkg/rpcinfo"
	server "github.com/cloudwego/kitex/server"
	registry "github.com/kitex-contrib/registry-nacos/registry"
)

// NewServer creates a server.Server with the given handler and options.
func NewServer(handler user_microservice.UserService, opts ...server.Option) server.Server {
	var options []server.Option
	r, err := registry.NewDefaultNacosRegistry()
	if err != nil {
		klog.Fatal(err)
	}
	options = append(options, server.WithRegistry(r), server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
		ServiceName: "test_service",
	}))

	options = append(options, opts...)
	options = append(options, server.WithCompatibleMiddlewareForUnary())

	svr := server.NewServer(options...)
	if err := svr.RegisterService(serviceInfo(), handler); err != nil {
		panic(err)
	}
	return svr
}

func RegisterService(svr server.Server, handler user_microservice.UserService, opts ...server.RegisterOption) error {
	return svr.RegisterService(serviceInfo(), handler, opts...)
}
