package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateIndexes(ctx context.Context, db *mongo.Database) error {
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
