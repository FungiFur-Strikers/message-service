package token

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// テストヘルパー関数
func createTestToken(t *testing.T) *Token {
	t.Helper()
	now := time.Now()
	return &Token{
		ID:        primitive.NewObjectID(),
		Token:     "test-token-123",
		Name:      "Test Token",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestToken_Creation(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Token
		validate    func(*testing.T, *Token)
		shouldError bool
	}{
		{
			name: "正常なトークンの作成",
			setup: func() *Token {
				return createTestToken(t)
			},
			validate: func(t *testing.T, token *Token) {
				assert.NotEmpty(t, token.ID)
				assert.Equal(t, "test-token-123", token.Token)
				assert.Equal(t, "Test Token", token.Name)
				assert.True(t, token.ExpiresAt.After(token.CreatedAt))
				assert.Equal(t, token.CreatedAt, token.UpdatedAt)
				assert.Nil(t, token.DeletedAt)
			},
		},
		{
			name: "削除済みトークンの作成",
			setup: func() *Token {
				token := createTestToken(t)
				deletedAt := token.CreatedAt.Add(time.Hour)
				token.DeletedAt = &deletedAt
				return token
			},
			validate: func(t *testing.T, token *Token) {
				assert.NotNil(t, token.DeletedAt)
				assert.True(t, token.DeletedAt.After(token.CreatedAt))
				assert.True(t, token.DeletedAt.After(token.UpdatedAt))
			},
		},
		{
			name: "最大長の名前を持つトークン",
			setup: func() *Token {
				token := createTestToken(t)
				token.Name = "This is a very long token name that tests the maximum length limit of the name field in our system"
				return token
			},
			validate: func(t *testing.T, token *Token) {
				assert.NotEmpty(t, token.Name)
				assert.LessOrEqual(t, len(token.Name), 100) // 仮の最大長制限
			},
		},
		{
			name: "空の名前を持つトークン",
			setup: func() *Token {
				token := createTestToken(t)
				token.Name = ""
				return token
			},
			validate: func(t *testing.T, token *Token) {
				assert.Empty(t, token.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setup()
			tt.validate(t, token)
		})
	}
}

func TestToken_Validation(t *testing.T) {
	t.Run("トークンの有効期限テスト", func(t *testing.T) {
		tests := []struct {
			name        string
			setupDates  func() (time.Time, time.Time)
			shouldValid bool
		}{
			{
				name: "有効な期限（未来）",
				setupDates: func() (time.Time, time.Time) {
					now := time.Now()
					return now, now.Add(24 * time.Hour)
				},
				shouldValid: true,
			},
			{
				name: "無効な期限（過去）",
				setupDates: func() (time.Time, time.Time) {
					now := time.Now()
					return now, now.Add(-24 * time.Hour)
				},
				shouldValid: false,
			},
			{
				name: "現在時刻と同じ期限",
				setupDates: func() (time.Time, time.Time) {
					now := time.Now()
					return now, now
				},
				shouldValid: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				createdAt, expiresAt := tt.setupDates()
				token := &Token{
					ID:        primitive.NewObjectID(),
					Token:     "test-token",
					Name:      "Test Token",
					ExpiresAt: expiresAt,
					CreatedAt: createdAt,
					UpdatedAt: createdAt,
				}

				if tt.shouldValid {
					assert.True(t, token.ExpiresAt.After(token.CreatedAt))
				} else {
					assert.False(t, token.ExpiresAt.After(token.CreatedAt))
				}
			})
		}
	})

	t.Run("トークン文字列のバリデーション", func(t *testing.T) {
		tests := []struct {
			name        string
			tokenString string
			shouldValid bool
		}{
			{
				name:        "有効なトークン文字列",
				tokenString: "valid-token-123",
				shouldValid: true,
			},
			{
				name:        "空のトークン文字列",
				tokenString: "",
				shouldValid: false,
			},
			{
				name:        "特殊文字を含むトークン",
				tokenString: "token@#$%^",
				shouldValid: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				token := createTestToken(t)
				token.Token = tt.tokenString

				if tt.shouldValid {
					assert.NotEmpty(t, token.Token)
					assert.Regexp(t, "^[a-zA-Z0-9-]+$", token.Token)
				} else {
					if token.Token != "" {
						assert.NotRegexp(t, "^[a-zA-Z0-9-]+$", token.Token)
					} else {
						assert.Empty(t, token.Token)
					}
				}
			})
		}
	})
}
