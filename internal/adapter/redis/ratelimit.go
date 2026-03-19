package redis

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimit struct {
	redisClient *RedisClient
}

func NewRateLimit(redisClient *RedisClient) *RateLimit {
	return &RateLimit{
		redisClient: redisClient,
	}
}

func (r *RateLimit) CheckEventTypeLimit(ctx context.Context, matchID string, eventType string) (bool, error) {
	limit := r.getEventTypeLimit(eventType)
	key := fmt.Sprintf("ratelimit:eventtype:%s,%s", matchID, eventType)
	counter, err := r.getCounter(ctx, key)
	if err != nil {
		return false, err
	}

	if counter >= int64(limit) {
		log.Printf("⏭️  [RATELIMIT] event-type blocked: match=%s type=%s used=%d limit=%d", matchID, eventType, counter, limit)
		return false, nil
	}

	if err := r.incrementCounter(ctx, key, 24*time.Hour); err != nil {
		return false, err
	}
	log.Printf("✅ [RATELIMIT] event-type pass: match=%s type=%s used=%d next=%d limit=%d", matchID, eventType, counter, counter+1, limit)

	return true, nil
}

func (r *RateLimit) CheckPersonaCooldown(ctx context.Context, personaID string, cooldownSeconds int) (bool, error) {
	key := fmt.Sprintf("cooldown:persona:%s", personaID)
	ttl := time.Duration(cooldownSeconds) * time.Second

	ok, err := r.redisClient.SetNX(ctx, key, "1", ttl)
	if err != nil {
		return false, fmt.Errorf("persona cooldown check: %w", err)
	}
	if !ok {
		log.Printf("⏭️  [RATELIMIT-T2] persona cooldown blocked: persona=%s", personaID)
		return false, nil
	}
	log.Printf("✅ [RATELIMIT-T2] persona pass: persona=%s cooldown=%ds", personaID, cooldownSeconds)
	return true, nil
}

func (r *RateLimit) CheckMatchLimit(ctx context.Context, matchID string, maxBots int) (bool, error) {
	if maxBots <= 0 {
		maxBots = 100
	}

	// `ratelimit:total_msgs:*` is currently incremented for bot messages, so use it as bot counter.
	key := fmt.Sprintf("ratelimit:total_msgs:%s", matchID)
	botCount, err := r.getCounter(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to get bot message count: %w", err)
	}

	if botCount >= int64(maxBots) {
		log.Printf("⏭️  [RATELIMIT-T3] match limit blocked: match=%s bots=%d limit=%d", matchID, botCount, maxBots)
		return false, nil
	}

	log.Printf("✅ [RATELIMIT-T3] match limit pass: match=%s bots=%d next=%d limit=%d", matchID, botCount, botCount+1, maxBots)
	return true, nil
}

func (r *RateLimit) getEventTypeLimit(eventType string) int {
	switch eventType {
	case "GOAL":
		return 3
	case "RED_CARD":
		return 1
	case "PENALTY":
		return 1
	case "PENALTY_MISSED":
		return 1
	case "SUBSTITUTION":
		return 2
	case "YELLOW_CARD":
		return 1
	default:
		return 2
	}

}

func (r *RateLimit) getCounter(ctx context.Context, key string) (int64, error) {
	val, err := r.redisClient.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("get counter %s: %w", key, err)
	}
	// FIX: dùng strconv thay fmt.Sscanf để bắt lỗi parse đúng cách
	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse counter %q: %w", val, err)
	}
	return count, nil
}
func (r *RateLimit) incrementCounter(ctx context.Context, key string, ttl time.Duration) error {
	count, err := r.redisClient.Incr(ctx, key)
	if err != nil {
		return err
	}

	if count == 1 {
		if err := r.redisClient.Expire(ctx, key, ttl); err != nil {
			fmt.Printf("warning: set ttl failed: %v\n", err)
		}
	}
	return nil
}

func (r *RateLimit) IncrTotalMessages(ctx context.Context, matchID string) error {
	key := fmt.Sprintf("ratelimit:total_msgs:%s", matchID)
	return r.incrementCounter(ctx, key, 24*time.Hour)
}
