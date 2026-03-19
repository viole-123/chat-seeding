package seeding

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"uniscore-seeding-bot/internal/domain/model"
	"uniscore-seeding-bot/internal/domain/service"
)

type PersonaSelector struct {
	personaStateService service.PersonaStateService
	personas            []model.Persona
}

func NewPersonaSelector(personaStateService service.PersonaStateService, personasFilePath string) (*PersonaSelector, error) {
	personas, err := loadPersonasFromFile(personasFilePath)
	if err != nil {
		return nil, fmt.Errorf("load personas failed: %w", err)
	}

	return &PersonaSelector{
		personaStateService: personaStateService,
		personas:            personas,
	}, nil
}

func (s *PersonaSelector) SelectPersona(ctx context.Context, contextBundle model.ContextBundle) (*model.Persona, error) {
	var candidates []model.PersonaScore

	eventType := contextBundle.CurrentEvent.Type
	if eventType == "" && len(contextBundle.Match.Events) > 0 {
		eventType = contextBundle.Match.Events[0].Type
	}
	if eventType == "" {
		return nil, fmt.Errorf("no event type for persona selection")
	}
	personaEventType := normalizeEventTypeForPersona(eventType)

	// ⭐ Bước 1: Tính điểm cho từng persona dựa trên context
	lastPersona, err := s.personaStateService.GetLastPersona(ctx, contextBundle.Match.MatchID)
	if err != nil {
		return nil, fmt.Errorf("get last persona failed: %w", err)
	}

	total := len(s.personas)
	filteredByEvent := 0
	filteredByCooldown := 0
	filteredByRepeat := 0
	filteredByScore := 0

	for _, p := range s.personas {
		if personaEventType != "" && !isPersonaEligibleForEvent(p, personaEventType) {
			filteredByEvent++
			continue
		}
		onCooldown, err := s.personaStateService.IsOnCoolDown(ctx, p.ID)
		if err != nil {
			return nil, fmt.Errorf("check persona cooldown failed: %w", err)
		}
		if onCooldown {
			filteredByCooldown++
			continue
		}
		if lastPersona == p.ID {
			filteredByRepeat++
			continue
		}
		score := s.calculatePersonaScore(p, contextBundle)
		if score <= 0 {
			filteredByScore++
			continue
		}
		candidates = append(candidates, model.PersonaScore{Persona: p, Score: score})

	}
	if len(candidates) == 0 {
		log.Printf("   ⚠️  [PERSONA] no candidate: match=%s event=%s total=%d filtered_event=%d cooldown=%d repeat=%d score=%d",
			contextBundle.Match.MatchID, personaEventType, total, filteredByEvent, filteredByCooldown, filteredByRepeat, filteredByScore)
		fallbackPersona := s.findFallbackPersona(ctx)
		if fallbackPersona == nil {
			return nil, fmt.Errorf("no eligible persona found: event=%s normalized=%s total_personas=%d", eventType, personaEventType, len(s.personas))
		}
		log.Printf("   ♻️  [PERSONA] fallback selected: %s", fallbackPersona.ID)
		return fallbackPersona, nil
	}

	log.Printf("   🔎 [PERSONA] candidates=%d total=%d filtered_event=%d cooldown=%d repeat=%d score=%d last=%s event=%s",
		len(candidates), total, filteredByEvent, filteredByCooldown, filteredByRepeat, filteredByScore, lastPersona, personaEventType)

	selected := pickWeightedRandom(candidates)
	if selected != nil {
		cooldown := selected.Policy.CooldownSeconds
		if cooldown <= 0 {
			cooldown = 30
		}
		_ = s.personaStateService.SetCoolDown(ctx, selected.ID, cooldown)
		_ = s.personaStateService.SaveLastPersona(ctx, contextBundle.Match.MatchID, selected.ID)
		log.Printf("   ✅ [PERSONA] selected=%s cooldown=%ds", selected.ID, cooldown)
	}
	return selected, nil
}
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
func isPersonaEligibleForEvent(p model.Persona, eventType string) bool {
	if len(p.Policy.AllowedEventTypes) > 0 && !contains(p.Policy.AllowedEventTypes, eventType) {
		return false
	}

	return true
}
func pickWeightedRandom(candidates []model.PersonaScore) *model.Persona {
	total := 0

	for _, c := range candidates {
		total += c.Score

	}
	if total <= 0 {
		idx := rand.Intn(len(candidates))
		return &candidates[idx].Persona
	}
	r := rand.Intn(total)
	sum := 0
	for _, c := range candidates {

		if c.Score < 0 {
			continue
		}
		sum += c.Score
		if r < sum {
			return &c.Persona
		}
	}
	return &candidates[0].Persona
}

func (s *PersonaSelector) calculatePersonaScore(p model.Persona, contextBundle model.ContextBundle) int {
	score := p.Policy.WeightBase

	eventType := contextBundle.CurrentEvent.Type
	if eventType == "" && len(contextBundle.Match.Events) > 0 {
		eventType = contextBundle.Match.Events[0].Type
	}
	eventType = normalizeEventTypeForPersona(eventType)

	switch eventType {
	case "GOAL":
		if p.Profile.Tone == "hype" {
			score += 20
			// score +=p.Policy.EventWeight["GOAL"]
		}
		if contextBundle.CurrentEvent.Minute >= 85 {
			score += 10
		}
	case "RED_CARD":
		if p.Profile.Tone == "analyst" {
			score += 20
		}
	case "SUBSTITUTION":
		if p.Profile.Tone == "calm" {
			score += 5
		}
	}
	switch contextBundle.CurrentEvent.Type {
	case "excited":
		if p.Profile.Tone == "hype" || p.Profile.Tone == "funny" {
			score += 15
		}
	case "negative", "angry":
		if p.Profile.Tone == "calm" || p.Profile.Tone == "analyst" {
			score += 15
		}
	}
	return score
}

func (s *PersonaSelector) findFallbackPersona(ctx context.Context) *model.Persona {
	for _, p := range s.personas {
		onCooldown, err := s.personaStateService.IsOnCoolDown(ctx, p.ID)
		if err != nil || onCooldown {
			continue
		}

		return &p
	}
	return nil
}

// Used as a safety fallback for prematch burst sending to avoid dead-end selection.
func (s *PersonaSelector) SelectPersonaAllowReuse(contextBundle model.ContextBundle) *model.Persona {
	eventType := contextBundle.CurrentEvent.Type
	if eventType == "" && len(contextBundle.Match.Events) > 0 {
		eventType = contextBundle.Match.Events[0].Type
	}
	if eventType == "" {
		return nil
	}
	personaEventType := normalizeEventTypeForPersona(eventType)

	eligible := make([]model.PersonaScore, 0, len(s.personas))
	for _, p := range s.personas {
		if !isPersonaEligibleForEvent(p, personaEventType) {
			continue
		}
		score := s.calculatePersonaScore(p, contextBundle)
		if score <= 0 {
			score = 1
		}
		eligible = append(eligible, model.PersonaScore{Persona: p, Score: score})
	}
	if len(eligible) == 0 {
		return nil
	}
	return pickWeightedRandom(eligible)
}

func normalizeEventTypeForPersona(eventType string) string {
	e := strings.ToUpper(strings.TrimSpace(eventType))
	switch e {
	case "PENALTY", "OWN_GOAL":
		return "GOAL"
	default:
		return e
	}
}

func loadPersonasFromFile(path string) ([]model.Persona, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var personas []model.Persona
	if err := yaml.Unmarshal(data, &personas); err != nil {
		return nil, err
	}

	return personas, nil
}
