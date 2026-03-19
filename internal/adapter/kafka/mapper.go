package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
	"uniscore-seeding-bot/internal/domain/model"
)

type RawAPIResponse struct {
	Code    int               `json:"code"`
	Results []RawMatchResults `json:"results"`
}

type RawMatchResults struct {
	ID        string        `json:"id"`
	Score     []interface{} `json:"score"`
	Stats     []interface{} `json:"stats"`
	Incidents []RawIncident `json:"incidents"`
	TLive     []interface{} `json:"tlive"`
}

type RawTLive struct {
	Time string      `json:"time"`
	Data interface{} `json:"data"`
}

type RawMatchResult struct {
	ID        string        `json:"id"`
	Score     []interface{} `json:"score"`
	Stats     []interface{} `json:"stats"`
	Incidents []RawIncident `json:"incidents"`
	TLive     []interface{} `json:"tlive"`
}

type RawStat struct {
	Type int `json:"type"`
	Home int `json:"home"`
	Away int `json:"away"`
}

type RawIncident struct {
	Type       int    `json:"type"`
	Position   int    `json:"position"`
	Time       int    `json:"time"`
	AddTime    int    `json:"add_time,omitempty"`
	PlayerID   string `json:"player_id,omitempty"`
	PlayerName string `json:"player_name,omitempty"`

	InPlayerID    string `json:"in_player_id,omitempty"`
	InPlayerName  string `json:"in_player_name,omitempty"`
	OutPlayerID   string `json:"out_player_id,omitempty"`
	OutPlayerName string `json:"out_player_name,omitempty"`

	ReasonType  int    `json:"reason_type,omitempty"`
	HomeScore   *int   `json:"home_score,omitempty"`
	AwayScore   *int   `json:"away_score,omitempty"`
	Assist1ID   string `json:"assist1_id,omitempty"`
	Assist1Name string `json:"assist1_name,omitempty"`
	VarReason   int    `json:"var_reason,omitempty"`
	VarResult   int    `json:"var_result,omitempty"`
}

type RawSingleWrapped struct {
	Code    int             `json:"code"`
	Results RawMatchResults `json:"results"`
}

func MapBytesToEventsBundle(data []byte) (*model.EventsBundle, error) {
	var apiRespSingular struct {
		Code   int               `json:"code"`
		Result []RawMatchResults `json:"result"`
	}
	if err := json.Unmarshal(data, &apiRespSingular); err == nil && len(apiRespSingular.Result) > 0 {
		return mapMultipleMatches(apiRespSingular.Result), nil
	}
	var apiResp RawAPIResponse
	if err := json.Unmarshal(data, &apiResp); err == nil && len(apiResp.Results) > 0 {
		return mapMultipleMatches(apiResp.Results), nil
	}

	var singleMatch RawSingleWrapped
	if err := json.Unmarshal(data, &singleMatch); err == nil && singleMatch.Results.ID != "" {
		return mapSingleMatch(singleMatch.Results), nil
	}

	log.Printf("❌ [MAPPER] Không match format nào. Raw JSON: %s", string(data))
	return nil, fmt.Errorf("Unknown JSON format - not API response or single match")
}

func MapBytesToEvents(data []byte) ([]model.MatchEvent, error) {
	bundle, err := MapBytesToEventsBundle(data)
	if err != nil {
		return nil, err
	}
	return bundle.MatchEvents, nil
}

func mapMultipleMatches(matches []RawMatchResults) *model.EventsBundle {
	bundle := &model.EventsBundle{}
	for _, match := range matches {
		status, ok := extractScoreStatus(match)
		if ok {
			log.Printf("🔎 [MAPPER] match=%s score_status=%d incidents=%d", match.ID, status, len(match.Incidents))
		} else {
			log.Printf("🔎 [MAPPER] match=%s score_status=unknown incidents=%d", match.ID, len(match.Incidents))
		}

		if isEndedMatch(match) {
			log.Printf("⏭️  [MAPPER] Skip ended match: match=%s", match.ID)
			continue
		}
		singleBundle := mapSingleMatch(match)
		bundle.MatchEvents = append(bundle.MatchEvents, singleBundle.MatchEvents...)
	}
	return bundle
}

func mapSingleMatch(match RawMatchResults) *model.EventsBundle {
	bundle := &model.EventsBundle{}
	if isEndedMatch(match) {
		log.Printf("⏭️  [MAPPER] Skip ended match: match=%s", match.ID)
		return bundle
	}

	leagueID := deriveLeagueID(match.ID)
	if len(match.Incidents) == 0 {
		log.Printf("⚠️  [MAPPER] Match %s có 0 incidents", match.ID)
	}
	log.Printf("Match = %s incident = %d ->processing...", match.ID, len(match.Incidents))
	for _, incident := range match.Incidents {
		eventType := mapIncidentType(incident.Type)
		event := model.MatchEvent{
			MatchID:       match.ID,
			LeagueID:      leagueID,
			Type:          eventType,
			Position:      incident.Position,
			Minute:        incident.Time,
			AddTime:       incident.AddTime,
			PlayerName:    incident.PlayerName,
			InPlayerName:  incident.InPlayerName,
			OutPlayerName: incident.OutPlayerName,
			ReasonType:    incident.ReasonType,
			Timestamp:     time.Now().Unix(),
		}
		if incident.HomeScore != nil {
			event.HomeScore = *incident.HomeScore
		}
		if incident.AwayScore != nil {
			event.AwayScore = *incident.AwayScore
		}
		bundle.MatchEvents = append(bundle.MatchEvents, event)
	}
	return bundle
}

func isEndedMatch(match RawMatchResults) bool {
	if isEndedByScoreStatus(match) {
		return true
	}

	// Fallback: if payload already has END incident, treat as ended snapshot and skip.
	for _, incident := range match.Incidents {
		if incident.Type == 12 {
			return true
		}
	}

	return false
}

func isEndedByScoreStatus(match RawMatchResults) bool {
	status, ok := extractScoreStatus(match)
	if !ok {
		return false
	}
	return status == 8
}

func extractScoreStatus(match RawMatchResults) (int, bool) {
	if len(match.Score) < 2 {
		return 0, false
	}

	statusRaw := match.Score[1]
	statusFloat, ok := statusRaw.(float64)
	if !ok {
		return 0, false
	}

	return int(statusFloat), true
}

func deriveLeagueID(matchID string) string {
	id := strings.TrimSpace(matchID)
	if id == "" {
		return "UNKNOWN"
	}
	parts := strings.SplitN(id, "_", 2)
	if len(parts) > 0 && parts[0] != "" {
		return strings.ToUpper(parts[0])
	}
	return "UNKNOWN"
}

func mapIncidentType(incidentType int) string {
	typeMap := map[int]string{
		1:  "GOAL",
		2:  "CORNER",
		3:  "YELLOW_CARD",
		4:  "RED_CARD",
		5:  "OFFSIDE",
		6:  "FREE_KICK",
		7:  "GOAL_KICK",
		8:  "PENALTY",
		9:  "SUBSTITUTION",
		10: "START",
		11: "MIDFIELD",
		12: "END",
		13: "HALFTIME_SCORE",
		15: "CARD_UPGRADE_CONFIRMED",
		16: "PENALTY_MISSED",
		17: "OWN_GOAL",
		19: "INJURY_TIME",
	}

	if eventType, ok := typeMap[incidentType]; ok {
		return eventType
	}
	return "UNKNOWN"
}
