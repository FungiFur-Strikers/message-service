package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"message-service/internal/domain/token"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mockTokenRepository struct {
	findByTokenFunc func(ctx context.Context, tokenStr string) (*token.Token, error)
}

func (m *mockTokenRepository) FindByToken(ctx context.Context, tokenStr string) (*token.Token, error) {
	return m.findByTokenFunc(ctx, tokenStr)
}

func (m *mockTokenRepository) Create(ctx context.Context, t *token.Token) error {
	return nil
}

func (m *mockTokenRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockTokenRepository) List(ctx context.Context) ([]token.Token, error) {
	return nil, nil
}

func (m *mockTokenRepository) FindByID(ctx context.Context, id string) (*token.Token, error) {
	return nil, nil
}

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		authHeader     string
		setupMock      func() token.Repository
		expectedStatus int
		expectedBody   string
		checkContext   func(*testing.T, *gin.Context)
	}{
		{
			name:   "トークン作成エンドポイントは認証をスキップ",
			method: "POST",
			path:   "/api/tokens",
			setupMock: func() token.Repository {
				return &mockTokenRepository{}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "認証ヘッダーが無い場合はエラー",
			method: "GET",
			path:   "/api/messages",
			setupMock: func() token.Repository {
				return &mockTokenRepository{}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Authentication required"}`,
		},
		{
			name:       "不正な認証フォーマット",
			method:     "GET",
			path:       "/api/messages",
			authHeader: "InvalidFormat token123",
			setupMock: func() token.Repository {
				return &mockTokenRepository{}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid authentication format"}`,
		},
		{
			name:       "トークンが存在しない場合",
			method:     "GET",
			path:       "/api/messages",
			authHeader: "Bearer token123",
			setupMock: func() token.Repository {
				return &mockTokenRepository{
					findByTokenFunc: func(ctx context.Context, tokenStr string) (*token.Token, error) {
						return nil, nil
					},
				}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid token"}`,
		},
		{
			name:       "リポジトリからのエラー発生",
			method:     "GET",
			path:       "/api/messages",
			authHeader: "Bearer token123",
			setupMock: func() token.Repository {
				return &mockTokenRepository{
					findByTokenFunc: func(ctx context.Context, tokenStr string) (*token.Token, error) {
						return nil, errors.New("repository error")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Error occurred during token validation"}`,
		},
		{
			name:       "有効なトークン",
			method:     "GET",
			path:       "/api/messages",
			authHeader: "Bearer validToken",
			setupMock: func() token.Repository {
				return &mockTokenRepository{
					findByTokenFunc: func(ctx context.Context, tokenStr string) (*token.Token, error) {
						now := time.Now()
						return &token.Token{
							ID:        primitive.NewObjectID(),
							Token:     "validToken",
							Name:      "Test Token",
							ExpiresAt: now.Add(24 * time.Hour),
							CreatedAt: now,
							UpdatedAt: now,
						}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
			checkContext: func(t *testing.T, c *gin.Context) {
				// コンテキストにトークン情報が正しく設定されているか確認
				tokenVal, exists := c.Get("token")
				assert.True(t, exists, "token should be set in context")
				assert.NotNil(t, tokenVal, "token value should not be nil")

				tokenID, exists := c.Get("token_id")
				assert.True(t, exists, "token_id should be set in context")
				assert.NotEmpty(t, tokenID, "token_id should not be empty")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			authMiddleware := NewAuthMiddleware(tt.setupMock())
			r.Use(authMiddleware.RequireAuth())

			r.Handle(tt.method, tt.path, func(c *gin.Context) {
				if tt.checkContext != nil {
					tt.checkContext(t, c)
				}
				c.Status(http.StatusOK)
			})

			req, _ := http.NewRequest(tt.method, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}
