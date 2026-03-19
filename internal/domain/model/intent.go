package model

type UserMessage struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Content  string `json:"content"`
}

type DetectIntent struct {
	Sentiment     string   `json:"sentiment"`
	Language      string   `json:"language"`
	TeamBias      string   `json:"team_bias"`
	MainTopic     []string `json:"main_topic"`
	RequiresReply bool     `json:"requires_reply"`
}
