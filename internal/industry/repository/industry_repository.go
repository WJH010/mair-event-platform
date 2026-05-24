package repository

import (
	"context"
	"event-platform/internal/industry/dto"
	"event-platform/internal/industry/model"
	"event-platform/internal/utils"
	"fmt"

	"gorm.io/gorm"
)

// IndustryRepository 行业仓库接口
type IndustryRepository interface {
	// Create 创建行业
	Create(ctx context.Context, industry *model.Industries) error
	// Update 更新行业
	Update(ctx context.Context, industryID int, updateFields map[string]interface{}) error
	// GetByID 根据ID获取行业
	GetByID(ctx context.Context, industryID int) (*model.Industries, error)
	// Delete 硬删除行业
	Delete(ctx context.Context, industryID int) error
	// ListIndustries 查询行业列表
	ListIndustries(ctx context.Context, req dto.ListIndustriesRequest) ([]*dto.ListIndustriesResponse, error)
}

// IndustryRepositoryImpl 行业仓库实现
type IndustryRepositoryImpl struct {
	db *gorm.DB
}

// NewIndustryRepository 创建行业仓库实例
func NewIndustryRepository(db *gorm.DB) IndustryRepository {
	return &IndustryRepositoryImpl{db: db}
}

// Create 创建行业
func (repo *IndustryRepositoryImpl) Create(ctx context.Context, industry *model.Industries) error {
	if err := repo.db.WithContext(ctx).Create(industry).Error; err != nil {
		ok, fieldName, _ := utils.IsUniqueConstraintError(err)
		if ok {
			if fieldName == "industry_code" {
				return utils.NewBusinessError(utils.ErrCodeResourceExists, "行业代码已存在")
			}
			return utils.NewBusinessError(utils.ErrCodeResourceExists, "行业已存在")
		}
		return utils.NewSystemError(fmt.Errorf("创建行业失败: %w", err))
	}
	return nil
}

// Update 更新行业
func (repo *IndustryRepositoryImpl) Update(ctx context.Context, industryID int, updateFields map[string]interface{}) error {
	var count int64
	if err := repo.db.WithContext(ctx).
		Model(&model.Industries{}).
		Where("id = ?", industryID).
		Count(&count).Error; err != nil {
		return utils.NewSystemError(fmt.Errorf("检查行业存在性失败: %w", err))
	}
	if count == 0 {
		return utils.NewBusinessError(utils.ErrCodeResourceNotFound, "行业不存在")
	}
	result := repo.db.WithContext(ctx).Model(&model.Industries{}).
		Where("id = ?", industryID).Updates(updateFields)
	if result.Error != nil {
		ok, fieldName, _ := utils.IsUniqueConstraintError(result.Error)
		if ok {
			if fieldName == "industry_code" {
				return utils.NewBusinessError(utils.ErrCodeResourceExists, "行业代码已存在")
			}
			return utils.NewBusinessError(utils.ErrCodeResourceExists, "行业已存在")
		}
		return utils.NewSystemError(fmt.Errorf("更新行业失败: %w", result.Error))
	}
	return nil
}

// GetByID 根据ID获取行业
func (repo *IndustryRepositoryImpl) GetByID(ctx context.Context, industryID int) (*model.Industries, error) {
	var industry model.Industries
	if err := repo.db.WithContext(ctx).Where("id = ?", industryID).First(&industry).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, utils.NewSystemError(fmt.Errorf("查询行业失败: %w", err))
	}
	return &industry, nil
}

// Delete 硬删除行业
func (repo *IndustryRepositoryImpl) Delete(ctx context.Context, industryID int) error {
	result := repo.db.WithContext(ctx).Where("id = ?", industryID).Delete(&model.Industries{})
	if result.Error != nil {
		return utils.NewSystemError(fmt.Errorf("删除行业失败: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return utils.NewBusinessError(utils.ErrCodeResourceNotFound, "行业不存在")
	}
	return nil
}

// ListIndustries 查询行业列表
func (repo *IndustryRepositoryImpl) ListIndustries(ctx context.Context, req dto.ListIndustriesRequest) ([]*dto.ListIndustriesResponse, error) {
	var industries []*dto.ListIndustriesResponse
	query := repo.db.WithContext(ctx).Table("industries").
		Select("id, industry_code, industry_name, `desc`, enable")
	if req.IndustryName != "" {
		query = query.Where("industry_name LIKE ?", "%"+req.IndustryName+"%")
	}
	if req.Enable != nil {
		query = query.Where("enable = ?", *req.Enable)
	}
	if err := query.Find(&industries).Error; err != nil {
		return nil, utils.NewSystemError(fmt.Errorf("查询行业列表失败: %w", err))
	}
	return industries, nil
}
