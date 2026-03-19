package seeding

import (
	"context"
	"testing"
	"uniscore-seeding-bot/internal/domain/model"
)

type fakeLLMGateway struct{}

func (f *fakeLLMGateway) Generate(ctx context.Context, bundle model.ContextBundle, persona model.Persona) (*model.LLMResponse, error) {
	return &model.LLMResponse{Text: "ok", Language: "vi"}, nil
}

func (f *fakeLLMGateway) DetectUserIntent(ctx context.Context, systemPrompt string, userMsg string, matchCtx model.MatchState) (*model.DetectIntent, error) {
	return &model.DetectIntent{Sentiment: "neutral", MainTopic: []string{"other"}, RequiresReply: false}, nil
}

func (f *fakeLLMGateway) AnalyzeSentiment(ctx context.Context, bundle model.ContextBundle) (string, error) {
	return "neutral", nil
}

func TestBotReplySystemSkipWhenNoReplyRequired(t *testing.T) {
	replySystem := NewBotReplySystem(NewIntentDetector(&fakeLLMGateway{}), &fakeLLMGateway{}, &PersonaSelector{})

	reply, err := replySystem.GenerateReply(context.Background(), model.UserMessage{Content: "ok"}, model.ContextBundle{Match: model.MatchState{MatchID: "m1"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply == nil || reply.ReplyType != model.ReplyTypeSkip {
		t.Fatalf("expected skip reply, got %#v", reply)
	}
}
