// Code generated by hertz generator.

package user_microservice

import (
	"context"
	"github.com/AdrianWangs/nexus/go-common/middleware"
	"github.com/AdrianWangs/nexus/go-service/user/biz/dal/model"
	"github.com/AdrianWangs/nexus/go-service/user/biz/service"
	user_microservice "github.com/AdrianWangs/nexus/go-service/user/kitex_gen/user_microservice"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Login 登录
// @Summary 登录
// @Description 登录并获取 token
// @Tags 用户服务
// @Accept application/json
// @Param UsernameOrEmail body string true "账号（邮箱、手机、用户名）"
// @Param Password body string true "密码"
// @router /login [POST]
func Login(ctx context.Context, c *app.RequestContext) {

	// 使用 jwt 中间件了，所以不需要再次验证
}

// Register .
// @router /register [POST]
func Register(ctx context.Context, c *app.RequestContext) {
	var err error
	var req user_microservice.RegisterRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}

	resp := new(user_microservice.RegisterResponse)

	c.JSON(consts.StatusOK, resp)
}

// ThirdPartyLogin .
// @router /third_party_login [POST]
func ThirdPartyLogin(ctx context.Context, c *app.RequestContext) {
	var err error
	var req user_microservice.ThirdPartyLoginRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}

	resp := new(user_microservice.ThirdPartyLoginResponse)

	c.JSON(consts.StatusOK, resp)
}

// UpdateUserProfile .
// @router /update_user_profile [POST]
func UpdateUserProfile(ctx context.Context, c *app.RequestContext) {
	var err error
	var req user_microservice.UpdateUserRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}

	resp := new(user_microservice.UpdateUserResponse)

	c.JSON(consts.StatusOK, resp)
}

// GetUser .
// @router /get_user [GET]
func GetUser(ctx context.Context, c *app.RequestContext) {
	var err error
	var req user_microservice.GetUserRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}

	// 获取用户信息中的 id
	user, exist := c.Get(middleware.IdentityKey)
	if !exist {
		c.String(consts.StatusNonAuthoritativeInfo, "用户未登录")
		return
	}

	req.UserId = user.(*model.User).ID

	resp, err := service.NewGetUserService(ctx).Run(&req)

	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}

	c.JSON(consts.StatusOK, resp)
}
