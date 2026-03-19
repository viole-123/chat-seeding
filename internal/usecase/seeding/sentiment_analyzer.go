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
	if bundle.Audience.ChatVelocity > 0 {
		return string(model.SentimentNegative)
	}

	if len(bundle.Match.Events) > 0 {
		switch bundle.Match.Events[0].Type {
		case "goal":
			return string(model.SentimentPositive)
		case "red_card":
			return string(model.SentimentNegative)
		}
	}
	return string(model.SentimentNeutral)

}
