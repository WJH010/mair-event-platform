package dto

import "time"

type DashboardUserListRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	FieldID    int    `form:"field_id" binding:"omitempty,min=1"`
	Unit       string `form:"unit" binding:"omitempty,max=255"`
	Department string `form:"department" binding:"omitempty,max=255"`
	Position   string `form:"position" binding:"omitempty,max=255"`
	IndustryID int    `form:"industry_id" binding:"omitempty,min=1"`
}

type DashboardUserItem struct {
	UserID       int         `json:"user_id"`
	Name         string      `json:"name"`
	PhoneNumber  string      `json:"phone_number"`
	Email        string      `json:"email"`
	Unit         string      `json:"unit"`
	Department   string      `json:"department"`
	Position     string      `json:"position"`
	IndustryID   int         `json:"industry_id"`
	IndustryName string      `json:"industry_name"`
	Fields       []FieldItem `json:"fields" gorm:"-"`
}

type FieldItem struct {
	FieldID   int    `json:"field_id"`
	FieldCode string `json:"field_code"`
	FieldName string `json:"field_name"`
}

type DashboardUserListResponse struct {
	TotalCount int64                `json:"total_count"`
	List       []*DashboardUserItem `json:"list"`
}

type DashboardEventListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Title    string `form:"title" binding:"omitempty"`
}

type DashboardEventItem struct {
	ID                 int       `json:"id"`
	Title              string    `json:"title"`
	EventStartTime     time.Time `json:"event_start_time"`
	EventEndTime       time.Time `json:"event_end_time"`
	MaxRegistrants     int       `json:"max_registrants"`
	CurrentRegistrants int       `json:"current_registrants"`
	Status             string    `json:"status"`
}

type DashboardEventIDRequest struct {
	ID int `uri:"id" binding:"required,numeric"`
}

type DashboardEventOverviewResponse struct {
	ID                    int       `json:"id"`
	Title                 string    `json:"title"`
	EventStartTime        time.Time `json:"event_start_time"`
	EventEndTime          time.Time `json:"event_end_time"`
	RegistrationStartTime time.Time `json:"registration_start_time"`
	RegistrationEndTime   time.Time `json:"registration_end_time"`
	MaxRegistrants        int       `json:"max_registrants"`
	CurrentRegistrants    int       `json:"current_registrants"`
	EventAddress          string    `json:"event_address"`
	Status                string    `json:"status"`
}

type DashboardRegUserListRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	FieldID    int    `form:"field_id" binding:"omitempty,min=1"`
	Unit       string `form:"unit" binding:"omitempty,max=255"`
	Department string `form:"department" binding:"omitempty,max=255"`
	Position   string `form:"position" binding:"omitempty,max=255"`
	IndustryID int    `form:"industry_id" binding:"omitempty,min=1"`
}

type DashboardRegUserItem struct {
	UserID       int    `json:"user_id"`
	Name         string `json:"name"`
	PhoneNumber  string `json:"phone_number"`
	Email        string `json:"email"`
	Unit         string `json:"unit"`
	Department   string `json:"department"`
	Position     string `json:"position"`
	IndustryID   int    `json:"industry_id"`
	IndustryName string `json:"industry_name"`
}

type StatItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type DashboardEventStatisticsResponse struct {
	ByField      []StatItem `json:"by_field"`
	ByUnit       []StatItem `json:"by_unit"`
	ByDepartment []StatItem `json:"by_department"`
	ByPosition   []StatItem `json:"by_position"`
}
