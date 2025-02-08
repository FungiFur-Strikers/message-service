package repository

import (
	"context"
	"fmt"
	"message-service/internal/domain/message"
	"message-service/internal/domain/token"
	"reflect"

	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestCollection はテスト用のモックコレクション
type TestCollection struct {
	mock.Mock
}

// InsertOne モックメソッド
func (m *TestCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*mongo.InsertOneResult), args.Error(1)
}

// UpdateOne モックメソッド
func (m *TestCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*mongo.UpdateResult), args.Error(1)
}

// Find モックメソッド
func (m *TestCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (CursorInterface, error) {
	args := m.Called(ctx, filter)
	if cursor, ok := args.Get(0).(CursorInterface); ok {
		return cursor, args.Error(1)
	}
	return nil, args.Error(1)
}

// mockSingleResult は SingleResult のモック実装
type mockSingleResult struct {
	result interface{}
	err    error
}

func (m *mockSingleResult) Decode(v interface{}) error {
	if m.err != nil {
		return m.err
	}
	if m.result == nil {
		return mongo.ErrNoDocuments
	}
	return copyValue(v, m.result)
}

// FindOne モックメソッド
func (m *TestCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult {
	args := m.Called(ctx, filter)
	if result, ok := args.Get(0).(*mockSingleResult); ok {
		return result
	}
	return &mockSingleResult{err: mongo.ErrNoDocuments}
}

// TestCursor はテスト用のモックカーソル
type TestCursor struct {
	Results  []interface{}
	Position int
}

func (m *TestCursor) Next(ctx context.Context) bool {
	m.Position++
	return m.Position <= len(m.Results)
}

func (m *TestCursor) Decode(val interface{}) error {
	if m.Position > 0 && m.Position <= len(m.Results) {
		if err := copyValue(val, m.Results[m.Position-1]); err != nil {
			return err
		}
	}
	return nil
}

func (m *TestCursor) Close(ctx context.Context) error {
	return nil
}

func (m *TestCursor) All(ctx context.Context, results interface{}) error {
	return copyValue(results, m.Results)
}

// NewTestCursor テストカーソル作成
func NewTestCursor[T any](results []T) CursorInterface {
	var interfaceSlice []interface{}
	for _, item := range results {
		itemCopy := item
		interfaceSlice = append(interfaceSlice, &itemCopy)
	}

	return &TestCursor{
		Results:  interfaceSlice,
		Position: 0,
	}
}

// NewTestSingleResult はテスト用の SingleResult を作成
func NewTestSingleResult(response interface{}, err error) *mockSingleResult {
	return &mockSingleResult{
		result: response,
		err:    err,
	}
}

// テストヘルパー関数
func copyValue(dst interface{}, src interface{}) error {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer")
	}
	dstVal = dstVal.Elem()

	switch dstVal.Kind() {
	case reflect.Struct:
		// 構造体の場合（単一の Token または Message）
		switch v := dst.(type) {
		case *token.Token:
			if srcVal, ok := src.(*token.Token); ok {
				*v = *srcVal
				return nil
			}
		case *message.Message:
			if srcVal, ok := src.(*message.Message); ok {
				*v = *srcVal
				return nil
			}
		}
	case reflect.Slice:
		// スライスの場合（Token または Message のスライス）
		switch src := src.(type) {
		case []token.Token:
			if tokens, ok := dst.(*[]token.Token); ok {
				*tokens = make([]token.Token, len(src))
				copy(*tokens, src)
				return nil
			}
		case []message.Message:
			if messages, ok := dst.(*[]message.Message); ok {
				*messages = make([]message.Message, len(src))
				copy(*messages, src)
				return nil
			}
		case []interface{}:
			// インターフェースのスライスの場合
			switch dstType := reflect.TypeOf(dst).Elem(); dstType.Elem().String() {
			case "token.Token":
				tokens := make([]token.Token, len(src))
				for i, item := range src {
					if token, ok := item.(*token.Token); ok {
						tokens[i] = *token
					}
				}
				reflect.ValueOf(dst).Elem().Set(reflect.ValueOf(tokens))
				return nil
			case "message.Message":
				messages := make([]message.Message, len(src))
				for i, item := range src {
					if message, ok := item.(*message.Message); ok {
						messages[i] = *message
					}
				}
				reflect.ValueOf(dst).Elem().Set(reflect.ValueOf(messages))
				return nil
			}
		}
	}
	return fmt.Errorf("unsupported type for copy")
}

// NewTestTokenRepository はテスト用のTokenRepositoryを作成
func NewTestTokenRepository() (*TokenRepository, *TestCollection) {
	mock := new(TestCollection)
	return &TokenRepository{
		collection: mock,
	}, mock
}

// NewTestRepository はテスト用のMessageRepositoryを作成
func NewTestRepository() (*MessageRepository, *TestCollection) {
	mock := new(TestCollection)
	return &MessageRepository{
		collection: mock,
	}, mock
}
