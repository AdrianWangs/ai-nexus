package service

import (
	"context"
	"github.com/AdrianWangs/nexus/go-service/user/biz/dal/model"
	"github.com/AdrianWangs/nexus/go-service/user/biz/dal/mysql"
	"github.com/AdrianWangs/nexus/go-service/user/biz/dal/query"
	user_microservice "github.com/AdrianWangs/nexus/go-service/user/kitex_gen/user_microservice"
	"time"
)

type UpdateUserProfileService struct {
	ctx context.Context
} // NewUpdateUserProfileService new UpdateUserProfileService
func NewUpdateUserProfileService(ctx context.Context) *UpdateUserProfileService {
	return &UpdateUserProfileService{ctx: ctx}
}

// Run create note info
func (s *UpdateUserProfileService) Run(request *user_microservice.UpdateUserRequest) (resp *user_microservice.UpdateUserResponse, err error) {

	// 默认返回成功
	resp = &user_microservice.UpdateUserResponse{}
	resp.Success = true
	resp.ErrorMessage = nil

	// 判断用户是否存在
	userQuery := query.Use(mysql.DB).User
	user, err := userQuery.WithContext(s.ctx).Where(userQuery.ID.Eq(request.UserId)).First()

	if err != nil {
		resp.Success = false
		resp.ErrorMessage = new(string)
		*resp.ErrorMessage = err.Error()
		return
	}

	if user == nil {
		resp.Success = false
		resp.ErrorMessage = new(string)
		*resp.ErrorMessage = "user not found"
		return
	}

	// 将生日从字符串转换为时间
	birthday, err := time.Parse("2006-01-02", *request.Birthday)

	updatedUser := &model.User{
		Username:        user.Username, //用户名不能修改
		Password:        *request.Password,
		Birthday:        birthday,
		Gender:          *request.Gender,
		RoleID:          user.RoleID, //角色不能修改
		PhoneNumber:     *request.PhoneNumber,
		Email:           *request.Email,
		ThirdPartyToken: user.ThirdPartyToken, //第三方登录token不能修改
	}

	// 更新用户信息
	userQuery = query.Use(mysql.DB).User
	err = userQuery.WithContext(s.ctx).Where(userQuery.ID.Eq(request.UserId)).Save(updatedUser)

	if err != nil {
		resp.Success = false
		resp.ErrorMessage = new(string)
		*resp.ErrorMessage = err.Error()
	}

	return
}
