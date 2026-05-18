package controller

import (
	"event-platform/internal/field/dto"
	"event-platform/internal/field/service"
	"event-platform/internal/utils"

	"github.com/gin-gonic/gin"
)

// FieldController 领域控制器
type FieldController struct {
	fieldService service.FieldService
}

// NewFieldController 创建领域控制器实例
func NewFieldController(fieldService service.FieldService) *FieldController {
	return &FieldController{fieldService: fieldService}
}

// CreateField 创建领域
func (ctr *FieldController) CreateField(ctx *gin.Context) {
	var req dto.CreateFieldRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	operator, err := utils.GetUserID(ctx)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	err = ctr.fieldService.CreateField(ctx, req, operator)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

// UpdateField 更新领域
func (ctr *FieldController) UpdateField(ctx *gin.Context) {
	var urlReq dto.FieldUrlID
	if !utils.BindUrl(ctx, &urlReq) {
		return
	}

	var req dto.UpdateFieldRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	operator, err := utils.GetUserID(ctx)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	err = ctr.fieldService.UpdateField(ctx, urlReq.ID, req, operator)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

// DeleteField 删除领域
func (ctr *FieldController) DeleteField(ctx *gin.Context) {
	var urlReq dto.FieldUrlID
	if !utils.BindUrl(ctx, &urlReq) {
		return
	}

	operator, err := utils.GetUserID(ctx)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	err = ctr.fieldService.DeleteField(ctx, urlReq.ID, operator)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

// UpdateFieldStatus 更新领域状态
func (ctr *FieldController) UpdateFieldStatus(ctx *gin.Context) {
	var urlReq dto.FieldUrlID
	if !utils.BindUrl(ctx, &urlReq) {
		return
	}

	var req dto.UpdateFieldStatusRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	operator, err := utils.GetUserID(ctx)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	err = ctr.fieldService.UpdateFieldStatus(ctx, urlReq.ID, req, operator)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

// ListFields 查询领域列表
func (ctr *FieldController) ListFields(ctx *gin.Context) {
	var req dto.ListFieldsRequest
	if !utils.BindQuery(ctx, &req) {
		return
	}

	fields, err := ctr.fieldService.ListFields(ctx, req)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", fields)
}
