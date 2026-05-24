package service

import (
	"context"
	"event-platform/internal/industry/dto"
	"event-platform/internal/industry/model"
	"event-platform/internal/industry/repository"
	"event-platform/internal/utils"
)

// IndustryService 行业服务接口
type IndustryService interface {
	// CreateIndustry 创建行业
	CreateIndustry(ctx context.Context, req dto.CreateIndustryRequest, operator int) error
	// UpdateIndustry 更新行业
	UpdateIndustry(ctx context.Context, industryID int, req dto.UpdateIndustryRequest, operator int) error
	// DeleteIndustry 删除行业
	DeleteIndustry(ctx context.Context, industryID int, operator int) error
	// UpdateIndustryStatus 更新行业状态
	UpdateIndustryStatus(ctx context.Context, industryID int, req dto.UpdateIndustryStatusRequest, operator int) error
	// ListIndustries 查询行业列表
	ListIndustries(ctx context.Context, req dto.ListIndustriesRequest) ([]*dto.ListIndustriesResponse, error)
}

// IndustryServiceImpl 行业服务实现
type IndustryServiceImpl struct {
	industryRepo repository.IndustryRepository
}

// NewIndustryService 创建行业服务实例
func NewIndustryService(industryRepo repository.IndustryRepository) IndustryService {
	return &IndustryServiceImpl{industryRepo: industryRepo}
}

// CreateIndustry 创建行业
func (svc *IndustryServiceImpl) CreateIndustry(ctx context.Context, req dto.CreateIndustryRequest, operator int) error {
	industry := &model.Industries{
		IndustryCode: req.IndustryCode,
		IndustryName: req.IndustryName,
		Desc:         req.Desc,
		Enable:       1,
		CreateUser:   operator,
		UpdateUser:   operator,
	}
	return svc.industryRepo.Create(ctx, industry)
}

// UpdateIndustry 更新行业
func (svc *IndustryServiceImpl) UpdateIndustry(ctx context.Context, industryID int, req dto.UpdateIndustryRequest, operator int) error {
	updateFields := make(map[string]interface{})
	if req.IndustryCode != "" {
		updateFields["industry_code"] = req.IndustryCode
	}
	if req.IndustryName != "" {
		updateFields["industry_name"] = req.IndustryName
	}
	if req.Desc != "" {
		updateFields["desc"] = req.Desc
	}
	if len(updateFields) == 0 {
		return nil
	}
	updateFields["update_user"] = operator
	return svc.industryRepo.Update(ctx, industryID, updateFields)
}

// DeleteIndustry 删除行业
func (svc *IndustryServiceImpl) DeleteIndustry(ctx context.Context, industryID int, operator int) error {
	return svc.industryRepo.Delete(ctx, industryID)
}

// UpdateIndustryStatus 更新行业状态
func (svc *IndustryServiceImpl) UpdateIndustryStatus(ctx context.Context, industryID int, req dto.UpdateIndustryStatusRequest, operator int) error {
	industry, err := svc.industryRepo.GetByID(ctx, industryID)
	if err != nil {
		return err
	}
	if industry == nil {
		return utils.NewBusinessError(utils.ErrCodeResourceNotFound, "行业不存在")
	}

	if req.Operation == "ENABLE" {
		if industry.Enable == 1 {
			return utils.NewBusinessError(utils.ErrCodeResourceConflict, "行业已启用，请勿重复操作")
		}
	} else {
		if industry.Enable == 2 {
			return utils.NewBusinessError(utils.ErrCodeResourceConflict, "行业已禁用，请勿重复操作")
		}
	}

	updateFields := make(map[string]interface{})
	if req.Operation == "ENABLE" {
		updateFields["enable"] = 1
	} else {
		updateFields["enable"] = 2
	}
	updateFields["update_user"] = operator
	return svc.industryRepo.Update(ctx, industryID, updateFields)
}

// ListIndustries 查询行业列表
func (svc *IndustryServiceImpl) ListIndustries(ctx context.Context, req dto.ListIndustriesRequest) ([]*dto.ListIndustriesResponse, error) {
	return svc.industryRepo.ListIndustries(ctx, req)
}
