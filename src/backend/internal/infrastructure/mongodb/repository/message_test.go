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

func TestMessageRepository_Search(t *testing.T) {
	now := time.Now()
	channelID := "test-channel"
	sender := "test-sender"
	fromDate := now.Add(-24 * time.Hour)
	toDate := now

	tests := []struct {
		name      string
		criteria  message.SearchCriteria
		mockFn    func(*TestCollection)
		want      []message.Message
		wantErr   bool
		assertErr func(*testing.T, error)
	}{
		{
			name: "正常系：全ての検索条件を指定",
			criteria: message.SearchCriteria{
				ChannelID: &channelID,
				Sender:    &sender,
				FromDate:  &fromDate,
				ToDate:    &toDate,
			},
			mockFn: func(m *TestCollection) {
				expectedMessages := []message.Message{
					{
						UID:       "msg1",
						ChannelID: channelID,
						Sender:    sender,
						SentAt:    now,
						Content:   "test message 1",
					},
				}
				cursor := NewTestCursor(expectedMessages)
				m.On("Find", mock.Anything, mock.MatchedBy(func(filter bson.M) bool {
					return filter["channel_id"] == channelID &&
						filter["sender"] == sender &&
						filter["deleted_at"] == nil
				}), mock.Anything).Return(cursor, nil)
			},
			want: []message.Message{
				{
					UID:       "msg1",
					ChannelID: channelID,
					Sender:    sender,
					SentAt:    now,
					Content:   "test message 1",
				},
			},
		},
		{
			name:     "正常系：検索条件なし",
			criteria: message.SearchCriteria{},
			mockFn: func(m *TestCollection) {
				expectedMessages := []message.Message{
					{
						UID:       "msg1",
						ChannelID: "channel1",
						Sender:    "sender1",
						SentAt:    now,
						Content:   "message 1",
					},
					{
						UID:       "msg2",
						ChannelID: "channel2",
						Sender:    "sender2",
						SentAt:    now,
						Content:   "message 2",
					},
				}
				cursor := NewTestCursor(expectedMessages)
				m.On("Find", mock.Anything, bson.M{"deleted_at": nil}, mock.Anything).Return(cursor, nil)
			},
			want: []message.Message{
				{
					UID:       "msg1",
					ChannelID: "channel1",
					Sender:    "sender1",
					SentAt:    now,
					Content:   "message 1",
				},
				{
					UID:       "msg2",
					ChannelID: "channel2",
					Sender:    "sender2",
					SentAt:    now,
					Content:   "message 2",
				},
			},
		},
		{
			name:     "異常系：Findメソッドでのエラー",
			criteria: message.SearchCriteria{},
			mockFn: func(m *TestCollection) {
				m.On("Find", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, mongo.CommandError{Message: "database error"})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := NewTestRepository()
			tt.mockFn(mock)

			got, err := repo.Search(context.Background(), tt.criteria)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.assertErr != nil {
					tt.assertErr(t, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			mock.AssertExpectations(t)
		})
	}
}

func TestMessageRepository_FindByUID(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		uid     string
		mockFn  func(*TestCollection)
		want    *message.Message
		wantErr bool
	}{
		{
			name: "正常系：存在するメッセージの取得",
			uid:  "existing-uid",
			mockFn: func(m *TestCollection) {
				expectedMsg := &message.Message{
					UID:       "existing-uid",
					ChannelID: "test-channel",
					Sender:    "test-sender",
					Content:   "test message",
					SentAt:    now,
				}
				m.On("FindOne", mock.Anything, mock.MatchedBy(func(filter bson.M) bool {
					return filter["uid"] == "existing-uid" && filter["deleted_at"] == nil
				})).Return(NewTestSingleResult(expectedMsg, nil))
			},
			want: &message.Message{
				UID:       "existing-uid",
				ChannelID: "test-channel",
				Sender:    "test-sender",
				Content:   "test message",
				SentAt:    now,
			},
			wantErr: false,
		},
		{
			name: "正常系：存在しないメッセージの取得",
			uid:  "non-existent-uid",
			mockFn: func(m *TestCollection) {
				m.On("FindOne", mock.Anything, mock.MatchedBy(func(filter bson.M) bool {
					return filter["uid"] == "non-existent-uid" && filter["deleted_at"] == nil
				})).Return(NewTestSingleResult(nil, mongo.ErrNoDocuments))
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "異常系：データベースエラー",
			uid:  "error-uid",
			mockFn: func(m *TestCollection) {
				m.On("FindOne", mock.Anything, mock.MatchedBy(func(filter bson.M) bool {
					return filter["uid"] == "error-uid" && filter["deleted_at"] == nil
				})).Return(NewTestSingleResult(nil, mongo.CommandError{Message: "database error"}))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := NewTestRepository()
			tt.mockFn(mock)

			got, err := repo.FindByUID(context.Background(), tt.uid)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			mock.AssertExpectations(t)
		})
	}
}
