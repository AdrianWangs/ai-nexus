package service

import (
	"context"
	common_config "github.com/AdrianWangs/nexus/go-common/conf"
	"github.com/AdrianWangs/nexus/go-service/user/biz/dal/mysql"
	user_microservice "github.com/AdrianWangs/nexus/go-service/user/kitex_gen/user_microservice"
	"github.com/AdrianWangs/nexus/go-service/user/model"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/jwt"
	"gorm.io/gorm"
	"time"
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

	//// 登录成功,生成token
	//token, err := GenerateToken(user)
	//if err != nil {
	//	resp = &user_microservice.LoginResponse{}
	//	resp.Success = false
	//	resp.ErrorMessage = new(string)
	//	*resp.ErrorMessage = "token 生成失败"
	//	return
	//}
	//
	//resp = &user_microservice.LoginResponse{
	//	Success: true,
	//	Token:   token,
	//	UserProfile: &user_microservice.User{
	//		UserId:      int64(user.ID),
	//		Username:    user.Username,
	//		Email:       user.Email,
	//		PhoneNumber: user.PhoneNumber,
	//	},
	//}
	return nil, nil
}

func LoginResponse(ctx context.Context, c *app.RequestContext, code int, token string, expire time.Time) {
	inputUserNameOrEmail := ctx.Value("usernameOrEmail").(string)
	inputPassword := ctx.Value("password").(string)

	user := model.User{}

	// 判断是否存在该用户
	mysql.DB.Find(&user, "username = ? OR email = ? OR phone_number = ?",
		inputUserNameOrEmail, inputUserNameOrEmail, inputUserNameOrEmail,
	).First(&user)

	// 判断密码是否正确
	if inputPassword != user.Password {
		c.JSON(code, map[string]string{
			"message": "用户名或密码错误",
			"code":    "400",
		})
		return
	}

	c.JSON(code, map[string]interface{}{
		"token":  token,
		"expire": expire.Format(time.RFC3339),
		"user": map[string]interface{}{
			"userId":      user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"phoneNumber": user.PhoneNumber,
		},
	})

}

// GetAuthMiddleWare get auth middleware
func GetAuthMiddleWare() (*jwt.HertzJWTMiddleware, error) {

	jwt_config := common_config.GetConf().JWT

	authMiddleware, err := jwt.New(&jwt.HertzJWTMiddleware{
		Key:        []byte(jwt_config.Secret),                // 用于签名的密钥
		Timeout:    time.Duration(jwt_config.Timeout),        // token 过期时间,
		MaxRefresh: time.Duration(jwt_config.RefreshTimeout), // token 最大刷新时间
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			// 将 User 结构体转换为 jwt.MapClaims
			if v, ok := data.(model.User); ok {
				return jwt.MapClaims{
					"userId":      v.ID,
					"username":    v.Username,
					"email":       v.Email,
					"phoneNumber": v.PhoneNumber,
					"roleId":      v.RoleId,
				}
			}
			return jwt.MapClaims{}
		}, // 用于生成 token 的函数
		IdentityHandler: func(ctx context.Context, c *app.RequestContext) interface{} {
			claims := jwt.ExtractClaims(ctx, c)
			return &model.User{
				Model: gorm.Model{
					ID: uint(claims["userId"].(float64)),
				},
				Username:    claims["username"].(string),
				Email:       claims["email"].(string),
				PhoneNumber: claims["phoneNumber"].(string),
				RoleId:      int32(claims["roleId"].(float64)),
			} // 用于提取 token 中的数据
		}, // 用于提取 token 中的数据
		Authenticator: func(ctx context.Context, c *app.RequestContext) (interface{}, error) {
			var loginVals user_microservice.LoginRequest
			if err := c.BindAndValidate(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			userName := loginVals.UsernameOrEmail
			password := loginVals.Password
			user := model.User{}
			mysql.DB.Find(&user, "username = ? OR email = ? OR phone_number = ?", userName, userName, userName).First(&user)
			if user.Password == password {
				return user, nil
			}
			return nil, jwt.ErrFailedAuthentication
		}, // 用于验证用户身份的函数
		Authorizator: func(data interface{}, ctx context.Context, c *app.RequestContext) bool {
			if _, ok := data.(*model.User); ok {
				return true
			}
			return false
		}, // 用于验证用户权限的函数
		Unauthorized: func(ctx context.Context, c *app.RequestContext, code int, message string) {
			c.JSON(code, map[string]string{"message": message})
		}, // 用于处理未授权的请求
		LoginResponse: LoginResponse, // 用于处理登录成功的请求
		RefreshResponse: func(ctx context.Context, c *app.RequestContext, code int, token string, expire time.Time) {
			c.JSON(code, map[string]string{
				"token":  token,
				"expire": expire.Format(time.RFC3339),
			})
		}, // 用于处理 token 刷新成功的请求
	})

	if err != nil {
		return nil, err
	}

	return authMiddleware, nil

}
