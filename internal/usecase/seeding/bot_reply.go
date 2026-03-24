// internal/usecase/seeding/bot_reply_system.go
package seeding

import (
	"context"
	"fmt"
	"strings"
	"time"
	"uniscore-seeding-bot/internal/domain/model"
	"uniscore-seeding-bot/internal/domain/service"
)

type BotReplySystem struct {
	intentDetector *IntentDetector
	llmGateway     service.LLMGatewayService
	personaLoader  *PersonaSelector
	rateLimiter    service.RateLimitService // FIX: thêm để check persona cooldown
}

func NewBotReplySystem(
	intentDetector *IntentDetector,
	llmGateway service.LLMGatewayService,
	personaLoader *PersonaSelector,
	rateLimiter service.RateLimitService,
) *BotReplySystem {
	return &BotReplySystem{
		intentDetector: intentDetector,
		llmGateway:     llmGateway,
		personaLoader:  personaLoader,
		rateLimiter:    rateLimiter,
	}
}

func (s *BotReplySystem) GenerateReply(ctx context.Context, userMsg model.UserMessage, bundle model.ContextBundle) (*model.BotReply, error) {
	start := time.Now()

	// Fast gate — tránh gọi LLM cho mọi tin nhắn
	if !shouldAttemptReply(userMsg.Content, bundle) {
		return skipReply(start), nil
	}

	// Detect intent — fallback về heuristic nếu LLM lỗi
	intent, err := s.intentDetector.AnalyzeIntent(ctx, userMsg, bundle.Match)
	if err != nil {
		intent = heuristicIntent(userMsg.Content)
	}

	if !intent.RequiresReply {
		return skipReply(start), nil
	}
	if shouldThrottleReply(intent, bundle) {
		return skipReply(start), nil
	}

	// Chọn persona
	persona, err := s.personaLoader.SelectPersona(ctx, bundle)
	if err != nil || persona == nil {
		return nil, fmt.Errorf("select persona failed: %w", err)
	}

	// FIX: check persona cooldown trước khi gọi LLM
	if s.rateLimiter != nil {
		allowed, rlErr := s.rateLimiter.CheckPersonaCooldown(ctx, persona.ID+"_reply", persona.Policy.CooldownSeconds)
		if rlErr != nil || !allowed {
			return skipReply(start), nil
		}
	}

	// Gọi LLM — fallback về quick reply nếu lỗi
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
			Meta:        map[string]string{"source": "fallback"},
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
		Meta:        map[string]string{"source": "llm", "language": resp.Language},
		GeneratedAt: time.Now(),
		LatencyMs:   time.Since(start).Milliseconds(),
	}, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func skipReply(start time.Time) *model.BotReply {
	return &model.BotReply{
		ReplyType:   model.ReplyTypeSkip,
		Priority:    model.PriorityLow,
		Confidence:  1.0,
		GeneratedAt: time.Now(),
		LatencyMs:   time.Since(start).Milliseconds(),
	}
}

// heuristicIntent — fallback khi LLM lỗi, dùng rule-based đơn giản
func heuristicIntent(content string) *model.DetectIntent {
	intent := &model.DetectIntent{
		Sentiment:     "neutral",
		Language:      "vi",
		TeamBias:      "none",
		MainTopic:     []string{"other"},
		RequiresReply: false,
	}
	if looksLikeQuestion(content) || hasDirectReplyCue(content) {
		intent.RequiresReply = true
	}
	lower := strings.ToLower(content)
	// FIX: check cả có dấu lẫn không dấu
	negativeKW := []string{"thua", "buồn", "tệ", "chán", "thất vọng", "bực"}
	for _, kw := range negativeKW {
		if strings.Contains(lower, kw) {
			intent.Sentiment = "negative"
			intent.RequiresReply = true
			break
		}
	}
	return intent
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
	lower := strings.ToLower(content)
	if strings.Contains(lower, "thua") || sentiment == "negative" || sentiment == "sad" {
		return "Bình tĩnh nhé, trận vẫn còn diễn biến phía trước."
	}
	if strings.Contains(lower, "goal") || strings.Contains(lower, "bàn thắng") || sentiment == "positive" {
		return "Không khí đang lên cao, đúng chất bóng đá!"
	}
	return "Mình đang theo dõi cùng bạn!"
}

func shouldAttemptReply(content string, bundle model.ContextBundle) bool {
	content = strings.TrimSpace(content)
	if len([]rune(content)) < 3 {
		return false
	}
	if looksLikeQuestion(content) || hasDirectReplyCue(content) {
		return true
	}
	if latestBotReplyAge(bundle) < 15*time.Second {
		return false
	}
	// FIX: thêm từ khóa có dấu tiếng Việt
	matchKeywords := []string{
		"goal", "bàn thắng", "ghi bàn", "thẻ đỏ", "thẻ vàng",
		"penalty", "phạt đền", "tỷ số", "score", "var",
	}
	lower := strings.ToLower(content)
	for _, kw := range matchKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func shouldThrottleReply(intent *model.DetectIntent, bundle model.ContextBundle) bool {
	age := latestBotReplyAge(bundle)
	if age < 0 {
		return false
	}
	if intent != nil && strings.ToLower(intent.Sentiment) == "negative" {
		return age < 6*time.Second
	}
	return age < 20*time.Second
}

func latestBotReplyAge(bundle model.ContextBundle) time.Duration {
	now := time.Now().Unix()
	var latest int64
	for _, msg := range bundle.Chat.RawMessages {
		if msg.IsBot && msg.Timestamp > latest {
			latest = msg.Timestamp
		}
	}
	if latest <= 0 {
		return -1
	}
	return time.Duration(now-latest) * time.Second
}

func looksLikeQuestion(content string) bool {
	lower := strings.ToLower(strings.TrimSpace(content))
	if strings.Contains(lower, "?") {
		return true
	}
	questionCues := []string{"sao vậy", "thế nào", "là gì", "vì sao", "tại sao", "why", "how", "what", "khi nào"}
	for _, cue := range questionCues {
		if strings.Contains(lower, cue) {
			return true
		}
	}
	return false
}

func hasDirectReplyCue(content string) bool {
	lower := strings.ToLower(strings.TrimSpace(content))
	// FIX: "oi" → "ơi" chính xác hơn, tránh match "toilet", "choice"...
	cues := []string{"bot ơi", "ad ơi", "admin ơi", "@bot", "hey bot"}
	for _, cue := range cues {
		if strings.Contains(lower, cue) {
			return true
		}
	}
	return false
}
