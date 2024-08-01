// Code generated by Kitex v0.9.1. DO NOT EDIT.

package userservice

import (
	"context"
	user_microservice "github.com/AdrianWangs/ai-nexus/go-service/user/kitex_gen/user_microservice"
	client "github.com/cloudwego/kitex/client"
	callopt "github.com/cloudwego/kitex/client/callopt"
)

// Client is designed to provide IDL-compatible methods with call-option parameter for kitex framework.
type Client interface {
	Login(ctx context.Context, request *user_microservice.LoginRequest, callOptions ...callopt.Option) (r *user_microservice.LoginResponse, err error)
	Register(ctx context.Context, request *user_microservice.RegisterRequest, callOptions ...callopt.Option) (r *user_microservice.RegisterResponse, err error)
	ThirdPartyLogin(ctx context.Context, request *user_microservice.ThirdPartyLoginRequest, callOptions ...callopt.Option) (r *user_microservice.ThirdPartyLoginResponse, err error)
	UpdateUserProfile(ctx context.Context, request *user_microservice.UpdateUserRequest, callOptions ...callopt.Option) (r *user_microservice.UpdateUserResponse, err error)
	GetUser(ctx context.Context, request *user_microservice.GetUserRequest, callOptions ...callopt.Option) (r *user_microservice.GetUserResponse, err error)
}

// NewClient creates a client for the service defined in IDL.
func NewClient(destService string, opts ...client.Option) (Client, error) {
	var options []client.Option
	options = append(options, client.WithDestService(destService))

	options = append(options, opts...)

	kc, err := client.NewClient(serviceInfoForClient(), options...)
	if err != nil {
		return nil, err
	}
	return &kUserServiceClient{
		kClient: newServiceClient(kc),
	}, nil
}

// MustNewClient creates a client for the service defined in IDL. It panics if any error occurs.
func MustNewClient(destService string, opts ...client.Option) Client {
	kc, err := NewClient(destService, opts...)
	if err != nil {
		panic(err)
	}
	return kc
}

type kUserServiceClient struct {
	*kClient
}

func (p *kUserServiceClient) Login(ctx context.Context, request *user_microservice.LoginRequest, callOptions ...callopt.Option) (r *user_microservice.LoginResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.Login(ctx, request)
}

func (p *kUserServiceClient) Register(ctx context.Context, request *user_microservice.RegisterRequest, callOptions ...callopt.Option) (r *user_microservice.RegisterResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.Register(ctx, request)
}

func (p *kUserServiceClient) ThirdPartyLogin(ctx context.Context, request *user_microservice.ThirdPartyLoginRequest, callOptions ...callopt.Option) (r *user_microservice.ThirdPartyLoginResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.ThirdPartyLogin(ctx, request)
}

func (p *kUserServiceClient) UpdateUserProfile(ctx context.Context, request *user_microservice.UpdateUserRequest, callOptions ...callopt.Option) (r *user_microservice.UpdateUserResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.UpdateUserProfile(ctx, request)
}

func (p *kUserServiceClient) GetUser(ctx context.Context, request *user_microservice.GetUserRequest, callOptions ...callopt.Option) (r *user_microservice.GetUserResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.GetUser(ctx, request)
}
