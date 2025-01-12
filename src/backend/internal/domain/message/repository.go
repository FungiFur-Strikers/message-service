package message

import "context"

type Repository interface {
	Create(ctx context.Context, msg *Message) error
	Delete(ctx context.Context, uid string) error
	Search(ctx context.Context, criteria SearchCriteria) ([]Message, error)
	FindByUID(ctx context.Context, uid string) (*Message, error)
}
