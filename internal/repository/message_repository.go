package repository

import (
	"discord-message-service/internal/models"
	"sort"

	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(message *models.Message) error {
	return r.db.Create(message).Error
}

func (r *MessageRepository) Search(keyword string, limit int, channelID string) ([]models.Message, error) {
	query := r.db.Model(&models.Message{})

	if channelID != "" {
		query = query.Where("channel_id = ?", channelID)
	}

	if keyword != "" {
		query = query.Where("content LIKE ?", "%"+keyword+"%")
	}

	var messages []models.Message
	err := query.Order("sent_at desc").Limit(limit).Find(&messages).Error

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].SentAt.Before(messages[j].SentAt)
	})

	return messages, err
}
