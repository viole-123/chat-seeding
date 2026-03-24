package redis

import (
	"context"
	"fmt"
	"time"
)

type DedupService struct {
	redisClient *RedisClient
	ttl         time.Duration
}

// Implement từ service DedupService, sử dụng Redis để lưu hash của event đã xử lý trong 1 khoảng thời gian nhất định (ttl) để tránh duplicate
func NewDedupService(redisClient *RedisClient, ttl time.Duration) *DedupService {
	return &DedupService{
		redisClient: redisClient,
		ttl:         ttl,
	}
}

// IsDuplicateEvent check xem event (matchID + minute + eventType) đã được xử lý chưa
func (d *DedupService) IsDuplicateEvent(ctx context.Context, matchID, minute, eventType string) (bool, error) {
	key := fmt.Sprintf("event:%s:%s:%s", matchID, minute, eventType)
	wasSet, err := d.redisClient.redis.SetNX(ctx, key, "1", d.ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis SETNX failed: %w", err)
	}
	return !wasSet, nil
}

func (d *DedupService) IsDuplicateMessage(ctx context.Context, matchID, msgHash string) (bool, error) {
	key := fmt.Sprintf("msg:dedup:%s:%s", matchID, msgHash)
	wasSet, err := d.redisClient.redis.SetNX(ctx, key, "1", d.ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis SETNX failed: %w", err)
	}
	return !wasSet, nil
}
