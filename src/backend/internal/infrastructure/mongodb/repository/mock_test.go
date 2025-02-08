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

// NewTestSingleResult はテスト用の SingleResult を作成
func NewTestSingleResult(response interface{}, err error) *mockSingleResult {
	return &mockSingleResult{
		result: response,
		err:    err,
	}
}

// copyValue はインターフェースの値をコピーするヘルパー関数
func copyValue(dst interface{}, src interface{}) error {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer")
	}
	dstVal = dstVal.Elem()

	switch dstVal.Kind() {
	case reflect.Struct:
		return copyStruct(dst, src)
	case reflect.Slice:
		return copySlice(dst, src)
	default:
		return fmt.Errorf("unsupported type for copy")
	}
}

// copyStruct は構造体をコピーする
func copyStruct(dst interface{}, src interface{}) error {
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
	return fmt.Errorf("unsupported struct type for copy")
}

// copySlice はスライスをコピーする
func copySlice(dst interface{}, src interface{}) error {
	if tokens, ok := dst.(*[]token.Token); ok {
		if srcSlice, ok := src.([]token.Token); ok {
			*tokens = make([]token.Token, len(srcSlice))
			copy(*tokens, srcSlice)
			return nil
		}
		return copyInterfaceSlice(tokens, src)
	}
	if messages, ok := dst.(*[]message.Message); ok {
		if srcSlice, ok := src.([]message.Message); ok {
			*messages = make([]message.Message, len(srcSlice))
			copy(*messages, srcSlice)
			return nil
		}
		return copyInterfaceSlice(messages, src)
	}
	return fmt.Errorf("unsupported slice type for copy")
}

// copyInterfaceSlice はインターフェーススライスをコピーする
func copyInterfaceSlice(dst interface{}, src interface{}) error {
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() != reflect.Slice {
		return fmt.Errorf("source must be a slice")
	}

	switch d := dst.(type) {
	case *[]token.Token:
		tokens := make([]token.Token, srcVal.Len())
		for i := 0; i < srcVal.Len(); i++ {
			if token, ok := srcVal.Index(i).Interface().(*token.Token); ok {
				tokens[i] = *token
			}
		}
		*d = tokens
		return nil
	case *[]message.Message:
		messages := make([]message.Message, srcVal.Len())
		for i := 0; i < srcVal.Len(); i++ {
			if message, ok := srcVal.Index(i).Interface().(*message.Message); ok {
				messages[i] = *message
			}
		}
		*d = messages
		return nil
	}
	return fmt.Errorf("unsupported interface slice type")
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
