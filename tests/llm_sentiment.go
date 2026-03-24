package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func getEnvOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getDurationEnvOrDefault(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := time.ParseDuration(raw)
	if err != nil {
		log.Printf("Gia tri %s khong hop le (%q), dung mac dinh %s", key, raw, fallback)
		return fallback
	}
	return v
}

func printQuotaHintIfNeeded(err error) {
	if err == nil {
		return
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "status=429") || strings.Contains(msg, "usage limit") {
		fmt.Println("   Goi y: Ban dang dung model cloud va bi het quota phien.")
		fmt.Println("   Cach 1 (cloud): kiem tra quota/API key cua nha cung cap.")
		fmt.Println("   Cach 2 (local): doi model sang qwen2.5:7b hoac model co trong 'ollama list'.")
	}
}

func printModelNotFoundHintIfNeeded(err error, apiURL string) {
	if err == nil {
		return
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "model") || !strings.Contains(msg, "not found") {
		return
	}

	fmt.Println("   Goi y: Model chua ton tai o endpoint hien tai.")
	if strings.Contains(strings.ToLower(apiURL), "localhost") || strings.Contains(apiURL, "127.0.0.1") {
		fmt.Println("   Ban dang goi local Ollama. Hay dung model co trong 'ollama list', vi du: qwen2.5:7b")
		fmt.Println("   Hoac neu muon dung cloud model thi set:")
		fmt.Println("   $env:LLM_API_URL='https://<your-cloud-openai-compatible-base-url>'")
		fmt.Println("   $env:LLM_API_KEY='<your_api_key>'")
		fmt.Println("   $env:LLM_MODEL='gpt-oss:120b-cloud'")
	}
}

// func main() {
// 	apiURL := getEnvOrDefault("LLM_API_URL", "http://localhost:11434")
// 	modelName := getEnvOrDefault("LLM_MODEL", "qwen2.5:7b")
// 	timeout := getDurationEnvOrDefault("LLM_TIMEOUT", 120*time.Second)

// 	fmt.Printf("=== Test ket noi va phan tich cam xuc voi %s ===\n", modelName)
// 	fmt.Printf("LLM_API_URL: %s\n", apiURL)

// 	client := vllm.NewVLLMClient(apiURL, timeout)
// 	gateway := vllm.NewVLLMGateway(apiURL, modelName, timeout)

// 	fmt.Println("\n--- Test 1: Tao phan hoi don gian ---")
// 	testSimpleResponse(client, apiURL, modelName)

// 	fmt.Println("\n--- Test 2: Phan tich intent ---")
// 	testIntentDetection(gateway, apiURL)

// 	fmt.Println("\n=== Hoan thanh tat ca cac test ===")
// }

// func testSimpleResponse(client *vllm.VLLMClient, apiURL string, modelName string) {
// 	ctx := context.Background()

// 	req := model.LLMRequest{
// 		Model: modelName,
// 		Messages: []model.LLMMessage{
// 			{Role: "system", Content: "Ban la mot tro ly AI huu ich."},
// 			{Role: "user", Content: "Xin chao, ban co the giup gi cho toi?"},
// 		},
// 		Temperature: 0.7,
// 		MaxTokens:   150,
// 	}

// 	response, err := client.Complete(ctx, req)
// 	if err != nil {
// 		log.Printf("Loi khi tao phan hoi: %v", err)
// 		printQuotaHintIfNeeded(err)
// 		printModelNotFoundHintIfNeeded(err, apiURL)
// 		return
// 	}

// 	fmt.Printf("Phan hoi tu mo hinh: %s\n", response)
// }

// func testIntentDetection(gateway *vllm.VLLMGateway, apiURL string) {
// 	ctx := context.Background()
// 	userMsg := "Minh rat buon vi doi nha vua bi thua!"

// 	matchCtx := model.MatchState{
// 		MatchID:  "test_match_001",
// 		HomeTeam: model.Team{Name: "Doi Nha", ShortName: "DN"},
// 		AwayTeam: model.Team{Name: "Doi Khach", ShortName: "DK"},
// 	}

// 	systemPrompt := `Tra ve CHI JSON hop le theo schema:
// {"sentiment":"positive|neutral|negative","team_bias":"string","main_topic":"goal|red_card|score|stats|lineup|substitution|time|greeting|other","requires_reply":true}`

// 	intent, err := gateway.DetectUserIntent(ctx, systemPrompt, userMsg, matchCtx)
// 	if err != nil {
// 		log.Printf("Loi khi phan tich intent: %v", err)
// 		printQuotaHintIfNeeded(err)
// 		printModelNotFoundHintIfNeeded(err, apiURL)
// 		return
// 	}

// 	fmt.Printf("Intent duoc phan tich: %+v\n", intent)
// }
