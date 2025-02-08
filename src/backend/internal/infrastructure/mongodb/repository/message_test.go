package repository

import (
	"context"
	"message-service/internal/domain/message"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMessageRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		msg     *message.Message
		mockFn  func(*TestCollection)
		wantErr bool
	}{
		{
			name: "正常系：メッセージの作成",
			msg: &message.Message{
				UID:       "test-uid",
				SentAt:    time.Now(),
				Sender:    "test-sender",
				ChannelID: "test-channel",
				Content:   "test message",
			},
			mockFn: func(m *TestCollection) {
				m.On("InsertOne", mock.Anything, mock.AnythingOfType("*message.Message")).
					Return(&mongo.InsertOneResult{InsertedID: "test-id"}, nil)
			},
			wantErr: false,
		},
		{
			name: "異常系：重複するUID",
			msg: &message.Message{
				UID:       "duplicate-uid",
				SentAt:    time.Now(),
				Sender:    "test-sender",
				ChannelID: "test-channel",
				Content:   "test message",
			},
			mockFn: func(m *TestCollection) {
				m.On("InsertOne", mock.Anything, mock.AnythingOfType("*message.Message")).
					Return(&mongo.InsertOneResult{}, mongo.WriteException{})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := NewTestRepository()
			tt.mockFn(mock)

			err := repo.Create(context.Background(), tt.msg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.msg.CreatedAt)
				assert.NotZero(t, tt.msg.UpdatedAt)
			}
			mock.AssertExpectations(t)
		})
	}
}

func TestMessageRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		uid     string
		mockFn  func(*TestCollection)
		wantErr bool
	}{
		{
			name: "正常系：メッセージの削除",
			uid:  "test-uid",
			mockFn: func(m *TestCollection) {
				m.On("UpdateOne", mock.Anything, mock.MatchedBy(func(filter bson.M) bool {
					return filter["uid"] == "test-uid" && filter["deleted_at"] == nil
				}), mock.AnythingOfType("primitive.M")).
					Return(&mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil)
			},
			wantErr: false,
		},
		{
			name: "異常系：存在しないメッセージ",
			uid:  "non-existent-uid",
			mockFn: func(m *TestCollection) {
				m.On("UpdateOne", mock.Anything, mock.MatchedBy(func(filter bson.M) bool {
					return filter["uid"] == "non-existent-uid" && filter["deleted_at"] == nil
				}), mock.AnythingOfType("primitive.M")).
					Return(&mongo.UpdateResult{MatchedCount: 0, ModifiedCount: 0}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := NewTestRepository()
			tt.mockFn(mock)

			err := repo.Delete(context.Background(), tt.uid)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, mongo.ErrNoDocuments, err)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectations(t)
		})
	}
}
