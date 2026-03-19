package safety

import (
	"context"
	"fmt"
	"time"
	"uniscore-seeding-bot/internal/adapter/redis"
)

// ShadowBanService enables ghost mode: process internally but skip publishing.
type ShadowBanService struct {
	redisClient *redis.RedisClient
}

func NewShadowBanService(redisClient *redis.RedisClient) *ShadowBanService {
	return &ShadowBanService{redisClient: redisClient}
}

func (s *ShadowBanService) Set(ctx context.Context, scope, id string, enabled bool) error {
	key := s.buildKey(scope, id)
	value := "0"
	if enabled {
		value = "1"
	}
	if err := s.redisClient.SetWithTTL(ctx, key, value, 7*24*time.Hour); err != nil {
		return fmt.Errorf("set shadow ban failed: %w", err)
	}
	return nil
}

func (s *ShadowBanService) IsEnabled(ctx context.Context, scope, id string) (bool, error) {
	key := s.buildKey(scope, id)
	val, err := s.redisClient.Get(ctx, key)
	if err != nil {
		if err.Error() == "redis: nil" {
			return false, nil
		}
		return false, err
	}
	return val == "1", nil
}

func (s *ShadowBanService) IsShadowed(ctx context.Context, matchID, leagueID string) (bool, error) {
	if global, err := s.IsEnabled(ctx, "global", ""); err != nil {
		return false, err
	} else if global {
		return true, nil
	}

	if leagueID != "" {
		if league, err := s.IsEnabled(ctx, "league", leagueID); err != nil {
			return false, err
		} else if league {
			return true, nil
		}
	}

	if matchID != "" {
		if match, err := s.IsEnabled(ctx, "match", matchID); err != nil {
			return false, err
		} else if match {
			return true, nil
		}
	}

	return false, nil
}

func (s *ShadowBanService) buildKey(scope, id string) string {
	if scope == "global" {
		return "shadowban:global"
	}
	return fmt.Sprintf("shadowban:%s:%s", scope, id)
}
