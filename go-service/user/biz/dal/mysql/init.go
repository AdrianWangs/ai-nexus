package mysql

import (
	"github.com/AdrianWangs/ai-nexus/go-service/user/biz/dal/model"
	"github.com/AdrianWangs/ai-nexus/go-service/user/conf"
	"github.com/cloudwego/kitex/pkg/klog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB  *gorm.DB
	err error
)

func Init() {

	DB, err = gorm.Open(mysql.Open(conf.GetConf().MySQL.DSN),
		&gorm.Config{
			PrepareStmt:            true,
			SkipDefaultTransaction: true,
		},
	)

	// 自动迁移模式
	DB.AutoMigrate(&model.User{})

	if err != nil {
		panic(err)
	}
	klog.Infof("mysql 初始化成功")
}
