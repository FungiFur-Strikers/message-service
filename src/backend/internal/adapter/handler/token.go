package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"message-service/internal/domain/token"
	"message-service/pkg/api"
	"time"
)

type TokenHandler struct {
	repo token.Repository
}

// テスト用にrand.Readをモック可能にする
var randRead = rand.Read

func NewTokenHandler(repo token.Repository) *TokenHandler {
	return &TokenHandler{repo: repo}
}

func (h *TokenHandler) PostApiTokens(ctx context.Context, req api.PostApiTokensRequestObject) (api.PostApiTokensResponseObject, error) {
	// トークンの生成
	tokenBytes := make([]byte, 32)
	if _, err := randRead(tokenBytes); err != nil {
		return api.PostApiTokens400Response{}, err
	}
	tokenString := base64.URLEncoding.EncodeToString(tokenBytes)

	// 有効期限の設定
	expiresIn := 2592000 // デフォルト30日
	if req.Body.ExpiresIn != nil {
		expiresIn = *req.Body.ExpiresIn
	}
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)

	tkn := &token.Token{
		Token:     tokenString,
		Name:      req.Body.Name,
		ExpiresAt: expiresAt,
	}

	if err := h.repo.Create(ctx, tkn); err != nil {
		return api.PostApiTokens400Response{}, err
	}

	// IDをstring型に変換
	idStr := tkn.ID.Hex()
	return api.PostApiTokens201JSONResponse{
		Id:        &idStr,
		Token:     &tkn.Token,
		Name:      &tkn.Name,
		ExpiresAt: &tkn.ExpiresAt,
	}, nil
}

func (h *TokenHandler) GetApiTokens(ctx context.Context, req api.GetApiTokensRequestObject) (api.GetApiTokensResponseObject, error) {
	tokens, err := h.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	response := make(api.GetApiTokens200JSONResponse, len(tokens))
	for i, tkn := range tokens {
		idStr := tkn.ID.Hex()
		response[i] = api.Token{
			Id:        &idStr,
			Token:     &tkn.Token,
			Name:      &tkn.Name,
			ExpiresAt: &tkn.ExpiresAt,
			CreatedAt: &tkn.CreatedAt,
			UpdatedAt: &tkn.UpdatedAt,
		}
	}

	return response, nil
}

func (h *TokenHandler) DeleteApiTokensId(ctx context.Context, req api.DeleteApiTokensIdRequestObject) (api.DeleteApiTokensIdResponseObject, error) {
	if err := h.repo.Delete(ctx, req.Id); err != nil {
		return api.DeleteApiTokensId404Response{}, err
	}
	return api.DeleteApiTokensId204Response{}, nil
}
