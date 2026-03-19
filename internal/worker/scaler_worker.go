package worker

import (
	"context"
	"log"
	"time"
	"uniscore-seeding-bot/internal/domain/model"
	"uniscore-seeding-bot/internal/domain/service"
	"uniscore-seeding-bot/internal/usecase/seeding"
)

// ScalerWorker periodically recommends bot pressure per match.
type ScalerWorker struct {
	store      service.ContextStore
	autoScaler *seeding.AutoScaler
	interval   time.Duration
	logger     *log.Logger
}

func NewScalerWorker(store service.ContextStore, autoScaler *seeding.AutoScaler, interval time.Duration, logger *log.Logger) *ScalerWorker {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	return &ScalerWorker{store: store, autoScaler: autoScaler, interval: interval, logger: logger}
}

func (w *ScalerWorker) Start(ctx context.Context) {
	w.logger.Printf("📈 [SCALER] started interval=%s", w.interval)
	w.tick(ctx)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.tick(ctx)
		case <-ctx.Done():
			w.logger.Println("📈 [SCALER] stopped")
			return
		}
	}
}

func (w *ScalerWorker) tick(ctx context.Context) {
	matches, err := w.store.GetAllTodayMatches(ctx)
	if err != nil {
		w.logger.Printf("📈 [SCALER] get matches failed: %v", err)
		return
	}

	now := time.Now()
	processed := 0
	prematchCount := 0
	liveCount := 0
	totalActiveUsers := 0
	for _, m := range matches {
		phase := ComputePhase(int64(m.MatchTime), now)
		if phase != model.PhaseLive && phase != model.PhasePrematch {
			continue
		}
		processed++
		if phase == model.PhasePrematch {
			prematchCount++
		} else if phase == model.PhaseLive {
			liveCount++
		}

		chatCtx, err := w.store.GetRecentChatWindow(m.MatchID, 200)
		if err != nil {
			continue
		}

		activeUsers := countRecentHumanMessages(chatCtx.RawMessages)
		totalActiveUsers += activeUsers
		_ = w.autoScaler.RecommendBots(activeUsers)
	}

	if processed > 0 {
		w.logger.Printf("📈 [SCALER] matches=%d prematch=%d live=%d total_active_users=%d",
			processed,
			prematchCount,
			liveCount,
			totalActiveUsers,
		)
	}
}

func countRecentHumanMessages(messages []model.ChatMessage) int {
	count := 0
	for _, msg := range messages {
		if !msg.IsBot {
			count++
		}
	}
	return count
}
