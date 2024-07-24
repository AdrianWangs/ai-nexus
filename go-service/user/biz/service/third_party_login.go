package service

import (
	"context"
	user_microservice "github.com/AdrianWangs/nexus/go-service/user/kitex_gen/user_microservice"
)

type ThirdPartyLoginService struct {
	ctx context.Context
} // NewThirdPartyLoginService new ThirdPartyLoginService
func NewThirdPartyLoginService(ctx context.Context) *ThirdPartyLoginService {
	return &ThirdPartyLoginService{ctx: ctx}
}

// Run create note info
func (s *ThirdPartyLoginService) Run(request *user_microservice.ThirdPartyLoginRequest) (resp *user_microservice.ThirdPartyLoginResponse, err error) {
	// Finish your business logic.

	return
}
