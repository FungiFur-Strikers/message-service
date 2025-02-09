// src\backend\internal\infrastructure\mongodb\repository\collection.go

package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SingleResult はドキュメント取得結果の操作に必要なメソッドを定義するインターフェース
type SingleResult interface {
	Decode(v interface{}) error
}

// CursorInterface はカーソル操作に必要なメソッドを定義するインターフェース
type CursorInterface interface {
	Next(ctx context.Context) bool
	Decode(val interface{}) error
	Close(ctx context.Context) error
	All(ctx context.Context, results interface{}) error
}

// MongoCollectionInterface はmongoドライバーの必要なメソッドを定義するインターフェース
type MongoCollectionInterface interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (CursorInterface, error)
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult
}

// MongoCursorWrapper は実際のmongo.Cursorをラップする構造体
type MongoCursorWrapper struct {
	Cursor CursorInterface
}

// CursorInterface の実装
func (w *MongoCursorWrapper) Next(ctx context.Context) bool {
	return w.Cursor.Next(ctx)
}

func (w *MongoCursorWrapper) Decode(val interface{}) error {
	return w.Cursor.Decode(val)
}

func (w *MongoCursorWrapper) Close(ctx context.Context) error {
	return w.Cursor.Close(ctx)
}

func (w *MongoCursorWrapper) All(ctx context.Context, results interface{}) error {
	return w.Cursor.All(ctx, results)
}

// MongoCollectionAdapter は*mongo.CollectionをMongoCollectionInterfaceに適合させるアダプター
type MongoCollectionAdapter struct {
	coll *mongo.Collection
}

func (a *MongoCollectionAdapter) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return a.coll.InsertOne(ctx, document, opts...)
}

func (a *MongoCollectionAdapter) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return a.coll.UpdateOne(ctx, filter, update, opts...)
}

func (a *MongoCollectionAdapter) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (CursorInterface, error) {
	cursor, err := a.coll.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return &MongoCursorWrapper{Cursor: cursor}, nil
}

func (a *MongoCollectionAdapter) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult {
	return a.coll.FindOne(ctx, filter, opts...)
}

// MongoCollectionWrapper は実際のmongo.Collectionをラップする構造体
type MongoCollectionWrapper struct {
	Collection MongoCollectionInterface
}

func (w *MongoCollectionWrapper) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (CursorInterface, error) {
	return w.Collection.Find(ctx, filter, opts...)
}

func (w *MongoCollectionWrapper) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult {
	return w.Collection.FindOne(ctx, filter, opts...)
}

func (w *MongoCollectionWrapper) FindByID(ctx context.Context, id interface{}, opts ...*options.FindOneOptions) SingleResult {
	return w.FindOne(ctx, bson.M{"_id": id}, opts...)
}

func (w *MongoCollectionWrapper) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return w.Collection.InsertOne(ctx, document, opts...)
}

func (w *MongoCollectionWrapper) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return w.Collection.UpdateOne(ctx, filter, update, opts...)
}

func NewMongoCollectionWrapper(coll *mongo.Collection) MongoCollectionInterface {
	adapter := &MongoCollectionAdapter{coll: coll}
	return &MongoCollectionWrapper{Collection: adapter}
}
