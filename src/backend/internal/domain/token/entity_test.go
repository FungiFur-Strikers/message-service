package token

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestToken(t *testing.T) {
	t.Run("新規トークンの作成", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(24 * time.Hour)
		token := &Token{
			ID:        primitive.NewObjectID(),
			Token:     "token-123",
			Name:      "API Access Token",
			ExpiresAt: expiresAt,
			CreatedAt: now,
			UpdatedAt: now,
		}

		assert.NotEmpty(t, token.ID)
		assert.Equal(t, "token-123", token.Token)
		assert.Equal(t, "API Access Token", token.Name)
		assert.Equal(t, expiresAt, token.ExpiresAt)
		assert.Equal(t, now, token.CreatedAt)
		assert.Equal(t, now, token.UpdatedAt)
		assert.Nil(t, token.DeletedAt)
	})

	t.Run("有効期限切れトークンの検証", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(-1 * time.Hour) // 1時間前に有効期限切れ
		token := &Token{
			ID:        primitive.NewObjectID(),
			Token:     "expired-token",
			Name:      "Expired Token",
			ExpiresAt: expiresAt,
			CreatedAt: now.Add(-24 * time.Hour),
			UpdatedAt: now.Add(-24 * time.Hour),
		}

		assert.NotEmpty(t, token.ID)
		assert.True(t, token.ExpiresAt.Before(now), "トークンは有効期限切れであるべき")
	})

	t.Run("削除済みトークンの作成", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(24 * time.Hour)
		deletedAt := now.Add(time.Hour)
		token := &Token{
			ID:        primitive.NewObjectID(),
			Token:     "deleted-token",
			Name:      "Deleted Token",
			ExpiresAt: expiresAt,
			CreatedAt: now,
			UpdatedAt: now,
			DeletedAt: &deletedAt,
		}

		assert.NotEmpty(t, token.ID)
		assert.Equal(t, "deleted-token", token.Token)
		assert.NotNil(t, token.DeletedAt)
		assert.Equal(t, deletedAt, *token.DeletedAt)
	})

	t.Run("トークン名の検証", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(24 * time.Hour)

		// 長い名前のトークン
		longName := "This is a very long token name that should still work fine in the system"
		token := &Token{
			ID:        primitive.NewObjectID(),
			Token:     "token-with-long-name",
			Name:      longName,
			ExpiresAt: expiresAt,
			CreatedAt: now,
			UpdatedAt: now,
		}

		assert.NotEmpty(t, token.ID)
		assert.Equal(t, longName, token.Name)

		// 空の名前のトークン
		emptyNameToken := &Token{
			ID:        primitive.NewObjectID(),
			Token:     "token-with-empty-name",
			Name:      "",
			ExpiresAt: expiresAt,
			CreatedAt: now,
			UpdatedAt: now,
		}

		assert.NotEmpty(t, emptyNameToken.ID)
		assert.Empty(t, emptyNameToken.Name)
	})
}
