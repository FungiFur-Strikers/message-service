package repository

import (
	"context"
	"errors"
	"message-service/internal/domain/token"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// モックカーソル
type mockTokenCursor struct {
	mock.Mock
	tokens  []token.Token
	current int
}

func (m *mockTokenCursor) Close(ctx context.Context) error {
	return nil
}

func (m *mockTokenCursor) Next(ctx context.Context) bool {
	m.current++
	return m.current <= len(m.tokens)
}

func (m *mockTokenCursor) Decode(val interface{}) error {
	if m.current > 0 && m.current <= len(m.tokens) {
		*(val.(*token.Token)) = m.tokens[m.current-1]
	}
	return nil
}

func (m *mockTokenCursor) Err() error {
	return nil
}

func (m *mockTokenCursor) All(ctx context.Context, results interface{}) error {
	*(results.(*[]token.Token)) = m.tokens
	return nil
}

// unsafeSetTokenCollection は非公開フィールドに値を設定するためのヘルパー関数
func unsafeSetTokenCollection(repo *TokenRepository, mock *mockCollection) {
	val := reflect.ValueOf(repo).Elem()
	field := val.FieldByName("collection")
	ptr := unsafe.Pointer(field.UnsafeAddr())
	realPtr := (*mongo.Collection)(ptr)
	*realPtr = *(*mongo.Collection)(unsafe.Pointer(mock))
}

func TestTokenRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		token   *token.Token
		mockFn  func(*mockCollection)
		wantErr bool
	}{
		{
			name: "正常系：トークンの作成",
			token: &token.Token{
				Token:     "test-token-value",
				Name:      "test-token",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			mockFn: func(m *mockCollection) {
				m.On("InsertOne", mock.Anything, mock.AnythingOfType("*token.Token")).
					Return(&mongo.InsertOneResult{InsertedID: primitive.NewObjectID()}, nil)
			},
			wantErr: false,
		},
		{
			name: "異常系：データベースエラー",
			token: &token.Token{
				Token:     "test-token-value",
				Name:      "test-token",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			mockFn: func(m *mockCollection) {
				m.On("InsertOne", mock.Anything, mock.AnythingOfType("*token.Token")).
					Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := new(mockCollection)
			tt.mockFn(mock)
			repo := &TokenRepository{}
			unsafeSetTokenCollection(repo, mock)

			err := repo.Create(context.Background(), tt.token)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.token.CreatedAt)
				assert.NotZero(t, tt.token.UpdatedAt)
			}
			mock.AssertExpectations(t)
		})
	}
}

func TestTokenRepository_List(t *testing.T) {
	testTime := time.Now()
	mockTokens := []token.Token{
		{
			ID:        primitive.NewObjectID(),
			Token:     "test-token-value-1",
			Name:      "test-token-1",
			ExpiresAt: testTime.Add(24 * time.Hour),
			CreatedAt: testTime,
			UpdatedAt: testTime,
		},
	}

	cursor := &mockTokenCursor{tokens: mockTokens}

	tests := []struct {
		name    string
		mockFn  func(*mockCollection)
		want    []token.Token
		wantErr bool
	}{
		{
			name: "正常系：トークン一覧取得",
			mockFn: func(m *mockCollection) {
				m.On("Find", mock.Anything, mock.AnythingOfType("primitive.M")).
					Return(cursor, nil)
			},
			want:    mockTokens,
			wantErr: false,
		},
		{
			name: "異常系：データベースエラー",
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
			repo := &TokenRepository{}
			unsafeSetTokenCollection(repo, mock)

			got, err := repo.List(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			mock.AssertExpectations(t)
		})
	}
}

func TestTokenRepository_Delete(t *testing.T) {
	validID := primitive.NewObjectID()
	tests := []struct {
		name    string
		id      string
		mockFn  func(*mockCollection)
		wantErr bool
	}{
		{
			name: "正常系：トークンの削除",
			id:   validID.Hex(),
			mockFn: func(m *mockCollection) {
				m.On("UpdateOne", mock.Anything, mock.AnythingOfType("primitive.M"), mock.AnythingOfType("primitive.M")).
					Return(&mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil)
			},
			wantErr: false,
		},
		{
			name: "異常系：存在しないトークン",
			id:   primitive.NewObjectID().Hex(),
			mockFn: func(m *mockCollection) {
				m.On("UpdateOne", mock.Anything, mock.AnythingOfType("primitive.M"), mock.AnythingOfType("primitive.M")).
					Return(&mongo.UpdateResult{MatchedCount: 0, ModifiedCount: 0}, nil)
			},
			wantErr: true,
		},
		{
			name:    "異常系：無効なID形式",
			id:      "invalid-id",
			mockFn:  func(m *mockCollection) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := new(mockCollection)
			tt.mockFn(mock)
			repo := &TokenRepository{}
			unsafeSetTokenCollection(repo, mock)

			err := repo.Delete(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectations(t)
		})
	}
}

func TestTokenRepository_FindByToken(t *testing.T) {
	testTime := time.Now()
	mockToken := &token.Token{
		ID:        primitive.NewObjectID(),
		Token:     "test-token-value",
		Name:      "test-token",
		ExpiresAt: testTime.Add(24 * time.Hour),
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}

	tests := []struct {
		name        string
		tokenString string
		mockFn      func(*mockCollection)
		want        *token.Token
		wantErr     bool
	}{
		{
			name:        "正常系：トークンの取得",
			tokenString: "test-token-value",
			mockFn: func(m *mockCollection) {
				m.On("FindOne", mock.Anything, mock.AnythingOfType("primitive.M")).
					Return(&mockSingleResult{res: mockToken})
			},
			want:    mockToken,
			wantErr: false,
		},
		{
			name:        "異常系：存在しないトークン",
			tokenString: "non-existent-token",
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
			repo := &TokenRepository{}
			unsafeSetTokenCollection(repo, mock)

			got, err := repo.FindByToken(context.Background(), tt.tokenString)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.want == nil {
					assert.Nil(t, got)
				} else {
					assert.Equal(t, tt.want.Token, got.Token)
				}
			}
			mock.AssertExpectations(t)
		})
	}
}

func TestTokenRepository_FindByID(t *testing.T) {
	testTime := time.Now()
	validID := primitive.NewObjectID()
	mockToken := &token.Token{
		ID:        validID,
		Token:     "test-token-value",
		Name:      "test-token",
		ExpiresAt: testTime.Add(24 * time.Hour),
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}

	tests := []struct {
		name    string
		id      string
		mockFn  func(*mockCollection)
		want    *token.Token
		wantErr bool
	}{
		{
			name: "正常系：トークンの取得",
			id:   validID.Hex(),
			mockFn: func(m *mockCollection) {
				m.On("FindOne", mock.Anything, mock.AnythingOfType("primitive.M")).
					Return(&mockSingleResult{res: mockToken})
			},
			want:    mockToken,
			wantErr: false,
		},
		{
			name: "異常系：存在しないトークン",
			id:   primitive.NewObjectID().Hex(),
			mockFn: func(m *mockCollection) {
				m.On("FindOne", mock.Anything, mock.AnythingOfType("primitive.M")).
					Return(&mockSingleResult{err: mongo.ErrNoDocuments})
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "異常系：無効なID形式",
			id:      "invalid-id",
			mockFn:  func(m *mockCollection) {},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := new(mockCollection)
			tt.mockFn(mock)
			repo := &TokenRepository{}
			unsafeSetTokenCollection(repo, mock)

			got, err := repo.FindByID(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.want == nil {
					assert.Nil(t, got)
				} else {
					assert.Equal(t, tt.want.ID, got.ID)
					assert.Equal(t, tt.want.Token, got.Token)
				}
			}
			mock.AssertExpectations(t)
		})
	}
}
