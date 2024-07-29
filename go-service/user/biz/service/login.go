package service

import (
	"context"
	"github.com/AdrianWangs/nexus/go-service/user/biz/dal/mysql"
	user_microservice "github.com/AdrianWangs/nexus/go-service/user/kitex_gen/user_microservice"
	"github.com/AdrianWangs/nexus/go-service/user/model"
)

type LoginService struct {
	ctx context.Context
} // NewLoginService new LoginService
func NewLoginService(ctx context.Context) *LoginService {
	return &LoginService{ctx: ctx}
}

// Run create note info
func (s *LoginService) Run(request *user_microservice.LoginRequest) (resp *user_microservice.LoginResponse, err error) {

	inputUserNameOrEmail := request.UsernameOrEmail
	inputPassword := request.Password

	user := model.User{}

	// 判断是否存在该用户
	mysql.DB.Find(&user, "username = ? OR email = ? OR phone_number = ?",
		inputUserNameOrEmail, inputUserNameOrEmail, inputUserNameOrEmail,
	).First(&user)

	// 判断密码是否正确
	if inputPassword != user.Password {
		resp = &user_microservice.LoginResponse{}
		resp.Success = false
		resp.ErrorMessage = new(string)
		*resp.ErrorMessage = "用户名或密码错误"
		return
	}

	// 登录成功,生成token
	token, err := model.GenerateToken(user.ID)
	if err != nil {
		resp = &user_microservice.LoginResponse{}
		resp.Success = false
		resp.ErrorMessage = new(string)
		*resp.ErrorMessage = "token 生成失败"
		return
	}

	resp = &user_microservice.LoginResponse{
		Success: true,
		Token: token,
		UserProfile: &user_microservice.User{
			Username: user.Username,
			Email: user.Email,
			PhoneNumber: user.PhoneNumber,
			ID: user.ID,
		}
	}

}
