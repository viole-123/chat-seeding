package seeding

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"uniscore-seeding-bot/internal/domain/model"
	"uniscore-seeding-bot/internal/domain/service"
	"uniscore-seeding-bot/internal/observability/metrics"
	"uniscore-seeding-bot/internal/usecase/template"
)

// GenerateMessage picks template and renders content.

type MessageGenerator struct {
	templateLoader   *template.TemplateLoader
	templateRenderer *template.TemplateRenderer
	llmGateway       service.LLMGatewayService // Phase 2: LLM Gateway for fallback
}

func NewMessageGenerator(templateLoader *template.TemplateLoader, templateRenderer *template.TemplateRenderer, llmGateway service.LLMGatewayService) *MessageGenerator {
	return &MessageGenerator{
		templateLoader:   templateLoader,
		templateRenderer: templateRenderer,
		llmGateway:       llmGateway,
	}
}

func (g *MessageGenerator) GenerateMessage(bundle *model.ContextBundle, persona *model.Persona) (*model.DraftMessage, error) {
	normalizedBundle := g.normalizeBundleForGeneration(bundle)
	bundle = &normalizedBundle

	var (
		tmpl *model.Template
		err  error
	)
	if g.templateLoader != nil {
		tmpl, err = g.templateLoader.GetMatchingTemplates(bundle, persona)
		if err != nil {
			log.Printf("⚠️  [GENERATOR] Template search failed: %v", err)
		}
	}

	if tmpl != nil && err == nil {
		text, err := g.templateRenderer.Render(tmpl.Text, bundle)
		if err != nil {
			log.Printf("⚠️  [GENERATOR] Render failed: %v", err)
		} else {
			eventType := g.effectiveEventType(bundle)
			return &model.DraftMessage{
				Text:      g.withMatchTeams(bundle, text),
				MatchID:   bundle.Match.MatchID,
				EventType: eventType,
				PersonaID: persona.ID,
				Meta: map[string]string{
					"source":      "template",
					"template_id": tmpl.ID,
					"language":    tmpl.Lang,
				},
				CreatedAt: time.Now(),
			}, nil
		}
	}

	if g.llmGateway != nil {
		// Cloud models often need >2s; keep fallback patient enough to avoid false timeouts.
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		llmResp, err := g.llmGateway.Generate(ctx, *bundle, *persona)
		if err == nil && llmResp != nil {
			llmText := sanitizeLLMText(llmResp.Text)
			if isUsableLLMText(llmText) {
				metrics.Inc("llm_requests_total", map[string]string{"result": "success"})
				eventType := g.effectiveEventType(bundle)
				return &model.DraftMessage{
					Text:      g.withMatchTeams(bundle, llmText),
					MatchID:   bundle.Match.MatchID,
					EventType: eventType,
					PersonaID: persona.ID,
					Meta: map[string]string{
						"source":          "llm",
						"model":           "gpt-oss", // hardcode model name for now
						"language":        llmResp.Language,
						"tone":            persona.Profile.Tone,
						"template_missed": "true",
						"event_type":      eventType,
					},
					CreatedAt: time.Now(),
				}, nil
			}
			err = fmt.Errorf("llm output looks malformed: %q", llmText)
		}
		metrics.Inc("llm_requests_total", map[string]string{"result": "failure"})
		log.Printf("   ⚠️  [GENERATOR] LLM failed: %v", err)
	} else {
		log.Printf("   ⚠️  [GENERATOR] LLM gateway is nil, skip AI fallback")
	}

	fallbackText := g.getGenericFallbackTemplate(bundle)
	eventType := g.effectiveEventType(bundle)
	return &model.DraftMessage{
		Text:      g.withMatchTeams(bundle, fallbackText),
		MatchID:   bundle.Match.MatchID,
		EventType: eventType,
		PersonaID: persona.ID,
		Meta: map[string]string{
			"source": "fallback",
			"tone":   persona.Profile.Tone,
		},
		CreatedAt: time.Now(),
	}, nil
}

func (g *MessageGenerator) effectiveEventType(bundle *model.ContextBundle) string {
	eventType := normalizeEventType(bundle.CurrentEvent.Type)
	if len(bundle.Match.Events) > 0 {
		eventType = normalizeEventType(bundle.Match.Events[0].Type)
	}
	if eventType == "" {
		return "OTHER"
	}
	return eventType
}

func (g *MessageGenerator) normalizeBundleForGeneration(bundle *model.ContextBundle) model.ContextBundle {
	if bundle == nil {
		return model.ContextBundle{}
	}
	normalized := *bundle
	normalized.CurrentEvent.Type = normalizeEventType(normalized.CurrentEvent.Type)
	if len(normalized.Match.Events) > 0 {
		normalized.Match.Events[0].Type = normalizeEventType(normalized.Match.Events[0].Type)
	}
	return normalized
}

func normalizeEventType(raw string) string {
	v := strings.ToUpper(strings.TrimSpace(raw))
	if v == "" {
		return ""
	}
	v = strings.ReplaceAll(v, "-", "_")
	v = strings.ReplaceAll(v, " ", "_")
	if strings.Contains(v, "UPCOMING") {
		return "MATCH_UPCOMING"
	}
	return v
}

func sanitizeLLMText(text string) string {
	t := strings.TrimSpace(text)
	t = strings.Trim(t, "\"`")

	if strings.HasSuffix(t, "}") && !strings.Contains(t, "{") {
		t = strings.TrimSpace(strings.TrimSuffix(t, "}"))
	}

	if strings.HasPrefix(t, "{") && strings.HasSuffix(t, "}") {
		inner := extractTextFromLooseJSON(t)
		if strings.TrimSpace(inner) != "" {
			t = inner
		}
	}

	return strings.TrimSpace(t)
}

func extractTextFromLooseJSON(raw string) string {
	lower := strings.ToLower(raw)
	idx := strings.Index(lower, "\"text\"")
	if idx < 0 {
		return ""
	}
	part := raw[idx:]
	colon := strings.Index(part, ":")
	if colon < 0 {
		return ""
	}
	part = strings.TrimSpace(part[colon+1:])
	if part == "" {
		return ""
	}
	if part[0] != '"' {
		end := strings.IndexAny(part, ",}")
		if end < 0 {
			return strings.TrimSpace(part)
		}
		return strings.TrimSpace(part[:end])
	}
	part = part[1:]
	end := strings.Index(part, "\"")
	if end < 0 {
		return strings.TrimSpace(strings.ReplaceAll(part, "\\n", " "))
	}
	return strings.TrimSpace(strings.ReplaceAll(part[:end], "\\n", " "))
}

func (g *MessageGenerator) getGenericFallbackTemplate(bundle *model.ContextBundle) string {
	event := bundle.CurrentEvent
	if event.Type == "" {
		variants := []string{
			"Trận đấu đang diễn ra, anh em giữ nhịp bình luận nhé!",
			"Không khí trận này đang nóng dần lên rồi đó!",
			"Giữ vị trí đi, lát nữa chắc chắn có biến!",
		}
		return variants[len(bundle.Chat.RawMessages)%len(variants)]
	}

	switch event.Type {
	case "MATCH_UPCOMING":
		variants := []string{
			"Kèo này căng đó, chuẩn bị vào nhịp thôi!",
			"Trận sắp bắt đầu, anh em vào chỗ ngồi chưa?",
			"Không khí trước giờ bóng lăn đang nóng lên rồi!",
			"Đội hình hai bên có vẻ thú vị, chờ bóng lăn thôi!",
			"Sắp kickoff rồi, cùng xem đội nào mở điểm trước nhé!",
		}
		return variants[len(bundle.Chat.RawMessages)%len(variants)]
	case "GOAL":
		return fmt.Sprintf("⚽ GOAL! %s ghi bàn phút %d!", event.PlayerName, event.Minute)
	case "YELLOW_CARD":
		return fmt.Sprintf("🟨 Thẻ vàng cho %s phút %d", event.PlayerName, event.Minute)
	case "RED_CARD":
		return fmt.Sprintf("🟥 Thẻ đỏ! %s bị truất quyền thi đấu phút %d", event.PlayerName, event.Minute)
	case "SUBSTITUTION":
		return fmt.Sprintf("🔄 Thay người: %s vào sân thay %s", event.InPlayerName, event.OutPlayerName)
	default:
		return fmt.Sprintf("⚽ Sự kiện %s diễn ra phút %d", event.Type, event.Minute)
	}
}

func (g *MessageGenerator) withMatchTeams(bundle *model.ContextBundle, text string) string {
	home := strings.TrimSpace(bundle.Match.HomeTeam.ShortName)
	away := strings.TrimSpace(bundle.Match.AwayTeam.ShortName)

	if home == "" && away == "" {
		return text
	}
	if home == "" {
		return fmt.Sprintf("[%s] %s", away, text)
	}
	if away == "" {
		return fmt.Sprintf("[%s] %s", home, text)
	}
	return fmt.Sprintf("[%s vs %s] %s", home, away, text)
}

func FormatUnixTime(unixTime int) string {
	return time.Unix(int64(unixTime), 0).Format("15:04:05 2006-01-02")
}

func isUsableLLMText(text string) bool {
	t := sanitizeLLMText(text)
	if t == "" {
		return false
	}
	if len([]rune(t)) < 8 {
		return false
	}
	if t == "{" || t == "}" || t == "{\"" || t == "\"}" {
		return false
	}
	if strings.HasPrefix(t, "{") && !strings.Contains(t, " ") {
		return false
	}
	if strings.HasSuffix(t, "}") && len([]rune(t)) <= 10 {
		return false
	}
	if strings.HasPrefix(t, "{") && strings.HasSuffix(t, "}") {
		return false
	}

	letterCount := 0
	for _, r := range t {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r >= 128 {
			letterCount++
		}
	}
	return letterCount >= 6
}
