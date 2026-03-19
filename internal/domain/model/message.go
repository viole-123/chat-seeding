package model

import "time"

type DraftMessage struct {
	Text      string            `json:"text"`
	MatchID   string            `json:"match_id"`
	EventType string            `json:"event_type"`
	PersonaID string            `json:"persona_id"`
	Meta      map[string]string `json:"meta"`
	CreatedAt time.Time         `json:"created_at"`
}

type ChatMessage struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Timestamp int64     `json:"timestamp"`
	IsBot     bool      `json:"is_bot"`
	Persona   string    `json:"persona"`
	MatchID   string    `json:"match_id"`
	RoomID    string    `json:"room_id"`
	EventType string    `json:"event_type,omitempty"` // GOAL, PREMATCH, etc.
	CreatedAt time.Time `json:"created_at"`
}
