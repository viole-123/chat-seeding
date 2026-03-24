package seeding

import (
	"context"
	"fmt"
	"strings"
	"time"
	"uniscore-seeding-bot/internal/adapter/redis"

	redispkg "github.com/redis/go-redis/v9"
)

// RoomManager stores a stable room mapping for each match in Redis.
type RoomManager struct {
	redisClient *redis.RedisClient
	ttl         time.Duration
}

func NewRoomManager(redisClient *redis.RedisClient, ttl time.Duration) *RoomManager {
	return &RoomManager{
		redisClient: redisClient,
		ttl:         ttl,
	}
}

func (r *RoomManager) GetOrCreate(ctx context.Context, matchID string) (string, error) {
	matchID = strings.TrimSpace(matchID)
	if matchID == "" {
		return "", fmt.Errorf("matchID is required")
	}

	key := fmt.Sprintf("room:%s", matchID)
	roomID, err := r.redisClient.Get(ctx, key)
	if err == nil && roomID != "" {
		return roomID, nil
	}
	if err != nil && err != redispkg.Nil {
		return "", fmt.Errorf("read room mapping failed: %w", err)
	}

	roomID = fmt.Sprintf("room-%s", matchID)
	if err := r.redisClient.Set(ctx, key, roomID); err != nil {
		return "", fmt.Errorf("save room mapping failed: %w", err)
	}
	return roomID, nil
}
func (r *RoomManager) SetRoomTTL(ctx context.Context, roomID string) error {
	key := fmt.Sprintf("room:ttl:%s", roomID)
	return r.redisClient.SetWithTTL(ctx, key, "active", r.ttl)
}

func (r *RoomManager) IsRoomActive(ctx context.Context, roomID string) (bool, error) {
	key := fmt.Sprintf("room:ttl:%s", roomID)
	_, err := r.redisClient.Get(ctx, key)

	return err == nil, err
}
