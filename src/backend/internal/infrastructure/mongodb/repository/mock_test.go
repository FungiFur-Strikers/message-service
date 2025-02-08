package repository

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoCollectionInterface はmongoドライバーの必要なメソッドを定義するインターフェース
type MongoCollectionInterface interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (mongo.Cursor, error)
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) mongo.SingleResult
}

// TestCollection はテスト用のモックコレクション
type TestCollection struct {
	mock.Mock
}

func (m *TestCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *TestCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *TestCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (mongo.Cursor, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(mongo.Cursor), args.Error(1)
}

func (m *TestCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(mongo.SingleResult)
}

// TestCursor はテスト用のモックカーソル
type TestCursor struct {
	mock.Mock
	Results  interface{}
	Position int
}

func (m *TestCursor) Close(ctx context.Context) error {
	return nil
}

func (m *TestCursor) Next(ctx context.Context) bool {
	m.Position++
	return m.Position <= m.getResultsLength()
}

func (m *TestCursor) Decode(val interface{}) error {
	if m.Position > 0 && m.Position <= m.getResultsLength() {
		switch results := m.Results.(type) {
		case []interface{}:
			copyValue(val, results[m.Position-1])
		}
	}
	return nil
}

func (m *TestCursor) getResultsLength() int {
	switch results := m.Results.(type) {
	case []interface{}:
		return len(results)
	default:
		return 0
	}
}

// TestSingleResult はテスト用のモックシングルリザルト
type TestSingleResult struct {
	mock.Mock
	error    error
	response interface{}
}

func (m *TestSingleResult) Decode(v interface{}) error {
	if m.error != nil {
		return m.error
	}
	copyValue(v, m.response)
	return nil
}

func (m *TestSingleResult) Err() error {
	return m.error
}

// テストヘルパー関数
func copyValue(dst interface{}, src interface{}) {
	switch d := dst.(type) {
	case *[]interface{}:
		if s, ok := src.([]interface{}); ok {
			*d = s
		}
	default:
		if d != nil && src != nil {
			*d.(*interface{}) = src
		}
	}
}

// NewTestRepository はテスト用のリポジトリとモックコレクションを作成します
func NewTestRepository() (*MessageRepository, *TestCollection) {
	mock := new(TestCollection)
	return &MessageRepository{
		collection: &mongo.Collection{},
	}, mock
}

// NewTestSingleResult はテスト用のシングルリザルトを作成します
func NewTestSingleResult(response interface{}, err error) *TestSingleResult {
	return &TestSingleResult{
		error:    err,
		response: response,
	}
}

// NewTestCursor はテスト用のカーソルを作成します
func NewTestCursor(results interface{}) *TestCursor {
	return &TestCursor{
		Results:  results,
		Position: 0,
	}
}
