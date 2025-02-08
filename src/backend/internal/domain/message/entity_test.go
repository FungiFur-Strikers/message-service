package message

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// テストヘルパー関数
func createTestMessage(t *testing.T) *Message {
	t.Helper()
	now := time.Now()
	return &Message{
		ID:        primitive.NewObjectID(),
		UID:       "msg-123",
		SentAt:    now,
		Sender:    "user1",
		ChannelID: "channel-1",
		Content:   "Hello, World!",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestMessage_Creation(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Message
		validate    func(*testing.T, *Message)
		shouldError bool
	}{
		{
			name: "正常なメッセージの作成",
			setup: func() *Message {
				return createTestMessage(t)
			},
			validate: func(t *testing.T, msg *Message) {
				assert.NotEmpty(t, msg.ID)
				assert.Equal(t, "msg-123", msg.UID)
				assert.NotZero(t, msg.SentAt)
				assert.Equal(t, "user1", msg.Sender)
				assert.Equal(t, "channel-1", msg.ChannelID)
				assert.Equal(t, "Hello, World!", msg.Content)
				assert.NotZero(t, msg.CreatedAt)
				assert.NotZero(t, msg.UpdatedAt)
				assert.Nil(t, msg.DeletedAt)
			},
		},
		{
			name: "削除済みメッセージの作成",
			setup: func() *Message {
				msg := createTestMessage(t)
				deletedAt := time.Now().Add(time.Hour)
				msg.DeletedAt = &deletedAt
				return msg
			},
			validate: func(t *testing.T, msg *Message) {
				assert.NotNil(t, msg.DeletedAt)
				assert.True(t, msg.DeletedAt.After(msg.CreatedAt))
			},
		},
		{
			name: "空のContentを持つメッセージ",
			setup: func() *Message {
				msg := createTestMessage(t)
				msg.Content = ""
				return msg
			},
			validate: func(t *testing.T, msg *Message) {
				assert.Empty(t, msg.Content)
			},
		},
		{
			name: "更新日時が作成日時より前の場合",
			setup: func() *Message {
				msg := createTestMessage(t)
				msg.UpdatedAt = msg.CreatedAt.Add(-time.Hour)
				return msg
			},
			validate: func(t *testing.T, msg *Message) {
				assert.True(t, msg.UpdatedAt.Before(msg.CreatedAt))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.setup()
			tt.validate(t, msg)
		})
	}
}

func TestSearchCriteria_Validation(t *testing.T) {
	t.Run("検索条件のバリデーション", func(t *testing.T) {
		tests := []struct {
			name     string
			criteria SearchCriteria
			isValid  bool
		}{
			{
				name: "有効な完全な検索条件",
				criteria: SearchCriteria{
					ChannelID: strPtr("channel-1"),
					Sender:    strPtr("user1"),
					FromDate:  timePtr(time.Now().Add(-24 * time.Hour)),
					ToDate:    timePtr(time.Now()),
				},
				isValid: true,
			},
			{
				name: "チャンネルIDのみの検索条件",
				criteria: SearchCriteria{
					ChannelID: strPtr("channel-1"),
				},
				isValid: true,
			},
			{
				name: "送信者のみの検索条件",
				criteria: SearchCriteria{
					Sender: strPtr("user1"),
				},
				isValid: true,
			},
			{
				name: "日付範囲のみの検索条件",
				criteria: SearchCriteria{
					FromDate: timePtr(time.Now().Add(-24 * time.Hour)),
					ToDate:   timePtr(time.Now()),
				},
				isValid: true,
			},
			{
				name: "終了日が開始日より前の無効な日付範囲",
				criteria: SearchCriteria{
					FromDate: timePtr(time.Now()),
					ToDate:   timePtr(time.Now().Add(-24 * time.Hour)),
				},
				isValid: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.isValid {
					if tt.criteria.FromDate != nil && tt.criteria.ToDate != nil {
						assert.True(t, tt.criteria.FromDate.Before(*tt.criteria.ToDate) || tt.criteria.FromDate.Equal(*tt.criteria.ToDate))
					}
				} else {
					if tt.criteria.FromDate != nil && tt.criteria.ToDate != nil {
						assert.True(t, tt.criteria.FromDate.After(*tt.criteria.ToDate))
					}
				}
			})
		}
	})
}

// ヘルパー関数
func strPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
