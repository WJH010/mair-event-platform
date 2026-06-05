package controller

import (
	"event-platform/internal/message/dto"
	"event-platform/internal/message/service"
	"event-platform/internal/utils"

	"github.com/gin-gonic/gin"
)

type MessageController struct {
	messageService service.MessageService
}

func NewMessageController(messageService service.MessageService) *MessageController {
	return &MessageController{messageService: messageService}
}

func (ctr *MessageController) CreateMessage(ctx *gin.Context) {
	var req dto.CreateMessageRequest
	if !utils.BindJSON(ctx, &req) {
		return
	}

	userID, err := utils.GetUserID(ctx)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	messageID, err := ctr.messageService.CreateMessage(ctx, &req, userID)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", gin.H{"message_id": messageID})
}

func (ctr *MessageController) RevokeMessage(ctx *gin.Context) {
	var req dto.MessageIDRequest
	if !utils.BindUrl(ctx, &req) {
		return
	}

	if err := ctr.messageService.RevokeMessage(ctx, req.ID); err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", nil)
}

func (ctr *MessageController) ListMessages(ctx *gin.Context) {
	var req dto.MessageListRequest
	if !utils.BindQuery(ctx, &req) {
		return
	}

	messages, total, err := ctr.messageService.ListMessages(ctx, &req)
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
	utils.SuccessPage(ctx, total, page, pageSize, messages)
}

func (ctr *MessageController) ListUserMessages(ctx *gin.Context) {
	var req dto.UserMessageListRequest
	if !utils.BindQuery(ctx, &req) {
		return
	}

	userID, err := utils.GetUserID(ctx)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	messages, total, err := ctr.messageService.ListUserMessages(ctx, &req, userID)
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
	utils.SuccessPage(ctx, total, page, pageSize, messages)
}

func (ctr *MessageController) GetUnreadCount(ctx *gin.Context) {
	userID, err := utils.GetUserID(ctx)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	count, err := ctr.messageService.GetUnreadCount(ctx, userID)
	if err != nil {
		utils.HandlerFunc(ctx, err)
		return
	}

	utils.Success(ctx, "success", dto.UnreadCountResponse{UnreadCount: count})
}
