package service

import (
	"context"
	"uniscore-seeding-bot/internal/domain/model"
)

type LLMGatewayService interface {
	Generate(ctx context.Context, bundle model.ContextBundle, persona model.Persona) (*model.LLMResponse, error)
	DetectUserIntent(ctx context.Context, systemPrompt string, userMsg string, matchCtx model.MatchState) (*model.DetectIntent, error)
	AnalyzeSentiment(ctx context.Context, bundle model.ContextBundle) (string, error)
}
