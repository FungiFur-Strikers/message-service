package handler

import (
	"context"
	"message-service/internal/domain/message"
	"message-service/internal/domain/token"

	"github.com/stretchr/testify/mock"
)

// mockMessageRepository はメッセージリポジトリのモック
type mockMessageRepository struct {
	mock.Mock
}

func (m *mockMessageRepository) Create(ctx context.Context, msg *message.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *mockMessageRepository) Delete(ctx context.Context, uid string) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}

func (m *mockMessageRepository) Search(ctx context.Context, criteria message.SearchCriteria) ([]message.Message, error) {
	args := m.Called(ctx, criteria)
	if msgs, ok := args.Get(0).([]message.Message); ok {
		return msgs, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockMessageRepository) FindByUID(ctx context.Context, uid string) (*message.Message, error) {
	args := m.Called(ctx, uid)
	if msg, ok := args.Get(0).(*message.Message); ok {
		return msg, args.Error(1)
	}
	return nil, args.Error(1)
}

// mockTokenRepository はトークンリポジトリのモック
type mockTokenRepository struct {
	mock.Mock
}

func (m *mockTokenRepository) Create(ctx context.Context, t *token.Token) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *mockTokenRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockTokenRepository) List(ctx context.Context) ([]token.Token, error) {
	args := m.Called(ctx)
	if tokens, ok := args.Get(0).([]token.Token); ok {
		return tokens, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTokenRepository) FindByID(ctx context.Context, id string) (*token.Token, error) {
	args := m.Called(ctx, id)
	if t, ok := args.Get(0).(*token.Token); ok {
		return t, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTokenRepository) FindByToken(ctx context.Context, tokenStr string) (*token.Token, error) {
	args := m.Called(ctx, tokenStr)
	if t, ok := args.Get(0).(*token.Token); ok {
		return t, args.Error(1)
	}
	return nil, args.Error(1)
}
