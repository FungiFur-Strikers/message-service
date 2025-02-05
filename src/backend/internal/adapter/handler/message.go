package handler

import (
	"context"
	"message-service/internal/domain/message"
	"message-service/pkg/api"
)

type MessageHandler struct {
	repo message.Repository
}

func NewMessageHandler(repo message.Repository) *MessageHandler {
	return &MessageHandler{repo: repo}
}

func (h *MessageHandler) PostApiMessages(ctx context.Context, req api.PostApiMessagesRequestObject) (api.PostApiMessagesResponseObject, error) {
	msg := &message.Message{
		UID:       req.Body.Uid,
		SentAt:    req.Body.SentAt,
		Sender:    req.Body.Sender,
		ChannelID: req.Body.ChannelId,
		Content:   req.Body.Content,
	}

	if err := h.repo.Create(ctx, msg); err != nil {
		return api.PostApiMessages400Response{}, err
	}

	return api.PostApiMessages201JSONResponse{
		ChannelId: &msg.ChannelID,
		Content:   &msg.Content,
		CreatedAt: &msg.CreatedAt,
		Sender:    &msg.Sender,
		SentAt:    &msg.SentAt,
		Uid:       &msg.UID,
		UpdatedAt: &msg.UpdatedAt,
	}, nil
}

func (h *MessageHandler) GetApiMessagesSearch(ctx context.Context, req api.GetApiMessagesSearchRequestObject) (api.GetApiMessagesSearchResponseObject, error) {
	criteria := message.SearchCriteria{
		ChannelID: req.Params.ChannelId,
		Sender:    req.Params.Sender,
		FromDate:  req.Params.FromDate,
		ToDate:    req.Params.ToDate,
	}

	messages, err := h.repo.Search(ctx, criteria)
	if err != nil {
		return nil, err
	}

	response := make(api.GetApiMessagesSearch200JSONResponse, len(messages))
	for i, msg := range messages {
		response[i] = api.Message{
			ChannelId: &msg.ChannelID,
			Content:   &msg.Content,
			CreatedAt: &msg.CreatedAt,
			Sender:    &msg.Sender,
			SentAt:    &msg.SentAt,
			Uid:       &msg.UID,
			UpdatedAt: &msg.UpdatedAt,
		}
	}

	return response, nil
}

func (h *MessageHandler) DeleteApiMessagesUid(ctx context.Context, req api.DeleteApiMessagesUidRequestObject) (api.DeleteApiMessagesUidResponseObject, error) {
	if err := h.repo.Delete(ctx, req.Uid); err != nil {
		return api.DeleteApiMessagesUid404Response{}, err
	}
	return api.DeleteApiMessagesUid204Response{}, nil
}
