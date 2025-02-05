// src/backend/internal/domain/token/repository.go
package token

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, token *Token) error
	Delete(ctx context.Context, id string) error // ObjectID.Hex()の文字列を受け取る
	List(ctx context.Context) ([]Token, error)
	FindByID(ctx context.Context, id string) (*Token, error)
	FindByToken(ctx context.Context, tokenString string) (*Token, error)
}
