package seeding

import (
	"context"
	"fmt"
	"strings"
	"uniscore-seeding-bot/internal/adapter/redis"

	redispkg "github.com/redis/go-redis/v9"
)

// RoomManager stores a stable room mapping for each match in Redis.
type RoomManager struct {
	redisClient *redis.RedisClient
}

func NewRoomManager(redisClient *redis.RedisClient) *RoomManager {
	return &RoomManager{redisClient: redisClient}
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
