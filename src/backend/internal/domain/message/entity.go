package message

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UID       string             `bson:"uid"`
	SentAt    time.Time          `bson:"sent_at"`
	Sender    string             `bson:"sender"`
	ChannelID string             `bson:"channel_id"`
	Content   string             `bson:"content"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty"`
}

type SearchCriteria struct {
	ChannelID *string
	Sender    *string
	FromDate  *time.Time
	ToDate    *time.Time
}
