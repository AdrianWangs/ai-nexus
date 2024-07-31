// Package middleware @Author Adrian.Wang 2024/7/29 下午7:07:00
package middleware

import (
	"context"
	"github.com/AdrianWangs/nexus/go-service/user/biz/dal/model"
	"github.com/AdrianWangs/nexus/go-service/user/biz/dal/mysql"
	"github.com/cloudwego/kitex/pkg/klog"
	"net/http"
	"time"

	common_config "github.com/AdrianWangs/nexus/go-common/conf"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/hertz-contrib/jwt"
)

var (
	JwtMiddleware *jwt.HertzJWTMiddleware
	IdentityKey   = "userId"
)

func InitJwt() {

	jwt_config := common_config.GetConf().JWT

	timeout := time.Duration(jwt_config.Timeout) * time.Second
	maxRefresh := time.Duration(jwt_config.RefreshTimeout) * time.Second

	var err error
	JwtMiddleware, err = jwt.New(&jwt.HertzJWTMiddleware{

		Key:           []byte(jwt_config.Secret),
		Timeout:       timeout,
		MaxRefresh:    maxRefresh,
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		LoginResponse: func(ctx context.Context, c *app.RequestContext, code int, token string, expire time.Time) {
			c.JSON(http.StatusOK, utils.H{
				"code":    code,
				"token":   token,
				"expire":  expire.Format(time.RFC3339),
				"message": "success",
			})
		},
		Authenticator: func(ctx context.Context, c *app.RequestContext) (interface{}, error) {
			var loginStruct struct {
				UsernameOrEmail string `form:"UsernameOrEmail" json:"UsernameOrEmail" query:"UsernameOrEmail" vd:"(len($) > 0 && len($) < 30); msg:'Illegal format'"`
				Password        string `form:"Password" json:"Password" query:"Password" vd:"(len($) > 0 && len($) < 30); msg:'Illegal format'"`
			}

			klog.Info("Authenticator")

			// 验证参数
			if err := c.BindAndValidate(&loginStruct); err != nil {
				return nil, err
			}
			user, err := mysql.CheckUser(loginStruct.UsernameOrEmail, loginStruct.Password)
			if err != nil {
				klog.Error("验证用户失败")
				return nil, err
			}

			return user, nil
		}, //验证用户，返回用户信息
		IdentityKey: IdentityKey, // 存放在 token 中的 key
		IdentityHandler: func(ctx context.Context, c *app.RequestContext) interface{} {
			claims := jwt.ExtractClaims(ctx, c)
			return &model.User{
				ID: int64(claims[IdentityKey].(float64)),
			}
		}, // 从 token 中提取信息,主要是用户的 id
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*model.User); ok {
				return jwt.MapClaims{
					IdentityKey: v.ID,
				}
			}
			return jwt.MapClaims{}
		}, // 存放在 token 中的信息
		HTTPStatusMessageFunc: func(e error, ctx context.Context, c *app.RequestContext) string {
			hlog.CtxErrorf(ctx, "jwt biz err = %+v", e.Error())
			return e.Error()
		}, // 错误信息
		Unauthorized: func(ctx context.Context, c *app.RequestContext, code int, message string) {
			c.JSON(http.StatusOK, utils.H{
				"code":    code,
				"message": message,
			})
		}, // 未授权
	})
	if err != nil {
		klog.Error("初始化 jwt 失败")
		panic(err)
	}

	klog.Info("初始化 jwt 成功")
}
