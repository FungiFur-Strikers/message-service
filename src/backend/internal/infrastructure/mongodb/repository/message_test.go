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

func TestNewMessageRepository(t *testing.T) {
	// モックコレクションの作成とリポジトリの初期化
	repo, mockCollection := NewTestRepository()

	// 検証
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.collection)
	assert.Equal(t, mockCollection, repo.collection)
}

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
			name: "異常系：データベースエラー",
			msg: &message.Message{
				UID: "error-uid",
			},
			mockFn: func(m *TestCollection) {
				m.On("InsertOne", mock.Anything, mock.AnythingOfType("*message.Message")).
					Return(nil, mongo.CommandError{Message: "database error"})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockCollection := NewTestRepository()
			tt.mockFn(mockCollection)

			err := repo.Create(context.Background(), tt.msg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.msg.CreatedAt)
				assert.NotZero(t, tt.msg.UpdatedAt)
			}
			mockCollection.AssertExpectations(t)
		})
	}
}

func TestMessageRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		uid     string
		mockFn  func(*TestCollection)
		wantErr bool
		errType error
	}{
		{
			name: "正常系：メッセージの削除",
			uid:  "test-uid",
			mockFn: func(m *TestCollection) {
				m.On("UpdateOne", mock.Anything,
					mock.MatchedBy(func(filter bson.M) bool {
						return filter["uid"] == "test-uid" && filter["deleted_at"] == nil
					}),
					mock.AnythingOfType("primitive.M")).
					Return(&mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil)
			},
			wantErr: false,
		},
		{
			name: "異常系：存在しないメッセージ",
			uid:  "non-existent-uid",
			mockFn: func(m *TestCollection) {
				m.On("UpdateOne", mock.Anything,
					mock.MatchedBy(func(filter bson.M) bool {
						return filter["uid"] == "non-existent-uid" && filter["deleted_at"] == nil
					}),
					mock.AnythingOfType("primitive.M")).
					Return(&mongo.UpdateResult{MatchedCount: 0, ModifiedCount: 0}, nil)
			},
			wantErr: true,
			errType: mongo.ErrNoDocuments,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockCollection := NewTestRepository()
			tt.mockFn(mockCollection)

			err := repo.Delete(context.Background(), tt.uid)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.Equal(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
			}
			mockCollection.AssertExpectations(t)
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
		name     string
		criteria message.SearchCriteria
		mockFn   func(*TestCollection)
		want     []message.Message
		wantErr  bool
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
			name:     "異常系：データベースエラー",
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
			repo, mockCollection := NewTestRepository()
			tt.mockFn(mockCollection)

			got, err := repo.Search(context.Background(), tt.criteria)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			mockCollection.AssertExpectations(t)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockCollection := NewTestRepository()
			tt.mockFn(mockCollection)

			got, err := repo.FindByUID(context.Background(), tt.uid)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			mockCollection.AssertExpectations(t)
		})
	}
}
