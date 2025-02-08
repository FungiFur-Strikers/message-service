package mongodb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		name    string
		setup   func() (Database, *mockCollection, *mockCollection, *mockIndexView, *mockIndexView)
		wantErr bool
	}{
		{
			name: "正常系：全てのインデックスが正常に作成される",
			setup: func() (Database, *mockCollection, *mockCollection, *mockIndexView, *mockIndexView) {
				db := new(mockDatabase)
				messagesCol := new(mockCollection)
				tokensCol := new(mockCollection)
				messagesIndexView := new(mockIndexView)
				tokensIndexView := new(mockIndexView)

				// データベースのモック設定
				db.On("Collection", "messages").Return(messagesCol)
				db.On("Collection", "tokens").Return(tokensCol)

				// コレクションのモック設定
				messagesCol.On("Indexes").Return(messagesIndexView)
				tokensCol.On("Indexes").Return(tokensIndexView)

				// インデックスビューのモック設定
				messagesIndexView.On("CreateMany", mock.Anything, mock.Anything).Return([]string{"index1"}, nil)
				tokensIndexView.On("CreateMany", mock.Anything, mock.Anything).Return([]string{"index1"}, nil)

				return db, messagesCol, tokensCol, messagesIndexView, tokensIndexView
			},
			wantErr: false,
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

				// データベースのモック設定
				db.On("Collection", "messages").Return(messagesCol)
				db.On("Collection", "tokens").Return(tokensCol)

				// コレクションのモック設定
				messagesCol.On("Indexes").Return(messagesIndexView)
				tokensCol.On("Indexes").Return(tokensIndexView)

				// メッセージインデックス作成成功
				messagesIndexView.On("CreateMany", mock.Anything, mock.Anything).Return([]string{"index1"}, nil)
				// トークンインデックス作成失敗
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
