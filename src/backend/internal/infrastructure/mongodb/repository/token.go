// src\backend\internal\infrastructure\mongodb\repository\token.go
package repository

import (
	"context"
	"message-service/internal/domain/token"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TokenRepository struct {
	collection MongoCollectionInterface
}

func NewTokenRepository(db *mongo.Database) *TokenRepository {
	return &TokenRepository{
		collection: NewMongoCollectionWrapper(db.Collection("tokens")),
	}
}

func (r *TokenRepository) Create(ctx context.Context, token *token.Token) error {
	now := time.Now()
	token.CreatedAt = now
	token.UpdatedAt = now

	result, err := r.collection.InsertOne(ctx, token)
	if err != nil {
		return err
	}

	if insertedID, ok := result.InsertedID.(primitive.ObjectID); ok {
		token.ID = insertedID
	}
	return nil
}

func (r *TokenRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	now := time.Now()
	filter := bson.M{"_id": objectID, "deleted_at": nil}
	update := bson.M{"$set": bson.M{
		"deleted_at": now,
		"updated_at": now,
	}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *TokenRepository) List(ctx context.Context) ([]token.Token, error) {
	filter := bson.M{
		"deleted_at": nil,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []token.Token
	if err := cursor.All(ctx, &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *TokenRepository) FindByID(ctx context.Context, id string) (*token.Token, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id":        objectID,
		"deleted_at": nil,
	}

	var tkn token.Token
	if err := r.collection.FindOne(ctx, filter).Decode(&tkn); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &tkn, nil
}

func (r *TokenRepository) FindByToken(ctx context.Context, tokenString string) (*token.Token, error) {
	filter := bson.M{
		"token":      tokenString,
		"deleted_at": nil,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	var tkn token.Token
	if err := r.collection.FindOne(ctx, filter).Decode(&tkn); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &tkn, nil
}
