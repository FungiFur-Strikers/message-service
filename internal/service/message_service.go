package service

import (
	"discord-message-service/internal/models"
	"discord-message-service/internal/repository"
	"strings"
)

type MessageService struct {
	repo *repository.MessageRepository
}

func NewMessageService(repo *repository.MessageRepository) *MessageService {
	return &MessageService{repo: repo}
}

func (s *MessageService) CreateMessage(message *models.Message) error {
	return s.repo.Create(message)
}

func (s *MessageService) SearchMessages(query map[string]interface{}) (map[string][]models.Message, error) {
	keywords := strings.Split(query["keywords"].(string), ",")
	limit := query["limit"].(int)
	channelID := query["channelID"].(string)
	searchMethod := query["keywordSearchMethod"].(string)

	var allMessages []models.Message
	var err error

	if searchMethod == "or" {
		for _, keyword := range keywords {
			messages, err := s.repo.Search(strings.TrimSpace(keyword), limit, channelID)
			if err != nil {
				return nil, err
			}
			allMessages = append(allMessages, messages...)
		}
	} else {
		// "and" search method
		allMessages, err = s.repo.Search(strings.Join(keywords, " "), limit, channelID)
		if err != nil {
			return nil, err
		}
	}

	// Group messages by channel
	groupedMessages := make(map[string][]models.Message)
	for _, msg := range allMessages {
		groupedMessages[msg.ChannelID] = append(groupedMessages[msg.ChannelID], msg)
	}

	return groupedMessages, nil
}
