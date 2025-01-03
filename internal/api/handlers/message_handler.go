package handlers

import (
	"discord-message-service/internal/models"
	"discord-message-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MessageHandler メッセージ関連のHTTPリクエストを処理するハンドラー構造体
type MessageHandler struct {
	service *service.MessageService
}

// MessageHandler インスタンスを作成
func NewMessageHandler(service *service.MessageService) *MessageHandler {
	return &MessageHandler{service: service}
}

// CreateMessage 新しいメッセージを作成するためのハンドラー
func (h *MessageHandler) CreateMessage(c *gin.Context) {
	var input struct {
		Data models.Message `json:"data"`
	}

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// メッセージを作成
	if err := h.service.CreateMessage(&input.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "メッセージの作成に失敗しました"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "メッセージが正常に作成されました"})
}

// SearchMessages メッセージを検索するためのハンドラー
func (h *MessageHandler) SearchMessages(c *gin.Context) {
	var query struct {
		Keywords            string `form:"keywords"`
		KeywordSearchMethod string `form:"keywordSearchMethod"`
		Limit               int    `form:"limit"`
		ChannelID           string `form:"channelID"`
	}

	// クエリパラメータをバインド
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// デフォルト値の設定
	if query.Limit == 0 {
		query.Limit = 30
	}

	if query.KeywordSearchMethod == "" {
		query.KeywordSearchMethod = "and"
	}

	// メッセージを検索
	results, err := h.service.SearchMessages(map[string]interface{}{
		"keywords":            query.Keywords,
		"keywordSearchMethod": query.KeywordSearchMethod,
		"limit":               query.Limit,
		"channelID":           query.ChannelID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "メッセージの検索に失敗しました"})
		return
	}

	// マークダウン形式の結果を文字列として返す
	c.String(http.StatusOK, results)
}
