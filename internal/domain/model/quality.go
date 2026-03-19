package model

type QualityResult struct {
	IsPass bool     `json:"is_pass"`
	Reason []string `json:"reason"`
	Action []string `json:"action"`
}

type QualityCheckConfig struct {
	MinLength   int      `json:"min_length"`
	MaxLength   int      `json:"max_length"`
	BannedWords []string `yaml:"banned_words"`
	DedupTTL    int      `yaml:"dedup_ttl"`
}
