package handler

import (
	"context"
	"message-service/internal/domain/message"
	"message-service/internal/domain/token"
	"message-service/pkg/api"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewHandler(t *testing.T) {
	messageRepo := new(mockMessageRepository)
	tokenRepo := new(mockTokenRepository)

	handler := NewHandler(messageRepo, tokenRepo)

	assert.NotNil(t, handler)
	assert.IsType(t, &Handler{}, handler)
}

func TestHandler_MessageMethods(t *testing.T) {
	ctx := context.Background()
	messageRepo := new(mockMessageRepository)
	tokenRepo := new(mockTokenRepository)
	handler := NewHandler(messageRepo, tokenRepo)

	t.Run("PostApiMessages", func(t *testing.T) {
		msg := &message.Message{
			UID:       "test-uid",
			ChannelID: "test-channel",
			Sender:    "test-sender",
			Content:   "test content",
			SentAt:    time.Now(),
		}

		messageRepo.On("Create", ctx, mock.AnythingOfType("*message.Message")).Return(nil)

		request := api.PostApiMessagesRequestObject{
			Body: &api.PostApiMessagesJSONRequestBody{
				ChannelId: msg.ChannelID,
				Content:   msg.Content,
				Sender:    msg.Sender,
				SentAt:    msg.SentAt,
				Uid:       msg.UID,
			},
		}

		response, err := handler.PostApiMessages(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		messageRepo.AssertExpectations(t)
	})

	t.Run("GetApiMessagesSearch", func(t *testing.T) {
		now := time.Now()
		channelID := "test-channel"
		sender := "test-sender"
		fromDate := now.Add(-24 * time.Hour)
		toDate := now

		messages := []message.Message{
			{
				UID:       "msg1",
				ChannelID: channelID,
				Sender:    sender,
				Content:   "test content 1",
				SentAt:    now,
			},
		}

		messageRepo.On("Search", ctx, mock.AnythingOfType("message.SearchCriteria")).Return(messages, nil)

		request := api.GetApiMessagesSearchRequestObject{
			Params: api.GetApiMessagesSearchParams{
				ChannelId: &channelID,
				Sender:    &sender,
				FromDate:  &fromDate,
				ToDate:    &toDate,
			},
		}

		response, err := handler.GetApiMessagesSearch(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		messageRepo.AssertExpectations(t)
	})

	t.Run("DeleteApiMessagesUid", func(t *testing.T) {
		uid := "test-uid"
		messageRepo.On("Delete", ctx, uid).Return(nil)

		request := api.DeleteApiMessagesUidRequestObject{
			Uid: uid,
		}

		response, err := handler.DeleteApiMessagesUid(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		messageRepo.AssertExpectations(t)
	})
}

func TestHandler_TokenMethods(t *testing.T) {
	ctx := context.Background()
	messageRepo := new(mockMessageRepository)
	tokenRepo := new(mockTokenRepository)
	handler := NewHandler(messageRepo, tokenRepo)

	t.Run("GetApiTokens", func(t *testing.T) {
		tokens := []token.Token{
			{
				Name:      "test-token",
				Token:     "token-string",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
		}

		tokenRepo.On("List", ctx).Return(tokens, nil)

		request := api.GetApiTokensRequestObject{}

		response, err := handler.GetApiTokens(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		tokenRepo.AssertExpectations(t)
	})

	t.Run("PostApiTokens", func(t *testing.T) {
		tokenRepo.On("Create", ctx, mock.AnythingOfType("*token.Token")).Return(nil)

		request := api.PostApiTokensRequestObject{
			Body: &api.PostApiTokensJSONRequestBody{
				Name: "test-token",
			},
		}

		response, err := handler.PostApiTokens(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		tokenRepo.AssertExpectations(t)
	})

	t.Run("DeleteApiTokensId", func(t *testing.T) {
		id := "test-id"
		tokenRepo.On("Delete", ctx, id).Return(nil)

		request := api.DeleteApiTokensIdRequestObject{
			Id: id,
		}

		response, err := handler.DeleteApiTokensId(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		tokenRepo.AssertExpectations(t)
	})
}
