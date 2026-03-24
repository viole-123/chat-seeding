package service

import (
	"context"
	"uniscore-seeding-bot/internal/config"
)

type AutoScalerService interface {
	GetStateBotPerRoom(ctx context.Context, roomID string) (config.ScalerState, int64, error)
	GetBotConfig(ctx context.Context, roomID string) (config.MaxBotsConfig, error)
}
