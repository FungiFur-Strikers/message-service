package message

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMessage(t *testing.T) {
	t.Run("新規メッセージの作成", func(t *testing.T) {
		now := time.Now()
		msg := &Message{
			ID:        primitive.NewObjectID(),
			UID:       "msg-123",
			SentAt:    now,
			Sender:    "user1",
			ChannelID: "channel-1",
			Content:   "Hello, World!",
			CreatedAt: now,
			UpdatedAt: now,
		}

		assert.NotEmpty(t, msg.ID)
		assert.Equal(t, "msg-123", msg.UID)
		assert.Equal(t, now, msg.SentAt)
		assert.Equal(t, "user1", msg.Sender)
		assert.Equal(t, "channel-1", msg.ChannelID)
		assert.Equal(t, "Hello, World!", msg.Content)
		assert.Equal(t, now, msg.CreatedAt)
		assert.Equal(t, now, msg.UpdatedAt)
		assert.Nil(t, msg.DeletedAt)
	})

	t.Run("削除済みメッセージの作成", func(t *testing.T) {
		now := time.Now()
		deletedAt := now.Add(time.Hour)
		msg := &Message{
			ID:        primitive.NewObjectID(),
			UID:       "msg-456",
			SentAt:    now,
			Sender:    "user2",
			ChannelID: "channel-2",
			Content:   "Deleted message",
			CreatedAt: now,
			UpdatedAt: now,
			DeletedAt: &deletedAt,
		}

		assert.NotEmpty(t, msg.ID)
		assert.Equal(t, "msg-456", msg.UID)
		assert.NotNil(t, msg.DeletedAt)
		assert.Equal(t, deletedAt, *msg.DeletedAt)
	})
}

func TestSearchCriteria(t *testing.T) {
	t.Run("完全な検索条件の作成", func(t *testing.T) {
		channelID := "channel-1"
		sender := "user1"
		fromDate := time.Now().Add(-24 * time.Hour)
		toDate := time.Now()

		criteria := &SearchCriteria{
			ChannelID: &channelID,
			Sender:    &sender,
			FromDate:  &fromDate,
			ToDate:    &toDate,
		}

		assert.NotNil(t, criteria.ChannelID)
		assert.Equal(t, channelID, *criteria.ChannelID)
		assert.NotNil(t, criteria.Sender)
		assert.Equal(t, sender, *criteria.Sender)
		assert.NotNil(t, criteria.FromDate)
		assert.Equal(t, fromDate, *criteria.FromDate)
		assert.NotNil(t, criteria.ToDate)
		assert.Equal(t, toDate, *criteria.ToDate)
	})

	t.Run("部分的な検索条件の作成", func(t *testing.T) {
		channelID := "channel-1"
		criteria := &SearchCriteria{
			ChannelID: &channelID,
		}

		assert.NotNil(t, criteria.ChannelID)
		assert.Equal(t, channelID, *criteria.ChannelID)
		assert.Nil(t, criteria.Sender)
		assert.Nil(t, criteria.FromDate)
		assert.Nil(t, criteria.ToDate)
	})

	t.Run("日付範囲のみの検索条件", func(t *testing.T) {
		fromDate := time.Now().Add(-24 * time.Hour)
		toDate := time.Now()

		criteria := &SearchCriteria{
			FromDate: &fromDate,
			ToDate:   &toDate,
		}

		assert.Nil(t, criteria.ChannelID)
		assert.Nil(t, criteria.Sender)
		assert.NotNil(t, criteria.FromDate)
		assert.Equal(t, fromDate, *criteria.FromDate)
		assert.NotNil(t, criteria.ToDate)
		assert.Equal(t, toDate, *criteria.ToDate)
	})
}
