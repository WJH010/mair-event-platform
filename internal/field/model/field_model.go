package model

import (
	"time"
)

type Field struct {
	ID         int       `json:"id" gorm:"primaryKey;column:id"`
	FieldCode  string    `json:"field_code" gorm:"not null;size:255;column:field_code;uniqueIndex"`
	FieldName  string    `json:"field_name" gorm:"not null;size:255;column:field_name"`
	Desc       string    `json:"desc" gorm:"size:255;column:desc"`
	Enable     int       `json:"enable" gorm:"column:enable;default:1"` // 1:启用 2:禁用
	CreateTime time.Time `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time;autoUpdateTime"`
	CreateUser int       `json:"create_user" gorm:"column:create_user"`
	UpdateUser int       `json:"update_user" gorm:"column:update_user"`
}

func (*Field) TableName() string {
	return "field"
}
