package model

import "time"

// EventFieldMapping 对应数据库中 event_field_mapping 数据表的数据模型
type EventFieldMapping struct {
	ID         int       `json:"id" gorm:"primaryKey;column:id"`           // 主键ID
	EventID    int       `json:"event_id" gorm:"not null;column:event_id"` // 活动ID
	FieldID    int       `json:"field_id" gorm:"not null;column:field_id"` // 领域ID
	CreateTime time.Time `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time;autoUpdateTime"`
	CreateUser int       `json:"create_user" gorm:"column:create_user"`
	UpdateUser int       `json:"update_user" gorm:"column:update_user"`
}

// TableName 设置当前模型对应的数据库表名
func (*EventFieldMapping) TableName() string {
	return "event_field_mappings"
}
