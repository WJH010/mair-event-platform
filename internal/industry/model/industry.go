package model

import (
	"time"
)

// Industries 对应 news_platform 数据库中 industries 数据表的数据模型
type Industries struct {
	ID           int       `json:"id" gorm:"primaryKey;column:id"`                              // 主键ID
	IndustryCode string    `json:"industry_code" gorm:"not null;size:50;column:industry_code"`  // 行业编码
	IndustryName string    `json:"industry_name" gorm:"not null;size:255;column:industry_name"` // 行业名称
	Desc         string    `json:"desc" gorm:"size:255;column:desc"`
	Enable       int       `json:"enable" gorm:"column:enable;default:1"` // 1:启用 2:禁用
	CreateTime   time.Time `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime   time.Time `json:"update_time" gorm:"column:update_time;autoUpdateTime"`
	CreateUser   int       `json:"create_user" gorm:"column:create_user"`
	UpdateUser   int       `json:"update_user" gorm:"column:update_user"`
}

// TableName 设置当前模型对应的数据库表名
func (*Industries) TableName() string {
	return "industries"
}
