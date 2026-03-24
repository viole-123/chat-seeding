package redis

import (
	"context"
	"fmt"
	"time"
)

type QualityStateImpl struct {
	redis *RedisClient
}

// Initialize QualityStateService with Redis client
func NewQualityStateService(redis *RedisClient) *QualityStateImpl {
	return &QualityStateImpl{
		redis: redis,
	}
}

func (q *QualityStateImpl) IsMessageDuplicated(matchID string, msgHash string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("msg:hash:%s:%s", matchID, msgHash)
	exist, err := q.redis.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return exist > 0, nil
}

func (q *QualityStateImpl) SaveMessageHash(matchID string, msgHash string, ttlSeconds int) error {
	ctx := context.Background()
	key := fmt.Sprintf("msg:hash:%s:%s", matchID, msgHash)
	ttl := time.Duration(ttlSeconds) * time.Second
	if err := q.redis.redis.Set(ctx, key, "1", ttl).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}
