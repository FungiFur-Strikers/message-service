package api

import (
	"discord-message-service/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, messageHandler *handlers.MessageHandler) {
	api := router.Group("/api")
	{
		api.POST("/messages", messageHandler.CreateMessage)
		api.GET("/messages/search", messageHandler.SearchMessages)
	}
}
