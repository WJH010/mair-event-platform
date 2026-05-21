package repository

import (
	"context"
	"event-platform/internal/field/dto"
	"event-platform/internal/field/model"
	"event-platform/internal/utils"
	"fmt"

	"gorm.io/gorm"
)

// FieldRepository 领域仓库接口
type FieldRepository interface {
	// Create 创建领域
	Create(ctx context.Context, field *model.Field) error
	// Update 更新领域
	Update(ctx context.Context, fieldID int, updateFields map[string]interface{}) error
	// GetByID 根据ID获取领域
	GetByID(ctx context.Context, fieldID int) (*model.Field, error)
	// Delete 硬删除领域
	Delete(ctx context.Context, fieldID int) error
	// ListFields 分页查询领域
	ListFields(ctx context.Context, req dto.ListFieldsRequest) ([]*dto.ListFieldsResponse, error)
}

// FieldRepositoryImpl 领域仓库实现
type FieldRepositoryImpl struct {
	db *gorm.DB
}

// NewFieldRepository 创建领域仓库实例
func NewFieldRepository(db *gorm.DB) FieldRepository {
	return &FieldRepositoryImpl{db: db}
}

// Create 创建领域
func (repo *FieldRepositoryImpl) Create(ctx context.Context, field *model.Field) error {
	if err := repo.db.WithContext(ctx).Create(field).Error; err != nil {
		ok, fieldName, _ := utils.IsUniqueConstraintError(err)
		if ok {
			if fieldName == "field_code" {
				return utils.NewBusinessError(utils.ErrCodeResourceExists, "领域代码已存在")
			}
			return utils.NewBusinessError(utils.ErrCodeResourceExists, "领域已存在")
		}
		return utils.NewSystemError(fmt.Errorf("创建领域失败: %w", err))
	}
	return nil
}

// Update 更新领域
func (repo *FieldRepositoryImpl) Update(ctx context.Context, fieldID int, updateFields map[string]interface{}) error {
	var count int64
	// 检查领域是否存在
	if err := repo.db.WithContext(ctx).
		Model(&model.Field{}).
		Where("id = ?", fieldID).
		Count(&count).Error; err != nil {
		return utils.NewSystemError(fmt.Errorf("检查领域存在性失败: %w", err))
	}
	if count == 0 {
		return utils.NewBusinessError(utils.ErrCodeResourceNotFound, "领域不存在")
	}
	// 更新领域
	result := repo.db.WithContext(ctx).Model(&model.Field{}).
		Where("id = ?", fieldID).Updates(updateFields)
	// 检查更新结果
	if result.Error != nil {
		ok, fieldName, _ := utils.IsUniqueConstraintError(result.Error)
		if ok {
			if fieldName == "field_code" {
				return utils.NewBusinessError(utils.ErrCodeResourceExists, "领域代码已存在")
			}
			return utils.NewBusinessError(utils.ErrCodeResourceExists, "领域已存在")
		}
		return utils.NewSystemError(fmt.Errorf("更新领域失败: %w", result.Error))
	}
	return nil
}

// GetByID 根据ID获取领域
func (repo *FieldRepositoryImpl) GetByID(ctx context.Context, fieldID int) (*model.Field, error) {
	var field model.Field
	if err := repo.db.WithContext(ctx).Where("id = ?", fieldID).First(&field).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, utils.NewSystemError(fmt.Errorf("查询领域失败: %w", err))
	}
	return &field, nil
}

// Delete 硬删除领域
func (repo *FieldRepositoryImpl) Delete(ctx context.Context, fieldID int) error {
	result := repo.db.WithContext(ctx).Where("id = ?", fieldID).Delete(&model.Field{})
	if result.Error != nil {
		return utils.NewSystemError(fmt.Errorf("删除领域失败: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return utils.NewBusinessError(utils.ErrCodeResourceNotFound, "领域不存在")
	}
	return nil
}

// ListFields 查询领域列表
func (repo *FieldRepositoryImpl) ListFields(ctx context.Context, req dto.ListFieldsRequest) ([]*dto.ListFieldsResponse, error) {
	var fields []*dto.ListFieldsResponse
	query := repo.db.WithContext(ctx).Table("field").
		Select("id, field_code, field_name, `desc`, enable")
	// 领域名称查询参数
	if req.FieldName != "" {
		query = query.Where("field_name LIKE ?", "%"+req.FieldName+"%")
	}
	// 状态查询参数
	if req.Enable != nil {
		query = query.Where("enable = ?", *req.Enable)
	}

	if err := query.Find(&fields).Error; err != nil {
		return nil, utils.NewSystemError(fmt.Errorf("查询领域列表失败: %w", err))
	}
	return fields, nil
}
