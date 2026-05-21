package service

import (
	"context"
	"event-platform/internal/field/dto"
	"event-platform/internal/field/model"
	"event-platform/internal/field/repository"
	"event-platform/internal/utils"
)

// FieldService 领域服务接口
type FieldService interface {
	// CreateField 创建领域
	CreateField(ctx context.Context, req dto.CreateFieldRequest, operator int) error
	// UpdateField 更新领域
	UpdateField(ctx context.Context, fieldID int, req dto.UpdateFieldRequest, operator int) error
	// DeleteField 删除领域
	DeleteField(ctx context.Context, fieldID int, operator int) error
	// UpdateFieldStatus 更新领域状态
	UpdateFieldStatus(ctx context.Context, fieldID int, req dto.UpdateFieldStatusRequest, operator int) error
	// ListFields 查询领域列表
	ListFields(ctx context.Context, req dto.ListFieldsRequest) ([]*dto.ListFieldsResponse, error)
}

// FieldServiceImpl 领域服务实现
type FieldServiceImpl struct {
	fieldRepo repository.FieldRepository
}

// NewFieldService 创建领域服务实例
func NewFieldService(fieldRepo repository.FieldRepository) FieldService {
	return &FieldServiceImpl{fieldRepo: fieldRepo}
}

// CreateField 创建领域
func (svc *FieldServiceImpl) CreateField(ctx context.Context, req dto.CreateFieldRequest, operator int) error {
	field := &model.Field{
		FieldCode:  req.FieldCode,
		FieldName:  req.FieldName,
		Desc:       req.Desc,
		Enable:     1,
		CreateUser: operator,
		UpdateUser: operator,
	}
	return svc.fieldRepo.Create(ctx, field)
}

// UpdateField 更新领域
func (svc *FieldServiceImpl) UpdateField(ctx context.Context, fieldID int, req dto.UpdateFieldRequest, operator int) error {
	updateFields := make(map[string]interface{})
	if req.FieldCode != "" {
		updateFields["field_code"] = req.FieldCode
	}
	if req.FieldName != "" {
		updateFields["field_name"] = req.FieldName
	}
	if req.Desc != "" {
		updateFields["desc"] = req.Desc
	}
	if len(updateFields) == 0 {
		return nil
	}
	updateFields["update_user"] = operator
	return svc.fieldRepo.Update(ctx, fieldID, updateFields)
}

func (svc *FieldServiceImpl) DeleteField(ctx context.Context, fieldID int, operator int) error {
	return svc.fieldRepo.Delete(ctx, fieldID)
}

// UpdateFieldStatus 更新领域状态
func (svc *FieldServiceImpl) UpdateFieldStatus(ctx context.Context, fieldID int, req dto.UpdateFieldStatusRequest, operator int) error {
	field, err := svc.fieldRepo.GetByID(ctx, fieldID)
	if err != nil {
		return err
	}
	if field == nil {
		return utils.NewBusinessError(utils.ErrCodeResourceNotFound, "领域不存在")
	}

	if req.Operation == "ENABLE" {
		if field.Enable == 1 {
			return utils.NewBusinessError(utils.ErrCodeResourceConflict, "领域已启用，请勿重复操作")
		}
	} else {
		if field.Enable == 2 {
			return utils.NewBusinessError(utils.ErrCodeResourceConflict, "领域已禁用，请勿重复操作")
		}
	}

	updateFields := make(map[string]interface{})
	if req.Operation == "ENABLE" {
		updateFields["enable"] = 1
	} else {
		updateFields["enable"] = 2
	}
	updateFields["update_user"] = operator
	return svc.fieldRepo.Update(ctx, fieldID, updateFields)
}

// ListFields 查询领域列表
func (svc *FieldServiceImpl) ListFields(ctx context.Context, req dto.ListFieldsRequest) ([]*dto.ListFieldsResponse, error) {
	return svc.fieldRepo.ListFields(ctx, req)
}
