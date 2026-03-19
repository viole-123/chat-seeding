package model

// MatchPhase là trạng thái vòng đời của một trận đấu
type MatchPhase string

const (
	PhaseWaiting  MatchPhase = "waiting"  // chưa đến T-30 phút
	PhasePrematch MatchPhase = "prematch" // TRUOC 30 phút đá
	PhaseLive     MatchPhase = "live"     // đang đá
	PhaseEnded    MatchPhase = "ended"    // kết thúc
)

type Team struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
}

type Country struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Competition struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Tier     int      `json:"tier"`
	Country  Country  `json:"country"`
	Category Category `json:"category"`
}

type MatchDailyCatchFromRedis struct {
	MatchID     string      `json:"id"`
	HomeTeam    Team        `json:"home_team"`
	AwayTeam    Team        `json:"away_team"`
	Competition Competition `json:"competition"`
	MatchTime   int         `json:"match_time"`
	Date        string      `json:"date"`
}

//Tong hop trạng thái hiện tại của trận đấu, bao gồm thông tin cơ bản (đội, giải đấu, thời gian), trạng thái vòng đời (phase), tỉ số, sự kiện đã xảy ra
type MatchState struct {
	MatchID     string      `json:"match_id"`
	RoomID      string      `json:"room_id"`
	HomeTeam    Team        `json:"home_team"`
	AwayTeam    Team        `json:"away_team"`
	Competition Competition `json:"competition"`
	MatchTime   int64       `json:"match_time"` // Unix timestamp giờ kick-off
	Date        string      `json:"date"`       // "2026-03-11"

	Phase MatchPhase `json:"phase"`

	Minute    int          `json:"minute"`
	HomeScore int          `json:"home_score"`
	AwayScore int          `json:"away_score"`
	Events    []MatchEvent `json:"events"` // incidents trong trận

	UpdatedAt int64 `json:"updated_at"`
}
