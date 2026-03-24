package model

type Template struct {
	ID         string                 `json:"id" db:"id"`
	Phase      MatchPhase             `json:"phase" db:"phase"`
	EventType  string                 `json:"event_type" db:"event_type"`
	Lang       string                 `json:"lang" db:"lang"`
	PersonaID  string                 `json:"persona_id" db:"persona_id"`
	Text       string                 `json:"text" db:"text"`
	Priority   int                    `json:"priority" db:"priority"`
	Enabled    bool                   `json:"enabled" db:"enabled"`
	Conditions map[string]interface{} `json:"conditions" db:"conditions"`
	CreatedAt  int64                  `json:"created_at" db:"created_at"`
	UpdatedAt  int64                  `json:"updated_at" db:"updated_at"`
}

type TemplateCondition struct {
	MinuteRange []int `json:"minute_range,omitempty"`
	ScoreDiff   []int `json:"score_diff,omitempty"`
}
