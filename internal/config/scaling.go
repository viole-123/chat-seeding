package config

type ScalerState string

const (
	ScalerStateLow    ScalerState = "low"
	ScalerStateMedium ScalerState = "medium"
	ScalerStateHigh   ScalerState = "high"
	ScalerStatePeak   ScalerState = "peak"
)

type MaxBotsConfig struct {
	MinBots int         `json:"min_bots"`
	MaxBots int         `json:"max_bots"`
	State   ScalerState `json:"state"`
}

func (c MaxBotsConfig) BotCountFor() int {
	if c.MinBots > c.MaxBots {
		return c.MinBots
	}
	return c.MinBots
}
