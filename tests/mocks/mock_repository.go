package mocks

import "uniscore-seeding-bot/internal/domain/model"

type MockRepository struct{}

func (m *MockRepository) SaveMessage(msg model.ChatMessage) error { return nil }

func (m *MockRepository) GetMessageHistory(matchID string, limit int) ([]model.ChatMessage, error) {
	return []model.ChatMessage{}, nil
}
