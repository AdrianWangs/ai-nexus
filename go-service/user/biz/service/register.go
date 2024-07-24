package service

import (
	"context"
	user_microservice "github.com/AdrianWangs/nexus/go-service/user/kitex_gen/user_microservice"
)

type RegisterService struct {
	ctx context.Context
} // NewRegisterService new RegisterService
func NewRegisterService(ctx context.Context) *RegisterService {
	return &RegisterService{ctx: ctx}
}

// Run create note info
func (s *RegisterService) Run(request *user_microservice.RegisterRequest) (resp *user_microservice.RegisterResponse, err error) {
	// Finish your business logic.

	return
}
