package service

import (
	"context"
	user_microservice "github.com/AdrianWangs/ai-nexus/go-service/user/kitex_gen/user_microservice"
)

type LoginService struct {
	ctx context.Context
} // NewLoginService new LoginService
func NewLoginService(ctx context.Context) *LoginService {
	return &LoginService{ctx: ctx}
}

// Run create note info
func (s *LoginService) Run(request *user_microservice.LoginRequest) (resp *user_microservice.LoginResponse, err error) {
	return nil, nil
}
