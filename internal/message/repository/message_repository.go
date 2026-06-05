package repository

import (
	"context"
	"errors"
	"event-platform/internal/message/dto"
	"event-platform/internal/message/model"
	"event-platform/internal/utils"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type MessageRepository interface {
	ExecTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error
	CreateMessage(ctx context.Context, tx *gorm.DB, message *model.Message) error
	BatchCreateUserMessageMappings(ctx context.Context, tx *gorm.DB, mappings []*model.UserMessageMapping) error
	GetMessageByID(ctx context.Context, messageID int) (*model.Message, error)
	SoftDeleteMessage(ctx context.Context, tx *gorm.DB, messageID int) error
	DeleteUserMessageMappingsByMessageID(ctx context.Context, tx *gorm.DB, messageID int) error
	ListMessages(ctx context.Context, page, pageSize int, title, targetType string, targetID, userID int) ([]*dto.MessageListResponse, int64, error)
	ListUserMessages(ctx context.Context, page, pageSize int, userID int) ([]*dto.UserMessageListResponse, int64, error)
	MarkAllAsRead(ctx context.Context, userID int) error
	GetUnreadCount(ctx context.Context, userID int) (int64, error)
}

type MessageRepositoryImpl struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &MessageRepositoryImpl{db: db}
}

func (repo *MessageRepositoryImpl) ExecTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return repo.db.WithContext(ctx).Transaction(fn)
}

func (repo *MessageRepositoryImpl) CreateMessage(ctx context.Context, tx *gorm.DB, message *model.Message) error {
	if err := tx.WithContext(ctx).Create(message).Error; err != nil {
		return utils.NewSystemError(fmt.Errorf("创建消息失败: %w", err))
	}
	return nil
}

func (repo *MessageRepositoryImpl) BatchCreateUserMessageMappings(ctx context.Context, tx *gorm.DB, mappings []*model.UserMessageMapping) error {
	if len(mappings) == 0 {
		return nil
	}
	batchSize := 500
	for i := 0; i < len(mappings); i += batchSize {
		end := i + batchSize
		if end > len(mappings) {
			end = len(mappings)
		}
		if err := tx.WithContext(ctx).Create(mappings[i:end]).Error; err != nil {
			return utils.NewSystemError(fmt.Errorf("批量创建用户消息关联失败: %w", err))
		}
	}
	return nil
}

func (repo *MessageRepositoryImpl) GetMessageByID(ctx context.Context, messageID int) (*model.Message, error) {
	var message model.Message
	if err := repo.db.WithContext(ctx).Where("id = ? AND is_deleted = ?", messageID, utils.DeletedFlagNo).First(&message).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewBusinessError(utils.ErrCodeResourceNotFound, "消息不存在或已被撤回")
		}
		return nil, utils.NewSystemError(fmt.Errorf("查询消息失败: %w", err))
	}
	return &message, nil
}

func (repo *MessageRepositoryImpl) SoftDeleteMessage(ctx context.Context, tx *gorm.DB, messageID int) error {
	result := tx.WithContext(ctx).Model(&model.Message{}).
		Where("id = ? AND is_deleted = ?", messageID, utils.DeletedFlagNo).
		Update("is_deleted", utils.DeletedFlagYes)
	if result.Error != nil {
		return utils.NewSystemError(fmt.Errorf("撤回消息失败: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return utils.NewBusinessError(utils.ErrCodeResourceNotFound, "消息不存在或已被撤回")
	}
	return nil
}

func (repo *MessageRepositoryImpl) DeleteUserMessageMappingsByMessageID(ctx context.Context, tx *gorm.DB, messageID int) error {
	if err := tx.WithContext(ctx).Where("message_id = ?", messageID).Delete(&model.UserMessageMapping{}).Error; err != nil {
		return utils.NewSystemError(fmt.Errorf("删除用户消息关联失败: %w", err))
	}
	return nil
}

func (repo *MessageRepositoryImpl) ListMessages(ctx context.Context, page, pageSize int, title, targetType string, targetID, userID int) ([]*dto.MessageListResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	var messages []*dto.MessageListResponse
	var total int64

	query := repo.db.WithContext(ctx).Table("messages m").
		Select("m.id, m.title, m.content, m.sender_id, m.target_type, m.target_id, m.send_time, m.create_time").
		Where("m.is_deleted = ?", utils.DeletedFlagNo)

	if title != "" {
		query = query.Where("m.title LIKE ?", "%"+title+"%")
	}
	if targetType != "" {
		query = query.Where("m.target_type = ?", targetType)
	}
	if targetID > 0 {
		query = query.Where("m.target_id = ?", targetID)
	}
	if userID > 0 {
		query = query.Joins("JOIN user_message_mappings umm ON m.id = umm.message_id").
			Where("umm.user_id = ?", userID)
	}

	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, utils.NewSystemError(fmt.Errorf("计算消息总数失败: %w", err))
	}

	if err := query.Order("m.send_time DESC").Offset(offset).Limit(pageSize).Find(&messages).Error; err != nil {
		return nil, 0, utils.NewSystemError(fmt.Errorf("查询消息列表失败: %w", err))
	}

	return messages, total, nil
}

func (repo *MessageRepositoryImpl) ListUserMessages(ctx context.Context, page, pageSize int, userID int) ([]*dto.UserMessageListResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	var messages []*dto.UserMessageListResponse
	var total int64

	query := repo.db.WithContext(ctx).Table("messages m").
		Select("m.id, m.title, m.content, m.sender_id, m.target_type, m.target_id, m.send_time, umm.is_read, umm.read_time").
		Joins("JOIN user_message_mappings umm ON m.id = umm.message_id").
		Where("umm.user_id = ? AND m.is_deleted = ?", userID, utils.DeletedFlagNo)

	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, utils.NewSystemError(fmt.Errorf("计算用户消息总数失败: %w", err))
	}

	if err := query.Order("m.send_time DESC").Offset(offset).Limit(pageSize).Find(&messages).Error; err != nil {
		return nil, 0, utils.NewSystemError(fmt.Errorf("查询用户消息列表失败: %w", err))
	}

	return messages, total, nil
}

func (repo *MessageRepositoryImpl) MarkAllAsRead(ctx context.Context, userID int) error {
	now := time.Now()
	result := repo.db.WithContext(ctx).
		Model(&model.UserMessageMapping{}).
		Where("user_id = ? AND is_read = 0", userID).
		Updates(map[string]interface{}{
			"is_read":   1,
			"read_time": now,
		})
	if result.Error != nil {
		return utils.NewSystemError(fmt.Errorf("标记消息已读失败: %w", result.Error))
	}
	return nil
}

func (repo *MessageRepositoryImpl) GetUnreadCount(ctx context.Context, userID int) (int64, error) {
	var count int64
	err := repo.db.WithContext(ctx).
		Table("user_message_mappings umm").
		Joins("JOIN messages m ON umm.message_id = m.id AND m.is_deleted = ?", utils.DeletedFlagNo).
		Where("umm.user_id = ? AND umm.is_read = 0", userID).
		Count(&count).Error
	if err != nil {
		return 0, utils.NewSystemError(fmt.Errorf("查询未读消息数失败: %w", err))
	}
	return count, nil
}
