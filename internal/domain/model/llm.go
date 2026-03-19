package model

type LLMResponse struct {
	Text      string   `json:"text"`
	Language  string   `json:"language"`
	StyleTags []string `json:"style_tags,omitempty"`
	RiskFlags []string `json:"risk_flags,omitempty"`
}
type LLMRequest struct {
	Model       string       `json:"model"`
	Messages    []LLMMessage `json:"messages"`
	Temperature float64      `json:"temperature,omitempty"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Stream      bool         `json:"stream,omitempty"`
}

type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type VLLMResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Index        int        `json:"index"`
	Message      LLMMessage `json:"message"`
	FinishReason string     `json:"finish_reason"`
}
