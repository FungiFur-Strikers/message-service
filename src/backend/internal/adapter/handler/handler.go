package handler

import (
	"context"
	"message-service/internal/domain/message"
	"message-service/internal/domain/token"
	"message-service/pkg/api"
)

type Handler struct {
	messageHandler *MessageHandler
	tokenHandler   *TokenHandler
}

func NewHandler(messageRepo message.Repository, tokenRepo token.Repository) api.StrictServerInterface {
	return &Handler{
		messageHandler: NewMessageHandler(messageRepo),
		tokenHandler:   NewTokenHandler(tokenRepo),
	}
}

// メッセージ関連のメソッド
func (h *Handler) PostApiMessages(ctx context.Context, request api.PostApiMessagesRequestObject) (api.PostApiMessagesResponseObject, error) {
	return h.messageHandler.PostApiMessages(ctx, request)
}

func (h *Handler) GetApiMessagesSearch(ctx context.Context, request api.GetApiMessagesSearchRequestObject) (api.GetApiMessagesSearchResponseObject, error) {
	return h.messageHandler.GetApiMessagesSearch(ctx, request)
}

func (h *Handler) DeleteApiMessagesUid(ctx context.Context, request api.DeleteApiMessagesUidRequestObject) (api.DeleteApiMessagesUidResponseObject, error) {
	return h.messageHandler.DeleteApiMessagesUid(ctx, request)
}

// トークン関連のメソッド
func (h *Handler) GetApiTokens(ctx context.Context, request api.GetApiTokensRequestObject) (api.GetApiTokensResponseObject, error) {
	return h.tokenHandler.GetApiTokens(ctx, request)
}

func (h *Handler) PostApiTokens(ctx context.Context, request api.PostApiTokensRequestObject) (api.PostApiTokensResponseObject, error) {
	return h.tokenHandler.PostApiTokens(ctx, request)
}

func (h *Handler) DeleteApiTokensId(ctx context.Context, request api.DeleteApiTokensIdRequestObject) (api.DeleteApiTokensIdResponseObject, error) {
	return h.tokenHandler.DeleteApiTokensId(ctx, request)
}
