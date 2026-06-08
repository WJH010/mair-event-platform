package model

import (
	"time"
)

// UserMessageMapping 对应user_messages_mapping表的数据模型
type UserMessageMapping struct {
	ID         int       `json:"id" gorm:"primaryKey;column:id"`
	UserID     int       `json:"user_id" gorm:"column:user_id"`
	MessageID  int       `json:"message_id" gorm:"column:message_id"`
	IsRead     int       `json:"is_read" gorm:"column:is_read;default:0"`               // 是否已读，默认值为0
	ReadTime   *time.Time `json:"read_time" gorm:"column:read_time"` // 读取时间，未读时为NULL
	CreateTime time.Time `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time;autoUpdateTime"`
	CreateUser int       `json:"create_user" gorm:"column:create_user"` // 数据创建用户ID
	UpdateUser int       `json:"update_user" gorm:"column:update_user"` // 最后更新数据用户ID
}

// TableName 设置表名
func (*UserMessageMapping) TableName() string {
	return "user_message_mappings" // 表名指定为user_message_mappings
}
