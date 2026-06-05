package service

import (
	"context"
	"event-platform/internal/dashboard/dto"
	"event-platform/internal/dashboard/repository"
	userrepo "event-platform/internal/user/repository"
)

type DashboardService interface {
	ListDashboardUsers(ctx context.Context, req *dto.DashboardUserListRequest) (*dto.DashboardUserListResponse, int64, error)
	ListDashboardEvents(ctx context.Context, req *dto.DashboardEventListRequest) ([]*dto.DashboardEventItem, int64, error)
	GetEventOverview(ctx context.Context, eventID int) (*dto.DashboardEventOverviewResponse, error)
	ListEventRegUsers(ctx context.Context, eventID int, req *dto.DashboardRegUserListRequest) ([]*dto.DashboardRegUserItem, int64, error)
	GetEventStatistics(ctx context.Context, eventID int) (*dto.DashboardEventStatisticsResponse, error)
}

type DashboardServiceImpl struct {
	dashboardRepo repository.DashboardRepository
	userRepo      userrepo.UserRepository
}

func NewDashboardService(dashboardRepo repository.DashboardRepository, userRepo userrepo.UserRepository) DashboardService {
	return &DashboardServiceImpl{
		dashboardRepo: dashboardRepo,
		userRepo:      userRepo,
	}
}

func (svc *DashboardServiceImpl) ListDashboardUsers(ctx context.Context, req *dto.DashboardUserListRequest) (*dto.DashboardUserListResponse, int64, error) {
	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}

	users, totalCount, filteredTotal, err := svc.dashboardRepo.ListDashboardUsers(ctx, page, pageSize, req)
	if err != nil {
		return nil, 0, err
	}

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
				u.Fields = make([]dto.FieldItem, 0, len(fields))
				for _, f := range fields {
					u.Fields = append(u.Fields, dto.FieldItem{
						FieldCode: f.FieldCode,
						FieldName: f.FieldName,
					})
				}
			} else {
				u.Fields = []dto.FieldItem{}
			}
		}
	}

	return &dto.DashboardUserListResponse{
		TotalCount: totalCount,
		List:       users,
	}, filteredTotal, nil
}

func (svc *DashboardServiceImpl) ListDashboardEvents(ctx context.Context, req *dto.DashboardEventListRequest) ([]*dto.DashboardEventItem, int64, error) {
	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}
	return svc.dashboardRepo.ListDashboardEvents(ctx, page, pageSize, req.Title)
}

func (svc *DashboardServiceImpl) GetEventOverview(ctx context.Context, eventID int) (*dto.DashboardEventOverviewResponse, error) {
	return svc.dashboardRepo.GetEventOverview(ctx, eventID)
}

func (svc *DashboardServiceImpl) ListEventRegUsers(ctx context.Context, eventID int, req *dto.DashboardRegUserListRequest) ([]*dto.DashboardRegUserItem, int64, error) {
	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}
	return svc.dashboardRepo.ListEventRegUsers(ctx, page, pageSize, eventID, req)
}

func (svc *DashboardServiceImpl) GetEventStatistics(ctx context.Context, eventID int) (*dto.DashboardEventStatisticsResponse, error) {
	return svc.dashboardRepo.GetEventStatistics(ctx, eventID)
}
