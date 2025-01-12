package main

import (
	"context"
	"log"
	"message-service/internal/adapter/handler"
	"message-service/internal/adapter/repository"
	"message-service/pkg/api"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// MongoDB接続設定
	mongoURI := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_NAME")

	opts := options.Client().
		ApplyURI(mongoURI).
		SetTimeout(10 * time.Second).
		SetServerSelectionTimeout(5 * time.Second)

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}

	db := client.Database(dbName)

	// 依存関係の構築
	repo := repository.NewMongoRepository(db)
	handler := handler.NewMessageHandler(repo)

	// Ginルーターの設定
	router := gin.Default()
	api.RegisterHandlers(router, api.NewStrictHandler(handler, nil))

	// サーバー起動
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
