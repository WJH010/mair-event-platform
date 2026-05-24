package dto

type FieldItem struct {
	FieldCode string `json:"field_code"`
	FieldName string `json:"field_name"`
}

type LoginRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,phone"`
	Password    string `json:"password" binding:"required"`
}

type UserIDRequest struct {
	UserID int `uri:"id" binding:"required"`
}

// UserUpdateRequest 用户信息更新请求
type UserUpdateRequest struct {
	Nickname   *string `json:"nickname" binding:"omitempty,nickname"`
	AvatarURL  *string `json:"avatar_url" binding:"omitempty,url"`
	Name       *string `json:"name" binding:"omitempty,real_name"`
	Gender     *string `json:"gender" binding:"omitempty,oneof=M F U"` // M: 男, F: 女, U: 未知
	Email      *string `json:"email" binding:"omitempty,email"`
	Unit       *string `json:"unit" binding:"omitempty"`
	Department *string `json:"department" binding:"omitempty"`
	Position   *string `json:"position" binding:"omitempty"`
	IndustryID *int `json:"industry_id" binding:"omitempty"`
	FieldIDs   []int   `json:"field_ids" binding:"omitempty"` // 领域ID列表
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	UserID       int         `json:"user_id"`
	Nickname     string      `json:"nickname"`
	AvatarURL    string      `json:"avatar_url"`
	Name         string      `json:"name"`
	GenderCode   string      `json:"gender_code"`
	Gender       string      `json:"gender"`
	CountryCode  string      `json:"country_code"`
	PhoneNumber  string      `json:"phone_number"`
	Email        string      `json:"email"`
	Unit         string      `json:"unit"`
	Department   string      `json:"department"`
	Position     string      `json:"position"`
	IndustryID   int    `json:"industry_id"`
	IndustryName string `json:"industry_name"`
	Role         string `json:"role"`
	RoleName     string `json:"role_name"`
	Status       int    `json:"status"`
	Fields       []FieldItem `json:"fields" gorm:"-"`
}

type ListUsersRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Name       string `form:"name" binding:"omitempty,max=255"`
	GenderCode string `form:"gender_code" binding:"omitempty,oneof=M F U"`
	Unit       string `form:"unit" binding:"omitempty,max=255"`
	Department string `form:"department" binding:"omitempty,max=255"`
	Position   string `form:"position" binding:"omitempty,max=255"`
	IndustryID string `form:"industry_id" binding:"omitempty,numeric"`
	Role       string `form:"role" binding:"omitempty"`
}

type ListUsersResponse struct {
	UserID       int         `json:"user_id"`
	Nickname     string      `json:"nickname"`
	AvatarURL    string      `json:"avatar_url"`
	Name         string      `json:"name"`
	GenderCode   string      `json:"gender_code"`
	Gender       string      `json:"gender"`
	CountryCode  string      `json:"country_code"`
	PhoneNumber  string      `json:"phone_number"`
	Email        string      `json:"email"`
	Unit         string      `json:"unit"`
	Department   string      `json:"department"`
	Position     string      `json:"position"`
	IndustryID   int         `json:"industry_id"`
	IndustryName string      `json:"industry_name"`
	RoleName     string      `json:"role_name"`
	Status       int         `json:"status"`
	UserStatus   string      `json:"user_status"`
	Fields       []FieldItem `json:"fields" gorm:"-"`
}

type RegisterRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,phone"`
	Password    string `json:"password" binding:"required"`
	VerifyToken string `json:"verify_token" binding:"required"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type ChangePasswordRequest struct {
	VerifyToken string `json:"verify_token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// UpdateAdminStatusRequest 更新管理员状态请求
type UpdateAdminStatusRequest struct {
	Operation string `json:"operation" binding:"required,oneof=ENABLE DISABLE"` // ENABLE：启用，DISABLE：禁用
}

// RefreshTokenRequest 刷新token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	SessionSign  string `json:"session_sign" binding:"omitempty"`
}

type SendSMSRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,phone"`
	Purpose     string `json:"purpose" binding:"required,oneof=REGISTER LOGIN CHANGE_PASSWORD RESET_PASSWORD"`
}

type VerifySMSRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,phone"`
	Code        string `json:"code" binding:"required,len=4"`
	Purpose     string `json:"purpose" binding:"required,oneof=REGISTER LOGIN CHANGE_PASSWORD RESET_PASSWORD"`
}

type SMSLoginRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,phone"`
	VerifyToken string `json:"verify_token" binding:"required"`
}

type ResetPasswordRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,phone"`
	VerifyToken string `json:"verify_token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}
