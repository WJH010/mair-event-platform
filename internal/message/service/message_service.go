package service

import (
	"context"
	"event-platform/internal/message/dto"
	"event-platform/internal/message/model"
	"event-platform/internal/message/repository"
	userrepo "event-platform/internal/user/repository"
	"event-platform/internal/utils"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MessageService interface {
	CreateMessage(ctx context.Context, req *dto.CreateMessageRequest, userID int) (int, error)
	RevokeMessage(ctx context.Context, messageID int) error
	ListMessages(ctx context.Context, req *dto.MessageListRequest) ([]*dto.MessageListResponse, int64, error)
	ListUserMessages(ctx context.Context, req *dto.UserMessageListRequest, userID int) ([]*dto.UserMessageListResponse, int64, error)
	GetUnreadCount(ctx context.Context, userID int) (int64, error)
}

type MessageServiceImpl struct {
	messageRepo repository.MessageRepository
	userRepo    userrepo.UserRepository
}

func NewMessageService(messageRepo repository.MessageRepository, userRepo userrepo.UserRepository) MessageService {
	return &MessageServiceImpl{
		messageRepo: messageRepo,
		userRepo:    userRepo,
	}
}

func (svc *MessageServiceImpl) CreateMessage(ctx context.Context, req *dto.CreateMessageRequest, userID int) (int, error) {
	if req.TargetType == "FIELD" && req.TargetID <= 0 {
		return 0, utils.NewBusinessError(utils.ErrCodeParamInvalid, "target_type为FIELD时，target_id必填")
	}

	message := &model.Message{
		Title:      req.Title,
		Content:    req.Content,
		SenderID:   userID,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		SendTime:   time.Now(),
		CreateUser: userID,
		UpdateUser: userID,
	}

	var userIDs []int
	var err error
	if req.TargetType == "ALL" {
		userIDs, err = svc.userRepo.GetAllEnabledUserIDs(ctx)
	} else {
		userIDs, err = svc.userRepo.GetUserIDsByFieldID(ctx, req.TargetID)
	}
	if err != nil {
		return 0, err
	}

	err = svc.messageRepo.ExecTransaction(ctx, func(tx *gorm.DB) error {
		if err := svc.messageRepo.CreateMessage(ctx, tx, message); err != nil {
			return err
		}
		mappings := make([]*model.UserMessageMapping, 0, len(userIDs))
		for _, uid := range userIDs {
			mappings = append(mappings, &model.UserMessageMapping{
				UserID:      uid,
				MessageID:   message.ID,
				CreateUser:  userID,
				UpdateUser:  userID,
			})
		}
		if err := svc.messageRepo.BatchCreateUserMessageMappings(ctx, tx, mappings); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return message.ID, nil
}

func (svc *MessageServiceImpl) RevokeMessage(ctx context.Context, messageID int) error {
	_, err := svc.messageRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return err
	}

	return svc.messageRepo.ExecTransaction(ctx, func(tx *gorm.DB) error {
		if err := svc.messageRepo.SoftDeleteMessage(ctx, tx, messageID); err != nil {
			return err
		}
		if err := svc.messageRepo.DeleteUserMessageMappingsByMessageID(ctx, tx, messageID); err != nil {
			return err
		}
		return nil
	})
}

func (svc *MessageServiceImpl) ListMessages(ctx context.Context, req *dto.MessageListRequest) ([]*dto.MessageListResponse, int64, error) {
	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}
	return svc.messageRepo.ListMessages(ctx, page, pageSize, req.Title, req.TargetType, req.TargetID, req.UserID)
}

func (svc *MessageServiceImpl) ListUserMessages(ctx context.Context, req *dto.UserMessageListRequest, userID int) ([]*dto.UserMessageListResponse, int64, error) {
	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}

	messages, total, err := svc.messageRepo.ListUserMessages(ctx, page, pageSize, userID)
	if err != nil {
		return nil, 0, err
	}

	if err := svc.messageRepo.MarkAllAsRead(ctx, userID); err != nil {
		logrus.Warnf("标记用户消息已读失败[userID=%d]: %v", userID, err)
	}

	return messages, total, nil
}

func (svc *MessageServiceImpl) GetUnreadCount(ctx context.Context, userID int) (int64, error) {
	return svc.messageRepo.GetUnreadCount(ctx, userID)
}
