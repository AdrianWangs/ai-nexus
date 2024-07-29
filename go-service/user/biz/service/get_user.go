package service

import (
	"context"
	user_microservice "github.com/AdrianWangs/nexus/go-service/user/kitex_gen/user_microservice"
)

type GetUserService struct {
	ctx context.Context
} // NewGetUserService new GetUserService
func NewGetUserService(ctx context.Context) *GetUserService {
	return &GetUserService{ctx: ctx}
}

// Run create note info
func (s *GetUserService) Run(request *user_microservice.GetUserRequest) (resp *user_microservice.GetUserResponse, err error) {

	resp = &user_microservice.GetUserResponse{}

	resp.Success = true
	resp.ErrorMessage = nil
	resp.UserProfile = &user_microservice.User{
		Username:    "test",
		Password:    "test",
		Email:       "",
		PhoneNumber: "",
	}

	err = nil
	return
}
