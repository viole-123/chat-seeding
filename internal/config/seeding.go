package config

import (
	"fmt"
	"time"
)

// SeedingPolicy contains tunable seeding parameters.
type SeedingPolicy struct {
	MaxMessagesBot   int           `yaml:"max_messages_bot"`
	BotRatio         float64       `yaml:"bot_ratio"`
	Cooldown         time.Duration `yaml:"cooldown"`
	MaxEventsPerHour int           `yaml:"max_events_per_hour"`
	DepupWindow      time.Duration `yaml:"depup_window"`
	EnableKillSwitch bool          `yaml:"enable_kill_switch"`
}

func DefaultSeedingPolicy() SeedingPolicy {
	return SeedingPolicy{
		MaxMessagesBot:   100,
		BotRatio:         0.3,
		Cooldown:         60 * time.Second,
		MaxEventsPerHour: 1000,
		DepupWindow:      24 * time.Hour,
		EnableKillSwitch: true,
	}

}

func MaxMessagesBotEventType(eventType string) int {
	switch eventType {
	case "goal":
		return 5
	case "red_card":
		return 3
	default:
		return 2
	}
}

func (p SeedingPolicy) Validate() error {
	if p.MaxMessagesBot <= 0 {
		return fmt.Errorf("MaxMessagesBot must be > 0")
	}
	if p.BotRatio < 0 || p.BotRatio > 1 {
		return fmt.Errorf("BotRatio must be between 0 and 1")
	}
	if p.Cooldown <= 0 {
		return fmt.Errorf("Cooldown must be > 0")
	}
	return nil
}
