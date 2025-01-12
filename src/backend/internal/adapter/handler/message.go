package handler

import (
	"context"
	"message-service/internal/domain/message"
	"message-service/pkg/api"
)

type MessageHandler struct {
	repo message.Repository
}

func NewMessageHandler(repo message.Repository) api.StrictServerInterface {
	return &MessageHandler{repo: repo}
}

func (h *MessageHandler) PostApiMessage(ctx context.Context, req api.PostApiMessageRequestObject) (api.PostApiMessageResponseObject, error) {
	msg := &message.Message{
		UID:       req.Body.Uid,
		SentAt:    req.Body.SentAt,
		Sender:    req.Body.Sender,
		ChannelID: req.Body.ChannelId,
		Content:   req.Body.Content,
	}

	if err := h.repo.Create(ctx, msg); err != nil {
		return api.PostApiMessage400Response{}, err
	}

	return api.PostApiMessage201JSONResponse{
		ChannelId: &msg.ChannelID,
		Content:   &msg.Content,
		CreatedAt: &msg.CreatedAt,
		Sender:    &msg.Sender,
		SentAt:    &msg.SentAt,
		Uid:       &msg.UID,
		UpdatedAt: &msg.UpdatedAt,
	}, nil
}

func (h *MessageHandler) DeleteApiMessageUid(ctx context.Context, req api.DeleteApiMessageUidRequestObject) (api.DeleteApiMessageUidResponseObject, error) {
	if err := h.repo.Delete(ctx, req.Uid); err != nil {
		return api.DeleteApiMessageUid404Response{}, err
	}
	return api.DeleteApiMessageUid204Response{}, nil
}

func (h *MessageHandler) GetApiMessageSearch(ctx context.Context, req api.GetApiMessageSearchRequestObject) (api.GetApiMessageSearchResponseObject, error) {
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

	response := make(api.GetApiMessageSearch200JSONResponse, len(messages))
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
