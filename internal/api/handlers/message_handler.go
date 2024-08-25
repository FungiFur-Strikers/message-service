package handlers

import (
	"discord-message-service/internal/models"
	"discord-message-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	service *service.MessageService
}

func NewMessageHandler(service *service.MessageService) *MessageHandler {
	return &MessageHandler{service: service}
}

func (h *MessageHandler) CreateMessage(c *gin.Context) {
	var input struct {
		Data models.Message `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateMessage(&input.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Message created successfully"})
}

func (h *MessageHandler) SearchMessages(c *gin.Context) {
	var query struct {
		Keywords            string `form:"keywords"`
		KeywordSearchMethod string `form:"keywordSearchMethod"`
		Limit               int    `form:"limit"`
		ChannelID           string `form:"channelID"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if query.Limit == 0 {
		query.Limit = 10
	}

	if query.KeywordSearchMethod == "" {
		query.KeywordSearchMethod = "and"
	}

	results, err := h.service.SearchMessages(map[string]interface{}{
		"keywords":            query.Keywords,
		"keywordSearchMethod": query.KeywordSearchMethod,
		"limit":               query.Limit,
		"channelID":           query.ChannelID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search messages"})
		return
	}

	c.JSON(http.StatusOK, results)
}
