package model

type ContextBundle struct {
	Match        MatchState     `json:"match"`
	RecentEvents []CompactEvent `json:"recent_events"`
	Chat         ChatContext    `json:"chat"`
	Audience     AudienceSignal `json:"audience"`
	CurrentEvent MatchEvent     `json:"current_event"`
}

type ChatContext struct {
	RoomID            string        `json:"room_id"`
	LastBotMessages   []string      `json:"last_bot_messages"`
	LastMessageHashes []string      `json:"last_message_hashes,omitempty"`
	LastMessageTime   int64         `json:"last_message_time"`
	LastPersonaUsed   string        `json:"last_persona_used"`
	RawMessages       []ChatMessage `json:"-"`
}

type AudienceSignal struct {
	Sentiment    Sentiment `json:"sentiment"`     // pos/neg/neutral (Stub cho MVP)
	ChatVelocity float64   `json:"chat_velocity"` // msg/min
	DominantTeam string    `json:"dominant_team"` // team_id or "none"
	HotTopics    []string  `json:"hot_topics"`    // ["goal", "red_card", ...] (Stub cho MVP)
}
