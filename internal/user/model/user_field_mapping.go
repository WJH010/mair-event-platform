package model

import "time"

// UserFieldMapping 对应数据库中 user_field_mapping 数据表的数据模型
type UserFieldMapping struct {
	ID         int       `json:"id" gorm:"primaryKey;column:id"`           // 主键ID
	UserID     int       `json:"user_id" gorm:"not null;column:user_id"`   // 用户ID
	FieldID    int       `json:"field_id" gorm:"not null;column:field_id"` // 领域ID
	CreateTime time.Time `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time;autoUpdateTime"`
	CreateUser int       `json:"create_user" gorm:"column:create_user"`
	UpdateUser int       `json:"update_user" gorm:"column:update_user"`
}

// TableName 设置当前模型对应的数据库表名
func (*UserFieldMapping) TableName() string {
	return "user_field_mappings"
}
