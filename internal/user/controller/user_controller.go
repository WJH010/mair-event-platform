package controller

import (
	"event-platform/internal/user/dto"
	"event-platform/internal/user/service"
	"event-platform/internal/utils"

	"github.com/gin-gonic/gin"
)

// UserController 用户控制器
type UserController struct {
	userService service.UserService
}

// NewUserController 创建用户控制器实例
func NewUserController(userService service.UserService) *UserController {
	return &UserController{userService: userService}
}

// RefreshToken 刷新token接口
func (ctr *UserController) RefreshToken(ctx *gin.Context) {
	// 初始化参数结构体并绑定查询参数
	var req dto.RefreshTokenRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	accessToken, refreshToken, err := ctr.userService.RefreshToken(ctx, req.RefreshToken)
	// 处理异常
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	result := gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	utils.Success(ctx, "success", result)
}

// Login 登录接口
func (ctr *UserController) Login(ctx *gin.Context) {
	// 绑定并验证请求参数
	var req dto.LoginRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	accessToken, refreshToken, err := ctr.userService.Login(ctx, req)
	// 处理异常
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	result := gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	utils.Success(ctx, "success", result)
}

// Logout 退出登录
func (ctr *UserController) Logout(ctx *gin.Context) {
	// 获取当前登录用户ID
	userID, err := utils.GetUserID(ctx)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	// 从上下文获取原始access_token（由AuthMiddleware存入）
	accessToken, _ := ctx.Get("access_token")
	accessTokenStr, _ := accessToken.(string)

	// 调用服务处理退出登录
	err = ctr.userService.Logout(ctx, userID, accessTokenStr)

	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

// UpdateUserInfo 更新用户信息接口
func (ctr *UserController) UpdateUserInfo(ctx *gin.Context) {
	// 获取userID
	userID, err := utils.GetUserID(ctx)
	// 处理异常
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	// 绑定并验证请求参数
	var req dto.UserUpdateRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	// 调用服务更新用户信息
	err = ctr.userService.UpdateUserInfo(ctx, userID, req)
	// 处理异常
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

// GetUserInfo 获取用户信息接口
func (ctr *UserController) GetUserInfo(ctx *gin.Context) {
	// 获取userID
	userID, err := utils.GetUserID(ctx)
	// 处理异常
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	// 调用服务获取用户信息
	user, err := ctr.userService.GetUserByID(ctx, userID)
	// 处理异常
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", user)
}

// ListAllUsers 列出所有用户接口（管理员权限）
func (ctr *UserController) ListAllUsers(ctx *gin.Context) {
	// 绑定并验证请求参数
	var req dto.ListUsersRequest
	if !utils.BindQuery(ctx, &req) {
		return
	}

	// 设置默认分页参数
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 调用服务获取用户列表
	users, total, err := ctr.userService.ListAllUsers(ctx, req.Page, req.PageSize, req)
	// 处理异常
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.SuccessPage(ctx, total, req.Page, req.PageSize, users)
}

// RegisterUser 用户注册
func (ctr *UserController) RegisterUser(ctx *gin.Context) {
	var req dto.RegisterRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	err := ctr.userService.RegisterUser(ctx, req)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

// SMSLogin 短信登录
func (ctr *UserController) SMSLogin(ctx *gin.Context) {
	var req dto.SMSLoginRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	accessToken, refreshToken, err := ctr.userService.SMSLogin(ctx, req)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	result := gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	utils.Success(ctx, "success", result)
}

// ResetPassword 重置密码
func (ctr *UserController) ResetPassword(ctx *gin.Context) {
	var req dto.ResetPasswordRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	err := ctr.userService.ResetPassword(ctx, req)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

// SendSMS 发送短信验证码
func (ctr *UserController) SendSMS(ctx *gin.Context) {
	var req dto.SendSMSRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	err := ctr.userService.SendSMSVerifyCode(ctx, req)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

// VerifySMS 验证短信验证码
func (ctr *UserController) VerifySMS(ctx *gin.Context) {
	var req dto.VerifySMSRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	verifyToken, err := ctr.userService.VerifySMSCode(ctx, req)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", gin.H{"verify_token": verifyToken})
}

// UpdateUserRole 用户角色变更
func (ctr *UserController) UpdateUserRole(ctx *gin.Context) {
	var urlReq dto.UserIDRequest
	if !utils.BindUrl(ctx, &urlReq) {
		return
	}

	var req dto.UpdateRoleRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	operator, err := utils.GetUserID(ctx)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	err = ctr.userService.UpdateUserRole(ctx, urlReq.UserID, req, operator)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

// ChangePassword 修改密码
func (ctr *UserController) ChangePassword(ctx *gin.Context) {
	userID, err := utils.GetUserID(ctx)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	var req dto.ChangePasswordRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	err = ctr.userService.ChangePassword(ctx, userID, req)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

// UpdateUserStatus 更新用户状态
func (ctr *UserController) UpdateUserStatus(ctx *gin.Context) {
	// 从路径参数获取userID
	var urlReq dto.UserIDRequest
	if !utils.BindUrl(ctx, &urlReq) {
		return
	}

	// 绑定并验证请求参数
	var req dto.UpdateAdminStatusRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	// 获取当前登录用户ID
	operator, err := utils.GetUserID(ctx)
	// 处理异常
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	// 调用服务更新用户状态
	err = ctr.userService.UpdateUserStatus(ctx, urlReq.UserID, req.Operation, operator)
	// 处理异常
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}
