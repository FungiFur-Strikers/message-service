// src\backend\internal\infrastructure\mongodb\repository\collection_test.go
package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

type mockMongoCollection struct {
	mock.Mock
}

func (m *mockMongoCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document, opts)
	if res := args.Get(0); res != nil {
		return res.(*mongo.InsertOneResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockMongoCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update, opts)
	if res := args.Get(0); res != nil {
		return res.(*mongo.UpdateResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockMongoCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	args := m.Called(ctx, filter, opts)
	if res := args.Get(0); res != nil {
		return res.(*mongo.Cursor), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockMongoCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	args := m.Called(ctx, filter, opts)
	if res := args.Get(0); res != nil {
		return res.(*mongo.SingleResult)
	}
	return nil
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
		mockColl := &mockMongoCollection{}

		mockCursor := &mongo.Cursor{}
		mockColl.On("Find", ctx, filter, mock.Anything).Return(mockCursor, nil)

		cursor, err := mockColl.Find(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, cursor)
		mockColl.AssertExpectations(t)
	})

	t.Run("エラー発生", func(t *testing.T) {
		mockColl := &mockMongoCollection{}

		expectedErr := errors.New("find error")
		mockColl.On("Find", ctx, filter, mock.Anything).Return(nil, expectedErr)

		cursor, err := mockColl.Find(ctx, filter)

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
		mockColl := &mockMongoCollection{}

		expectedResult := &mongo.SingleResult{}
		mockColl.On("FindOne", ctx, filter, mock.Anything).Return(expectedResult)

		result := mockColl.FindOne(ctx, filter)

		assert.NotNil(t, result)
		mockColl.AssertExpectations(t)
	})
}

func TestNewMongoCollectionWrapper(t *testing.T) {
	coll := &mongo.Collection{}
	wrapper := NewMongoCollectionWrapper(coll)

	assert.NotNil(t, wrapper)
	assert.IsType(t, &MongoCollectionWrapper{}, wrapper)

	// 型アサーションでCollectionフィールドを確認
	collWrapper, ok := wrapper.(*MongoCollectionWrapper)
	assert.True(t, ok)
	assert.Equal(t, coll, collWrapper.Collection)
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
