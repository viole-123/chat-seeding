package repository

import "uniscore-seeding-bot/internal/domain/model"

type MessageRepository interface {
	SaveMessage(m model.ChatMessage) error
	GetMessageHistory(matchID string, limit int) ([]model.ChatMessage, error)
}
