package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"uniscore-seeding-bot/internal/config"
)

// RedisClient wraps redis.Client for domain operations.
type RedisClient struct {
	redis *redis.Client
}

// NewRedisClient tạo Redis client và test connection ngay
func NewRedisClient(ctx context.Context, redisConfig config.RedisConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Username: redisConfig.Username,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &RedisClient{redis: rdb}, nil
}

// Get lấy giá trị theo key, trả về redis.Nil nếu key không tồn tại
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.redis.Get(ctx, key).Result()
}

// Set lưu key-value không có TTL (tồn tại vĩnh viễn)
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}) error {
	return r.redis.Set(ctx, key, value, 0).Err()
}

// SetWithTTL lưu key-value với TTL (tự xóa sau ttl)
func (r *RedisClient) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.redis.Set(ctx, key, value, ttl).Err()
}

// Close đóng kết nối Redis (gọi khi shutdown)
func (r *RedisClient) Close() error {
	return r.redis.Close()
}

func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.redis.Incr(ctx, key).Result()
}
func (r *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.redis.Expire(ctx, key, ttl).Err()
}
func (r *RedisClient) HGet(ctx context.Context, key string, field string) (string, error) {
	return r.redis.HGet(ctx, key, field).Result()
}
func (r *RedisClient) SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	return r.redis.SetNX(ctx, key, value, ttl).Result()
}
