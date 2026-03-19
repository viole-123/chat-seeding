package vllm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"uniscore-seeding-bot/internal/domain/model"
)

type VLLMClient struct {
	apiURL     string
	httpClient *http.Client
	apiKey     string
	timeout    time.Duration
}

// Init VLLMClient with API URL and timeout
func NewVLLMClient(apiURL string, timeout time.Duration) *VLLMClient {
	return &VLLMClient{
		apiURL: strings.TrimSpace(apiURL),
		apiKey: strings.TrimSpace(os.Getenv("LLM_API_KEY")),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *VLLMClient) completionURL() string {
	base := strings.TrimRight(strings.TrimSpace(c.apiURL), "/")
	base = strings.ReplaceAll(base, "/v1/chatt/completions", "/v1/chat/completions")
	base = strings.ReplaceAll(base, "/v1/chatt", "/v1/chat")
	lower := strings.ToLower(base)

	if strings.HasSuffix(lower, "/v1/chat/completions") {
		return base
	}
	if strings.HasSuffix(lower, "/v1") {
		return base + "/chat/completions"
	}
	return base + "/v1/chat/completions"
}

// ham goi API LLM tra ve string
func (c *VLLMClient) Complete(ctx context.Context, req model.LLMRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request failed: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.completionURL(), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("vLLM API error: status=%d body=%s", resp.StatusCode, string(bodyBytes))
	}
	var llmResp model.VLLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return "", fmt.Errorf("Decode response failed :%w", err)
	}
	if len(llmResp.Choices) == 0 {
		return "", fmt.Errorf("vLLM returned no choices")
	}

	return llmResp.Choices[0].Message.Content, nil

}
