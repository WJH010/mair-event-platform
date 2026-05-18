package model

import (
	"time"
)

// MessageGroup 对应 message_groups 表
type MessageGroup struct {
	ID             int       `json:"id" gorm:"primaryKey;column:id"`
	GroupName      string    `json:"group_name" gorm:"not null;column:group_name;type:varchar(255)"`
	Desc           string    `json:"desc" gorm:"column:desc;type:text"`
	FieldID        int       `json:"field_id" gorm:"column:field_id;default:NULL"`
	IncludeAllUser string    `json:"include_all_user" gorm:"not null;default:N;column:include_all_user;type:varchar(5)"` // 全体用户包含标记：默认 N
	LatestMsgID    int       `json:"latest_msg_id" gorm:"column:latest_msg_id;default:0"`
	IsDeleted      string    `json:"is_deleted" gorm:"not null;default:N;column:is_deleted;type:varchar(5)"` // 软删除标记：默认 N
	CreateTime     time.Time `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime     time.Time `json:"update_time" gorm:"column:update_time;autoUpdateTime"`
	CreateUser     int       `json:"create_user" gorm:"column:create_user"`
	UpdateUser     int       `json:"update_user" gorm:"column:update_user"`
}

// TableName 指定模型对应的数据表名为 message_groups
func (*MessageGroup) TableName() string {
	return "message_groups"
}
