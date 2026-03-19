package model

type Sentiment string

const (
	SentimentPositive Sentiment = "positive"
	SentimentNegative Sentiment = "negative"
	SentimentNeutral  Sentiment = "neutral"
)
