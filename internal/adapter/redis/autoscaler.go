package redis

import (
	"context"
	"fmt"
	"uniscore-seeding-bot/internal/config"

	goredis "github.com/redis/go-redis/v9"
)

type AutoScalerImpl struct {
	redis *goredis.Client
}

func NewAutoScalerImpl(redisClient *RedisClient) (*AutoScalerImpl, error) {
	return &AutoScalerImpl{redis: redisClient.redis}, nil
}

func (a *AutoScalerImpl) GetStateBotPerRoom(ctx context.Context, roomID string) (config.ScalerState, int64, error) {
	// Lấy số lượng user hiện tại từ Redis
	key := fmt.Sprintf("room:online:%s", roomID)
	count, err := a.redis.SCard(ctx, key).Result()
	if err != nil && err != goredis.Nil {
		return config.ScalerStateLow, 0, fmt.Errorf("redis SCARD %s: %w", key, err)
	}
	state := stateFromCount(count)
	return state, count, nil
}

func stateFromCount(count int64) config.ScalerState {
	switch {
	case count < 50:
		return config.ScalerStateLow
	case count < 100:
		return config.ScalerStateMedium
	case count < 500:
		return config.ScalerStateHigh
	case count >= 500:
		return config.ScalerStatePeak
	default:
		return config.ScalerStatePeak
	}
}

func (a *AutoScalerImpl) GetBotConfig(ctx context.Context, roomID string) (config.MaxBotsConfig, error) {
	state, count, err := a.GetStateBotPerRoom(ctx, roomID)
	if err != nil {
		// Fallback về low khi Redis lỗi — không block pipeline
		return lowConfig(), nil
	}

	cfg := configFromState(state)
	_ = count // có thể dùng để log
	return cfg, nil
}

func lowConfig() config.MaxBotsConfig {
	return config.MaxBotsConfig{MinBots: 2, MaxBots: 3, State: config.ScalerStateLow}
}
func configFromState(state config.ScalerState) config.MaxBotsConfig {
	switch state {
	case config.ScalerStateLow:
		return lowConfig()
	case config.ScalerStateMedium:
		return config.MaxBotsConfig{MinBots: 4, MaxBots: 6, State: state}
	case config.ScalerStateHigh:
		return config.MaxBotsConfig{MinBots: 5, MaxBots: 8, State: state}
	case config.ScalerStatePeak:
		return config.MaxBotsConfig{MinBots: 3, MaxBots: 5, State: state}
	default:
		return lowConfig()
	}
}
