package service

import "context"

// RateLimitService enforces per-match rate limits.
type RateLimitService interface {
	CheckEventTypeLimit(ctx context.Context, matchID string, eventType string) (bool, error)
	CheckPersonaCooldown(ctx context.Context, personaID string, cooldownSeconds int) (bool, error)
	CheckMatchLimit(ctx context.Context, matchID string, maxBots int) (bool, error)
	IncrTotalMessages(ctx context.Context, matchID string) error
}
