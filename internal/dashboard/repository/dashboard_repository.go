package repository

import (
	"context"
	"event-platform/internal/dashboard/dto"
	"event-platform/internal/utils"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type DashboardRepository interface {
	ListDashboardUsers(ctx context.Context, page, pageSize int, req *dto.DashboardUserListRequest) ([]*dto.DashboardUserItem, int64, int64, error)
	ListDashboardEvents(ctx context.Context, page, pageSize int, title string) ([]*dto.DashboardEventItem, int64, error)
	GetEventOverview(ctx context.Context, eventID int) (*dto.DashboardEventOverviewResponse, error)
	ListEventRegUsers(ctx context.Context, page, pageSize, eventID int, req *dto.DashboardRegUserListRequest) ([]*dto.DashboardRegUserItem, int64, error)
	GetEventStatistics(ctx context.Context, eventID int) (*dto.DashboardEventStatisticsResponse, error)
}

type DashboardRepositoryImpl struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &DashboardRepositoryImpl{db: db}
}

func (repo *DashboardRepositoryImpl) ListDashboardUsers(ctx context.Context, page, pageSize int, req *dto.DashboardUserListRequest) ([]*dto.DashboardUserItem, int64, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var totalCount int64
	if err := repo.db.WithContext(ctx).Model(&struct {
		UserID int `gorm:"column:user_id"`
	}{}).
		Table("users").
		Where("status = ?", utils.UserStatusEnabled).
		Count(&totalCount).Error; err != nil {
		return nil, 0, 0, utils.NewSystemError(fmt.Errorf("查询总用户数失败: %w", err))
	}

	var users []*dto.DashboardUserItem
	var filteredTotal int64

	query := repo.db.WithContext(ctx).Table("users u").
		Select(`u.user_id, u.name, u.phone_number, u.email, u.unit, u.department, u.position, u.industry_id, i.industry_name`).
		Joins("LEFT JOIN industries i ON u.industry_id = i.id").
		Where("u.status = ?", utils.UserStatusEnabled)

	if req.FieldID > 0 {
		query = query.Joins("JOIN user_field_mappings ufm ON u.user_id = ufm.user_id").
			Where("ufm.field_id = ?", req.FieldID)
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
	if req.IndustryID > 0 {
		query = query.Where("u.industry_id = ?", req.IndustryID)
	}

	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&filteredTotal).Error; err != nil {
		return nil, 0, 0, utils.NewSystemError(fmt.Errorf("计算筛选用户数失败: %w", err))
	}

	if err := query.Group("u.user_id").Order("u.user_id DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, 0, utils.NewSystemError(fmt.Errorf("查询用户列表失败: %w", err))
	}

	return users, totalCount, filteredTotal, nil
}

func (repo *DashboardRepositoryImpl) ListDashboardEvents(ctx context.Context, page, pageSize int, title string) ([]*dto.DashboardEventItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var events []*dto.DashboardEventItem
	var total int64

	now := time.Now()
	query := repo.db.WithContext(ctx).Table("events e").
		Select(`e.id, e.title, e.event_start_time, e.event_end_time, e.max_registrants, e.current_registrants,
			CASE
				WHEN e.registration_start_time > ? THEN 'NotBegun'
				WHEN e.registration_start_time <= ? AND e.registration_end_time >= ? THEN 'InProgress'
				ELSE 'Completed'
			END AS status`, now, now, now).
		Where("e.is_deleted = ?", utils.DeletedFlagNo)

	if title != "" {
		query = query.Where("e.title LIKE ?", "%"+title+"%")
	}

	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, utils.NewSystemError(fmt.Errorf("计算活动总数失败: %w", err))
	}

	if err := query.Order("e.event_start_time DESC").Offset(offset).Limit(pageSize).Find(&events).Error; err != nil {
		return nil, 0, utils.NewSystemError(fmt.Errorf("查询活动列表失败: %w", err))
	}

	return events, total, nil
}

func (repo *DashboardRepositoryImpl) GetEventOverview(ctx context.Context, eventID int) (*dto.DashboardEventOverviewResponse, error) {
	var result dto.DashboardEventOverviewResponse
	now := time.Now()
	err := repo.db.WithContext(ctx).Table("events e").
		Select(`e.id, e.title, e.event_start_time, e.event_end_time, e.registration_start_time, e.registration_end_time,
			e.max_registrants, e.current_registrants, e.event_address,
			CASE
				WHEN e.registration_start_time > ? THEN 'NotBegun'
				WHEN e.registration_start_time <= ? AND e.registration_end_time >= ? THEN 'InProgress'
				ELSE 'Completed'
			END AS status`, now, now, now).
		Where("e.id = ? AND e.is_deleted = ?", eventID, utils.DeletedFlagNo).
		Take(&result).Error
	if err != nil {
		return nil, utils.NewSystemError(fmt.Errorf("查询活动概览失败: %w", err))
	}
	return &result, nil
}

func (repo *DashboardRepositoryImpl) ListEventRegUsers(ctx context.Context, page, pageSize, eventID int, req *dto.DashboardRegUserListRequest) ([]*dto.DashboardRegUserItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var users []*dto.DashboardRegUserItem
	var total int64

	query := repo.db.WithContext(ctx).Table("event_registration_info eri").
		Select(`eri.user_id, eri.name, eri.phone_number, eri.email, eri.unit, eri.department, eri.position, eri.industry_id, i.industry_name`).
		Joins("LEFT JOIN industries i ON eri.industry_id = i.id").
		Where("eri.event_id = ?", eventID)

	if req.FieldID > 0 {
		query = query.Joins("JOIN user_field_mappings ufm ON eri.user_id = ufm.user_id").
			Where("ufm.field_id = ?", req.FieldID)
	}
	if req.Unit != "" {
		query = query.Where("eri.unit LIKE ?", "%"+req.Unit+"%")
	}
	if req.Department != "" {
		query = query.Where("eri.department LIKE ?", "%"+req.Department+"%")
	}
	if req.Position != "" {
		query = query.Where("eri.position LIKE ?", "%"+req.Position+"%")
	}
	if req.IndustryID > 0 {
		query = query.Where("eri.industry_id = ?", req.IndustryID)
	}

	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, utils.NewSystemError(fmt.Errorf("计算报名用户总数失败: %w", err))
	}

	if err := query.Group("eri.user_id").Order("eri.create_time DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, utils.NewSystemError(fmt.Errorf("查询报名用户列表失败: %w", err))
	}

	return users, total, nil
}

func (repo *DashboardRepositoryImpl) GetEventStatistics(ctx context.Context, eventID int) (*dto.DashboardEventStatisticsResponse, error) {
	result := &dto.DashboardEventStatisticsResponse{
		ByField:      make([]dto.StatItem, 0),
		ByUnit:       make([]dto.StatItem, 0),
		ByDepartment: make([]dto.StatItem, 0),
		ByPosition:   make([]dto.StatItem, 0),
	}

	if err := repo.db.WithContext(ctx).Table("event_registration_info eri").
		Select("f.field_name AS name, COUNT(DISTINCT eri.user_id) AS count").
		Joins("JOIN user_field_mappings ufm ON eri.user_id = ufm.user_id").
		Joins("JOIN field f ON ufm.field_id = f.id").
		Where("eri.event_id = ?", eventID).
		Group("f.field_name").
		Order("count DESC").
		Find(&result.ByField).Error; err != nil {
		return nil, utils.NewSystemError(fmt.Errorf("统计领域分布失败: %w", err))
	}

	if err := repo.db.WithContext(ctx).Table("event_registration_info").
		Select("unit AS name, COUNT(DISTINCT user_id) AS count").
		Where("event_id = ? AND unit IS NOT NULL AND unit != ''", eventID).
		Group("unit").
		Order("count DESC").
		Find(&result.ByUnit).Error; err != nil {
		return nil, utils.NewSystemError(fmt.Errorf("统计单位分布失败: %w", err))
	}

	if err := repo.db.WithContext(ctx).Table("event_registration_info").
		Select("department AS name, COUNT(DISTINCT user_id) AS count").
		Where("event_id = ? AND department IS NOT NULL AND department != ''", eventID).
		Group("department").
		Order("count DESC").
		Find(&result.ByDepartment).Error; err != nil {
		return nil, utils.NewSystemError(fmt.Errorf("统计部门分布失败: %w", err))
	}

	if err := repo.db.WithContext(ctx).Table("event_registration_info").
		Select("position AS name, COUNT(DISTINCT user_id) AS count").
		Where("event_id = ? AND position IS NOT NULL AND position != ''", eventID).
		Group("position").
		Order("count DESC").
		Find(&result.ByPosition).Error; err != nil {
		return nil, utils.NewSystemError(fmt.Errorf("统计职位分布失败: %w", err))
	}

	return result, nil
}
