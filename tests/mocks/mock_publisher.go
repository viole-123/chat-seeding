package mocks

import "uniscore-seeding-bot/internal/domain/model"

type MockPublisher struct{}

func (m *MockPublisher) Publish(msg model.ChatMessage) error { return nil }
