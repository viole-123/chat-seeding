package seeding

import (
	"uniscore-seeding-bot/internal/adapter/vllm"
	"uniscore-seeding-bot/internal/domain/model"
)

type SentimentAnalyzer struct {
	model           *vllm.VLLMGateway
	personaSelector *PersonaSelector
	ReadJSONFile    func(path string) ([]byte, error)
}

func NewSentimentAnalyzer(model *vllm.VLLMGateway, personaSelector *PersonaSelector) *SentimentAnalyzer {
	return &SentimentAnalyzer{
		model:           model,
		personaSelector: personaSelector,
	}
}

func (s *SentimentAnalyzer) AnalyzeSentiment(bundle model.ContextBundle) string {
	if bundle.Audience.ChatVelocity > 5.0 {
		return string(model.SentimentPositive)
	}

	if len(bundle.Match.Events) > 0 {
		switch bundle.Match.Events[0].Type {
		case "GOAL", "PENALTY", "OWN_GOAL":
			return string(model.SentimentPositive)
		case "RED_CARD":
			return string(model.SentimentNegative)
		}
	}
	return string(model.SentimentNeutral)

}
