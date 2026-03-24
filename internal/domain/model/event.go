package model

// MatchEvent represents an in-game event.
type MatchEvent struct {
	MatchID  string `json:"match_id"`
	LeagueID string `json:"league_id"`

	Type          string `json:"type"`
	Position      int    `json:"position"`
	Minute        int    `json:"minute"`
	AddTime       int    `json:"add_time"`
	HomeScore     int    `json:"home_score"`
	AwayScore     int    `json:"away_score"`
	PlayerName    string `json:"player_name"`
	InPlayerName  string `json:"in_player_name"`
	OutPlayerName string `json:"out_player_name"`
	ReasonType    int    `json:"reason_type"`
	HomeTeam      string `json:"home_team"`
	AwayTeam      string `json:"away_team"`
	TeamSide      string `json:"team_side,omitempty"`

	Timestamp int64 `json:"timestamp"`
}

type CompactEvent struct {
	Type       string `json:"type"`
	Minute     int    `json:"minute"`
	AddTime    int    `json:"add_time"`
	TeamSide   string `json:"team_side"`
	PlayerName string `json:"player_name"`
	HomeScore  int    `json:"home_score"`
	AwayScore  int    `json:"away_score"`
	Summary    string `json:"summary"`
}
type EventsBundle struct {
	MatchEvents []MatchEvent
}
