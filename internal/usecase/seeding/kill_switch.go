package seeding

import (
	"context"
	"fmt"
	"time"
	"uniscore-seeding-bot/internal/adapter/redis"
)

type KillSwitchService struct {
	redisClient *redis.RedisClient
}

func NewKillSwitchService(redisClient *redis.RedisClient) *KillSwitchService {
	return &KillSwitchService{
		redisClient: redisClient,
	}
}

func (k *KillSwitchService) SetKillSwitch(ctx context.Context, scope, id string, isKilled bool) error {
	key := k.buildKey(scope, id)
	value := "0"
	if isKilled {
		value = "1"
	}
	if err := k.redisClient.SetWithTTL(ctx, key, value, 7*24*time.Hour); err != nil {
		return fmt.Errorf("Set kill switch failed: %w", err)
	}
	return nil
}

func (k *KillSwitchService) IsKilled(ctx context.Context, scope, id string) (bool, error) {
	if killed, err := k.checkKey(ctx, "killswitch:global"); err != nil {
		return false, fmt.Errorf("check global kill switch failed: %w", err)
	} else if killed {
		return true, nil
	}
	switch scope {
	case "global":
		return false, nil
	case "league":
		if id == "" {
			return false, fmt.Errorf("league id is required")
		}
		return k.checkKey(ctx, k.buildKey("league", id))
	case "match":
		if id == "" {
			return false, fmt.Errorf("match id is required")
		}
		return k.checkKey(ctx, k.buildKey("match", id))
	default:
		return false, fmt.Errorf("invalid scope: %s", scope)
	}
}

func (k *KillSwitchService) buildKey(scope, id string) string {
	if scope == "global" {
		return "killswitch:global"
	}
	return fmt.Sprintf("killswitch:%s:%s", scope, id)
}

func (k *KillSwitchService) checkKey(ctx context.Context, key string) (bool, error) {
	val, err := k.redisClient.Get(ctx, key)
	if err != nil {
		if err.Error() == "redis: nil" {
			return false, nil
		}
		return false, err
	}
	return val == "1", nil
}
