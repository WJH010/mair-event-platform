package dto

// IndustryUrlID 行业URL ID参数
type IndustryUrlID struct {
	ID int `uri:"id" binding:"required"`
}

// CreateIndustryRequest 创建行业请求参数
type CreateIndustryRequest struct {
	IndustryCode string `json:"industry_code" binding:"required,max=50"`
	IndustryName string `json:"industry_name" binding:"required,max=255"`
	Desc         string `json:"desc" binding:"omitempty,max=255"`
}

// UpdateIndustryRequest 更新行业请求参数
type UpdateIndustryRequest struct {
	IndustryCode string `json:"industry_code" binding:"omitempty,non_empty_string,max=50"`
	IndustryName string `json:"industry_name" binding:"omitempty,non_empty_string,max=255"`
	Desc         string `json:"desc" binding:"omitempty,max=255"`
}

// UpdateIndustryStatusRequest 更新行业状态请求参数
type UpdateIndustryStatusRequest struct {
	Operation string `json:"operation" binding:"required,oneof=ENABLE DISABLE"`
}

// ListIndustriesRequest 分页查询行业请求参数
type ListIndustriesRequest struct {
	IndustryName string `form:"industry_name" binding:"omitempty,max=255"`
	Enable       *int   `form:"enable" binding:"omitempty,oneof=1 2"`
}

// ListIndustriesResponse 分页查询行业响应结构体
type ListIndustriesResponse struct {
	ID           int    `json:"id"`
	IndustryCode string `json:"industry_code"`
	IndustryName string `json:"industry_name"`
	Desc         string `json:"desc"`
	Enable       int    `json:"enable"`
}
