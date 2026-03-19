package service

import "uniscore-seeding-bot/internal/domain/model"

type PublisherService interface {
	Publish(msg model.ChatMessage) error
}
