package service

import (
	"context"
	"github.com/AdrianWangs/nexus/go-service/user/biz/dal/mysql"
	user_microservice "github.com/AdrianWangs/nexus/go-service/user/kitex_gen/user_microservice"
	"github.com/AdrianWangs/nexus/go-service/user/model"
)

type RegisterService struct {
	ctx context.Context
} // NewRegisterService new RegisterService
func NewRegisterService(ctx context.Context) *RegisterService {
	return &RegisterService{ctx: ctx}
}

// Run RegisterService 是用户注册方法，通过这里只能注册普通用户，也不能进行第三方账号注册
func (s *RegisterService) Run(request *user_microservice.RegisterRequest) (resp *user_microservice.RegisterResponse, err error) {

	resp = &user_microservice.RegisterResponse{}

	resp.Success = true
	resp.ErrorMessage = nil
	err = nil

	user := model.User{
		Username:        request.Username,
		Password:        request.Password,
		Birthday:        request.Birthday,
		Gender:          request.Gender,
		RoleId:          0, // 普通用户
		PhoneNumber:     request.PhoneNumber,
		Email:           request.Email,
		ThirdPartyToken: "",
	}

	err = mysql.DB.Create(&user).Error

	// 如果有错误，设置错误信息
	if err != nil {
		resp.Success = false
		resp.ErrorMessage = new(string)
		*resp.ErrorMessage = err.Error()
	}

	return
}
