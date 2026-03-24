package service

import (
	"context"
	"uniscore-seeding-bot/internal/domain/model"
)

type ContextStore interface {
	GetMatchState(matchId string) (model.MatchState, error)
	SetMatchState(matchID string, matchCtx model.MatchState) error

	// GetRecentEvents(matchId string, limit int) ([]model.MatchEvent, error)
	GetRecentEvents(matchId string, limit int) ([]model.CompactEvent, error)

	GetRecentChatWindow(matchId string, limit int) (model.ChatContext, error)

	GetMatchByID(ctx context.Context, matchID string) (*model.MatchDailyCatchFromRedis, error)
	GetAllTodayMatches(ctx context.Context) ([]*model.MatchDailyCatchFromRedis, error)

	HasSentPrematch(ctx context.Context, matchID string) (bool, error)

	PushEvent(matchID string, event model.MatchEvent) error
	PushChatMessage(roomID string, msg model.ChatMessage) error
	MarkSentPrematch(ctx context.Context, matchID string) error

	GetBotCount(ctx context.Context, roomID string) (int64, error)
}
