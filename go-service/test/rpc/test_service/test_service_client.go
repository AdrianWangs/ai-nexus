package test_service

import (
	"context"
	user_microservice "github.com/AdrianWangs/ai-nexus/go-service/test/kitex_gen/user_microservice"

	"github.com/AdrianWangs/ai-nexus/go-service/test/kitex_gen/user_microservice/userservice"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
)

type RPCClient interface {
	KitexClient() userservice.Client
	Service() string
	Login(ctx context.Context, request *user_microservice.LoginRequest, callOptions ...callopt.Option) (r *user_microservice.LoginResponse, err error)
	Register(ctx context.Context, request *user_microservice.RegisterRequest, callOptions ...callopt.Option) (r *user_microservice.RegisterResponse, err error)
	ThirdPartyLogin(ctx context.Context, request *user_microservice.ThirdPartyLoginRequest, callOptions ...callopt.Option) (r *user_microservice.ThirdPartyLoginResponse, err error)
	UpdateUserProfile(ctx context.Context, request *user_microservice.UpdateUserRequest, callOptions ...callopt.Option) (r *user_microservice.UpdateUserResponse, err error)
	GetUser(ctx context.Context, request *user_microservice.GetUserRequest, callOptions ...callopt.Option) (r *user_microservice.GetUserResponse, err error)
}

func NewRPCClient(dstService string, opts ...client.Option) (RPCClient, error) {
	kitexClient, err := userservice.NewClient(dstService, opts...)
	if err != nil {
		return nil, err
	}
	cli := &clientImpl{
		service:     dstService,
		kitexClient: kitexClient,
	}

	return cli, nil
}

type clientImpl struct {
	service     string
	kitexClient userservice.Client
}

func (c *clientImpl) Service() string {
	return c.service
}

func (c *clientImpl) KitexClient() userservice.Client {
	return c.kitexClient
}

func (c *clientImpl) Login(ctx context.Context, request *user_microservice.LoginRequest, callOptions ...callopt.Option) (r *user_microservice.LoginResponse, err error) {
	return c.kitexClient.Login(ctx, request, callOptions...)
}

func (c *clientImpl) Register(ctx context.Context, request *user_microservice.RegisterRequest, callOptions ...callopt.Option) (r *user_microservice.RegisterResponse, err error) {
	return c.kitexClient.Register(ctx, request, callOptions...)
}

func (c *clientImpl) ThirdPartyLogin(ctx context.Context, request *user_microservice.ThirdPartyLoginRequest, callOptions ...callopt.Option) (r *user_microservice.ThirdPartyLoginResponse, err error) {
	return c.kitexClient.ThirdPartyLogin(ctx, request, callOptions...)
}

func (c *clientImpl) UpdateUserProfile(ctx context.Context, request *user_microservice.UpdateUserRequest, callOptions ...callopt.Option) (r *user_microservice.UpdateUserResponse, err error) {
	return c.kitexClient.UpdateUserProfile(ctx, request, callOptions...)
}

func (c *clientImpl) GetUser(ctx context.Context, request *user_microservice.GetUserRequest, callOptions ...callopt.Option) (r *user_microservice.GetUserResponse, err error) {
	return c.kitexClient.GetUser(ctx, request, callOptions...)
}
