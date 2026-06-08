package dto

import "time"

type CreateMessageRequest struct {
	Title      string `json:"title" binding:"required,max=255"`
	Content    string `json:"content" binding:"required"`
	TargetType string `json:"target_type" binding:"required,oneof=ALL FIELD"`
	TargetID   int    `json:"target_id" binding:"omitempty,min=1"`
}

type MessageListRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Title      string `form:"title" binding:"omitempty"`
	TargetType string `form:"target_type" binding:"omitempty,oneof=ALL FIELD"`
	TargetID   int    `form:"target_id" binding:"omitempty,min=1"`
	UserID     int    `form:"user_id" binding:"omitempty,min=1"`
}

type UserMessageListRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type MessageIDRequest struct {
	ID int `uri:"id" binding:"required,numeric"`
}

type MessageListResponse struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	SenderID   int       `json:"sender_id"`
	TargetType string    `json:"target_type"`
	TargetID   int       `json:"target_id"`
	SendTime   time.Time `json:"send_time"`
	CreateTime time.Time `json:"create_time"`
}

type UserMessageListResponse struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	SenderID   int       `json:"sender_id"`
	TargetType string    `json:"target_type"`
	TargetID   int       `json:"target_id"`
	SendTime   time.Time `json:"send_time"`
	IsRead     int       `json:"is_read"`
	ReadTime   *time.Time `json:"read_time"`
}

type UnreadCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
}
