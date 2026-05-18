package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"event-platform/internal/config"
	msgsvc "event-platform/internal/message/service"
	rd "event-platform/internal/redis"
	"event-platform/internal/user/dto"
	"event-platform/internal/user/model"
	"event-platform/internal/user/repository"
	"event-platform/internal/utils"
	"fmt"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/argon2"
)

// WxLoginResponse 微信登录请求参数
type WxLoginResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid,omitempty"`
	ErrCode    int    `json:"errcode,omitempty"`
	ErrMsg     string `json:"errmsg,omitempty"`
}

// UserService 用户服务接口
type UserService interface {
	// RefreshToken 通过刷新令牌获取新的access token和refresh token
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	// Login 后台登录
	Login(ctx context.Context, req dto.LoginRequest) (string, string, error)
	// Logout 退出登录
	Logout(ctx context.Context, userID int, accessToken string) error
	// UpdateUserInfo 更新用户信息
	UpdateUserInfo(ctx context.Context, userID int, req dto.UserUpdateRequest) error
	// GetUserByID 根据用户ID获取用户信息
	GetUserByID(ctx context.Context, userID int) (*dto.UserInfoResponse, error)
	// ListAllUsers 分页查询所有用户
	ListAllUsers(ctx context.Context, page, pageSize int, req dto.ListUsersRequest) ([]*dto.ListUsersResponse, int64, error)
	// RegisterUser 用户注册
	RegisterUser(ctx context.Context, req dto.RegisterRequest) error
	// UpdateUserRole 用户角色变更
	UpdateUserRole(ctx context.Context, userID int, req dto.UpdateRoleRequest, operator int) error
	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, userID int, req dto.ChangePasswordRequest) error
	// UpdateUserStatus 更新用户状态
	UpdateUserStatus(ctx context.Context, userID int, Operation string, operator int) error
}

// UserServiceImpl 用户服务实现
type UserServiceImpl struct {
	userRepo repository.UserRepository
	msgSvc   msgsvc.MsgGroupService
	cfg      *config.Config
}

// Argon2参数配置
const (
	// 内存成本：哈希过程中使用的内存量（字节）
	argonMemory uint32 = 65536 // 64MB
	// 时间成本：计算迭代次数
	argonTime uint32 = 3
	// 并行度：使用的CPU核心数
	argonThreads uint8 = 4
	// 生成的哈希长度（字节）
	argonKeyLen uint32 = 32
	// 盐值长度（字节）
	argonSaltLen uint32 = 16
)

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, msgSvc msgsvc.MsgGroupService, cfg *config.Config) UserService {
	return &UserServiceImpl{userRepo: userRepo, msgSvc: msgSvc, cfg: cfg}
}

// 生成access token（短期有效）
func (svc *UserServiceImpl) generateAccessToken(userID int, userRole string) (string, error) {
	claims := jwt.MapClaims{
		"userid":    userID,
		"user_role": userRole,
		"exp":       time.Now().Add(time.Hour * 24).Unix(), // 1天
		"iat":       time.Now().Unix(),
		"type":      "access", // 标记令牌类型
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstr, err := token.SignedString([]byte(svc.cfg.JWT.JwtSecret))
	if err != nil {
		return "", utils.NewSystemError(fmt.Errorf("生成access token失败: %w", err))
	}
	return tokenstr, nil
}

// 生成refresh token（长期有效）
func (svc *UserServiceImpl) generateRefreshToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"userid": userID,
		"exp":    time.Now().Add(time.Hour * 168).Unix(), // 2周
		"iat":    time.Now().Unix(),
		"type":   "refresh", // 标记令牌类型
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstr, err := token.SignedString([]byte(svc.cfg.JWT.RefreshSecret))
	if err != nil {
		return "", utils.NewSystemError(fmt.Errorf("生成refresh token失败: %w", err))
	}
	return tokenstr, nil
}

// RefreshClaims 自定义Claims结构体，明确指定字段类型
type RefreshClaims struct {
	UserID             int    `json:"userid"`
	Type               string `json:"type"`
	jwt.StandardClaims        // 嵌入标准声明
}

// parseAccessToken 解析access_token获取Claims（用于计算剩余有效期）
func parseAccessToken(tokenString string, secret string) (*jwt.StandardClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("无效的签名方法")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("无效的token")
}

// parseRefreshToken 解析refresh token
func (svc *UserServiceImpl) parseRefreshToken(tokenString string) (*RefreshClaims, error) {
	secret := []byte(svc.cfg.JWT.RefreshSecret)

	// 解析令牌时指定自定义Claims
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, utils.NewSystemError(fmt.Errorf("无效的签名方法"))
		}
		return secret, nil
	})

	if err != nil {
		var validationErr *jwt.ValidationError
		if errors.As(err, &validationErr) {
			if validationErr.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, utils.NewBusinessError(utils.ErrCodeRefreshTokenExpired, "认证信息已失效，请重新登录")
			}
		}
		return nil, err
	}

	// 验证令牌有效性并转换为自定义Claims
	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		// 检查令牌类型
		if claims.Type != "refresh" {
			return nil, utils.NewBusinessError(utils.ErrCodeTokenTypeInvalid, "无效的token类型")
		}
		return claims, nil
	}

	return nil, utils.NewSystemError(fmt.Errorf("无效的token"))
}

// RefreshToken 通过刷新令牌获取新的access token和refresh token
func (svc *UserServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	// 解析refresh token
	claims, err := svc.parseRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	// 验证用户存在性
	userID := claims.UserID
	user, err := svc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	// 验证refresh token是否匹配
	if user.RefreshToken != refreshToken {
		return "", "", utils.NewBusinessError(utils.ErrCodeRefreshTokenExpired, "认证信息已失效，请重新登录")
	}

	// 生成新的access token
	accessToken, err := svc.generateAccessToken(userID, user.Role)
	if err != nil {
		return "", "", err
	}
	// 同时生成新的refresh token，实现滚动更新
	newRefreshToken, err := svc.generateRefreshToken(userID)
	if err != nil {
		return "", "", err
	}
	// 存储新的refresh token到数据库
	err = svc.userRepo.UpdateRefreshToken(ctx, userID, newRefreshToken)
	if err != nil {
		return "", "", err
	}
	return accessToken, newRefreshToken, nil
}

// 生成密码哈希
func hashPassword(password string) (string, error) {
	// 生成随机盐值
	salt := make([]byte, argonSaltLen)
	_, err := rand.Read(salt)
	if err != nil {
		return "", utils.NewSystemError(fmt.Errorf("生成随机盐值失败: %w", err))
	}

	// 使用Argon2id变体进行哈希（推荐用于密码哈希）
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	// 组合盐值和哈希值，并进行Base64编码以便存储
	// 格式: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// 包含算法参数以便验证时使用
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads, b64Salt, b64Hash)

	return encodedHash, nil
}

// 常量时间比较函数，防止时序攻击
func constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	result := 0
	for i := range a {
		result |= int(a[i] ^ b[i])
	}
	return result == 0
}

// 验证密码
func verifyPassword(encodedHash, password string) (bool, error) {
	// 解析格式: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	// 按 $ 分割字符串，得到各部分
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, utils.NewSystemError(fmt.Errorf("哈希格式错误"))
	}
	// 验证算法是否为 argon2id
	if parts[1] != "argon2id" {
		return false, utils.NewSystemError(fmt.Errorf("不支持的算法: %s", parts[1]))
	}

	// 解析版本号（如 v=19）
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, utils.NewSystemError(fmt.Errorf("解析版本失败: %v", err))
	}

	// 解析参数（m=内存, t=时间, p=并行度）
	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, utils.NewSystemError(fmt.Errorf("解析参数失败: %v", err))
	}

	// 提取盐值和哈希值（直接从分割结果中获取，避免解析错误）
	saltStr := parts[4]
	hashStr := parts[5]

	// 解码盐值和哈希值
	saltBytes, err := base64.RawStdEncoding.DecodeString(saltStr)
	if err != nil {
		return false, utils.NewSystemError(fmt.Errorf("解码盐值失败: %w", err))
	}

	hashBytes, err := base64.RawStdEncoding.DecodeString(hashStr)
	if err != nil {
		return false, utils.NewSystemError(fmt.Errorf("解码哈希值失败: %w", err))
	}

	// 使用相同的参数计算输入密码的哈希
	inputHash := argon2.IDKey([]byte(password), saltBytes, time, memory, threads, uint32(len(hashBytes)))

	// 比较计算出的哈希和存储的哈希
	return constantTimeCompare(inputHash, hashBytes), nil
}

// Login 登录接口
func (svc *UserServiceImpl) Login(ctx context.Context, req dto.LoginRequest) (string, string, error) {
	// 从数据库中根据用户名查询密码
	userInfo, err := svc.userRepo.GetPasswordByUserName(ctx, req.Username)
	if err != nil {
		return "", "", err
	}
	// 用户密码为空，不允许登录后台系统
	if userInfo.Password == "" {
		return "", "", utils.NewBusinessError(utils.ErrCodeAuthFailed, "账号未设置密码，无法登录后台系统")
	}

	// 验证密码
	ok, err := verifyPassword(userInfo.Password, req.Password)
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", utils.NewBusinessError(utils.ErrCodeAuthFailed, "密码错误")
	}

	// 检查用户状态
	if userInfo.Status != utils.UserStatusEnabled {
		return "", "", utils.NewBusinessError(utils.ErrCodeAuthFailed, "账号已被禁用，无法登录后台系统")
	}

	// 更新最后登录时间
	updateFields := make(map[string]any)
	updateFields["last_login_time"] = time.Now()
	if len(updateFields) > 0 {
		if err := svc.userRepo.Update(ctx, userInfo.UserID, updateFields); err != nil {
			// return "", err
			// 只记录日志，不影响登录成功
			logrus.Errorf("更新用户[%d]最后登录时间失败: %v", userInfo.UserID, err)
		}
	}

	// 登录成功，生成JWT Token
	token, err := svc.generateAccessToken(userInfo.UserID, userInfo.Role)
	if err != nil {
		return "", "", err
	}

	// 生成refresh token
	refreshToken, err := svc.generateRefreshToken(userInfo.UserID)
	if err != nil {
		return "", "", err
	}

	// 存储refresh token到数据库
	err = svc.userRepo.UpdateRefreshToken(ctx, userInfo.UserID, refreshToken)
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

// Logout 退出登录：清除refresh_token + 将access_token加入黑名单
func (svc *UserServiceImpl) Logout(ctx context.Context, userID int, accessToken string) error {
	// 1. 清除数据库中的refresh_token，阻止刷新
	if err := svc.userRepo.UpdateRefreshToken(ctx, userID, ""); err != nil {
		return err
	}

	// 2. 将access_token加入Redis黑名单
	rdb := rd.GetClient()
	fmt.Println("获取rd:", rdb)
	fmt.Println("token:", accessToken)
	if rdb != nil && accessToken != "" {
		// 计算token的SHA256哈希作为Redis key
		tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(accessToken)))
		blacklistKey := fmt.Sprintf("token:blacklist:%s", tokenHash)

		// 解析token获取剩余有效期
		claims, err := parseAccessToken(accessToken, svc.cfg.JWT.JwtSecret)
		fmt.Println("claims:", claims)
		if err == nil {
			// 计算token剩余有效期
			exp := time.Unix(claims.ExpiresAt, 0)
			ttl := time.Until(exp)
			fmt.Println("ttl:", ttl)
			if ttl > 0 {
				// 设置黑名单，TTL为token剩余有效期，过期自动清除
				fmt.Println("设置黑名单")
				rdb.Set(ctx, blacklistKey, "1", ttl)
			}
			// 如果ttl <= 0，token已过期，无需加入黑名单
		}
	}

	return nil
}

// UpdateUserInfo 更新用户信息
func (svc *UserServiceImpl) UpdateUserInfo(ctx context.Context, userID int, req dto.UserUpdateRequest) error {
	// 查询用户是否存在
	user, err := svc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return utils.NewBusinessError(utils.ErrCodeResourceNotFound, "用户不存在，请刷新后重试")
	}

	// 构建更新字段映射
	updateFields := make(map[string]interface{})
	if req.Nickname != nil {
		updateFields["nickname"] = *req.Nickname
	}
	if req.AvatarURL != nil {
		updateFields["avatar_url"] = *req.AvatarURL
	}
	if req.Name != nil {
		updateFields["name"] = *req.Name
	}
	if req.Gender != nil {
		updateFields["gender"] = *req.Gender
	}
	// if req.PhoneNumber != nil {
	// 	updateFields["phone_number"] = *req.PhoneNumber
	// }
	if req.Email != nil {
		updateFields["email"] = *req.Email
	}
	if req.Unit != nil {
		updateFields["unit"] = *req.Unit
	}
	if req.Department != nil {
		updateFields["department"] = *req.Department
	}
	if req.Position != nil {
		updateFields["position"] = *req.Position
	}
	if req.Industry != nil {
		updateFields["industry"] = *req.Industry
	}

	// 执行更新
	if len(updateFields) > 0 {
		if err := svc.userRepo.Update(ctx, userID, updateFields); err != nil {
			return err
		}
	}

	return nil
}

func (svc *UserServiceImpl) GetUserByID(ctx context.Context, userID int) (*dto.UserInfoResponse, error) {
	// 查询用户信息
	user, err := svc.userRepo.GetUserInfoByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, utils.NewBusinessError(utils.ErrCodeResourceNotFound, "用户不存在，请刷新后重试")
	}

	return user, nil
}

// ListAllUsers 分页查询用户列表
func (svc *UserServiceImpl) ListAllUsers(ctx context.Context, page, pageSize int, req dto.ListUsersRequest) ([]*dto.ListUsersResponse, int64, error) {
	return svc.userRepo.ListAllUsers(ctx, page, pageSize, req)
}

// RegisterUser 用户注册
func (svc *UserServiceImpl) RegisterUser(ctx context.Context, req dto.RegisterRequest) error {
	// 默认头像
	avatar := "http://47.113.194.28:9000/news-platform/images/20 2508/1754126743005963551.webp"

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return err
	}

	user := &model.User{
		Username:  req.Username,
		Password:  hashedPassword,
		AvatarURL: avatar,
		Role:      "USER",
	}

	if err := svc.userRepo.Create(ctx, user); err != nil {
		return err
	}
	return nil
}

// UpdateUserRole 用户角色变更
func (svc *UserServiceImpl) UpdateUserRole(ctx context.Context, userID int, req dto.UpdateRoleRequest, operator int) error {
	user, err := svc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return utils.NewBusinessError(utils.ErrCodeResourceNotFound, "用户不存在，请刷新后重试")
	}

	updateFields := make(map[string]interface{})
	updateFields["role"] = req.Role
	updateFields["update_user"] = operator

	if err := svc.userRepo.Update(ctx, userID, updateFields); err != nil {
		return err
	}
	return nil
}

// ChangePassword 修改密码
func (svc *UserServiceImpl) ChangePassword(ctx context.Context, userID int, req dto.ChangePasswordRequest) error {
	user, err := svc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return utils.NewBusinessError(utils.ErrCodeResourceNotFound, "用户不存在，请刷新后重试")
	}

	ok, err := verifyPassword(user.Password, req.OldPassword)
	if err != nil {
		return err
	}
	if !ok {
		return utils.NewBusinessError(utils.ErrCodeAuthFailed, "旧密码错误")
	}

	hashedPassword, err := hashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	updateFields := make(map[string]interface{})
	updateFields["password"] = hashedPassword
	updateFields["update_user"] = userID

	if err := svc.userRepo.Update(ctx, userID, updateFields); err != nil {
		return err
	}
	return nil
}

// UpdateUserStatus 禁用/启用用户账号
func (svc *UserServiceImpl) UpdateUserStatus(ctx context.Context, userID int, Operation string, operator int) error {
	// 查询用户是否存在
	user, err := svc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return utils.NewBusinessError(utils.ErrCodeResourceNotFound, "用户不存在，请刷新后重试")
	}
	if Operation == "DISABLE" {
		if user.Status == utils.UserStatusDisabled {
			return utils.NewBusinessError(utils.ErrCodeResourceConflict, "用户已被禁用，请勿重复操作")
		}
	} else if Operation == "ENABLE" {
		if user.Status == utils.UserStatusEnabled {
			return utils.NewBusinessError(utils.ErrCodeResourceConflict, "用户已被启用，请勿重复操作")
		}
	} else {
		return utils.NewBusinessError(utils.ErrCodeParamInvalid, "操作类型错误")
	}

	// 构建更新字段映射
	updateFields := make(map[string]any)
	if Operation == "DISABLE" {
		updateFields["status"] = utils.UserStatusDisabled
	} else if Operation == "ENABLE" {
		updateFields["status"] = utils.UserStatusEnabled
	}
	updateFields["update_user"] = operator

	// 执行更新
	if len(updateFields) > 0 {
		if err := svc.userRepo.Update(ctx, userID, updateFields); err != nil {
			return err
		}
	}
	return nil
}
