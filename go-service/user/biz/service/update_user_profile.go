package service

import (
	"context"
	user_microservice "github.com/AdrianWangs/nexus/go-service/user/kitex_gen/user_microservice"
)

type UpdateUserProfileService struct {
	ctx context.Context
} // NewUpdateUserProfileService new UpdateUserProfileService
func NewUpdateUserProfileService(ctx context.Context) *UpdateUserProfileService {
	return &UpdateUserProfileService{ctx: ctx}
}

// Run create note info
func (s *UpdateUserProfileService) Run(request *user_microservice.UpdateUserRequest) (resp *user_microservice.UpdateUserResponse, err error) {
	// Finish your business logic.

	return
}
