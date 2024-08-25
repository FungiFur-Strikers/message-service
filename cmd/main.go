package main

import (
	"discord-message-service/internal/api"
	"discord-message-service/internal/api/handlers"
	"discord-message-service/internal/config"
	"discord-message-service/internal/repository"
	"discord-message-service/internal/service"
	"discord-message-service/pkg/database"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repository
	messageRepo := repository.NewMessageRepository(db)

	// Initialize service
	messageService := service.NewMessageService(messageRepo)

	// Initialize handler
	messageHandler := handlers.NewMessageHandler(messageService)

	// Setup Gin router
	router := gin.Default()

	// Setup routes
	api.SetupRoutes(router, messageHandler)

	// Start server
	log.Printf("Starting server on %s", cfg.ServerAddress)
	if err := router.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
