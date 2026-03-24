package seeding

import (
	"context"
	"log"
	"math/rand"
	"uniscore-seeding-bot/internal/config"
	"uniscore-seeding-bot/internal/domain/service"
)

type AutoScalerLogic struct {
	scalerSvc service.AutoScalerService
}

func NewAutoScalerLogic(autoscalerService service.AutoScalerService) *AutoScalerLogic {
	return &AutoScalerLogic{
		scalerSvc: autoscalerService,
	}
}

func (u *AutoScalerLogic) ShouldSpawnBot(ctx context.Context, roomID string, currentBotCount int) (bool, config.MaxBotsConfig, error) {
	ctf, err := u.scalerSvc.GetBotConfig(ctx, roomID)
	if err != nil {
		log.Printf("⚠️  [AutoScale] GetBotConfig failed room=%s: %v — using fallback", roomID, err)
		return currentBotCount < 2, config.MaxBotsConfig{MinBots: 1, MaxBots: 2}, nil
	}
	should := currentBotCount - ctf.MaxBots
	log.Printf("🤖 [AutoScale] room=%s state=%s online→bots %d/%d spawn=%v",
		roomID, ctf.State, currentBotCount, ctf.MaxBots, should)
	return should < 0, ctf, nil

}

func (u *AutoScalerLogic) RandomBotCount(cfg config.MaxBotsConfig) int {
	if cfg.MaxBots <= cfg.MinBots {
		return cfg.MinBots
	}
	return cfg.MinBots + rand.Intn(cfg.MaxBots-cfg.MinBots+1)
}
