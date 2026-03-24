package redis

import (
	"context"
	"fmt"
	"time"
	"uniscore-seeding-bot/internal/domain/service"

	"github.com/redis/go-redis/v9"
)

type personaStateServiceImpl struct {
	redisClient *RedisClient
}

func NewPersonaStateService(redisClient *RedisClient) service.PersonaStateService {
	return &personaStateServiceImpl{
		redisClient: redisClient,
	}
}

func (p *personaStateServiceImpl) IsOnCoolDown(ctx context.Context, personaID string) (bool, error) {
	key := fmt.Sprintf("cooldown:persona:%s", personaID)
	val, err := p.redisClient.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	return val != "", nil
}

func (p *personaStateServiceImpl) SetCoolDown(ctx context.Context, personaID string, durationSeconds int) error {
	key := fmt.Sprintf("cooldown:persona:%s", personaID)
	ttl := time.Duration(durationSeconds) * time.Second

	return p.redisClient.SetWithTTL(ctx, key, "1", ttl)
}

func (p *personaStateServiceImpl) IsAntiRepeat(ctx context.Context, matchId string, personaID string, msgHash string) (bool, error) {
	key := fmt.Sprintf("antirepeat:match:%s:persona:%s:msg:%s", matchId, personaID, msgHash)
	val, err := p.redisClient.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	return val != "", nil
}

func (p *personaStateServiceImpl) SetLastMessageHash(ctx context.Context, matchID, personaID, msgHash string, ttlSeconds int) error {
	key := fmt.Sprintf("antirepeat:%s:%s", matchID, personaID)
	ttl := time.Duration(ttlSeconds) * time.Second

	return p.redisClient.SetWithTTL(ctx, key, msgHash, ttl)
}

func (p *personaStateServiceImpl) SaveLastPersona(ctx context.Context, matchID, personaID string) error {
	key := fmt.Sprintf("lastpersona:%s", matchID)
	ttl := 24 * time.Hour // Lưu 24h

	return p.redisClient.SetWithTTL(ctx, key, personaID, ttl)
}

func (p *personaStateServiceImpl) GetLastPersona(ctx context.Context, matchID string) (string, error) {
	key := fmt.Sprintf("lastpersona:%s", matchID)

	val, err := p.redisClient.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return "", nil // Chưa có persona nào
		}
		return "", err
	}

	return val, nil
}
