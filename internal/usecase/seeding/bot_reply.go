package seeding

import (
	"context"
	"fmt"
	"strings"
	"time"
	"uniscore-seeding-bot/internal/domain/model"
	"uniscore-seeding-bot/internal/domain/service"
)

// BotReplySystem generates two-way replies for user messages.
type BotReplySystem struct {
	intentDetector *IntentDetector
	llmGateway     service.LLMGatewayService
	personaLoader  *PersonaSelector
}

func NewBotReplySystem(intentDetector *IntentDetector, llmGateway service.LLMGatewayService, personaLoader *PersonaSelector) *BotReplySystem {
	return &BotReplySystem{
		intentDetector: intentDetector,
		llmGateway:     llmGateway,
		personaLoader:  personaLoader,
	}
}

func (s *BotReplySystem) GenerateReply(ctx context.Context, userMsg model.UserMessage, bundle model.ContextBundle) (*model.BotReply, error) {
	start := time.Now()

	intent, err := s.intentDetector.AnalyzeIntent(ctx, userMsg, bundle.Match)
	if err != nil {
		intent = &model.DetectIntent{
			Sentiment:     "neutral",
			Language:      "vi",
			TeamBias:      "none",
			MainTopic:     []string{"other"},
			RequiresReply: true,
		}
	}

	if !intent.RequiresReply {
		return &model.BotReply{
			Text:        "",
			ReplyType:   model.ReplyTypeSkip,
			Priority:    model.PriorityLow,
			Confidence:  1.0,
			Intent:      intent,
			GeneratedAt: time.Now(),
			LatencyMs:   time.Since(start).Milliseconds(),
		}, nil
	}

	if bundle.CurrentEvent.Type == "" {
		bundle.CurrentEvent.Type = mapTopicToEventType(intent.MainTopic)
	}
	if len(bundle.Match.Events) == 0 {
		bundle.Match.Events = []model.MatchEvent{bundle.CurrentEvent}
	}

	persona, err := s.personaLoader.SelectPersona(ctx, bundle)
	if err != nil || persona == nil {
		return nil, fmt.Errorf("select persona failed: %w", err)
	}

	resp, err := s.llmGateway.Generate(ctx, bundle, *persona)
	if err != nil || resp == nil || strings.TrimSpace(resp.Text) == "" {
		text := quickFallbackReply(intent, userMsg.Content)
		return &model.BotReply{
			Text:        text,
			PersonaID:   persona.ID,
			ReplyType:   model.ReplyTypeQuick,
			Priority:    calculatePriority(intent),
			Confidence:  0.65,
			Intent:      intent,
			Meta:        map[string]string{"source": "reply_fallback"},
			GeneratedAt: time.Now(),
			LatencyMs:   time.Since(start).Milliseconds(),
		}, nil
	}

	return &model.BotReply{
		Text:        resp.Text,
		PersonaID:   persona.ID,
		ReplyType:   model.ReplyTypeQuality,
		Priority:    calculatePriority(intent),
		Confidence:  0.82,
		Intent:      intent,
		Meta:        map[string]string{"source": "reply_llm", "language": resp.Language},
		GeneratedAt: time.Now(),
		LatencyMs:   time.Since(start).Milliseconds(),
	}, nil
}

func mapTopicToEventType(topics []string) string {
	if len(topics) == 0 {
		return "GOAL"
	}
	switch strings.ToLower(strings.TrimSpace(topics[0])) {
	case "goal", "score":
		return "GOAL"
	case "card", "red_card", "yellow_card":
		return "RED_CARD"
	case "substitution":
		return "SUBSTITUTION"
	default:
		return "GOAL"
	}
}

func calculatePriority(intent *model.DetectIntent) model.ReplyPriority {
	if intent == nil {
		return model.PriorityMedium
	}
	switch strings.ToLower(intent.Sentiment) {
	case "negative", "angry", "sad":
		return model.PriorityHigh
	case "positive", "excited":
		return model.PriorityMedium
	default:
		return model.PriorityLow
	}
}

func quickFallbackReply(intent *model.DetectIntent, content string) string {
	sentiment := "neutral"
	if intent != nil {
		sentiment = strings.ToLower(intent.Sentiment)
	}
	if strings.Contains(strings.ToLower(content), "thua") || sentiment == "negative" || sentiment == "sad" {
		return "Bình tĩnh nhé, trận vẫn còn diễn biến phía trước."
	}
	if strings.Contains(strings.ToLower(content), "goal") || sentiment == "positive" || sentiment == "excited" {
		return "Không khí đang lên cao, đúng chất bóng đá!"
	}
	return "Mình đang theo dõi cùng bạn, có gì hot mình cập nhật ngay."
}
