package service

import (
	"context"
	"github.com/AdrianWangs/ai-nexus/go-service/user/biz/dal/model"
	"github.com/AdrianWangs/ai-nexus/go-service/user/biz/dal/mysql"
	"github.com/AdrianWangs/ai-nexus/go-service/user/biz/dal/query"
	user_microservice "github.com/AdrianWangs/ai-nexus/go-service/user/kitex_gen/user_microservice"
	"time"
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

	// 默认返回成功
	resp.Success = true
	resp.ErrorMessage = nil
	err = nil

	// 将 birthday 字符串转化为时间
	birthday, err := time.Parse("2006-01-02", request.Birthday)
	if err != nil {
		resp.Success = false
		resp.ErrorMessage = new(string)
		*resp.ErrorMessage = err.Error()
		return
	}
	user := model.User{
		Username:        request.Username,
		Password:        request.Password,
		Birthday:        birthday,
		Gender:          request.Gender,
		RoleID:          0, // 普通用户
		PhoneNumber:     request.PhoneNumber,
		Email:           request.Email,
		ThirdPartyToken: "",
	}

	// 获取用户表的查询对象
	userQuery := query.Use(mysql.DB).User

	userExistQuery := userQuery.WithContext(s.ctx).
		Where(userQuery.Username.Eq(user.Username))

	if user.PhoneNumber != "" {
		userExistQuery = userExistQuery.Or(userQuery.PhoneNumber.Eq(user.PhoneNumber))
	}

	if user.Email != "" {
		userExistQuery = userExistQuery.Or(userQuery.Email.Eq(user.Email))
	}

	// 判断是否已经存在
	userNum, err := userExistQuery.Count()

	if err != nil {
		resp.Success = false
		resp.ErrorMessage = new(string)
		*resp.ErrorMessage = err.Error()
		return
	}

	if userNum > 0 {
		resp.Success = false
		resp.ErrorMessage = new(string)
		*resp.ErrorMessage = "用户名或手机或邮箱已存在"
		return
	}

	// 插入用户
	err = userQuery.WithContext(s.ctx).Create(&user)

	if err != nil {
		resp.Success = false
		resp.ErrorMessage = new(string)
		*resp.ErrorMessage = err.Error()
		return
	}

	return
}
