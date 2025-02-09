package repository

import (
	"context"
	"message-service/internal/domain/message"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageRepository struct {
	collection MongoCollectionInterface
}

func NewMessageRepository(db *mongo.Database) message.Repository {
	if db == nil {
		panic("database connection is required")
	}
	collection := db.Collection("messages")
	if collection == nil {
		panic("failed to get messages collection")
	}
	return &MessageRepository{
		collection: NewMongoCollectionWrapper(collection),
	}
}

func (r *MessageRepository) Create(ctx context.Context, msg *message.Message) error {
	now := time.Now()
	msg.CreatedAt = now
	msg.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, msg)
	return err
}

func (r *MessageRepository) Delete(ctx context.Context, uid string) error {
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

func (r *MessageRepository) Search(ctx context.Context, criteria message.SearchCriteria) ([]message.Message, error) {
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

	opts := options.Find().SetSort(bson.D{{Key: "sent_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
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

func (r *MessageRepository) FindByUID(ctx context.Context, uid string) (*message.Message, error) {
	filter := bson.M{"uid": uid, "deleted_at": nil}

	var msg message.Message
	if err := r.collection.FindOne(ctx, filter).Decode(&msg); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &msg, nil
}
