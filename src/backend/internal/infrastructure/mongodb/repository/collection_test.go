// src\backend\internal\infrastructure\mongodb\repository\collection_test.go
package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
)

// モックの定義
type mockMongoCursor struct {
	mock.Mock
}

func (m *mockMongoCursor) Next(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *mockMongoCursor) Decode(val interface{}) error {
	args := m.Called(val)
	return args.Error(0)
}

func (m *mockMongoCursor) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockMongoCursor) All(ctx context.Context, results interface{}) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}

// テストケース
func TestMongoCursorWrapper_All(t *testing.T) {
	ctx := context.Background()

	t.Run("正常系", func(t *testing.T) {
		mockCursor := &mockMongoCursor{}
		wrapper := &MongoCursorWrapper{Cursor: mockCursor}

		var results []interface{}
		mockCursor.On("All", ctx, &results).Return(nil)

		err := wrapper.All(ctx, &results)

		assert.NoError(t, err)
		mockCursor.AssertExpectations(t)
	})

	t.Run("エラー発生", func(t *testing.T) {
		mockCursor := &mockMongoCursor{}
		wrapper := &MongoCursorWrapper{Cursor: mockCursor}

		var results []interface{}
		expectedErr := errors.New("cursor error")
		mockCursor.On("All", ctx, &results).Return(expectedErr)

		err := wrapper.All(ctx, &results)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockCursor.AssertExpectations(t)
	})
}

func TestMongoCollectionWrapper_Find(t *testing.T) {
	ctx := context.Background()
	filter := map[string]interface{}{"key": "value"}

	t.Run("正常系", func(t *testing.T) {
		mockColl := new(TestCollection)
		wrapper := &MongoCollectionWrapper{Collection: mockColl}

		mockCursor := &mockMongoCursor{}
		mockColl.On("Find", ctx, filter).Return(mockCursor, nil)

		cursor, err := wrapper.Find(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, cursor)
		mockColl.AssertExpectations(t)
	})

	t.Run("エラー発生", func(t *testing.T) {
		mockColl := new(TestCollection)
		wrapper := &MongoCollectionWrapper{Collection: mockColl}

		expectedErr := errors.New("find error")
		mockColl.On("Find", ctx, filter).Return(nil, expectedErr)

		cursor, err := wrapper.Find(ctx, filter)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, cursor)
		mockColl.AssertExpectations(t)
	})
}

func TestMongoCollectionWrapper_FindOne(t *testing.T) {
	ctx := context.Background()
	filter := map[string]interface{}{"key": "value"}

	t.Run("正常系", func(t *testing.T) {
		mockColl := new(TestCollection)
		wrapper := &MongoCollectionWrapper{Collection: mockColl}

		expectedResult := NewTestSingleResult(&struct{}{}, nil)
		mockColl.On("FindOne", ctx, filter).Return(expectedResult)

		result := wrapper.FindOne(ctx, filter)

		assert.NotNil(t, result)
		mockColl.AssertExpectations(t)
	})

	t.Run("結果が見つからない場合", func(t *testing.T) {
		mockColl := new(TestCollection)
		wrapper := &MongoCollectionWrapper{Collection: mockColl}

		mockColl.On("FindOne", ctx, filter).Return(NewTestSingleResult(nil, mongo.ErrNoDocuments))

		result := wrapper.FindOne(ctx, filter)

		assert.NotNil(t, result)
		mockColl.AssertExpectations(t)
	})
}

func TestMongoCollectionWrapper_FindByID(t *testing.T) {
	ctx := context.Background()
	id := "test_id"

	t.Run("正常系", func(t *testing.T) {
		mockColl := new(TestCollection)
		wrapper := &MongoCollectionWrapper{Collection: mockColl}

		expectedResult := NewTestSingleResult(&struct{}{}, nil)
		mockColl.On("FindOne", ctx, mock.Anything).Return(expectedResult)

		result := wrapper.FindByID(ctx, id)

		assert.NotNil(t, result)
		mockColl.AssertExpectations(t)
	})

	t.Run("IDが見つからない場合", func(t *testing.T) {
		mockColl := new(TestCollection)
		wrapper := &MongoCollectionWrapper{Collection: mockColl}

		mockColl.On("FindOne", ctx, mock.Anything).Return(NewTestSingleResult(nil, mongo.ErrNoDocuments))

		result := wrapper.FindByID(ctx, id)

		assert.NotNil(t, result)
		mockColl.AssertExpectations(t)
	})
}

func TestNewMongoCollectionWrapper(t *testing.T) {
	coll := &mongo.Collection{}
	wrapper := NewMongoCollectionWrapper(coll)

	assert.NotNil(t, wrapper)
	assert.IsType(t, &MongoCollectionWrapper{}, wrapper)
}

func TestMongoCursorWrapper_Next(t *testing.T) {
	ctx := context.Background()
	mockCursor := &mockMongoCursor{}
	wrapper := &MongoCursorWrapper{Cursor: mockCursor}

	mockCursor.On("Next", ctx).Return(true)

	result := wrapper.Next(ctx)

	assert.True(t, result)
	mockCursor.AssertExpectations(t)
}

func TestMongoCursorWrapper_Decode(t *testing.T) {
	mockCursor := &mockMongoCursor{}
	wrapper := &MongoCursorWrapper{Cursor: mockCursor}

	var result interface{}
	mockCursor.On("Decode", &result).Return(nil)

	err := wrapper.Decode(&result)

	assert.NoError(t, err)
	mockCursor.AssertExpectations(t)
}

func TestMongoCursorWrapper_Close(t *testing.T) {
	ctx := context.Background()
	mockCursor := &mockMongoCursor{}
	wrapper := &MongoCursorWrapper{Cursor: mockCursor}

	mockCursor.On("Close", ctx).Return(nil)

	err := wrapper.Close(ctx)

	assert.NoError(t, err)
	mockCursor.AssertExpectations(t)
}

func TestMongoCollectionAdapter(t *testing.T) {
	ctx := context.Background()
	coll := &mongo.Collection{}
	adapter := &MongoCollectionAdapter{coll: coll}

	t.Run("インターフェースの実装を確認", func(t *testing.T) {
		var _ MongoCollectionInterface = adapter
	})

	// Note: 実際のMongoDBへの接続が必要なため、
	// これらのメソッドの詳細なテストは統合テストで行うべきです
	t.Run("メソッドの存在を確認", func(t *testing.T) {
		assert.NotPanics(t, func() {
			adapter.InsertOne(ctx, nil)
			adapter.UpdateOne(ctx, nil, nil)
			adapter.Find(ctx, nil)
			adapter.FindOne(ctx, nil)
		})
	})
}
