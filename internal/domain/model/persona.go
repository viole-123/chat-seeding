package model

type Persona struct {
	ID      string         `yaml:"id"`
	Enabled bool           `yaml:"enabled"`
	Profile PersonaProfile `yaml:"profile"`
	Policy  PersonaPolicy  `yaml:"rules"`
}

type PersonaProfile struct {
	Name           string   `yaml:"name"`
	Language       []string `yaml:"language"` // vi/en
	Tone           string   `yaml:"tone"`     // hype/calm/analyst/funny (Dựa theo 4 nhân vật [2])
	EmojiRate      float64  `yaml:"emoji_rate"`
	SentenceLength string   `yaml:"sentence_length"` // short/medium/long
	SlangLevel     string   `yaml:"slang_level"`
	SarcasmLevel   string   `yaml:"sarcasm_level"`
	TeamBias       string   `yaml:"team_bias"` // neutral/home/away
	SeedPhrases    []string `yaml:"seed_phrases"`
	Description    string   `yaml:"description"`
}

type PersonaPolicy struct {
	WeightBase        int            `yaml:"weight_base"`
	AllowedEventTypes []string       `yaml:"allowed_event_types"` // VD: ["GOAL", "RED_CARD"]
	BlockedEventTypes []string       `yaml:"blocked_event_types"`
	CooldownSeconds   int            `yaml:"cooldown_seconds"` // VD: 30-120s
	EventWeight       map[string]int `yaml:"event_weight"`
}

// PersonaScore holds a persona with its calculated score for selection
type PersonaScore struct {
	Persona Persona
	Score   int
}
