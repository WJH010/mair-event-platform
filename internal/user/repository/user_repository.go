package repository

import (
	"context"
	"errors"
	"event-platform/internal/user/dto"
	"event-platform/internal/user/model"
	"event-platform/internal/utils"
	"fmt"

	"gorm.io/gorm"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	// Create 创建新用户
	Create(ctx context.Context, user *model.User) error
	// Update 更新用户信息
	Update(ctx context.Context, tx *gorm.DB, userID int, updateFields map[string]any) error
	// GetUserInfoByID 根据用户ID获取用户信息
	GetUserInfoByID(ctx context.Context, userID int) (*dto.UserInfoResponse, error)
	// ListAllUsers 列出所有用户接口（管理员权限）
	ListAllUsers(ctx context.Context, page, pageSize int, req dto.ListUsersRequest) ([]*dto.ListUsersResponse, int64, error)
	// UpdateRefreshToken 更新用户刷新token
	UpdateRefreshToken(ctx context.Context, userID int, refreshToken string) error
	// GetUserByID 根据用户ID获取所有用户信息
	GetUserByID(ctx context.Context, userID int) (*model.User, error)
	// GetPasswordByUserName 根据用户名获取密码
	GetPasswordByUserName(ctx context.Context, userName string) (*model.User, error)
	// GetPasswordByPhoneNumber 根据手机号获取用户信息（用于登录验证）
	GetPasswordByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error)
	// GetByPhoneNumber 根据手机号查询用户是否存在
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error)
	// GetUserFieldMappings 根据用户ID获取领域映射列表
	GetUserFieldMappings(ctx context.Context, userID int) ([]*model.UserFieldMapping, error)
	// BatchCreateUserFieldMappings 批量创建用户领域映射
	BatchCreateUserFieldMappings(ctx context.Context, tx *gorm.DB, mappings []*model.UserFieldMapping) error
	// DeleteUserFieldMappings 删除用户的所有领域映射
	DeleteUserFieldMappings(ctx context.Context, tx *gorm.DB, userID int) error
	// GetUserFieldsByUserID 根据用户ID获取领域列表（含领域名称）
	GetUserFieldsByUserID(ctx context.Context, userID int) ([]dto.FieldItem, error)
	// GetUserFieldsByUserIDs 根据多个用户ID获取领域列表（含领域名称）
	GetUserFieldsByUserIDs(ctx context.Context, userIDs []int) (map[int][]dto.FieldItem, error)
}

// UserRepositoryImpl 用户仓库实现
type UserRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

// Create 创建新用户
func (repo *UserRepositoryImpl) Create(ctx context.Context, user *model.User) error {
	err := repo.db.WithContext(ctx).Create(user).Error
	if err != nil {
		exist, fieldName, _ := utils.IsUniqueConstraintError(err)
		if exist {
			switch fieldName {
			case "username":
				return utils.NewBusinessError(utils.ErrCodeResourceExists, "用户名已被注册")
			case "phone_number":
				return utils.NewBusinessError(utils.ErrCodeResourceExists, "该手机号已注册")
			default:
				return utils.NewBusinessError(utils.ErrCodeResourceExists, "用户数据已存在")
			}
		}
		return utils.NewSystemError(fmt.Errorf("创建用户失败: %w", err))
	}
	return err
}

// GetUserInfoByID 获取用户信息
func (repo *UserRepositoryImpl) GetUserInfoByID(ctx context.Context, userID int) (*dto.UserInfoResponse, error) {
	var user dto.UserInfoResponse
	query := repo.db.WithContext(ctx)

	result := query.Table("users u").
		Select(`u.user_id, u.nickname, u.avatar_url, u.name, u.gender AS gender_code,
				CASE
					WHEN gender = 'M' THEN
					'男'
					WHEN gender = 'F' THEN
					'女'
					ELSE
					'未知'
				END AS gender,u.country_code, u.phone_number, u.email, u.unit, u.department, u.position, u.industry_id, i.industry_name, u.role, ur.role_name, u.status`).
			Joins("LEFT JOIN industries i ON u.industry_id = i.id").
		Joins("LEFT JOIN user_role ur ON ur.role_code = u.role").
		Where("user_id = ?", userID).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, utils.NewSystemError(fmt.Errorf("查询用户信息失败: %v", result.Error))
	}
	return &user, nil
}

// Update 更新用户信息
func (repo *UserRepositoryImpl) Update(ctx context.Context, tx *gorm.DB, userID int, updateFields map[string]any) error {
	if tx == nil {
		tx = repo.db
	}

	result := tx.WithContext(ctx).Model(&model.User{}).
		Where("user_id = ?", userID).
		Updates(updateFields)

	if result.Error != nil {
		return utils.NewSystemError(fmt.Errorf("更新用户信息失败: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return utils.NewBusinessError(utils.ErrCodeResourceNotFound, "更新用户信息失败，用户数据异常，请刷新页面后重试")
	}
	return nil
}

// ListAllUsers 分页查询用户列表
func (repo *UserRepositoryImpl) ListAllUsers(ctx context.Context, page, pageSize int, req dto.ListUsersRequest) ([]*dto.ListUsersResponse, int64, error) {
	var users []*dto.ListUsersResponse
	var total int64

	query := repo.db.WithContext(ctx).Table("users u").
		Select(`u.user_id, u.nickname, u.avatar_url, u.name, u.gender AS gender_code,
				CASE
					WHEN gender = 'M' THEN
					'男'
					WHEN gender = 'F' THEN
					'女'
					ELSE
					'未知'
				END AS gender,
				u.country_code, u.phone_number, u.email, u.unit, u.department, u.position, u.industry_id, i.industry_name, ur.role_name,
				u.status,
				CASE
					u.status
					WHEN 1 THEN
					"已启用"
					WHEN 2 THEN
					"已禁用"
				END AS user_status
				`).
			Joins("LEFT JOIN industries i ON u.industry_id = i.id").
		Joins("LEFT JOIN user_role ur ON ur.role_code = u.role")

	// 拼接查询条件
	if req.Name != "" {
		query = query.Where("u.name LIKE ?", "%"+req.Name+"%")
	}
	if req.GenderCode != "" {
		query = query.Where("u.gender = ?", req.GenderCode)
	}
	if req.Unit != "" {
		query = query.Where("u.unit LIKE ?", "%"+req.Unit+"%")
	}
	if req.Department != "" {
		query = query.Where("u.department LIKE ?", "%"+req.Department+"%")
	}
	if req.Position != "" {
		query = query.Where("u.position LIKE ?", "%"+req.Position+"%")
	}
	if req.IndustryID != "" {
		query = query.Where("u.industry_id = ?", req.IndustryID)
	}
	if req.Role != "" {
		query = query.Where("u.role = ?", req.Role)
	}

	// 计算总记录数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.NewSystemError(fmt.Errorf("计算用户总数失败: %w", err))
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, utils.NewSystemError(fmt.Errorf("查询用户列表失败: %w", err))
	}

	return users, total, nil
}

// GetPasswordByUserName 根据用户名获取密码
func (repo *UserRepositoryImpl) GetPasswordByUserName(ctx context.Context, userName string) (*model.User, error) {
	var user model.User
	if err := repo.db.WithContext(ctx).Where("username = ? AND status = ?", userName, 1).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewBusinessError(utils.ErrCodeResourceNotFound, "用户不存在或账号已被禁用")
		}
		return nil, utils.NewSystemError(fmt.Errorf("查询用户失败: %w", err))
	}
	return &user, nil
}

// UpdateRefreshToken 更新用户刷新token
func (repo *UserRepositoryImpl) UpdateRefreshToken(ctx context.Context, userID int, refreshToken string) error {
	if err := repo.db.WithContext(ctx).Model(&model.User{}).
		Where("user_id = ?", userID).
		Update("refresh_token", refreshToken).Error; err != nil {
		return utils.NewSystemError(fmt.Errorf("更新刷新token失败: %w", err))
	}
	return nil
}

// GetUserByID 根据用户ID获取所有用户信息
func (repo *UserRepositoryImpl) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	var user model.User
	if err := repo.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, 1).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewBusinessError(utils.ErrCodeResourceNotFound, "用户不存在或账号已被禁用")
		}
		return nil, utils.NewSystemError(fmt.Errorf("查询用户失败: %w", err))
	}
	return &user, nil
}

// GetPasswordByPhoneNumber 根据手机号获取用户信息
func (repo *UserRepositoryImpl) GetPasswordByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error) {
	var user model.User
	if err := repo.db.WithContext(ctx).Where("phone_number = ? AND status = ?", phoneNumber, 1).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewBusinessError(utils.ErrCodeResourceNotFound, "手机号未注册或账号已被禁用")
		}
		return nil, utils.NewSystemError(fmt.Errorf("查询用户失败: %w", err))
	}
	return &user, nil
}

// GetByPhoneNumber 根据手机号查询用户是否存在
func (repo *UserRepositoryImpl) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error) {
	var user model.User
	if err := repo.db.WithContext(ctx).Where("phone_number = ?", phoneNumber).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, utils.NewSystemError(fmt.Errorf("查询用户失败: %w", err))
	}
	return &user, nil
}

// GetUserFieldMappings 根据用户ID获取领域映射列表
func (repo *UserRepositoryImpl) GetUserFieldMappings(ctx context.Context, userID int) ([]*model.UserFieldMapping, error) {
	var mappings []*model.UserFieldMapping
	if err := repo.db.WithContext(ctx).Where("user_id = ?", userID).Find(&mappings).Error; err != nil {
		return nil, utils.NewSystemError(fmt.Errorf("查询用户领域映射失败: %w", err))
	}
	return mappings, nil
}

// BatchCreateUserFieldMappings 批量创建用户领域映射
func (repo *UserRepositoryImpl) BatchCreateUserFieldMappings(ctx context.Context, tx *gorm.DB, mappings []*model.UserFieldMapping) error {
	if len(mappings) == 0 {
		return nil
	}
	if tx == nil {
		tx = repo.db
	}
	if err := tx.WithContext(ctx).Create(&mappings).Error; err != nil {
		exist, _, _ := utils.IsUniqueConstraintError(err)
		if exist {
			return utils.NewBusinessError(utils.ErrCodeResourceExists, "已添加该领域，请勿重复添加")
		}
		return utils.NewSystemError(fmt.Errorf("批量创建用户领域映射失败: %w", err))
	}
	return nil
}

// DeleteUserFieldMappings 删除用户的所有领域映射
func (repo *UserRepositoryImpl) DeleteUserFieldMappings(ctx context.Context, tx *gorm.DB, userID int) error {
	if tx == nil {
		tx = repo.db
	}
	if err := tx.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.UserFieldMapping{}).Error; err != nil {
		return utils.NewSystemError(fmt.Errorf("删除用户领域映射失败: %w", err))
	}
	return nil
}

// GetUserFieldMappingsByUserIDs 根据多个用户ID获取领域映射列表
func (repo *UserRepositoryImpl) GetUserFieldMappingsByUserIDs(ctx context.Context, userIDs []int) ([]*model.UserFieldMapping, error) {
	var mappings []*model.UserFieldMapping
	if len(userIDs) == 0 {
		return mappings, nil
	}
	if err := repo.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&mappings).Error; err != nil {
		return nil, utils.NewSystemError(fmt.Errorf("批量查询用户领域映射失败: %w", err))
	}
	return mappings, nil
}

// GetUserFieldsByUserID 根据用户ID获取领域列表（含领域名称）
func (repo *UserRepositoryImpl) GetUserFieldsByUserID(ctx context.Context, userID int) ([]dto.FieldItem, error) {
	var fields []dto.FieldItem
	err := repo.db.WithContext(ctx).
		Table("user_field_mappings ufm").
		Select("f.field_code, f.field_name").
		Joins("JOIN field f ON ufm.field_id = f.id").
		Where("ufm.user_id = ?", userID).
		Find(&fields).Error
	if err != nil {
		return nil, utils.NewSystemError(fmt.Errorf("查询用户领域信息失败: %w", err))
	}
	return fields, nil
}

// GetUserFieldsByUserIDs 根据多个用户ID获取领域列表（含领域名称），返回 map[userID][]FieldItem
func (repo *UserRepositoryImpl) GetUserFieldsByUserIDs(ctx context.Context, userIDs []int) (map[int][]dto.FieldItem, error) {
	type userFieldRow struct {
		UserID    int    `gorm:"column:user_id"`
		FieldCode string `gorm:"column:field_code"`
		FieldName string `gorm:"column:field_name"`
	}
	var rows []userFieldRow
	if len(userIDs) == 0 {
		return make(map[int][]dto.FieldItem), nil
	}
	err := repo.db.WithContext(ctx).
		Table("user_field_mappings ufm").
		Select("ufm.user_id, f.field_code, f.field_name").
		Joins("JOIN field f ON ufm.field_id = f.id").
		Where("ufm.user_id IN ?", userIDs).
		Find(&rows).Error
	if err != nil {
		return nil, utils.NewSystemError(fmt.Errorf("批量查询用户领域信息失败: %w", err))
	}
	result := make(map[int][]dto.FieldItem)
	for _, row := range rows {
		result[row.UserID] = append(result[row.UserID], dto.FieldItem{
			FieldCode: row.FieldCode,
			FieldName: row.FieldName,
		})
	}
	return result, nil
}
