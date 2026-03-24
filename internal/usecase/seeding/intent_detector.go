package seeding

import (
	"context"
	"uniscore-seeding-bot/internal/domain/model"
	"uniscore-seeding-bot/internal/domain/service"
)

type IntentDetector struct {
	llmGateway service.LLMGatewayService
}

func NewIntentDetector(llmGateway service.LLMGatewayService) *IntentDetector {
	return &IntentDetector{
		llmGateway: llmGateway,
	}
}

func (d *IntentDetector) AnalyzeIntent(ctx context.Context, msg model.UserMessage, matchCtx model.MatchState) (*model.DetectIntent, error) {
	// FIX: dùng positive/neutral/negative thay vì Sad|Excited|Angry
	// để khớp với sanitizeDetectIntent trong VLLMGateway
	systemPrompt := `Bạn là chuyên gia phân tích tâm lý bóng đá.
Phân tích tin nhắn người dùng và trả về JSON:
{"sentiment":"positive|neutral|negative","language":"vi|en","team_bias":"<tên đội hoặc none>","main_topic":"goal|card|score|stats|other","requires_reply":true|false}`

	return d.llmGateway.DetectUserIntent(ctx, systemPrompt, msg.Content, matchCtx)
}

func (d *IntentDetector) DetectUserEmotion(ctx context.Context, messages []model.ChatMessage) (string, error) {
	return d.llmGateway.AnalyzeSentiment(ctx, model.ContextBundle{
		Chat: model.ChatContext{RawMessages: messages},
	})
}
