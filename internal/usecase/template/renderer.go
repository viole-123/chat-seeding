package template

import (
	"fmt"
	"strconv"
	"strings"
	"uniscore-seeding-bot/internal/domain/model"
)

// Render replaces placeholders in template.
type TemplateRenderer struct {
}

// Init TemplateRenderer
func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{}
}

func (r *TemplateRenderer) Render(templateText string, bundle *model.ContextBundle) (string, error) {
	text := templateText
	homeTeam := bundle.Match.HomeTeam.ShortName
	awayTeam := bundle.Match.AwayTeam.ShortName
	if homeTeam == "" {
		homeTeam = bundle.Match.HomeTeam.Name
	}
	if awayTeam == "" {
		awayTeam = bundle.Match.AwayTeam.Name
	}

	if len(bundle.Match.Events) == 0 {
		text = strings.ReplaceAll(text, "{{home_team}}", homeTeam)
		text = strings.ReplaceAll(text, "{{away_team}}", awayTeam)
		return text, nil
	}

	event := bundle.Match.Events[0]
	// Replace placeholders
	text = strings.ReplaceAll(text, "{{player}}", event.PlayerName)
	text = strings.ReplaceAll(text, "{{player_name}}", event.PlayerName)
	text = strings.ReplaceAll(text, "{{time}}", strconv.Itoa(event.Minute))
	text = strings.ReplaceAll(text, "{{minute}}", strconv.Itoa(event.Minute))
	text = strings.ReplaceAll(text, "{{position}}", strconv.Itoa(event.Position))
	text = strings.ReplaceAll(text, "{{home_score}}", strconv.Itoa(event.HomeScore))
	text = strings.ReplaceAll(text, "{{away_score}}", strconv.Itoa(event.AwayScore))
	text = strings.ReplaceAll(text, "{{score}}", fmt.Sprintf("%d-%d", event.HomeScore, event.AwayScore))
	teamName := event.AwayTeam
	if event.Position == 1 {
		teamName = event.HomeTeam
	}
	text = strings.ReplaceAll(text, "{{team}}", teamName)
	if event.HomeTeam != "" {
		text = strings.ReplaceAll(text, "{{home_team}}", event.HomeTeam)
	} else {
		text = strings.ReplaceAll(text, "{{home_team}}", homeTeam)
	}
	if event.AwayTeam != "" {
		text = strings.ReplaceAll(text, "{{away_team}}", event.AwayTeam)
	} else {
		text = strings.ReplaceAll(text, "{{away_team}}", awayTeam)
	}

	text = strings.ReplaceAll(text, "{{in_player}}", event.InPlayerName)
	text = strings.ReplaceAll(text, "{{out_player}}", event.OutPlayerName)
	return text, nil
}
