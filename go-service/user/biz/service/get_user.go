package service

import (
	"context"
	"github.com/AdrianWangs/ai-nexus/go-service/user/biz/dal/mysql"
	"github.com/AdrianWangs/ai-nexus/go-service/user/biz/dal/query"
	user_microservice "github.com/AdrianWangs/ai-nexus/go-service/user/kitex_gen/user_microservice"
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

	userId := request.UserId

	userQuery := query.Use(mysql.DB).User

	user, err := userQuery.WithContext(s.ctx).Where(userQuery.ID.Eq(userId)).First()

	if err != nil {
		resp.Success = false
		*resp.ErrorMessage = err.Error()
		resp.UserProfile = nil
		return
	}

	resp.Success = true
	resp.ErrorMessage = nil
	resp.UserProfile = &user_microservice.User{
		UserId:          user.ID,
		Username:        user.Username,
		Birthday:        user.Birthday.Format("2006-01-02"),
		Gender:          user.Gender,
		RoleId:          user.RoleID,
		PhoneNumber:     user.PhoneNumber,
		Email:           user.Email,
		ThirdPartyToken: &user.ThirdPartyToken,
	}

	err = nil
	return
}
