//go:build ignore
// +build ignore

// Chay: go run .\tests\llm_sentiment_cloud.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type intentOutput struct {
	Sentiment     string `json:"sentiment"`
	MainTopic     string `json:"main_topic"`
	RequiresReply *bool  `json:"requires_reply"`
}

func main() {
	// ── Cau hinh ──────────────────────────────────────────────────────────────
	// Mac dinh: dung local Ollama voi cloud model (Ollama tu route len cloud)
	// De doi sang model khac, set bien moi truong truoc khi chay:
	//   $env:LLM_MODEL = "qwen3-coder:480b-cloud"
	//   $env:LLM_MODEL = "deepseek-v3.1:671b-cloud"
	//   $env:LLM_MODEL = "gpt-oss:120b-cloud"

	serverURL := getEnv("LLM_SERVER_URL", getEnv("LLM_API_URL", "http://localhost:11434"))
	modelName := getEnv("LLM_MODEL", "gpt-oss:120b-cloud")

	fmt.Println("=== Test ket noi Ollama Cloud ===")
	fmt.Printf("Server : %s\n", serverURL)
	fmt.Printf("Model  : %s\n\n", modelName)
	preflightModel(serverURL, modelName)

	// ── Khoi tao LLM ─────────────────────────────────────────────────────────
	llm, err := ollama.New(
		ollama.WithModel(modelName),
		ollama.WithServerURL(serverURL),
	)
	if err != nil {
		log.Fatalf("[FAIL] Khoi tao LLM: %v", err)
	}

	// ── Test 1: Cau hoi don gian ──────────────────────────────────────────────
	fmt.Println("--- Test 1: Cau hoi don gian (co streaming) ---")
	runStreamTest(llm, modelName, serverURL)

	// ── Test 2: Intent detection cho seeding bot ─────────────────────────────
	fmt.Println("\n--- Test 2: Intent detection (JSON output) ---")
	runIntentTest(llm, modelName, serverURL)
}

// ─────────────────────────────────────────────────────────────────────────────

func runStreamTest(llm *ollama.LLM, modelName, serverURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	fmt.Print("Phan hoi: ")
	_, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		"Xin chao! Hay noi 1 cau ngan bang tieng Viet trạng thái tức giận về thẻ đỏ người chơi ABC có được",
		llms.WithTemperature(0.7),
		llms.WithMaxTokens(200),
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}),
	)
	fmt.Println()
	if err != nil {
		log.Printf("[FAIL] %v", err)
		printHint(err, modelName, serverURL)
	} else {
		fmt.Println("[OK] Test 1 thanh cong!")
	}
}

func runIntentTest(llm *ollama.LLM, modelName, serverURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Dung GenerateContent de truyen system prompt + user message rieng biet
	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: `Phan tich tin nhan nguoi dung trong phong chat bong da.
Tra ve CHI JSON hop le, khong them chu thich, theo dung schema:
{"sentiment":"positive|neutral|negative","main_topic":"goal|red_card|score|stats|other","requires_reply":true}`},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Minh thấy bình thường khi xem 2 đội đấu "},
			},
		},
	}

	var lastRaw string
	var lastErr error

	for attempt := 1; attempt <= 3; attempt++ {
		resp, err := llm.GenerateContent(ctx, content,
			llms.WithTemperature(0.0), // de output on dinh hon
			llms.WithMaxTokens(220),   // tang max token de giam cat JSON
		)
		if err != nil {
			lastErr = err
			continue
		}

		if len(resp.Choices) == 0 {
			lastErr = fmt.Errorf("model tra ve khong co choices")
			continue
		}

		raw := strings.TrimSpace(resp.Choices[0].Content)
		lastRaw = raw
		if raw == "" {
			lastErr = fmt.Errorf("model tra ve rong (empty content)")
			continue
		}

		clean := normalizeJSONText(raw)
		if repaired, ok := tryRepairIntentJSON(clean); ok {
			clean = repaired
		}

		out, err := parseIntentJSON(clean)
		if err != nil {
			lastErr = err
			continue
		}

		normalized, _ := json.Marshal(out)
		fmt.Printf("JSON tra ve: %s\n", string(normalized))
		fmt.Println("[OK] Test 2 thanh cong!")
		return
	}

	if lastErr != nil {
		log.Printf("[FAIL] JSON intent khong hop le sau 3 lan thu: %v", lastErr)
	} else {
		log.Printf("[FAIL] Khong lay duoc output intent hop le")
	}
	if strings.TrimSpace(lastRaw) != "" {
		fmt.Printf("Raw output cuoi: %s\n", lastRaw)
	}
}

// ─────────────────────────────────────────────────────────────────────────────

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func normalizeJSONText(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSpace(s)
		if strings.HasPrefix(strings.ToLower(s), "json") {
			s = strings.TrimSpace(s[4:])
		}
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return strings.TrimSpace(s[start : end+1])
	}
	return s
}

func tryRepairIntentJSON(raw string) (string, bool) {
	s := strings.TrimSpace(raw)
	if s == "" || !strings.Contains(s, "{") {
		return raw, false
	}

	// Common truncation case: output ends around "requires_reply
	if strings.Contains(s, "\"requires_reply") && !strings.Contains(s, "\"requires_reply\":") {
		s = strings.TrimSpace(s)
		if strings.HasSuffix(s, "\"requires_reply") {
			s += "\":true"
		} else if strings.HasSuffix(s, "\"requires_reply\"") {
			s += ":true"
		}
	}

	open := strings.Count(s, "{")
	close := strings.Count(s, "}")
	for close < open {
		s += "}"
		close++
	}

	if s != raw {
		return s, true
	}
	return raw, false
}

func parseIntentJSON(raw string) (intentOutput, error) {
	var out intentOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return intentOutput{}, err
	}

	out.Sentiment = strings.ToLower(strings.TrimSpace(out.Sentiment))
	switch out.Sentiment {
	case "positive", "neutral", "negative":
	default:
		return intentOutput{}, fmt.Errorf("sentiment khong hop le: %q", out.Sentiment)
	}

	out.MainTopic = strings.ToLower(strings.TrimSpace(out.MainTopic))
	switch out.MainTopic {
	case "goal", "red_card", "score", "stats", "other":
	default:
		return intentOutput{}, fmt.Errorf("main_topic khong hop le: %q", out.MainTopic)
	}

	if out.RequiresReply == nil {
		return intentOutput{}, fmt.Errorf("thieu truong requires_reply")
	}

	return out, nil
}

func preflightModel(serverURL, modelName string) {
	type modelsResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	url := strings.TrimRight(serverURL, "/") + "/v1/models"
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("[WARN] Khong doc duoc %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("[WARN] %s tra ve %d: %s\n", url, resp.StatusCode, string(body))
		return
	}

	var out modelsResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		fmt.Printf("[WARN] Parse %s that bai: %v\n", url, err)
		return
	}

	ids := make([]string, 0, len(out.Data))
	found := false
	for _, m := range out.Data {
		ids = append(ids, m.ID)
		if strings.EqualFold(strings.TrimSpace(m.ID), strings.TrimSpace(modelName)) {
			found = true
		}
	}

	if found {
		fmt.Printf("[OK] Model '%s' co tren endpoint nay.\n\n", modelName)
		return
	}

	fmt.Printf("[WARN] Endpoint hien tai CHUA co model '%s'.\n", modelName)
	if len(ids) > 0 {
		fmt.Printf("[INFO] Models tren endpoint: %s\n", strings.Join(ids, ", "))
	}
	fmt.Println("[INFO] Neu ban dang dung Docker Ollama, can pull/signin TRONG container.")
	fmt.Println("[INFO] Thu: docker exec -it ollama ollama signin")
	fmt.Println("[INFO] Thu: docker exec ollama ollama pull " + modelName)
	fmt.Println()
}

func printHint(err error, modelName, serverURL string) {
	msg := strings.ToLower(err.Error())
	fmt.Println("\n  === Goi y sua loi ===")

	switch {
	case strings.Contains(msg, "connection refused"):
		fmt.Println("  Ollama chua chay. Mo terminal moi va chay: ollama serve")

	case strings.Contains(msg, "not found") || strings.Contains(msg, "unknown model"):
		fmt.Printf("  Endpoint '%s' khong co model '%s'.\n", serverURL, modelName)
		fmt.Println("  Neu endpoint la Docker Ollama, chay:")
		fmt.Printf("  docker exec ollama ollama pull %s\n", modelName)
		fmt.Println("  Hoac doi endpoint qua daemon da login cloud: set LLM_SERVER_URL.")

	case strings.Contains(msg, "unauthorized") || strings.Contains(msg, "401"):
		fmt.Printf("  Endpoint '%s' chua duoc dang nhap Ollama Cloud.\n", serverURL)
		fmt.Println("  Neu dung Docker Ollama:")
		fmt.Println("  docker exec -it ollama ollama signin")
		fmt.Println("  docker exec ollama ollama pull " + modelName)
		fmt.Println("  Neu dung Ollama Desktop: dung endpoint cua Desktop bang LLM_SERVER_URL.")

	case strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline"):
		fmt.Println("  Timeout. Cloud model co the mat 30-60s lan dau.")
		fmt.Println("  Thu tang LLM_TIMEOUT: $env:LLM_TIMEOUT='180s'")

	default:
		fmt.Printf("  Loi: %v\n", err)
		fmt.Printf("  Server: %s | Model: %s\n", serverURL, modelName)
	}
}
