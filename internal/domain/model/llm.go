// internal/domain/model/llm.go
package model

// LLMRequest — format chuẩn OpenAI-compatible (Ollama, vLLM đều dùng được)
type LLMRequest struct {
	Model       string       `json:"model"`
	Messages    []LLMMessage `json:"messages"`
	Temperature float64      `json:"temperature,omitempty"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Stream      bool         `json:"stream,omitempty"`
}

type LLMMessage struct {
	Role    string `json:"role"` // "system" | "user" | "assistant"
	Content string `json:"content"`
}

// LLMResponse — output từ VLLMGateway.Generate()
// Bot dùng Text để gửi vào phòng chat
type LLMResponse struct {
	Text      string   `json:"text"`
	Language  string   `json:"language"`
	StyleTags []string `json:"style_tags,omitempty"`
	RiskFlags []string `json:"risk_flags,omitempty"` // nếu có → reject message
}

// VLLMRawResponse — raw format trả về từ Ollama/OpenAI API
// VLLMClient parse cái này, không expose ra ngoài
type VLLMRawResponse struct {
	Choices []struct {
		Message      LLMMessage `json:"message"`
		FinishReason string     `json:"finish_reason"`
	} `json:"choices"`
}
