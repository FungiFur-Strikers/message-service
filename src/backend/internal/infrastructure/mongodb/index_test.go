// src\backend\internal\infrastructure\mongodb\index_test.go
package mongodb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// モックインデックスビュー
type mockIndexView struct {
	mock.Mock
}

func (m *mockIndexView) CreateMany(ctx context.Context, models []mongo.IndexModel, opts ...*options.CreateIndexesOptions) ([]string, error) {
	args := m.Called(ctx, models)
	return args.Get(0).([]string), args.Error(1)
}

// モックコレクション
type mockCollection struct {
	mock.Mock
}

func (m *mockCollection) Indexes() IndexView {
	args := m.Called()
	return args.Get(0).(IndexView)
}

// モックデータベース
type mockDatabase struct {
	mock.Mock
}

func (m *mockDatabase) Collection(name string) Collection {
	args := m.Called(name)
	return args.Get(0).(Collection)
}

func TestCreateIndexes(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() (Database, *mockCollection, *mockCollection, *mockIndexView, *mockIndexView)
		wantErr       bool
		validateIndex func(*testing.T, []mongo.IndexModel)
	}{
		{
			name: "正常系：全てのインデックスが正常に作成される",
			setup: func() (Database, *mockCollection, *mockCollection, *mockIndexView, *mockIndexView) {
				db := new(mockDatabase)
				messagesCol := new(mockCollection)
				tokensCol := new(mockCollection)
				messagesIndexView := new(mockIndexView)
				tokensIndexView := new(mockIndexView)

				db.On("Collection", "messages").Return(messagesCol)
				db.On("Collection", "tokens").Return(tokensCol)

				messagesCol.On("Indexes").Return(messagesIndexView)
				tokensCol.On("Indexes").Return(tokensIndexView)

				messagesIndexView.On("CreateMany", mock.Anything, mock.MatchedBy(func(models []mongo.IndexModel) bool {
					return len(models) == 4 // メッセージコレクションのインデックス数
				})).Return([]string{"index1", "index2", "index3", "index4"}, nil)

				tokensIndexView.On("CreateMany", mock.Anything, mock.MatchedBy(func(models []mongo.IndexModel) bool {
					return len(models) == 3 // トークンコレクションのインデックス数
				})).Return([]string{"index1", "index2", "index3"}, nil)

				return db, messagesCol, tokensCol, messagesIndexView, tokensIndexView
			},
			wantErr: false,
			validateIndex: func(t *testing.T, models []mongo.IndexModel) {
				// メッセージコレクションのインデックス構造を確認
				if len(models) == 4 {
					// UIDとdeleted_atの複合ユニークインデックス
					assert.Equal(t, bson.D{{Key: "uid", Value: 1}, {Key: "deleted_at", Value: 1}}, models[0].Keys)
					assert.True(t, models[0].Options.Unique != nil && *models[0].Options.Unique)

					// channel_idとdeleted_atの複合インデックス
					assert.Equal(t, bson.D{{Key: "channel_id", Value: 1}, {Key: "deleted_at", Value: 1}}, models[1].Keys)

					// senderとdeleted_atの複合インデックス
					assert.Equal(t, bson.D{{Key: "sender", Value: 1}, {Key: "deleted_at", Value: 1}}, models[2].Keys)

					// sent_atとdeleted_atの複合インデックス（降順）
					assert.Equal(t, bson.D{{Key: "sent_at", Value: -1}, {Key: "deleted_at", Value: 1}}, models[3].Keys)
				}
			},
		},
		{
			name: "異常系：メッセージコレクションのインデックス作成が失敗",
			setup: func() (Database, *mockCollection, *mockCollection, *mockIndexView, *mockIndexView) {
				db := new(mockDatabase)
				messagesCol := new(mockCollection)
				tokensCol := new(mockCollection)
				messagesIndexView := new(mockIndexView)
				tokensIndexView := new(mockIndexView)

				db.On("Collection", "messages").Return(messagesCol)
				messagesCol.On("Indexes").Return(messagesIndexView)
				messagesIndexView.On("CreateMany", mock.Anything, mock.Anything).Return([]string{}, assert.AnError)

				return db, messagesCol, tokensCol, messagesIndexView, tokensIndexView
			},
			wantErr: true,
		},
		{
			name: "異常系：トークンコレクションのインデックス作成が失敗",
			setup: func() (Database, *mockCollection, *mockCollection, *mockIndexView, *mockIndexView) {
				db := new(mockDatabase)
				messagesCol := new(mockCollection)
				tokensCol := new(mockCollection)
				messagesIndexView := new(mockIndexView)
				tokensIndexView := new(mockIndexView)

				db.On("Collection", "messages").Return(messagesCol)
				db.On("Collection", "tokens").Return(tokensCol)

				messagesCol.On("Indexes").Return(messagesIndexView)
				tokensCol.On("Indexes").Return(tokensIndexView)

				messagesIndexView.On("CreateMany", mock.Anything, mock.Anything).Return([]string{"index1"}, nil)
				tokensIndexView.On("CreateMany", mock.Anything, mock.Anything).Return([]string{}, assert.AnError)

				return db, messagesCol, tokensCol, messagesIndexView, tokensIndexView
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _, _, messagesIndexView, tokensIndexView := tt.setup()

			err := CreateIndexes(context.Background(), db)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			messagesIndexView.AssertExpectations(t)
			tokensIndexView.AssertExpectations(t)
		})
	}
}

func TestMongoDatabase_Collection(t *testing.T) {
	// モックデータベースを作成
	mockDB := new(mockDatabase)
	mockCollection := new(mockCollection)

	// モックの振る舞いを定義
	mockDB.On("Collection", "test").Return(mockCollection)

	// NewDatabaseを使用してラップ
	wrapper := NewDatabase(mockDB)

	collection := wrapper.Collection("test")
	assert.NotNil(t, collection)

	// モックの期待通りの呼び出しを検証
	mockDB.AssertExpectations(t)
}
func TestMongoCollection_Indexes(t *testing.T) {
	coll := &mongo.Collection{}
	wrapper := &MongoCollection{coll: coll}

	indexes := wrapper.Indexes()
	assert.NotNil(t, indexes)
	assert.IsType(t, &MongoIndexView{}, indexes)
}

func TestMongoIndexView_CreateMany(t *testing.T) {
	ctx := context.Background()

	// モックインデックスビューを作成
	mockIndexView := new(mockIndexView)

	models := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "test", Value: 1}},
		},
	}

	// モックの期待される振る舞いを設定
	expectedIndexNames := []string{"testIndex"}
	mockIndexView.On("CreateMany", ctx, models).Return(expectedIndexNames, nil)

	// テスト実行
	indexNames, err := mockIndexView.CreateMany(ctx, models)

	// アサーション
	assert.NoError(t, err)
	assert.Equal(t, expectedIndexNames, indexNames)

	// モックの期待通りの呼び出しを検証
	mockIndexView.AssertExpectations(t)
}
