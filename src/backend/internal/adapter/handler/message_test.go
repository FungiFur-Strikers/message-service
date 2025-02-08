package handler

import (
	"context"
	"errors"
	"message-service/internal/domain/message"
	"message-service/pkg/api"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// テストヘルパー関数
func createTestMessage() *message.Message {
	now := time.Now()
	return &message.Message{
		UID:       "test-uid",
		SentAt:    now,
		Sender:    "test-sender",
		ChannelID: "test-channel",
		Content:   "test message",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func createTestPostRequest(body *api.PostApiMessagesJSONRequestBody) api.PostApiMessagesRequestObject {
	if body == nil {
		now := time.Now()
		body = &api.PostApiMessagesJSONRequestBody{
			Uid:       "test-uid",
			SentAt:    now,
			Sender:    "test-sender",
			ChannelId: "test-channel",
			Content:   "test message",
		}
	}
	return api.PostApiMessagesRequestObject{Body: body}
}

func createTestSearchRequest(params api.GetApiMessagesSearchParams) api.GetApiMessagesSearchRequestObject {
	return api.GetApiMessagesSearchRequestObject{Params: params}
}

func TestMessageHandler_PostApiMessages(t *testing.T) {
	tests := []struct {
		name          string
		request       api.PostApiMessagesRequestObject
		mockSetup     func(*mockMessageRepository)
		expectedError bool
		expectedCode  int
		errorMessage  string
	}{
		{
			name:    "正常系：メッセージ作成成功",
			request: createTestPostRequest(nil),
			mockSetup: func(m *mockMessageRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*message.Message")).Return(nil)
			},
			expectedError: false,
			expectedCode:  201,
		},
		{
			name: "異常系：必須フィールド（Content）が空",
			request: createTestPostRequest(&api.PostApiMessagesJSONRequestBody{
				Uid:       "test-uid",
				SentAt:    time.Now(),
				Sender:    "test-sender",
				ChannelId: "test-channel",
				Content:   "",
			}),
			mockSetup: func(m *mockMessageRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*message.Message")).
					Return(errors.New("content is required"))
			},
			expectedError: true,
			expectedCode:  400,
			errorMessage:  "content is required",
		},
		{
			name: "異常系：重複するUID",
			request: createTestPostRequest(&api.PostApiMessagesJSONRequestBody{
				Uid:       "duplicate-uid",
				SentAt:    time.Now(),
				Sender:    "test-sender",
				ChannelId: "test-channel",
				Content:   "test message",
			}),
			mockSetup: func(m *mockMessageRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*message.Message")).
					Return(errors.New("duplicate uid"))
			},
			expectedError: true,
			expectedCode:  409,
			errorMessage:  "duplicate uid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockMessageRepository)
			tt.mockSetup(mockRepo)
			handler := NewMessageHandler(mockRepo)

			resp, err := handler.PostApiMessages(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
				response, ok := resp.(api.PostApiMessages201JSONResponse)
				assert.True(t, ok)
				assert.Equal(t, tt.request.Body.Uid, *response.Uid)
				assert.Equal(t, tt.request.Body.ChannelId, *response.ChannelId)
				assert.Equal(t, tt.request.Body.Content, *response.Content)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMessageHandler_GetApiMessagesSearch(t *testing.T) {
	testTime := time.Now()
	futureTime := testTime.Add(24 * time.Hour)
	tests := []struct {
		name          string
		request       api.GetApiMessagesSearchRequestObject
		mockSetup     func(*mockMessageRepository)
		expectedError bool
		expectedLen   int
		errorMessage  string
	}{
		{
			name: "正常系：検索結果あり",
			request: createTestSearchRequest(api.GetApiMessagesSearchParams{
				ChannelId: stringPtr("test-channel"),
				FromDate:  &testTime,
				ToDate:    &futureTime,
			}),
			mockSetup: func(m *mockMessageRepository) {
				m.On("Search", mock.Anything, mock.AnythingOfType("message.SearchCriteria")).
					Return([]message.Message{*createTestMessage()}, nil)
			},
			expectedError: false,
			expectedLen:   1,
		},
		{
			name: "正常系：日付範囲指定なしの検索",
			request: createTestSearchRequest(api.GetApiMessagesSearchParams{
				ChannelId: stringPtr("test-channel"),
			}),
			mockSetup: func(m *mockMessageRepository) {
				m.On("Search", mock.Anything, mock.AnythingOfType("message.SearchCriteria")).
					Return([]message.Message{*createTestMessage()}, nil)
			},
			expectedError: false,
			expectedLen:   1,
		},
		{
			name: "異常系：無効な日付範囲",
			request: createTestSearchRequest(api.GetApiMessagesSearchParams{
				ChannelId: stringPtr("test-channel"),
				FromDate:  &futureTime,
				ToDate:    &testTime,
			}),
			mockSetup: func(m *mockMessageRepository) {
				m.On("Search", mock.Anything, mock.AnythingOfType("message.SearchCriteria")).
					Return(nil, errors.New("invalid date range"))
			},
			expectedError: true,
			expectedLen:   0,
			errorMessage:  "invalid date range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockMessageRepository)
			tt.mockSetup(mockRepo)
			handler := NewMessageHandler(mockRepo)

			resp, err := handler.GetApiMessagesSearch(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
				response, ok := resp.(api.GetApiMessagesSearch200JSONResponse)
				assert.True(t, ok)
				assert.Len(t, response, tt.expectedLen)
				if tt.expectedLen > 0 {
					assert.NotEmpty(t, response[0].Uid)
					assert.NotEmpty(t, response[0].ChannelId)
					assert.NotEmpty(t, response[0].Content)
				}
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMessageHandler_DeleteApiMessagesUid(t *testing.T) {
	tests := []struct {
		name          string
		uid           string
		mockSetup     func(*mockMessageRepository)
		expectedError bool
		expectedCode  int
		errorMessage  string
	}{
		{
			name: "正常系：メッセージ削除成功",
			uid:  "test-uid",
			mockSetup: func(m *mockMessageRepository) {
				m.On("Delete", mock.Anything, "test-uid").Return(nil)
			},
			expectedError: false,
			expectedCode:  204,
		},
		{
			name: "異常系：存在しないメッセージ",
			uid:  "non-existent-uid",
			mockSetup: func(m *mockMessageRepository) {
				m.On("Delete", mock.Anything, "non-existent-uid").
					Return(errors.New("message not found"))
			},
			expectedError: true,
			expectedCode:  404,
			errorMessage:  "message not found",
		},
		{
			name: "異常系：無効なUID形式",
			uid:  "invalid@uid",
			mockSetup: func(m *mockMessageRepository) {
				m.On("Delete", mock.Anything, "invalid@uid").
					Return(errors.New("invalid uid format"))
			},
			expectedError: true,
			expectedCode:  400,
			errorMessage:  "invalid uid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockMessageRepository)
			tt.mockSetup(mockRepo)
			handler := NewMessageHandler(mockRepo)

			resp, err := handler.DeleteApiMessagesUid(context.Background(), api.DeleteApiMessagesUidRequestObject{
				Uid: tt.uid,
			})

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
				_, ok := resp.(api.DeleteApiMessagesUid204Response)
				assert.True(t, ok)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
