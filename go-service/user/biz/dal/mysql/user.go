// @Author Adrian.Wang 2024/7/29 下午10:19:00
package mysql

import (
	"errors"
	"github.com/AdrianWangs/ai-nexus/go-service/user/biz/dal/model"
	"github.com/cloudwego/kitex/pkg/klog"
)

func CreateUsers(users []*model.User) error {
	return DB.Create(users).Error
}

func FindUserByNameOrEmail(userName, email string) ([]*model.User, error) {
	res := make([]*model.User, 0)
	if err := DB.Where(DB.Or("user_name = ?", userName).
		Or("email = ?", email)).
		Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

// CheckUser 是用来检查用户是否存在的
func CheckUser(account, password string) (*model.User, error) {
	user := model.User{}
	DB.Find(&user, "username = ? OR email = ? OR phone_number = ?",
		account, account, account,
	).First(&user)

	klog.Info("CheckUser", user, password)

	if password != user.Password {
		return nil, errors.New("密码错误")
	}
	return &user, nil
}
