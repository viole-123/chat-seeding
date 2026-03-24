package seeding

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"uniscore-seeding-bot/internal/domain/model"
	"uniscore-seeding-bot/internal/domain/service"
)

// ContextBuilder xây dựng ContextBundle từ nhiều nguồn dữ liệu khác nhau
type ContextBuilder struct {
	store          service.ContextStore
	intentDetector *IntentDetector
}

// Init ContextBuilder với ContextStore để lấy dữ liệu cần thiết
func NewContextBuilder(store service.ContextStore, intentDetector *IntentDetector) *ContextBuilder {
	return &ContextBuilder{
		store:          store,
		intentDetector: intentDetector,
	}
}

func (b *ContextBuilder) BuildBundle(ctx context.Context, matchID, roomID string) (*model.ContextBundle, error) {
	matchState, err := b.store.GetMatchState(matchID) // ok
	if err != nil {
		return nil, fmt.Errorf("get match state [%s]: %w", matchID, err)
	}
	if matchState.RoomID == "" {
		matchState.RoomID = roomID
	}

	recentEvents, _ := b.store.GetRecentEvents(matchID, 20)
	chatCtx, _ := b.store.GetRecentChatWindow(roomID, 20)
	//enrich chat context với các thông tin cần thiết để phân tích audience và intent
	chatCtx.LastMessageTime = getLastMessageTime(chatCtx.RawMessages)
	chatCtx.LastPersonaUsed = getLastPersona(chatCtx.RawMessages)
	chatCtx.LastMessageHashes = buildMessageHashes(chatCtx.RawMessages)
	if len(chatCtx.LastBotMessages) == 0 {
		chatCtx.LastBotMessages = chatCtx.LastMessageHashes
	}
	currentEvent := b.buildCurrentEvent(matchState, recentEvents)
	if currentEvent.Type != "" {
		matchState.Events = []model.MatchEvent{currentEvent}
	}

	bundle := model.ContextBundle{
		Match:        matchState,
		RecentEvents: recentEvents,
		Chat:         chatCtx,
		CurrentEvent: currentEvent,
	}
	bundle.Audience = b.analyzeAudience(ctx, bundle)

	log.Printf("✅ [ContextBuilder] bundle match=%s phase=%s recent_events=%d chat_msgs=%d current_event=%s",
		bundle.Match.MatchID,
		bundle.Match.Phase,
		len(bundle.RecentEvents),
		len(bundle.Chat.RawMessages),
		bundle.CurrentEvent.Type,
	)
	return &bundle, nil
}
func (b *ContextBuilder) buildCurrentEvent(matchState model.MatchState, events []model.CompactEvent) model.MatchEvent {
	if len(events) == 0 {
		return model.MatchEvent{}
	}
	latest := events[0]
	position := 0
	switch latest.TeamSide {
	case "home":
		position = 1
	case "away":
		position = 2
	}

	return model.MatchEvent{
		MatchID:    matchState.MatchID,
		LeagueID:   matchState.Competition.ID,
		Type:       latest.Type,
		Position:   position,
		Minute:     latest.Minute,
		AddTime:    latest.AddTime,
		HomeScore:  latest.HomeScore,
		AwayScore:  latest.AwayScore,
		PlayerName: latest.PlayerName,
		TeamSide:   latest.TeamSide,
		HomeTeam:   matchState.HomeTeam.ShortName,
		AwayTeam:   matchState.AwayTeam.ShortName,
		Timestamp:  normalizeEventTimestamp(matchState.UpdatedAt),
	}
}
func (b *ContextBuilder) analyzeAudience(ctx context.Context, bundle model.ContextBundle) model.AudienceSignal {
	rawMsgs := bundle.Chat.RawMessages
	dominantTeam := "none"
	if len(bundle.RecentEvents) > 0 {
		dominantTeam = bundle.RecentEvents[0].TeamSide
	}

	hotTopics := deriveHotTopics(bundle)
	velocity := 0.0
	if len(rawMsgs) > 0 {
		velocity = float64(len(rawMsgs)) / 5.0
	}
	if len(rawMsgs) == 0 || b.intentDetector == nil {
		return model.AudienceSignal{
			Sentiment:    model.SentimentNeutral,
			ChatVelocity: velocity,
			DominantTeam: dominantTeam,
			HotTopics:    hotTopics,
		}
	}
	sentimentStr, err := b.intentDetector.DetectUserEmotion(ctx, rawMsgs)
	if err != nil {
		log.Printf("⚠️  [ContextBuilder] AnalyzeSentiment failed: %v", err)
		sentimentStr = string(model.SentimentNeutral)
	}
	sentiment := model.SentimentNeutral
	switch sentimentStr {
	case "positive":
		sentiment = model.SentimentPositive
	case "negative":
		sentiment = model.SentimentNegative
	}
	return model.AudienceSignal{
		Sentiment:    sentiment,
		ChatVelocity: velocity,
		DominantTeam: dominantTeam,
		HotTopics:    hotTopics,
	}
}

func buildMessageHashes(msgs []model.ChatMessage) []string {
	hashes := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		key := msg.ID
		if key == "" {
			// fallback: dùng content làm key để dedup
			key = fmt.Sprintf("hash:%s", msg.Content)
		}
		hashes = append(hashes, key)
	}
	return hashes
}

func getLastMessageTime(msgs []model.ChatMessage) int64 {
	if len(msgs) == 0 {
		return 0
	}
	if msgs[0].Timestamp > 0 {
		return msgs[0].Timestamp
	}
	if !msgs[0].CreatedAt.IsZero() {
		return msgs[0].CreatedAt.Unix()
	}
	return 0
}

func getLastPersona(msgs []model.ChatMessage) string {
	if len(msgs) == 0 {
		return ""
	}
	return msgs[0].Persona
}

func normalizeEventTimestamp(ts int64) int64 {
	if ts > 0 {
		return ts
	}
	return time.Now().Unix()
}

func deriveHotTopics(bundle model.ContextBundle) []string {
	topics := make(map[string]int)

	for _, event := range bundle.RecentEvents {
		topics[event.Type]++
	}

	for _, msg := range bundle.Chat.RawMessages {
		topics["chat"]++
		content := strings.ToLower(msg.Content)
		// Tiếng Anh
		if strings.Contains(content, "goal") {
			topics["goal"]++
		}
		if strings.Contains(content, "red card") || strings.Contains(content, "redcard") {
			topics["red_card"]++
		}
		// FIX: thêm tiếng Việt
		if strings.Contains(content, "bàn thắng") || strings.Contains(content, "ghi bàn") || strings.Contains(content, "goallll") {
			topics["goal"]++
		}
		if strings.Contains(content, "thẻ đỏ") {
			topics["red_card"]++
		}
		if strings.Contains(content, "thẻ vàng") || strings.Contains(content, "yellow card") {
			topics["yellow_card"]++
		}
		if strings.Contains(content, "penalty") || strings.Contains(content, "phạt đền") {
			topics["penalty"]++
		}
		if strings.Contains(content, "var") {
			topics["var"]++
		}
	}
	result := make([]string, 0, len(topics))
	for topic := range topics {
		result = append(result, topic)
	}
	return result
}
