package vllm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
	"uniscore-seeding-bot/internal/domain/model"
)

type VLLMGateway struct {
	baseURL string
	client  *VLLMClient
	model   string
	timeout time.Duration
}

// Init VLLMGateway with API URL and model name
// func NewVLLMGateway(apiURL string, model string, timeout time.Duration) *VLLMGateway {
// 	return &VLLMGateway{
// 		baseURL: apiURL,
// 		model:   model,
// 		timeout: timeout,
// 		client:  NewVLLMClient(apiURL, timeout),
// 	}
// }

func NewVLLMGateway(apiURL string, model string, timeout time.Duration) *VLLMGateway {
	resolvedURL := strings.TrimSpace(apiURL)
	if resolvedURL == "" {
		resolvedURL = "http://localhost:11434"
	}
	resolvedModel := strings.TrimSpace(model)
	if resolvedModel == "" {
		resolvedModel = "qwen2.5:7b"
	}
	log.Printf("🌐 [VLLMGateway] model=%s endpoint=%s", resolvedModel, resolvedURL)

	return &VLLMGateway{
		baseURL: resolvedURL,
		model:   resolvedModel,
		timeout: timeout,
		client:  NewVLLMClient(resolvedURL, timeout),
	}
}

// Implement LLMGatewayService interface

func (g *VLLMGateway) Generate(ctx context.Context, bundle model.ContextBundle, persona model.Persona) (*model.LLMResponse, error) {
	systemPrompt := g.buildSystemPrompt(persona)

	userPrompt, err := g.buildUserPrompt(bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to build user prompt: %w", err)
	}
	log.Printf("🔍 [LLM] model=%s persona=%s event=%s", g.model, persona.ID, bundle.CurrentEvent.Type)
	req := model.LLMRequest{
		Model: g.model,
		Messages: []model.LLMMessage{{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.8,
		MaxTokens:   180,
		Stream:      false,
	}

	var rawResponse string
	for attempt := 1; attempt <= 2; attempt++ {
		resp, reqErr := g.client.Complete(ctx, req)
		if reqErr != nil {
			return nil, fmt.Errorf("❌ VLLM API Call Failed: %w", reqErr)
		}
		rawResponse = resp
		if strings.TrimSpace(rawResponse) != "" {
			break
		}
		log.Printf("⚠️  [LLM] empty response attempt=%d persona=%s event=%s", attempt, persona.ID, bundle.CurrentEvent.Type)
	}
	if strings.TrimSpace(rawResponse) == "" {
		return nil, fmt.Errorf("❌ vLLM returned empty response")
	}

	log.Printf("🤖 [LLM] Raw response: %.200s", rawResponse)

	llmOut, parseErr := parseLLMGenerateOutput(rawResponse)
	if parseErr != nil {
		return nil, fmt.Errorf("parse LLM response failed: %w (raw=%.100s)", parseErr, rawResponse)
	}

	if strings.TrimSpace(llmOut.Text) == "" {
		return nil, fmt.Errorf("❌ vLLM parsed empty text")
	}
	if strings.TrimSpace(llmOut.Language) == "" {
		llmOut.Language = "vi"
	}
	if len(llmOut.RiskFlags) > 0 {
		log.Printf("   ⚠️  [LLM] Risk flags detected: %v", llmOut.RiskFlags)
		return nil, fmt.Errorf("❌ unsafe content detected: %v", llmOut.RiskFlags)
	}
	// tuy chon
	log.Printf("   ✅ [LLM] Generated: %s", llmOut.Text)
	return llmOut, nil
}

func parseLLMGenerateOutput(raw string) (*model.LLMResponse, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("empty response")
	}

	var out model.LLMResponse
	if err := json.Unmarshal([]byte(raw), &out); err == nil && strings.TrimSpace(out.Text) != "" {
		return &out, nil
	}

	clean := normalizeLLMJSON(raw)
	if repaired, ok := tryRepairLLMJSON(clean); ok {
		clean = repaired
	}
	if err := json.Unmarshal([]byte(clean), &out); err == nil && strings.TrimSpace(out.Text) != "" {
		return &out, nil
	}

	// Fallback: accept plain text output from model instead of hard-failing.
	text := extractTextField(clean)
	if text == "" {
		text = raw
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("no usable text in response")
	}

	return &model.LLMResponse{
		Text:      text,
		Language:  "vi",
		StyleTags: []string{"llm_fallback"},
		RiskFlags: nil,
	}, nil
}

func tryRepairLLMJSON(raw string) (string, bool) {
	s := strings.TrimSpace(raw)
	if s == "" || !strings.Contains(s, "{") {
		return raw, false
	}
	open := strings.Count(s, "{")
	close := strings.Count(s, "}")
	for close < open {
		s += "}"
		close++
	}
	if s != raw {
		return s, true
	}
	return raw, false
}

func extractTextField(s string) string {
	lower := strings.ToLower(s)
	idx := strings.Index(lower, `"text"`)
	if idx < 0 {
		idx = strings.Index(lower, `"content"`)
	}
	if idx < 0 {
		return ""
	}
	segment := s[idx:]
	colon := strings.Index(segment, ":")
	if colon < 0 {
		return ""
	}
	segment = strings.TrimSpace(segment[colon+1:])
	if segment == "" {
		return ""
	}
	if segment[0] == '"' {
		segment = segment[1:]
		if end := strings.Index(segment, `"`); end >= 0 {
			return strings.ReplaceAll(segment[:end], `\n`, " ")
		}
		return strings.ReplaceAll(segment, `\n`, " ")
	}
	if end := strings.IndexAny(segment, ",}"); end >= 0 {
		return strings.TrimSpace(segment[:end])
	}
	return strings.TrimSpace(segment)
}

func (g *VLLMGateway) buildSystemPrompt(persona model.Persona) string {
	language := "vi"
	if len(persona.Profile.Language) > 0 {
		language = persona.Profile.Language[0]
	}
	return fmt.Sprintf(`You are a football fan chatting in a live match room.

Persona:
- ID: %s
- Tone: %s
- Language: %s
- Slang level: %s

Output rules:
- Write in %s language only
- Max 100 characters
- Sound like a real fan, NOT a bot
- Use emojis naturally (not excessively)
- No hate speech, no personal attacks

Return ONLY valid JSON, no markdown:
{"text":"<your message>","language":"%s","style_tags":["<tag>"],"risk_flags":[]}`,
		persona.ID,
		persona.Profile.Tone,
		language,
		persona.Profile.SlangLevel,
		language,
		language,
	)
}

func (g *VLLMGateway) buildUserPrompt(bundle model.ContextBundle) (string, error) {
	// Lấy event hiện tại
	var currentEvent interface{}
	if bundle.CurrentEvent.Type != "" {
		currentEvent = map[string]interface{}{
			"type":   bundle.CurrentEvent.Type,
			"minute": bundle.CurrentEvent.Minute,
			"player": bundle.CurrentEvent.PlayerName,
			"team":   bundle.CurrentEvent.TeamSide,
			"score":  fmt.Sprintf("%d-%d", bundle.CurrentEvent.HomeScore, bundle.CurrentEvent.AwayScore),
		}
	}

	// Lấy 3 tin nhắn gần nhất để tránh spam prompt quá dài
	recentChat := []string{}
	msgs := bundle.Chat.RawMessages
	if len(msgs) > 3 {
		msgs = msgs[:3]
	}
	for _, m := range msgs {
		recentChat = append(recentChat, m.Content)
	}

	payload, err := json.Marshal(map[string]interface{}{
		"match": map[string]interface{}{
			"home_team": bundle.Match.HomeTeam.ShortName,
			"away_team": bundle.Match.AwayTeam.ShortName,
			"score":     fmt.Sprintf("%d-%d", bundle.Match.HomeScore, bundle.Match.AwayScore),
			"minute":    bundle.Match.Minute,
			"phase":     string(bundle.Match.Phase),
		},
		"current_event":      currentEvent,
		"recent_chat":        recentChat,
		"audience_sentiment": string(bundle.Audience.Sentiment),
	})
	if err != nil {
		return "", fmt.Errorf("marshal context: %w", err)
	}

	return fmt.Sprintf("Generate a fan chat message for this football moment:\n\n%s", string(payload)), nil
}

func (g *VLLMGateway) DetectUserIntent(ctx context.Context, systemPrompt, userMsg string, matchCtx model.MatchState) (*model.DetectIntent, error) {
	matchJSON, _ := json.Marshal(matchCtx)

	// FIX: system prompt nhất quán với sanitizeDetectIntent (positive/neutral/negative)
	intentInstruction := `You are an intent analyzer for a football match chat.
Return ONLY valid JSON, no markdown, no code fences:

{
  "sentiment": "positive|neutral|negative",
  "team_bias": "<team_name or none>",
  "main_topic": "score|stats|lineup|goal|card|substitution|time|greeting|other",
  "requires_reply": true|false
}

Rules:
- sentiment: positive (happy/excited), negative (sad/angry), neutral
- team_bias: team the user supports/mentions, or "none"
- requires_reply: false only for random noise or very short messages`

	sys := fmt.Sprintf("%s\n\n%s", strings.TrimSpace(systemPrompt), intentInstruction)
	user := fmt.Sprintf("User message: %s\nMatch: %s", userMsg, string(matchJSON))

	req := model.LLMRequest{
		Model: g.model,
		Messages: []model.LLMMessage{
			{Role: "system", Content: sys},
			{Role: "user", Content: user},
		},
		Temperature: 0.0, // deterministic cho intent
		MaxTokens:   150,
		Stream:      false,
	}

	raw, err := g.client.Complete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("vLLM intent call failed: %w", err)
	}

	// Thử parse vLLM response wrapper trước
	if intent, ok, err := tryParseFromVLLMResponse(raw); err != nil {
		return nil, err
	} else if ok {
		log.Printf("✅ [Intent] sentiment=%s topic=%v bias=%s", intent.Sentiment, intent.MainTopic, intent.TeamBias)
		return intent, nil
	}

	// Fallback: parse trực tiếp
	clean := normalizeLLMJSON(raw)
	out, err := parseDetectIntentJSON(clean)
	if err != nil {
		return nil, fmt.Errorf("parse intent JSON failed: %w (raw=%.100s)", err, raw)
	}

	log.Printf("✅ [Intent] sentiment=%s topic=%v bias=%s", out.Sentiment, out.MainTopic, out.TeamBias)
	return &out, nil
}
func (g *VLLMGateway) AnalyzeSentiment(ctx context.Context, bundle model.ContextBundle) (string, error) {
	if len(bundle.Chat.RawMessages) == 0 {
		return string(model.SentimentNeutral), nil
	}

	msgs := bundle.Chat.RawMessages
	if len(msgs) > 8 {
		msgs = msgs[:8]
	}

	var sb strings.Builder
	for _, m := range msgs {
		sb.WriteString("- ")
		sb.WriteString(m.Content)
		sb.WriteByte('\n')
	}

	req := model.LLMRequest{
		Model: g.model,
		Messages: []model.LLMMessage{
			{
				Role:    "system",
				Content: "Classify the overall sentiment of these football chat messages. Return exactly one word: positive, neutral, or negative.",
			},
			{Role: "user", Content: sb.String()},
		},
		Temperature: 0.0,
		MaxTokens:   8,
		Stream:      false,
	}

	raw, err := g.client.Complete(ctx, req)
	if err != nil {
		return "", fmt.Errorf("vLLM sentiment call: %w", err)
	}

	out := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case strings.Contains(out, "positive"):
		return string(model.SentimentPositive), nil
	case strings.Contains(out, "negative"):
		return string(model.SentimentNegative), nil
	default:
		return string(model.SentimentNeutral), nil
	}
}

func tryParseFromVLLMResponse(raw string) (*model.DetectIntent, bool, error) {
	var vr model.VLLMRawResponse
	if err := json.Unmarshal([]byte(raw), &vr); err != nil {
		return nil, false, nil // not vLLM response format
	}
	if len(vr.Choices) == 0 {
		return nil, false, nil // likely plain intent JSON content
	}

	content := vr.Choices[0].Message.Content
	clean := normalizeLLMJSON(content)

	out, err := parseDetectIntentJSON(clean)
	if err != nil {
		return nil, true, fmt.Errorf("failed to parse DetectIntent from vLLM choice content: %w, content=%q", err, content)
	}
	return &out, true, nil
}

func parseDetectIntentJSON(raw string) (model.DetectIntent, error) {
	var out model.DetectIntent
	if err := json.Unmarshal([]byte(raw), &out); err == nil {
		sanitizeDetectIntent(&out)
		return out, nil
	}

	var compat struct {
		Sentiment     string      `json:"sentiment"`
		Language      string      `json:"language"`
		TeamBias      string      `json:"team_bias"`
		MainTopic     interface{} `json:"main_topic"`
		RequiresReply bool        `json:"requires_reply"`
	}
	if err := json.Unmarshal([]byte(raw), &compat); err != nil {
		return model.DetectIntent{}, err
	}

	out = model.DetectIntent{
		Sentiment:     compat.Sentiment,
		Language:      compat.Language,
		TeamBias:      compat.TeamBias,
		RequiresReply: compat.RequiresReply,
	}

	switch topic := compat.MainTopic.(type) {
	case string:
		if strings.TrimSpace(topic) != "" {
			out.MainTopic = []string{topic}
		}
	case []interface{}:
		for _, item := range topic {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out.MainTopic = append(out.MainTopic, s)
			}
		}
	}

	sanitizeDetectIntent(&out)
	return out, nil
}

func sanitizeDetectIntent(d *model.DetectIntent) {
	// FIX: normalize về positive/neutral/negative (khớp với IntentDetector system prompt)
	d.Sentiment = strings.ToLower(strings.TrimSpace(d.Sentiment))
	switch d.Sentiment {
	case "positive", "neutral", "negative":
		// ok
	case "excited", "happy":
		d.Sentiment = "positive"
	case "sad", "angry", "frustrated":
		d.Sentiment = "negative"
	default:
		d.Sentiment = "neutral"
	}

	d.TeamBias = strings.TrimSpace(d.TeamBias)
	if d.TeamBias == "" {
		d.TeamBias = "none"
	}

	if len(d.MainTopic) == 0 {
		d.MainTopic = []string{"other"}
		return
	}
	first := strings.ToLower(strings.TrimSpace(d.MainTopic[0]))
	switch first {
	case "score", "stats", "lineup", "goal", "card", "substitution", "time", "greeting", "other":
		d.MainTopic[0] = first
	default:
		d.MainTopic[0] = "other"
	}
	d.MainTopic = d.MainTopic[:1]
}

// normalizeLLMJSON removes ```json fences and trims extra text.
func normalizeLLMJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSpace(s)
		if strings.HasPrefix(strings.ToLower(s), "json") {
			s = s[4:] // bỏ "json"
		}
		s = strings.TrimSpace(s)
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		s = s[start : end+1]
	}
	return strings.TrimSpace(s)
}
