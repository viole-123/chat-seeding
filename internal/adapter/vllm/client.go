package vllm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"uniscore-seeding-bot/internal/domain/model"
)

type VLLMClient struct {
	apiURL     string
	httpClient *http.Client
	apiKey     string
	timeout    time.Duration
	limiter    chan struct{}
}

// Init VLLMClient with API URL and timeout
func NewVLLMClient(apiURL string, timeout time.Duration) *VLLMClient {
	maxConcurrency := 2
	if raw := strings.TrimSpace(os.Getenv("LLM_MAX_CONCURRENCY")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			maxConcurrency = parsed
		}
	}

	return &VLLMClient{
		apiURL: strings.TrimSpace(apiURL),
		apiKey: strings.TrimSpace(os.Getenv("LLM_API_KEY")),
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
		limiter: make(chan struct{}, maxConcurrency),
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
	if err := c.acquire(ctx); err != nil {
		return "", err
	}
	defer c.release()

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request failed: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
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
			lastErr = fmt.Errorf("http request failed: %w", err)
			if attempt < 3 {
				if sleepErr := sleepWithContext(ctx, time.Duration(attempt)*300*time.Millisecond); sleepErr != nil {
					return "", sleepErr
				}
				continue
			}
			return "", lastErr
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("vLLM API error: status=%d body=%s", resp.StatusCode, string(bodyBytes))
			if shouldRetryStatus(resp.StatusCode) && attempt < 3 {
				retryDelay := retryAfterOrDefault(resp.Header.Get("Retry-After"), attempt)
				if sleepErr := sleepWithContext(ctx, retryDelay); sleepErr != nil {
					return "", sleepErr
				}
				continue
			}
			return "", lastErr
		}

		var raw model.VLLMRawResponse
		if err := json.Unmarshal(bodyBytes, &raw); err != nil {
			return "", fmt.Errorf("decode response failed: %w", err)
		}
		if len(raw.Choices) == 0 {
			return "", fmt.Errorf("vLLM returned no choices")
		}

		return raw.Choices[0].Message.Content, nil
	}

	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("vLLM request failed after retries")

}

func (c *VLLMClient) acquire(ctx context.Context) error {
	select {
	case c.limiter <- struct{}{}:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("llm request canceled before acquire: %w", ctx.Err())
	}
}

func (c *VLLMClient) release() {
	select {
	case <-c.limiter:
	default:
	}
}

func shouldRetryStatus(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusBadGateway || status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout
}

func retryAfterOrDefault(headerVal string, attempt int) time.Duration {
	if secs, err := strconv.Atoi(strings.TrimSpace(headerVal)); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	return time.Duration(attempt) * 400 * time.Millisecond
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
