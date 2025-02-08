package repository

import (
	"context"
	"errors"
	"message-service/internal/domain/message"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// モックカーソル
type mockCursor struct {
	mock.Mock
	messages []message.Message
	current  int
}

func (m *mockCursor) Close(ctx context.Context) error {
	return nil
}

func (m *mockCursor) Next(ctx context.Context) bool {
	m.current++
	return m.current <= len(m.messages)
}

func (m *mockCursor) Decode(val interface{}) error {
	if m.current > 0 && m.current <= len(m.messages) {
		*(val.(*message.Message)) = m.messages[m.current-1]
	}
	return nil
}

func (m *mockCursor) Err() error {
	return nil
}

func (m *mockCursor) All(ctx context.Context, results interface{}) error {
	*(results.(*[]message.Message)) = m.messages
	return nil
}

// モックシングルリザルト
type mockSingleResult struct {
	mock.Mock
	err error
	res interface{}
}

func (m *mockSingleResult) Decode(v interface{}) error {
	if m.err != nil {
		return m.err
	}
	if msg, ok := m.res.(*message.Message); ok {
		*(v.(*message.Message)) = *msg
	}
	return nil
}

func (m *mockSingleResult) Err() error {
	return m.err
}

// モックコレクション
type mockCollection struct {
	mock.Mock
}

func (m *mockCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *mockCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *mockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.Cursor), args.Error(1)
}

func (m *mockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	args := m.Called(ctx, filter)
	if sr, ok := args.Get(0).(*mongo.SingleResult); ok {
		return sr
	}
	if msr, ok := args.Get(0).(*mockSingleResult); ok {
		return mongo.NewSingleResultFromDocument(msr.res, msr.err, nil)
	}
	return mongo.NewSingleResultFromDocument(nil, errors.New("mock error"), nil)
}

// unsafeSetCollection は非公開フィールドに値を設定するためのヘルパー関数
func unsafeSetCollection(repo *MessageRepository, mock *mockCollection) {
	val := reflect.ValueOf(repo).Elem()
	field := val.FieldByName("collection")
	ptr := unsafe.Pointer(field.UnsafeAddr())
	realPtr := (*mongo.Collection)(ptr)
	*realPtr = *(*mongo.Collection)(unsafe.Pointer(mock))
}

func TestMessageRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		msg     *message.Message
		mockFn  func(*mockCollection)
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
			mockFn: func(m *mockCollection) {
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
			mockFn: func(m *mockCollection) {
				m.On("InsertOne", mock.Anything, mock.AnythingOfType("*message.Message")).
					Return(&mongo.InsertOneResult{}, mongo.WriteException{})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := new(mockCollection)
			tt.mockFn(mock)
			repo := &MessageRepository{}
			unsafeSetCollection(repo, mock)

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
		mockFn  func(*mockCollection)
		wantErr bool
	}{
		{
			name: "正常系：メッセージの削除",
			uid:  "test-uid",
			mockFn: func(m *mockCollection) {
				m.On("UpdateOne", mock.Anything, mock.AnythingOfType("primitive.M"), mock.AnythingOfType("primitive.M")).
					Return(&mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil)
			},
			wantErr: false,
		},
		{
			name: "異常系：存在しないメッセージ",
			uid:  "non-existent-uid",
			mockFn: func(m *mockCollection) {
				m.On("UpdateOne", mock.Anything, mock.AnythingOfType("primitive.M"), mock.AnythingOfType("primitive.M")).
					Return(&mongo.UpdateResult{MatchedCount: 0, ModifiedCount: 0}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := new(mockCollection)
			tt.mockFn(mock)
			repo := &MessageRepository{}
			unsafeSetCollection(repo, mock)

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
	testTime := time.Now()
	mockMessages := []message.Message{
		{
			UID:       "test-uid-1",
			SentAt:    testTime,
			Sender:    "test-sender",
			ChannelID: "test-channel",
			Content:   "test message 1",
			CreatedAt: testTime,
			UpdatedAt: testTime,
		},
	}

	cursor := &mockCursor{messages: mockMessages}

	tests := []struct {
		name     string
		criteria message.SearchCriteria
		mockFn   func(*mockCollection)
		want     []message.Message
		wantErr  bool
	}{
		{
			name: "正常系：検索結果あり",
			criteria: message.SearchCriteria{
				ChannelID: stringPtr("test-channel"),
			},
			mockFn: func(m *mockCollection) {
				m.On("Find", mock.Anything, mock.AnythingOfType("primitive.M")).
					Return(cursor, nil)
			},
			want:    mockMessages,
			wantErr: false,
		},
		{
			name:     "異常系：データベースエラー",
			criteria: message.SearchCriteria{},
			mockFn: func(m *mockCollection) {
				m.On("Find", mock.Anything, mock.AnythingOfType("primitive.M")).
					Return(nil, errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := new(mockCollection)
			tt.mockFn(mock)
			repo := &MessageRepository{}
			unsafeSetCollection(repo, mock)

			got, err := repo.Search(context.Background(), tt.criteria)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectations(t)
		})
	}
}

func TestMessageRepository_FindByUID(t *testing.T) {
	testTime := time.Now()
	mockMessage := &message.Message{
		UID:       "test-uid",
		SentAt:    testTime,
		Sender:    "test-sender",
		ChannelID: "test-channel",
		Content:   "test message",
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}

	tests := []struct {
		name    string
		uid     string
		mockFn  func(*mockCollection)
		want    *message.Message
		wantErr bool
	}{
		{
			name: "正常系：メッセージの取得",
			uid:  "test-uid",
			mockFn: func(m *mockCollection) {
				m.On("FindOne", mock.Anything, mock.AnythingOfType("primitive.M")).
					Return(&mockSingleResult{res: mockMessage})
			},
			want:    mockMessage,
			wantErr: false,
		},
		{
			name: "異常系：存在しないUID",
			uid:  "non-existent-uid",
			mockFn: func(m *mockCollection) {
				m.On("FindOne", mock.Anything, mock.AnythingOfType("primitive.M")).
					Return(&mockSingleResult{err: mongo.ErrNoDocuments})
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := new(mockCollection)
			tt.mockFn(mock)
			repo := &MessageRepository{}
			unsafeSetCollection(repo, mock)

			got, err := repo.FindByUID(context.Background(), tt.uid)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.want == nil {
					assert.Nil(t, got)
				} else {
					assert.Equal(t, tt.want.UID, got.UID)
				}
			}
			mock.AssertExpectations(t)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
