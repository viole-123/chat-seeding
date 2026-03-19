package seeding

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"uniscore-seeding-bot/internal/adapter/redis"
	"uniscore-seeding-bot/internal/config"
	"uniscore-seeding-bot/internal/domain/model"
	"uniscore-seeding-bot/internal/domain/service"
)

// PolicyChecker evaluates rate limits, cooldowns, bot ratio.
type PolicyChecker struct {
	dedupService      *redis.DedupService
	killSwitchService service.KillSwitchService
	rateLimitService  service.RateLimitService
	seedingPolicy     config.SeedingPolicy
	personaService    service.PersonaStateService
	contextStore      service.ContextStore
}

// Init NewPolicy
func NewPolicyChecker(
	dedupService *redis.DedupService,
	killSwitchService service.KillSwitchService,
	rateLimitService service.RateLimitService, // ← THÊM
	seedingPolicy config.SeedingPolicy,
	personaService service.PersonaStateService,
	contextStore service.ContextStore,
) *PolicyChecker {
	return &PolicyChecker{
		dedupService:      dedupService,
		killSwitchService: killSwitchService,
		rateLimitService:  rateLimitService,
		seedingPolicy:     seedingPolicy,
		personaService:    personaService,
		contextStore:      contextStore,
	}
}

func (p *PolicyChecker) CheckPolicy(ctx context.Context, event model.MatchEvent) (bool, error) {
	log.Printf("   🔍 [POLICY] begin: match=%s type=%s minute=%d+%d player=%q score=%d-%d",
		event.MatchID, event.Type, event.Minute, event.AddTime, event.PlayerName, event.HomeScore, event.AwayScore)

	// ⭐ Bước 0: Validate dữ liệu cơ bản của event
	if !p.validateEventData(event) {
		log.Printf("   ⏭️  [POLICY] skip invalid event data: match=%s type=%s minute=%d", event.MatchID, event.Type, event.Minute)
		return false, nil
	}
	if !isSeedableEventType(event.Type) {
		log.Printf("   ⏭️  [POLICY] skip non-seedable event type: match=%s type=%s", event.MatchID, event.Type)
		return false, nil
	}

	weightPass, weight, roll := passesEventWeighting(event.Type)
	if !weightPass {
		log.Printf("   ⏭️  [POLICY] skip by event weighting: match=%s type=%s roll=%.4f threshold=%.2f", event.MatchID, event.Type, roll, weight)
		return false, nil
	}
	log.Printf("   ✅ [POLICY] weighting pass: match=%s type=%s roll=%.4f threshold=%.2f", event.MatchID, event.Type, roll, weight)
	isKilled := false
	if leagueKilled, err := p.killSwitchService.IsKilled(ctx, "league", event.LeagueID); err != nil {
		log.Printf("   ⚠️  [POLICY] Failed to check league kill switch: %v", err)
	} else if leagueKilled {
		isKilled = true
	}
	if !isKilled {
		if matchKilled, err := p.killSwitchService.IsKilled(ctx, "match", event.MatchID); err != nil {
			log.Printf("   ⚠️  [POLICY] Failed to check match kill switch: %v", err)
		} else if matchKilled {
			isKilled = true
		}
	}
	if isKilled {
		log.Printf("   ⏭️  [POLICY] skip by kill switch: match=%s league=%s", event.MatchID, event.LeagueID)
		return false, nil
	}

	minuteKey := fmt.Sprintf("%d+%d", event.Minute, event.AddTime)
	eventSignature := buildEventDedupSignature(event)
	isDup, err := p.dedupService.IsDuplicateEvent(ctx, event.MatchID, minuteKey, eventSignature)
	if err != nil {
		return false, fmt.Errorf("check duplicated false, %w", err)
	}
	if isDup {
		log.Printf("   ⏭️  [POLICY] skip duplicate event: match=%s minute=%s fp=%s", event.MatchID, minuteKey, eventSignature)
		return false, nil
	}
	log.Printf("   ✅ [POLICY] dedup pass: match=%s minute=%s fp=%s", event.MatchID, minuteKey, eventSignature)

	// Bước 3: Check event type limit
	allowed, err := p.rateLimitService.CheckEventTypeLimit(ctx, event.MatchID, event.Type)
	if err != nil {
		return false, fmt.Errorf("check event type limit failed: %w", err)
	}
	if !allowed {
		log.Printf("   ⏭️  [POLICY] skip by event type limit: match=%s type=%s", event.MatchID, event.Type)
		return false, nil
	}
	log.Printf("   ✅ [POLICY] event type limit pass: match=%s type=%s", event.MatchID, event.Type)

	// Bước 5: Check match limit
	allowed, err = p.rateLimitService.CheckMatchLimit(ctx, event.MatchID, p.seedingPolicy.MaxMessagesBot)
	if err != nil {
		return false, fmt.Errorf("check match limit failed: %w", err)
	}
	if !allowed {
		log.Printf("   ⏭️  [POLICY] skip by match limit: match=%s max=%d", event.MatchID, p.seedingPolicy.MaxMessagesBot)
		return false, nil
	}
	log.Printf("   ✅ [POLICY] match limit pass: match=%s", event.MatchID)

	if p.contextStore != nil {
		roomID := event.MatchID
		if !strings.HasPrefix(roomID, "room-") {
			roomID = "room-" + roomID
		}
		chatCtx, err := p.contextStore.GetRecentChatWindow(roomID, 200)
		if err != nil {
			log.Printf("   ⚠️  [POLICY] bot ratio check failed: %v", err)
		} else if blocked, currentBot, currentUser, projected := exceedsProjectedBotRatio(chatCtx.RawMessages, p.seedingPolicy.BotRatio); blocked {
			log.Printf("   ⏭️  [POLICY] skip by bot ratio: match=%s room=%s threshold=%.2f", event.MatchID, roomID, p.seedingPolicy.BotRatio)
			return false, nil
		} else {
			log.Printf("   ✅ [POLICY] bot ratio pass: match=%s bot=%d user=%d projected=%.4f threshold=%.2f",
				event.MatchID, currentBot, currentUser, projected, p.seedingPolicy.BotRatio)
		}
	}

	return true, nil
}

func (p *PolicyChecker) validateEventData(event model.MatchEvent) bool {
	if event.LeagueID == "" {
		fmt.Println("Error : LeagueID chua duoc gan vao")
		return false
	}
	if !isValidEventType(event.Type) {
		return false
	}
	if needsPlayer(event.Type) && event.PlayerName == "" {
		return false
	}
	if event.Type == "SUBSTITUTION" && event.PlayerName == "" && (event.InPlayerName == "" || event.OutPlayerName == "") {
		return false
	}

	return true
}

func isValidEventType(eventType string) bool {
	validTypes := map[string]bool{
		"GOAL":                   true,
		"CORNER":                 true,
		"YELLOW_CARD":            true,
		"RED_CARD":               true,
		"OFFSIDE":                true,
		"FREE_KICK":              true,
		"GOAL_KICK":              true,
		"PENALTY":                true,
		"SUBSTITUTION":           true,
		"START":                  true,
		"MIDFIELD":               true,
		"END":                    true,
		"HALFTIME_SCORE":         true,
		"HALF_TIME":              true,
		"FULL_TIME":              true,
		"ADDED_TIME":             true,
		"INJURY":                 true,
		"INJURY_TIME":            true,
		"PENALTY_MISSED":         true,
		"OWN_GOAL":               true,
		"CARD_UPGRADE_CONFIRMED": true,
		"UNKNOWN":                true,
	}
	return validTypes[eventType]
}
func needsPlayer(eventType string) bool {
	return eventType == "GOAL" || eventType == "YELLOW_CARD" || eventType == "RED_CARD"
}

func isSeedableEventType(eventType string) bool {
	seedable := map[string]bool{
		"GOAL":                   true,
		"RED_CARD":               true,
		"PENALTY":                true,
		"PENALTY_MISSED":         true,
		"OWN_GOAL":               true,
		"CARD_UPGRADE_CONFIRMED": true,
		"UNKNOWN":                true,

		// Match milestones (nên seed)
		"START":          true,
		"HALFTIME_SCORE": true,
		"MIDFIELD":       true,
		"END":            true,
		"INJURY_TIME":    true,

		// Medium (vẫn seed nhưng sẽ bị weight lọc)
		"YELLOW_CARD":  true,
		"SUBSTITUTION": true,

		// Low/noisy → KHÔNG seed (để user thật nói)
		"CORNER":    false,
		"OFFSIDE":   false,
		"FREE_KICK": false,
		"GOAL_KICK": false,
	}
	return seedable[eventType]
}

func buildEventDedupSignature(event model.MatchEvent) string {
	// Build a stable raw signature from key fields, then hash by md5 to keep Redis key short and consistent.
	rawParts := []string{
		event.MatchID,
		event.Type,
		strings.ToLower(strings.TrimSpace(event.PlayerName)),
		strings.ToLower(strings.TrimSpace(event.InPlayerName)),
		strings.ToLower(strings.TrimSpace(event.OutPlayerName)),
		fmt.Sprintf("%d", event.Minute),
		fmt.Sprintf("%d", event.AddTime),
		fmt.Sprintf("%d", event.Position),
		fmt.Sprintf("%d", event.ReasonType),
		fmt.Sprintf("%d", event.HomeScore),
		fmt.Sprintf("%d", event.AwayScore),
		strings.ToLower(strings.TrimSpace(event.TeamSide)),
	}
	raw := strings.Join(rawParts, "|")
	hash := md5.Sum([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// ddieeuf chinh trong so xuat hien
func passesEventWeighting(eventType string) (bool, float64, float64) {
	weights := map[string]float64{
		"GOAL":                   1.00,
		"RED_CARD":               1.00,
		"PENALTY":                1.00,
		"PENALTY_MISSED":         0.95,
		"OWN_GOAL":               0.95,
		"CARD_UPGRADE_CONFIRMED": 0.90,
		"UNKNOWN":                0.90,

		// match milestones
		"START":          0.85,
		"HALFTIME_SCORE": 0.80,
		"MIDFIELD":       0.75,
		"END":            0.90,
		"INJURY_TIME":    0.70,

		// medium impact
		"SUBSTITUTION": 0.65,
		"YELLOW_CARD":  0.60,

		// noisy events
		"CORNER":    0.35,
		"OFFSIDE":   0.30,
		"FREE_KICK": 0.30,
		"GOAL_KICK": 0.20,
	}

	weight, ok := weights[eventType]
	if !ok {
		weight = 0.20
	}
	roll := rand.Float64()
	if weight >= 1 {
		return true, weight, 0
	}
	return roll < weight, weight, roll
}

func exceedsProjectedBotRatio(messages []model.ChatMessage, maxRatio float64) (bool, int, int, float64) {
	if maxRatio <= 0 {
		return true, 0, 0, 1
	}

	var botCount int
	var userCount int
	for _, msg := range messages {
		if msg.IsBot {
			botCount++
		} else {
			userCount++
		}
	}

	// No user traffic yet: allow event-driven bot seeding flow.
	if userCount == 0 {
		projected := float64(botCount+1) / float64(botCount+1)
		return false, botCount, userCount, projected
	}

	// Cold-start window: allow catalyst messages when room is still empty.
	total := botCount + userCount
	if total < 5 {
		projected := float64(botCount+1) / float64(total+1)
		return false, botCount, userCount, projected
	}

	projectedBot := botCount + 1
	projectedTotal := total + 1
	projectedRatio := float64(projectedBot) / float64(projectedTotal)
	return projectedRatio > maxRatio, botCount, userCount, projectedRatio
}
