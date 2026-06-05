package controller

import (
	"event-platform/internal/dashboard/dto"
	"event-platform/internal/dashboard/service"
	"event-platform/internal/utils"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	dashboardService service.DashboardService
}

func NewDashboardController(dashboardService service.DashboardService) *DashboardController {
	return &DashboardController{dashboardService: dashboardService}
}

func (ctr *DashboardController) ListDashboardUsers(ctx *gin.Context) {
	var req dto.DashboardUserListRequest
	if !utils.BindQuery(ctx, &req) {
		return
	}

	result, filteredTotal, err := ctr.dashboardService.ListDashboardUsers(ctx, &req)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}

	utils.Success(ctx, "success", gin.H{
		"total_count":    result.TotalCount,
		"filtered_total": filteredTotal,
		"page":           page,
		"page_size":       pageSize,
		"list":            result.List,
	})
}

func (ctr *DashboardController) ListDashboardEvents(ctx *gin.Context) {
	var req dto.DashboardEventListRequest
	if !utils.BindQuery(ctx, &req) {
		return
	}

	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}

	events, total, err := ctr.dashboardService.ListDashboardEvents(ctx, &req)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.SuccessPage(ctx, total, page, pageSize, events)
}

func (ctr *DashboardController) GetEventOverview(ctx *gin.Context) {
	var req dto.DashboardEventIDRequest
	if !utils.BindUrl(ctx, &req) {
		return
	}

	overview, err := ctr.dashboardService.GetEventOverview(ctx, req.ID)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", overview)
}

func (ctr *DashboardController) ListEventRegUsers(ctx *gin.Context) {
	var urlReq dto.DashboardEventIDRequest
	if !utils.BindUrl(ctx, &urlReq) {
		return
	}

	var req dto.DashboardRegUserListRequest
	if !utils.BindQuery(ctx, &req) {
		return
	}

	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}

	users, total, err := ctr.dashboardService.ListEventRegUsers(ctx, urlReq.ID, &req)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.SuccessPage(ctx, total, page, pageSize, users)
}

func (ctr *DashboardController) GetEventStatistics(ctx *gin.Context) {
	var req dto.DashboardEventIDRequest
	if !utils.BindUrl(ctx, &req) {
		return
	}

	stats, err := ctr.dashboardService.GetEventStatistics(ctx, req.ID)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", stats)
}