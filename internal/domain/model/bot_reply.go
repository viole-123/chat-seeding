package model

import "time"

type ReplyType string

const (
	ReplyTypeQuick   ReplyType = "quick"   // Template-based, < 500ms
	ReplyTypeQuality ReplyType = "quality" // LLM-based, 1-3s
	ReplyTypeSkip    ReplyType = "skip"    // Không reply
)

type ReplyPriority int

const (
	PriorityLow    ReplyPriority = 1  // Greeting, casual chat
	PriorityMedium ReplyPriority = 5  // Question about match
	PriorityHigh   ReplyPriority = 8  // Negative sentiment,需要 defuse
	PriorityUrgent ReplyPriority = 10 // Toxic, spam
)

type BotReply struct {
	Text        string            `json:"text"`
	PersonaID   string            `json:"persona_id"`
	ReplyType   ReplyType         `json:"reply_type"`
	Priority    ReplyPriority     `json:"priority"`
	Confidence  float64           `json:"confidence"` // 0.0-1.0
	Intent      *DetectIntent     `json:"intent"`
	Meta        map[string]string `json:"meta"`
	GeneratedAt time.Time         `json:"generated_at"`
	LatencyMs   int64             `json:"latency_ms"`
}

type UserSentimentHistory struct {
	UserID         string    `json:"user_id"`
	MatchID        string    `json:"match_id"`
	RecentMessages []string  `json:"recent_messages"` // Last 5 messages
	Sentiment      string    `json:"sentiment"`       // positive/neutral/negative
	TeamBias       string    `json:"team_bias"`       // team_name hoặc "none"
	ToxicityScore  float64   `json:"toxicity_score"`  // 0.0-1.0
	LastUpdated    time.Time `json:"last_updated"`
}

type ReplyTemplate struct {
	ID            string            `json:"id"`
	IntentPattern string            `json:"intent_pattern"` // "greeting", "question_time", "celebration"
	Persona       string            `json:"persona"`
	Templates     []string          `json:"templates"` // Multiple variants
	Conditions    map[string]string `json:"conditions"`
	Priority      ReplyPriority     `json:"priority"`
}
