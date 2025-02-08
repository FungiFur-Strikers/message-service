package handler

import (
	"context"
	"errors"
	"message-service/internal/domain/token"
	"message-service/pkg/api"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// テストヘルパー関数
func createTestToken() *token.Token {
	now := time.Now()
	return &token.Token{
		ID:        primitive.NewObjectID(),
		Token:     "test-token",
		Name:      "test-token-name",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func createTestTokenRequest(body *api.PostApiTokensJSONRequestBody) api.PostApiTokensRequestObject {
	if body == nil {
		body = &api.PostApiTokensJSONRequestBody{
			Name: "test-token",
		}
	}
	return api.PostApiTokensRequestObject{Body: body}
}

func TestTokenHandler_PostApiTokens(t *testing.T) {
	tests := []struct {
		name          string
		request       api.PostApiTokensRequestObject
		mockSetup     func(*mockTokenRepository)
		expectedError bool
		expectedCode  int
		errorMessage  string
	}{
		{
			name:    "正常系：デフォルト有効期限でトークン生成",
			request: createTestTokenRequest(nil),
			mockSetup: func(m *mockTokenRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*token.Token")).Return(nil)
			},
			expectedError: false,
			expectedCode:  201,
		},
		{
			name: "正常系：カスタム有効期限でトークン生成",
			request: createTestTokenRequest(&api.PostApiTokensJSONRequestBody{
				Name:      "test-token",
				ExpiresIn: intPtr(3600),
			}),
			mockSetup: func(m *mockTokenRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*token.Token")).Return(nil)
			},
			expectedError: false,
			expectedCode:  201,
		},
		{
			name: "異常系：空のトークン名",
			request: createTestTokenRequest(&api.PostApiTokensJSONRequestBody{
				Name: "",
			}),
			mockSetup: func(m *mockTokenRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*token.Token")).
					Return(errors.New("token name is required"))
			},
			expectedError: true,
			expectedCode:  400,
			errorMessage:  "token name is required",
		},
		{
			name: "異常系：無効な有効期限",
			request: createTestTokenRequest(&api.PostApiTokensJSONRequestBody{
				Name:      "test-token",
				ExpiresIn: intPtr(-1),
			}),
			mockSetup: func(m *mockTokenRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*token.Token")).
					Return(errors.New("invalid expiration time"))
			},
			expectedError: true,
			expectedCode:  400,
			errorMessage:  "invalid expiration time",
		},
		{
			name: "異常系：トークン名が最大長を超過",
			request: createTestTokenRequest(&api.PostApiTokensJSONRequestBody{
				Name: "this-is-a-very-long-token-name-that-exceeds-the-maximum-allowed-length-limit",
			}),
			mockSetup: func(m *mockTokenRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*token.Token")).
					Return(errors.New("token name too long"))
			},
			expectedError: true,
			expectedCode:  400,
			errorMessage:  "token name too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockTokenRepository)
			tt.mockSetup(mockRepo)
			handler := NewTokenHandler(mockRepo)

			resp, err := handler.PostApiTokens(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
				_, ok := resp.(api.PostApiTokens400Response)
				assert.True(t, ok)
			} else {
				assert.NoError(t, err)
				response, ok := resp.(api.PostApiTokens201JSONResponse)
				assert.True(t, ok)
				assert.NotEmpty(t, response.Token)
				assert.Equal(t, tt.request.Body.Name, *response.Name)
				assert.NotNil(t, response.ExpiresAt)
				if tt.request.Body.ExpiresIn != nil {
					expectedExpiry := time.Now().Add(time.Duration(*tt.request.Body.ExpiresIn) * time.Second)
					assert.WithinDuration(t, expectedExpiry, *response.ExpiresAt, time.Second)
				}
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTokenHandler_GetApiTokens(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*mockTokenRepository)
		expectedError bool
		expectedLen   int
		errorMessage  string
	}{
		{
			name: "正常系：有効なトークン一覧取得",
			mockSetup: func(m *mockTokenRepository) {
				tokens := []token.Token{*createTestToken(), *createTestToken()}
				m.On("List", mock.Anything).Return(tokens, nil)
			},
			expectedError: false,
			expectedLen:   2,
		},
		{
			name: "正常系：トークンが存在しない",
			mockSetup: func(m *mockTokenRepository) {
				m.On("List", mock.Anything).Return([]token.Token{}, nil)
			},
			expectedError: false,
			expectedLen:   0,
		},
		{
			name: "異常系：データベースエラー",
			mockSetup: func(m *mockTokenRepository) {
				m.On("List", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedError: true,
			expectedLen:   0,
			errorMessage:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockTokenRepository)
			tt.mockSetup(mockRepo)
			handler := NewTokenHandler(mockRepo)

			resp, err := handler.GetApiTokens(context.Background(), api.GetApiTokensRequestObject{})

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				response, ok := resp.(api.GetApiTokens200JSONResponse)
				assert.True(t, ok)
				assert.Len(t, response, tt.expectedLen)
				if tt.expectedLen > 0 {
					for _, token := range response {
						assert.NotEmpty(t, token.Token)
						assert.NotEmpty(t, token.Name)
						assert.NotNil(t, token.ExpiresAt)
						assert.True(t, token.ExpiresAt.After(time.Now()))
					}
				}
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTokenHandler_DeleteApiTokensId(t *testing.T) {
	validObjectID := primitive.NewObjectID().Hex()

	tests := []struct {
		name          string
		id            string
		mockSetup     func(*mockTokenRepository)
		expectedError bool
		expectedCode  int
		errorMessage  string
	}{
		{
			name: "正常系：トークン削除成功",
			id:   validObjectID,
			mockSetup: func(m *mockTokenRepository) {
				m.On("Delete", mock.Anything, validObjectID).Return(nil)
			},
			expectedError: false,
			expectedCode:  204,
		},
		{
			name: "異常系：存在しないトークン",
			id:   validObjectID,
			mockSetup: func(m *mockTokenRepository) {
				m.On("Delete", mock.Anything, validObjectID).
					Return(errors.New("token not found"))
			},
			expectedError: true,
			expectedCode:  404,
			errorMessage:  "token not found",
		},
		{
			name: "異常系：不正なObjectID形式",
			id:   "invalid-id",
			mockSetup: func(m *mockTokenRepository) {
				m.On("Delete", mock.Anything, "invalid-id").
					Return(errors.New("invalid object id format"))
			},
			expectedError: true,
			expectedCode:  404,
			errorMessage:  "invalid object id format",
		},
		{
			name: "異常系：データベースエラー",
			id:   validObjectID,
			mockSetup: func(m *mockTokenRepository) {
				m.On("Delete", mock.Anything, validObjectID).
					Return(errors.New("database error"))
			},
			expectedError: true,
			expectedCode:  404,
			errorMessage:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockTokenRepository)
			tt.mockSetup(mockRepo)
			handler := NewTokenHandler(mockRepo)

			resp, err := handler.DeleteApiTokensId(context.Background(), api.DeleteApiTokensIdRequestObject{
				Id: tt.id,
			})

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
				_, ok := resp.(api.DeleteApiTokensId404Response)
				assert.True(t, ok)
			} else {
				assert.NoError(t, err)
				_, ok := resp.(api.DeleteApiTokensId204Response)
				assert.True(t, ok)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func intPtr(i int) *int {
	return &i
}
