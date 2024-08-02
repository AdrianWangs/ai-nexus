package test_service

import (
	"context"
	user_microservice "github.com/AdrianWangs/ai-nexus/go-service/test/kitex_gen/user_microservice"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/klog"
)

func Login(ctx context.Context, request *user_microservice.LoginRequest, callOptions ...callopt.Option) (resp *user_microservice.LoginResponse, err error) {
	resp, err = defaultClient.Login(ctx, request, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "Login call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}

func Register(ctx context.Context, request *user_microservice.RegisterRequest, callOptions ...callopt.Option) (resp *user_microservice.RegisterResponse, err error) {
	resp, err = defaultClient.Register(ctx, request, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "Register call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}

func ThirdPartyLogin(ctx context.Context, request *user_microservice.ThirdPartyLoginRequest, callOptions ...callopt.Option) (resp *user_microservice.ThirdPartyLoginResponse, err error) {
	resp, err = defaultClient.ThirdPartyLogin(ctx, request, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "ThirdPartyLogin call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}

func UpdateUserProfile(ctx context.Context, request *user_microservice.UpdateUserRequest, callOptions ...callopt.Option) (resp *user_microservice.UpdateUserResponse, err error) {
	resp, err = defaultClient.UpdateUserProfile(ctx, request, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "UpdateUserProfile call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}

func GetUser(ctx context.Context, request *user_microservice.GetUserRequest, callOptions ...callopt.Option) (resp *user_microservice.GetUserResponse, err error) {
	resp, err = defaultClient.GetUser(ctx, request, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "GetUser call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}
