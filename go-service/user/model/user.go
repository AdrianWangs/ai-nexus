package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username    string `gorm:"type:varchar(20) not null;uniqueIndex" json:"Username"`
	Password    string `gorm:"type:varchar(20) not null" json:"Password"`
	Birthday    string `gorm:"type:date" json:"Birthday"`
	Gender      string `gorm:"type:varchar(10)" json:"Gender"`
	RoleId      int32  `gorm:"type:int" json:"RoleId"`
	PhoneNumber string `gorm:"type:varchar(20)" json:"PhoneNumber"`
	Email       string `gorm:"type:varchar(50)" json:"Email"`
	// 第三方登录token，可选
	ThirdPartyToken string `gorm:"type:varchar(100)" json:"ThirdPartyToken"`
}

// TableName 表名
func (User) TableName() string {
	return "user"
}
