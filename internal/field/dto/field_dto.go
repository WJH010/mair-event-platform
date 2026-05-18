package dto

// FieldUrlID 领域URLID参数
type FieldUrlID struct {
	ID int `uri:"id" binding:"required"`
}

// CreateFieldRequest 创建领域请求参数
type CreateFieldRequest struct {
	FieldCode string `json:"field_code" binding:"required,max=255"`
	FieldName string `json:"field_name" binding:"required,max=255"`
	Desc      string `json:"desc" binding:"omitempty,max=255"`
}

// UpdateFieldRequest 更新领域请求参数
type UpdateFieldRequest struct {
	FieldCode string `json:"field_code" binding:"omitempty,non_empty_string,max=255"`
	FieldName string `json:"field_name" binding:"omitempty,non_empty_string,max=255"`
	Desc      string `json:"desc" binding:"omitempty,max=255"`
}

// UpdateFieldStatusRequest 更新领域状态请求参数
type UpdateFieldStatusRequest struct {
	Operation string `json:"operation" binding:"required,oneof=ENABLE DISABLE"`
}

// ListFieldsRequest 分页查询领域请求参数
type ListFieldsRequest struct {
	FieldName string `form:"field_name" binding:"omitempty,max=255"`
	Enable    *int   `form:"enable" binding:"omitempty,oneof=1 2"`
}

// ListFieldsResponse 分页查询领域响应结构体
type ListFieldsResponse struct {
	ID        int    `json:"id"`
	FieldCode string `json:"field_code"`
	FieldName string `json:"field_name"`
	Desc      string `json:"desc"`
	Enable    int    `json:"enable"`
}
