package repository

import (
	"context"
	"message-service/internal/domain/message"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) message.Repository {
	return &mongoRepository{
		collection: db.Collection("messages"),
	}
}

func (r *mongoRepository) Create(ctx context.Context, msg *message.Message) error {
	msg.CreatedAt = time.Now()
	msg.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, msg)
	return err
}

func (r *mongoRepository) Delete(ctx context.Context, uid string) error {
	now := time.Now()
	filter := bson.M{"uid": uid, "deleted_at": nil}
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

func (r *mongoRepository) Search(ctx context.Context, criteria message.SearchCriteria) ([]message.Message, error) {
	filter := bson.M{"deleted_at": nil}

	if criteria.ChannelID != nil {
		filter["channel_id"] = *criteria.ChannelID
	}
	if criteria.Sender != nil {
		filter["sender"] = *criteria.Sender
	}
	if criteria.FromDate != nil || criteria.ToDate != nil {
		dateFilter := bson.M{}
		if criteria.FromDate != nil {
			dateFilter["$gte"] = criteria.FromDate
		}
		if criteria.ToDate != nil {
			dateFilter["$lte"] = criteria.ToDate
		}
		filter["sent_at"] = dateFilter
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []message.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *mongoRepository) FindByUID(ctx context.Context, uid string) (*message.Message, error) {
	filter := bson.M{"uid": uid, "deleted_at": nil}

	var msg message.Message
	if err := r.collection.FindOne(ctx, filter).Decode(&msg); err != nil {
		return nil, err
	}

	return &msg, nil
}
