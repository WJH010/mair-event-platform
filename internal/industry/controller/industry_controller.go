package controller

import (
	"event-platform/internal/industry/dto"
	"event-platform/internal/industry/service"
	"event-platform/internal/utils"

	"github.com/gin-gonic/gin"
)

// IndustryController 行业控制器
type IndustryController struct {
	industryService service.IndustryService
}

// NewIndustryController 创建行业控制器实例
func NewIndustryController(industryService service.IndustryService) *IndustryController {
	return &IndustryController{industryService: industryService}
}

// CreateIndustry 创建行业
func (ctr *IndustryController) CreateIndustry(ctx *gin.Context) {
	var req dto.CreateIndustryRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	operator, err := utils.GetUserID(ctx)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	err = ctr.industryService.CreateIndustry(ctx, req, operator)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

// UpdateIndustry 更新行业
func (ctr *IndustryController) UpdateIndustry(ctx *gin.Context) {
	var urlReq dto.IndustryUrlID
	if !utils.BindUrl(ctx, &urlReq) {
		return
	}

	var req dto.UpdateIndustryRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	operator, err := utils.GetUserID(ctx)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	err = ctr.industryService.UpdateIndustry(ctx, urlReq.ID, req, operator)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

// DeleteIndustry 删除行业
func (ctr *IndustryController) DeleteIndustry(ctx *gin.Context) {
	var urlReq dto.IndustryUrlID
	if !utils.BindUrl(ctx, &urlReq) {
		return
	}

	operator, err := utils.GetUserID(ctx)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	err = ctr.industryService.DeleteIndustry(ctx, urlReq.ID, operator)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

// UpdateIndustryStatus 更新行业状态
func (ctr *IndustryController) UpdateIndustryStatus(ctx *gin.Context) {
	var urlReq dto.IndustryUrlID
	if !utils.BindUrl(ctx, &urlReq) {
		return
	}

	var req dto.UpdateIndustryStatusRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	operator, err := utils.GetUserID(ctx)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	err = ctr.industryService.UpdateIndustryStatus(ctx, urlReq.ID, req, operator)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

// ListIndustries 查询行业列表
func (ctr *IndustryController) ListIndustries(ctx *gin.Context) {
	var req dto.ListIndustriesRequest
	if !utils.BindQuery(ctx, &req) {
		return
	}

	industries, err := ctr.industryService.ListIndustries(ctx, req)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", industries)
}
