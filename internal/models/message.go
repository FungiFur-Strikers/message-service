package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	SentAt    time.Time      `json:"sent_at"`
	Sender    string         `json:"sender"`
	ChannelID string         `json:"channel_id"`
	Content   string         `json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
