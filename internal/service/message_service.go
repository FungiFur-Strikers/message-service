package service

import (
	"discord-message-service/internal/models"
	"discord-message-service/internal/repository"
	"fmt"
	"strings"
)

// MessageService メッセージ関連の操作を行うサービス構造体
type MessageService struct {
	repo *repository.MessageRepository
}

// MessageService インスタンスを作成
func NewMessageService(repo *repository.MessageRepository) *MessageService {
	return &MessageService{repo: repo}
}

// CreateMessage 新しいメッセージを作成
func (s *MessageService) CreateMessage(message *models.Message) error {
	return s.repo.Create(message)
}

// SearchMessages クエリに基づいてメッセージを検索し、結果をマークダウン形式で返す
func (s *MessageService) SearchMessages(query map[string]interface{}) (string, error) {
	keywords := strings.Split(query["keywords"].(string), ",")
	limit := query["limit"].(int)
	channelID := query["channelID"].(string)
	searchMethod := query["keywordSearchMethod"].(string)

	var allMessages []models.Message
	var err error

	// キーワード検索方法に応じて検索を実行
	if searchMethod == "or" {
		// OR検索: 各キーワードで個別に検索し、結果を結合
		for _, keyword := range keywords {
			messages, err := s.repo.Search(strings.TrimSpace(keyword), limit, channelID)
			if err != nil {
				return "", err
			}
			allMessages = append(allMessages, messages...)
		}
	} else {
		// AND検索: すべてのキーワードを含むメッセージを検索
		allMessages, err = s.repo.Search(strings.Join(keywords, " "), limit, channelID)
		if err != nil {
			return "", err
		}
	}

	// メッセージをチャンネルごとにグループ化
	groupedMessages := make(map[string][]models.Message)
	for _, msg := range allMessages {
		groupedMessages[msg.ChannelID] = append(groupedMessages[msg.ChannelID], msg)
	}

	// 結果をマークダウン形式にフォーマット
	return formatMessagesAsMarkdown(groupedMessages), nil
}

// formatMessagesAsMarkdown グループ化されたメッセージをマークダウン形式の文字列に変換
func formatMessagesAsMarkdown(groupedMessages map[string][]models.Message) string {
	var result strings.Builder

	for channelID, messages := range groupedMessages {
		// チャンネルIDをヘッダーとして追加
		result.WriteString(fmt.Sprintf("# %s\n", channelID))

		for _, msg := range messages {
			// 各メッセージの送信日時と送信者を追加
			result.WriteString(fmt.Sprintf("------------- %s %s\n",
				msg.SentAt.Format("2006/01/02 15:04:05"),
				msg.Sender))
			// メッセージの内容を追加
			result.WriteString(fmt.Sprintf("%s\n\n", msg.Content))
		}

		result.WriteString("\n")
	}

	return result.String()
}
