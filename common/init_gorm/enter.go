package init_gorm

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitGorm(MysqlDataSource string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(MysqlDataSource), &gorm.Config{})

	if err != nil {
		panic("连接mysql数据库失败,error:" + err.Error())
	} else {
		fmt.Println("连接mysql数据库成功")
	}

	return db
}
