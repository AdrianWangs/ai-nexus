// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameUser = "user"

// User mapped from table <user>
type User struct {
	ID              int64          `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	CreatedAt       time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
	Username        string         `gorm:"column:username;not null" json:"username"`
	Password        string         `gorm:"column:password;not null" json:"password"`
	Birthday        time.Time      `gorm:"column:birthday" json:"birthday"`
	Gender          string         `gorm:"column:gender" json:"gender"`
	RoleID          int32          `gorm:"column:role_id" json:"role_id"`
	PhoneNumber     string         `gorm:"column:phone_number" json:"phone_number"`
	Email           string         `gorm:"column:email" json:"email"`
	ThirdPartyToken string         `gorm:"column:third_party_token" json:"third_party_token"`
}

// TableName User's table name
func (*User) TableName() string {
	return TableNameUser
}
