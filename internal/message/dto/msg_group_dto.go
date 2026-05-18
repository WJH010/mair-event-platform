package dto

import "time"

// MsgGroupIDRequest 用户入群请求
type MsgGroupIDRequest struct {
	MsgGroupID int `uri:"id" binding:"required,numeric"` // 消息组ID，必须为数字
}

// UserListForGroupRequest 用户ID列表请求
type UserListForGroupRequest struct {
	UserIDs []int `json:"user_ids" binding:"required,dive,numeric"` // 用户ID列表，必须为数字
}

// CreateMsgGroupRequest 创建消息群组请求
type CreateMsgGroupRequest struct {
	GroupName      string `json:"group_name" binding:"required,max=255"`          // 群组名称，必填，最大长度255
	Desc           string `json:"desc" binding:"omitempty"`                       // 群组描述，选填
	IncludeAllUser string `json:"include_all_user" binding:"omitempty,oneof=Y N"` // 是否包含所有用户，选填，默认N
	UserIDs        []int  `json:"user_ids" binding:"omitempty,dive,numeric"`      // 初始用户ID列表，选填，必须为数字
}

// UpdateMsgGroupRequest 更新消息群组请求
type UpdateMsgGroupRequest struct {
	GroupName *string `json:"group_name" binding:"omitempty,non_empty_string,max=255"` // 群组名称
	Desc      *string `json:"desc" binding:"omitempty,non_empty_string"`               // 群组描述
}

// ListMessageGroupRequest 消息群组列表请求参数
type ListMessageGroupRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`                  // 页码，最小为1
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`     // 页大小，1-100
	TypeCode string `form:"type_code" binding:"required,group_message_type"` // 消息类型代码
}

// ListMsgGroupRequest 分页查询消息群组请求
type ListMsgGroupRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`              // 页码，默认1
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"` // 每页数量，默认10，最大100
	GroupName  string `form:"group_name" binding:"omitempty,max=255"`      // 群组名称
	FieldID    int    `form:"field_id" binding:"omitempty,numeric"`        // 字段ID，必须为数字
	QueryScope string `form:"query_scope" binding:"omitempty,query_scope"` // 查询范围
}

// DeleteMsgGroupMapRequest 撤回组内消息请求
type DeleteMsgGroupMapRequest struct {
	MapID int `uri:"id" binding:"required,numeric"`
}

// ListNotInGroupUsersRequest 查询不在指定组内的用户请求参数
type ListNotInGroupUsersRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Name       string `form:"name" binding:"omitempty,max=255"`
	GenderCode string `form:"gender_code" binding:"omitempty,oneof=M F U"`
	Unit       string `form:"unit" binding:"omitempty,max=255"`
	Department string `form:"department" binding:"omitempty,max=255"`
	Position   string `form:"position" binding:"omitempty,max=255"`
	Industry   string `form:"industry" binding:"omitempty,numeric"`
}

// ListMsgGroupResponse 消息群组列表响应
type ListMsgGroupResponse struct {
	ID             int    `json:"id"`
	GroupName      string `json:"group_name"`
	Desc           string `json:"desc"`
	FieldID        int    `json:"field_id"`
	FieldName      string `json:"field_name"`
	IncludeAllUser string `json:"include_all_user"`
	IsDeleted      string `json:"is_deleted"`
	MemberCount    int    `json:"member_count"`
}

// ListGroupsUsersResponse 消息群组用户列表响应
type GetMsgGroupDetailResponse struct {
	ID             int    `json:"id"`
	GroupName      string `json:"group_name"`
	Desc           string `json:"desc"`
	FieldID        int    `json:"field_id"`
	FieldName      string `json:"field_name"`
	IncludeAllUser string `json:"include_all_user"`
	LatestMsgID    int    `json:"latest_msg_id"`
	IsDeleted      string `json:"is_deleted"`
	CreateTime     string `json:"create_time"`
	UpdateTime     string `json:"update_time"`
	CreateUser     int    `json:"create_user"`
	UpdateUser     int    `json:"update_user"`
}

type ListGroupsUsersResponse struct {
	UserID       int    `json:"user_id"`
	Nickname     string `json:"nickname"`
	Name         string `json:"name"`
	GenderCode   string `json:"gender_code"`
	Gender       string `json:"gender"`
	PhoneNumber  string `json:"phone_number"`
	Email        string `json:"email"`
	Unit         string `json:"unit"`
	Department   string `json:"department"`
	Position     string `json:"position"`
	Industry     string `json:"industry"`
	IndustryName string `json:"industry_name"`
}

// MessageGroupDTO 消息群组响应结构体
type MessageGroupDTO struct {
	MsgGroupID     int       `json:"msg_group_id"`
	GroupName      string    `json:"group_name"`
	LatestTitle    string    `json:"latest_title"`
	LatestContent  string    `json:"latest_content"`
	LatestSendTime time.Time `json:"latest_send_time"`
	HasUnread      string    `json:"has_unread"`
	MemberCount    int       `json:"member_count"`
}
