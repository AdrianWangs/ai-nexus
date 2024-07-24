package main

import (
	"context"
	"github.com/AdrianWangs/nexus/go-service/user/biz/service"
	user_microservice "github.com/AdrianWangs/nexus/go-service/user/kitex_gen/user_microservice"
)

// UserServiceImpl implements the last service interface defined in the IDL.
type UserServiceImpl struct{}

// Login implements the UserServiceImpl interface.
func (s *UserServiceImpl) Login(ctx context.Context, request *user_microservice.LoginRequest) (resp *user_microservice.LoginResponse, err error) {
	resp, err = service.NewLoginService(ctx).Run(request)

	return resp, err
}

// Register implements the UserServiceImpl interface.
func (s *UserServiceImpl) Register(ctx context.Context, request *user_microservice.RegisterRequest) (resp *user_microservice.RegisterResponse, err error) {
	resp, err = service.NewRegisterService(ctx).Run(request)

	return resp, err
}

// ThirdPartyLogin implements the UserServiceImpl interface.
func (s *UserServiceImpl) ThirdPartyLogin(ctx context.Context, request *user_microservice.ThirdPartyLoginRequest) (resp *user_microservice.ThirdPartyLoginResponse, err error) {
	resp, err = service.NewThirdPartyLoginService(ctx).Run(request)

	return resp, err
}

// UpdateUserProfile implements the UserServiceImpl interface.
func (s *UserServiceImpl) UpdateUserProfile(ctx context.Context, request *user_microservice.UpdateUserRequest) (resp *user_microservice.UpdateUserResponse, err error) {
	resp, err = service.NewUpdateUserProfileService(ctx).Run(request)

	return resp, err
}

// GetUser implements the UserServiceImpl interface.
func (s *UserServiceImpl) GetUser(ctx context.Context, request *user_microservice.GetUserRequest) (resp *user_microservice.GetUserResponse, err error) {
	resp, err = service.NewGetUserService(ctx).Run(request)

	return resp, err
}
