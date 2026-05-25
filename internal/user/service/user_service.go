package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"event-platform/internal/config"
	db "event-platform/internal/database"
	msgsvc "event-platform/internal/message/service"
	rd "event-platform/internal/redis"
	"event-platform/internal/sms"
	"event-platform/internal/user/dto"
	"event-platform/internal/user/model"
	"event-platform/internal/user/repository"
	"event-platform/internal/utils"
	"fmt"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	// RefreshToken 通过刷新令牌获取新的access token和refresh token
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	// Login 登录
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
	// SMSLogin 通过短信验证码登录
	SMSLogin(ctx context.Context, req dto.SMSLoginRequest) (string, string, error)
	// ResetPassword 重置密码
	ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error
	// SendSMSVerifyCode 发送短信验证码
	SendSMSVerifyCode(ctx context.Context, req dto.SendSMSRequest) error
	// VerifySMSCode 验证短信验证码
	VerifySMSCode(ctx context.Context, req dto.VerifySMSRequest) (string, error)
}

// UserServiceImpl 用户服务实现
type UserServiceImpl struct {
	userRepo repository.UserRepository
	msgSvc   msgsvc.MsgGroupService
	cfg      *config.Config
	smsSvc   *sms.SMSService
}

const (
	// Argon2参数配置
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
	// sms参数配置
	// 短信验证码有效期（分钟）
	smsCodeTTL = 5 * time.Minute
	// 短信验证码错误次数最大值
	smsCodeErrMax = 5
	// 短信验证码有效期（分钟）
	smsTokenTTL = 10 * time.Minute
	// 短信验证码发送间隔（秒）
	smsSendInterval = 60 * time.Second
	// 短信验证码发送次数最大值（每日）
	smsDailyMax = 10
)

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, msgSvc msgsvc.MsgGroupService, cfg *config.Config) UserService {
	return &UserServiceImpl{userRepo: userRepo, msgSvc: msgSvc, cfg: cfg, smsSvc: sms.NewSMSService(cfg)}
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
	// 从数据库中根据手机号查询密码
	userInfo, err := svc.userRepo.GetPasswordByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {
		return "", "", err
	}
	// 用户密码为空，不允许登录后台系统
	if userInfo.Password == "" {
		return "", "", utils.NewBusinessError(utils.ErrCodeAuthFailed, "账号未设置密码，无法登录")
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
		return "", "", utils.NewBusinessError(utils.ErrCodeAuthFailed, "账号已被禁用，无法登录")
	}

	// 更新最后登录时间
	updateFields := make(map[string]any)
	updateFields["last_login_time"] = time.Now()
	if len(updateFields) > 0 {
		if err := svc.userRepo.Update(ctx, nil, userInfo.UserID, updateFields); err != nil {
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
	if req.IndustryID != nil {
		updateFields["industry_id"] = *req.IndustryID
	}

	// 开启事务
	return db.WithTx(db.GetDB(), func(tx *gorm.DB) error {
		// 执行更新
		if len(updateFields) > 0 {
			if err := svc.userRepo.Update(ctx, tx, userID, updateFields); err != nil {
				return err
			}
		}
		// 处理领域更新
		if req.FieldIDs != nil {
			// 先删除旧的领域映射
			if err := svc.userRepo.DeleteUserFieldMappings(ctx, tx, userID); err != nil {
				return err
			}
			// 批量创建新的领域映射
			if len(req.FieldIDs) > 0 {
				mappings := make([]*model.UserFieldMapping, 0, len(req.FieldIDs))
				for _, fieldID := range req.FieldIDs {
					mappings = append(mappings, &model.UserFieldMapping{
						UserID:     userID,
						FieldID:    fieldID,
						CreateUser: userID,
						UpdateUser: userID,
					})
				}
				if err := svc.userRepo.BatchCreateUserFieldMappings(ctx, tx, mappings); err != nil {
					return err
				}
			}
		}
		return nil
	})
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

	// 查询用户领域信息
	fields, err := svc.userRepo.GetUserFieldsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if fields == nil {
		fields = []dto.FieldItem{}
	}
	user.Fields = fields

	return user, nil
}

// ListAllUsers 分页查询用户列表
func (svc *UserServiceImpl) ListAllUsers(ctx context.Context, page, pageSize int, req dto.ListUsersRequest) ([]*dto.ListUsersResponse, int64, error) {
	users, total, err := svc.userRepo.ListAllUsers(ctx, page, pageSize, req)
	if err != nil {
		return nil, 0, err
	}

	// 批量查询用户领域信息
	if len(users) > 0 {
		userIDs := make([]int, len(users))
		for i, u := range users {
			userIDs[i] = u.UserID
		}
		fieldsMap, err := svc.userRepo.GetUserFieldsByUserIDs(ctx, userIDs)
		if err != nil {
			return nil, 0, err
		}
		for _, u := range users {
			if fields, ok := fieldsMap[u.UserID]; ok {
				u.Fields = fields
			} else {
				u.Fields = []dto.FieldItem{}
			}
		}
	}

	return users, total, nil
}

// RegisterUser 用户注册
func (svc *UserServiceImpl) RegisterUser(ctx context.Context, req dto.RegisterRequest) error {
	// 验证令牌
	phone, purpose, err := svc.validateVerifyToken(req.VerifyToken)
	if err != nil {
		return err
	}
	if purpose != "REGISTER" {
		return utils.NewBusinessError(utils.ErrCodeParamInvalid, "验证令牌用途不匹配")
	}
	if phone != req.PhoneNumber {
		return utils.NewBusinessError(utils.ErrCodeParamInvalid, "验证令牌与手机号不匹配")
	}
	// 消费令牌
	svc.consumeVerifyToken(req.VerifyToken)

	// 密码哈希
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return err
	}

	user := &model.User{
		PhoneNumber: req.PhoneNumber,
		Password:    hashedPassword,
		Role:        "USER",
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

	if err := svc.userRepo.Update(ctx, nil, userID, updateFields); err != nil {
		return err
	}
	return nil
}

// ChangePassword 修改密码
func (svc *UserServiceImpl) ChangePassword(ctx context.Context, userID int, req dto.ChangePasswordRequest) error {
	// 验证令牌
	phone, purpose, err := svc.validateVerifyToken(req.VerifyToken)
	if err != nil {
		return err
	}
	if purpose != "CHANGE_PASSWORD" {
		return utils.NewBusinessError(utils.ErrCodeParamInvalid, "验证令牌用途不匹配")
	}
	// 查询用户是否存在
	user, err := svc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return utils.NewBusinessError(utils.ErrCodeResourceNotFound, "用户不存在，请刷新后重试")
	}
	// 验证手机号是否匹配
	if user.PhoneNumber != phone {
		return utils.NewBusinessError(utils.ErrCodeParamInvalid, "验证令牌与当前用户手机号不匹配")
	}
	// 消费令牌
	svc.consumeVerifyToken(req.VerifyToken)

	hashedPassword, err := hashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	updateFields := make(map[string]interface{})
	updateFields["password"] = hashedPassword
	updateFields["update_user"] = userID

	if err := svc.userRepo.Update(ctx, nil, userID, updateFields); err != nil {
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
		if err := svc.userRepo.Update(ctx, nil, userID, updateFields); err != nil {
			return err
		}
	}
	return nil
}

// SMSLogin 短信登录
func (svc *UserServiceImpl) SMSLogin(ctx context.Context, req dto.SMSLoginRequest) (string, string, error) {
	phone, purpose, err := svc.validateVerifyToken(req.VerifyToken)
	if err != nil {
		return "", "", err
	}
	if purpose != "LOGIN" {
		return "", "", utils.NewBusinessError(utils.ErrCodeParamInvalid, "验证令牌用途不匹配")
	}
	if phone != req.PhoneNumber {
		return "", "", utils.NewBusinessError(utils.ErrCodeParamInvalid, "验证令牌与手机号不匹配")
	}

	userInfo, err := svc.userRepo.GetPasswordByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {
		return "", "", err
	}

	if userInfo.Status != utils.UserStatusEnabled {
		return "", "", utils.NewBusinessError(utils.ErrCodeAuthFailed, "账号已被禁用，无法登录")
	}

	svc.consumeVerifyToken(req.VerifyToken)

	updateFields := make(map[string]any)
	updateFields["last_login_time"] = time.Now()
	if err := svc.userRepo.Update(ctx, nil, userInfo.UserID, updateFields); err != nil {
		logrus.Errorf("更新用户[%d]最后登录时间失败: %v", userInfo.UserID, err)
	}

	token, err := svc.generateAccessToken(userInfo.UserID, userInfo.Role)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := svc.generateRefreshToken(userInfo.UserID)
	if err != nil {
		return "", "", err
	}

	err = svc.userRepo.UpdateRefreshToken(ctx, userInfo.UserID, refreshToken)
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

// ResetPassword 重置密码
func (svc *UserServiceImpl) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error {
	phone, purpose, err := svc.validateVerifyToken(req.VerifyToken)
	if err != nil {
		return err
	}
	if purpose != "RESET_PASSWORD" {
		return utils.NewBusinessError(utils.ErrCodeParamInvalid, "验证令牌用途不匹配")
	}
	if phone != req.PhoneNumber {
		return utils.NewBusinessError(utils.ErrCodeParamInvalid, "验证令牌与手机号不匹配")
	}

	user, err := svc.userRepo.GetByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {
		return err
	}
	if user == nil {
		return utils.NewBusinessError(utils.ErrCodeResourceNotFound, "该手机号未注册")
	}

	svc.consumeVerifyToken(req.VerifyToken)

	hashedPassword, err := hashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	updateFields := make(map[string]interface{})
	updateFields["password"] = hashedPassword
	updateFields["update_user"] = user.UserID

	if err := svc.userRepo.Update(ctx, nil, user.UserID, updateFields); err != nil {
		return err
	}
	return nil
}

// SendSMSVerifyCode 发送短信验证码
func (svc *UserServiceImpl) SendSMSVerifyCode(ctx context.Context, req dto.SendSMSRequest) error {
	rdb := rd.GetClient()
	if rdb == nil {
		return utils.NewSystemError(fmt.Errorf("Redis服务不可用"))
	}

	if req.Purpose == "REGISTER" {
		existUser, err := svc.userRepo.GetByPhoneNumber(ctx, req.PhoneNumber)
		if err != nil {
			return err
		}
		if existUser != nil {
			return utils.NewBusinessError(utils.ErrCodeResourceExists, "该手机号已注册")
		}
	} else {
		existUser, err := svc.userRepo.GetByPhoneNumber(ctx, req.PhoneNumber)
		if err != nil {
			return err
		}
		if existUser == nil {
			return utils.NewBusinessError(utils.ErrCodeResourceNotFound, "该手机号未注册")
		}
	}

	sendLimitKey := fmt.Sprintf("sms:send:limit:%s", req.PhoneNumber)
	if rdb.Exists(ctx, sendLimitKey).Val() > 0 {
		return utils.NewBusinessError(utils.ErrCodeRateLimitExceeded, "发送过于频繁，请60秒后重试")
	}

	dailyKey := fmt.Sprintf("sms:send:daily:%s:%s", req.PhoneNumber, time.Now().Format("20060102"))
	dailyCount := rdb.Incr(ctx, dailyKey).Val()
	if dailyCount == 1 {
		now := time.Now()
		endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
		rdb.Expire(ctx, dailyKey, time.Until(endOfDay))
	}
	if dailyCount > int64(smsDailyMax) {
		return utils.NewBusinessError(utils.ErrCodeRateLimitExceeded, "今日发送次数已达上限")
	}

	// TODO 测试环境使用固定验证码
	var code string
	if svc.cfg.App.Env != "production" {
		code = "1234"
	} else {
		code = svc.smsSvc.GenerateVerifyCode()
	}

	codeKey := fmt.Sprintf("sms:code:%s", req.PhoneNumber)
	err := rdb.Set(ctx, codeKey, code, smsCodeTTL).Err()
	if err != nil {
		return utils.NewSystemError(fmt.Errorf("存储验证码失败: %w", err))
	}

	errCountKey := fmt.Sprintf("sms:code:count:%s", req.PhoneNumber)
	rdb.Del(ctx, errCountKey)

	rdb.Set(ctx, sendLimitKey, "1", smsSendInterval)

	if err := svc.smsSvc.SendVerifyCode(req.PhoneNumber, code); err != nil {
		rdb.Del(ctx, codeKey)
		rdb.Del(ctx, sendLimitKey)
		rdb.Decr(ctx, dailyKey)
		return utils.NewBusinessError(utils.ErrCodeDependencyServiceError, "短信发送失败，请稍后重试")
	}

	return nil
}

// VerifySMSCode 验证短信验证码
func (svc *UserServiceImpl) VerifySMSCode(ctx context.Context, req dto.VerifySMSRequest) (string, error) {
	rdb := rd.GetClient()
	if rdb == nil {
		return "", utils.NewSystemError(fmt.Errorf("Redis服务不可用"))
	}

	if req.Purpose == "LOGIN" {
		existUser, err := svc.userRepo.GetByPhoneNumber(ctx, req.PhoneNumber)
		if err != nil {
			return "", err
		}
		if existUser == nil {
			return "", utils.NewBusinessError(utils.ErrCodeResourceNotFound, "该手机号未注册，请先注册")
		}
	}

	codeKey := fmt.Sprintf("sms:code:%s", req.PhoneNumber)
	storedCode, err := rdb.Get(ctx, codeKey).Result()
	if err != nil {
		return "", utils.NewBusinessError(utils.ErrCodeAuthFailed, "验证码已过期，请重新获取")
	}

	errCountKey := fmt.Sprintf("sms:code:count:%s", req.PhoneNumber)
	if req.Code != storedCode {
		errCount := rdb.Incr(ctx, errCountKey).Val()
		rdb.Expire(ctx, errCountKey, smsCodeTTL)
		if errCount >= int64(smsCodeErrMax) {
			rdb.Del(ctx, codeKey)
			rdb.Del(ctx, errCountKey)
			return "", utils.NewBusinessError(utils.ErrCodeAuthFailed, "验证码错误次数过多，请重新获取")
		}
		return "", utils.NewBusinessError(utils.ErrCodeAuthFailed, "验证码错误")
	}

	rdb.Del(ctx, codeKey)
	rdb.Del(ctx, errCountKey)

	verifyToken := uuid.New().String()
	tokenKey := fmt.Sprintf("sms:token:%s", verifyToken)
	tokenValue := fmt.Sprintf("%s:%s", req.PhoneNumber, req.Purpose)
	if err := rdb.Set(ctx, tokenKey, tokenValue, smsTokenTTL).Err(); err != nil {
		return "", utils.NewSystemError(fmt.Errorf("存储验证令牌失败: %w", err))
	}

	return verifyToken, nil
}

// validateVerifyToken 验证短信验证码
func (svc *UserServiceImpl) validateVerifyToken(token string) (phone, purpose string, err error) {
	rdb := rd.GetClient()
	if rdb == nil {
		return "", "", utils.NewSystemError(fmt.Errorf("Redis服务不可用"))
	}

	tokenKey := fmt.Sprintf("sms:token:%s", token)
	val, err := rdb.Get(context.Background(), tokenKey).Result()
	if err != nil {
		return "", "", utils.NewBusinessError(utils.ErrCodeAuthFailed, "验证令牌已过期，请重新验证")
	}

	parts := strings.SplitN(val, ":", 2)
	if len(parts) != 2 {
		return "", "", utils.NewBusinessError(utils.ErrCodeAuthTokenInvalid, "验证令牌无效")
	}

	return parts[0], parts[1], nil
}

// consumeVerifyToken 消费短信验证码
func (svc *UserServiceImpl) consumeVerifyToken(token string) {
	rdb := rd.GetClient()
	if rdb != nil {
		tokenKey := fmt.Sprintf("sms:token:%s", token)
		rdb.Del(context.Background(), tokenKey)
	}
}
