package seeding

import (
	"testing"
	"uniscore-seeding-bot/internal/domain/model"
)

func TestExceedsProjectedBotRatio_ColdStartAllowed(t *testing.T) {
	messages := []model.ChatMessage{
		{IsBot: true},
		{IsBot: true},
	}

	blocked, _, _, _ := exceedsProjectedBotRatio(messages, 0.3)
	if blocked {
		t.Fatalf("expected cold-start room to be allowed")
	}
}

func TestExceedsProjectedBotRatio_BlocksWhenExceeded(t *testing.T) {
	messages := []model.ChatMessage{
		{IsBot: true},
		{IsBot: true},
		{IsBot: true},
		{IsBot: false},
		{IsBot: false},
	}

	blocked, _, _, _ := exceedsProjectedBotRatio(messages, 0.3)
	if !blocked {
		t.Fatalf("expected projected ratio to exceed threshold")
	}
}

func TestExceedsProjectedBotRatio_AllowsWhenWithinThreshold(t *testing.T) {
	messages := []model.ChatMessage{
		{IsBot: true},
		{IsBot: false},
		{IsBot: false},
		{IsBot: false},
		{IsBot: false},
	}

	blocked, _, _, _ := exceedsProjectedBotRatio(messages, 0.5)
	if blocked {
		t.Fatalf("expected projected ratio to stay within threshold")
	}
}
