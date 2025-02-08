package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database インターフェース
type Database interface {
	Collection(name string) Collection
}

// Collection インターフェース
type Collection interface {
	Indexes() IndexView
}

// IndexView インターフェース
type IndexView interface {
	CreateMany(ctx context.Context, models []mongo.IndexModel, opts ...*options.CreateIndexesOptions) ([]string, error)
}

// MongoDatabase は実際のmongo.Databaseのラッパー
type MongoDatabase struct {
	db *mongo.Database
}

func (m *MongoDatabase) Collection(name string) Collection {
	return &MongoCollection{coll: m.db.Collection(name)}
}

// MongoCollection は実際のmongo.Collectionのラッパー
type MongoCollection struct {
	coll *mongo.Collection
}

func (m *MongoCollection) Indexes() IndexView {
	return &MongoIndexView{view: m.coll.Indexes()}
}

// MongoIndexView は実際のmongo.IndexViewのラッパー
type MongoIndexView struct {
	view mongo.IndexView
}

func (m *MongoIndexView) CreateMany(ctx context.Context, models []mongo.IndexModel, opts ...*options.CreateIndexesOptions) ([]string, error) {
	return m.view.CreateMany(ctx, models, opts...)
}

// NewDatabase は*mongo.DatabaseからDatabaseインターフェースを作成
func NewDatabase(db *mongo.Database) Database {
	return &MongoDatabase{db: db}
}

func CreateIndexes(ctx context.Context, db Database) error {
	// メッセージコレクションのインデックス
	messageIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "uid", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "channel_id", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "sender", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "sent_at", Value: -1},
				{Key: "deleted_at", Value: 1},
			},
		},
	}

	// トークンコレクションのインデックス
	tokenIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "token", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "expires_at", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "created_at", Value: -1},
				{Key: "deleted_at", Value: 1},
			},
		},
	}

	// メッセージインデックスの作成
	_, err := db.Collection("messages").Indexes().CreateMany(ctx, messageIndexes)
	if err != nil {
		return err
	}

	// トークンインデックスの作成
	_, err = db.Collection("tokens").Indexes().CreateMany(ctx, tokenIndexes)
	if err != nil {
		return err
	}

	return nil
}
