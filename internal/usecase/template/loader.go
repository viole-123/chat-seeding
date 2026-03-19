package template

import (
	"log"
	"math/rand"
	"sync"
	"uniscore-seeding-bot/internal/domain/model"
	"uniscore-seeding-bot/internal/domain/repository"
)

type TemplateLoader struct {
	repo      repository.TemplateRepository
	cache     []model.Template
	cacheLock sync.RWMutex
}

// Init TemplateLoader with repository and load templates into RAM
func NewTemplateLoader(repo repository.TemplateRepository) *TemplateLoader {
	loader := &TemplateLoader{
		repo:  repo,
		cache: []model.Template{},
	}

	if err := loader.LoadAll(); err != nil {
		log.Printf("Template failed to load templates :%v", err)
	}
	return loader

}

func (l *TemplateLoader) LoadAll() error {
	templates, err := l.repo.GetAllTemplates()
	if err != nil {
		return err
	}
	l.cacheLock.Lock()
	l.cache = templates
	l.cacheLock.Unlock()
	log.Printf("✅ [TEMPLATE] Loaded %d templates into RAM", len(templates))
	return nil
}

func (l *TemplateLoader) GetMatchingTemplates(bundle *model.ContextBundle, persona *model.Persona) (*model.Template, error) {
	l.cacheLock.RLock()
	defer l.cacheLock.RUnlock()

	var candidates []model.Template
	currentPhase := bundle.Match.Phase
	totalTemplates := len(l.cache)
	enabledCount := 0
	phaseCount := 0
	langCount := 0
	personaCount := 0
	eventCount := 0
	conditionCount := 0
	for _, t := range l.cache {
		if !t.Enabled {
			continue
		}
		enabledCount++
		if t.Phase != currentPhase {
			continue
		}
		phaseCount++
		if t.Lang != "" && !containsString(persona.Profile.Language, t.Lang) {
			continue
		}
		langCount++
		if t.PersonaID != "" && t.PersonaID != persona.ID {
			continue
		}
		personaCount++
		if t.EventType != "" && t.EventType != bundle.CurrentEvent.Type {
			continue
		}
		eventCount++

		if !checkConditions(t.Conditions, bundle) {
			continue
		}
		conditionCount++
		candidates = append(candidates, t)
	}
	if len(candidates) == 0 {
		log.Printf("⚠️  [TEMPLATE] No match: total=%d enabled=%d phase=%d lang=%d persona=%d event=%d conditions=%d | phase=%s event=%s persona=%s languages=%v",
			totalTemplates, enabledCount, phaseCount, langCount, personaCount, eventCount, conditionCount,
			currentPhase, bundle.CurrentEvent.Type, persona.ID, persona.Profile.Language)
		return nil, nil
	}

	switch currentPhase {
	case model.PhasePrematch:
		return pickBestTemplateForPrematch(candidates), nil
	default:
		return pickBestTemplate(candidates), nil
	}

}

func checkConditions(conditions map[string]interface{}, bundle *model.ContextBundle) bool {
	if len(conditions) == 0 {
		return true
	}
	if v, ok := conditions["minute_range"]; ok {
		if raw, ok := v.([]interface{}); ok && len(raw) == 2 {
			min := int(raw[0].(float64))
			max := int(raw[1].(float64))
			if bundle.Match.Minute < min || bundle.Match.Minute > max {
				return false
			}
		}
	}
	if v, ok := conditions["score_diff"]; ok {
		if raw, ok := v.([]interface{}); ok && len(raw) == 2 {
			min := int(raw[0].(float64))
			max := int(raw[1].(float64))
			diff := abs(bundle.Match.HomeScore - bundle.Match.AwayScore)
			if diff < min || diff > max {
				return false
			}
		}
	}
	return true
}
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func containsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

func pickBestTemplate(candidates []model.Template) *model.Template {
	if len(candidates) == 0 {
		return nil
	}
	maxPriority := candidates[0].Priority
	for _, t := range candidates {
		if t.Priority > maxPriority {
			maxPriority = t.Priority
		}
	}
	var top []model.Template
	for _, t := range candidates {
		if t.Priority == maxPriority {
			top = append(top, t)
		}
	}
	chosen := top[rand.Intn(len(top))]
	return &chosen
}

func pickBestTemplateForPrematch(candidates []model.Template) *model.Template {
	if len(candidates) == 0 {
		return nil
	}
	totalWeight := 0
	for _, t := range candidates {
		totalWeight += t.Priority
	}
	if totalWeight == 0 {
		chosen := candidates[rand.Intn(len(candidates))]
		return &chosen
	}

	pick := rand.Intn(totalWeight)
	current := 0
	for i, t := range candidates {
		current += t.Priority
		if pick < current {
			return &candidates[i]
		}
	}
	chosen := candidates[len(candidates)-1]
	return &chosen
}
