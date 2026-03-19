package worker

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"uniscore-seeding-bot/internal/domain/model"
	"uniscore-seeding-bot/internal/domain/service"
)

// PrematchHandler là interface để poller gọi trực tiếp mà không cần import usecase.
type PrematchHandler interface {
	Handle(ctx context.Context, event model.MatchEvent) error
}

// PrematchPoller quét redis_matches mỗi interval, phát hiện trận sắp đấu
// và gọi thẳng PrematchHandler để xử lý (không cần qua Kafka).
type PrematchPoller struct {
	handler      PrematchHandler
	interval     time.Duration
	logger       *log.Logger
	contextStore service.ContextStore
}

func NewPrematchPoller(
	handler PrematchHandler,
	interval time.Duration,
	logger *log.Logger,
	contextStore service.ContextStore,
) *PrematchPoller {
	return &PrematchPoller{
		handler:      handler,
		interval:     interval,
		logger:       logger,
		contextStore: contextStore,
	}
}

func (p *PrematchPoller) Start(ctx context.Context) {
	p.logger.Printf("🗓️  [PREMATCH POLLER] Started, interval=%s", p.interval)
	p.poll(ctx)
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.poll(ctx)
		case <-ctx.Done():
			p.logger.Println("🗓️  [PREMATCH POLLER] Stopped")
			return
		}
	}
}

func (p *PrematchPoller) poll(ctx context.Context) {
	matches, err := p.contextStore.GetAllTodayMatches(ctx)
	if err != nil {
		p.logger.Printf("🗓️  [PREMATCH POLLER] ❌ scan redis_matches failed: %v", err)
		return
	}
	if len(matches) == 0 {
		p.logger.Printf("🗓️  [PREMATCH POLLER] ⚠️  No matches found in redis_matches")
		return
	}
	p.logger.Printf("🗓️  [PREMATCH POLLER] ✅ Loaded %d matches from redis_matches", len(matches))

	now := time.Now()
	countPhasePrematch := 0
	countPhaseLive := 0
	countPhaseEnded := 0
	countPhaseWaiting := 0
	var prematchList []string
	var liveList []string

	for _, m := range matches {
		phase := ComputePhase(int64(m.MatchTime), now)
		// In dữ liệu các trận đấu map qua phase ,có thời gian,kickoff
		// p.logger.Printf("🗓️  match=%-20s phase=%-10s kickoff=%s [%s vs %s]",
		// 	m.MatchID, phase,
		// 	time.Unix(int64(m.MatchTime), 0).Format("15:04 2006-01-02"),
		// 	m.HomeTeam.ShortName, m.AwayTeam.ShortName)

		switch phase {
		case model.PhasePrematch:
			countPhasePrematch++
			prematchList = append(prematchList, fmt.Sprintf("%s [%s vs %s]", m.MatchID, m.HomeTeam.ShortName, m.AwayTeam.ShortName))
		case model.PhaseLive:
			countPhaseLive++
			liveList = append(liveList, fmt.Sprintf("%s [%s vs %s]", m.MatchID, m.HomeTeam.ShortName, m.AwayTeam.ShortName))
		case model.PhaseEnded:
			countPhaseEnded++
		case model.PhaseWaiting:
			countPhaseWaiting++
		}

		if phase != model.PhasePrematch {
			continue
		}

		sent, err := p.contextStore.HasSentPrematch(ctx, m.MatchID)
		if err != nil {
			p.logger.Printf("🗓️  ⚠️  check sent failed match=%s: %v", m.MatchID, err)
			continue
		}
		if sent {
			roomID := fmt.Sprintf("room-%s", m.MatchID)
			chatCtx, chatErr := p.contextStore.GetRecentChatWindow(roomID, 1)
			if chatErr != nil {
				p.logger.Printf("🗓️  ⚠️  check chat failed match=%s: %v", m.MatchID, chatErr)
				continue
			}
			if len(chatCtx.RawMessages) > 0 {
				p.logger.Printf("🗓️  ⏭️  match=%s already sent prematch, skip", m.MatchID)
				continue
			}
			p.logger.Printf("🗓️  🔁 stale sent marker found (no chat), retry match=%s", m.MatchID)
		}

		// Gọi trực tiếp PrematchHandler (không qua Kafka)
		event := model.MatchEvent{
			MatchID:   m.MatchID,
			Type:      "MATCH_UPCOMING",
			Timestamp: time.Now().Unix(),
		}
		matchID := m.MatchID
		home := m.HomeTeam.ShortName
		away := m.AwayTeam.ShortName
		p.logger.Printf("🗓️  🚀 PREMATCH triggering handler match=%s [%s vs %s]",
			matchID, home, away)
		go func(evt model.MatchEvent) {
			if handleErr := p.handler.Handle(ctx, evt); handleErr != nil {
				p.logger.Printf("🗓️  ❌ PREMATCH handler failed match=%s: %v", matchID, handleErr)
				return
			}
			if markErr := p.contextStore.MarkSentPrematch(ctx, matchID); markErr != nil {
				p.logger.Printf("🗓️  ⚠️  mark sent failed match=%s: %v", matchID, markErr)
				return
			}
			p.logger.Printf("🗓️  ✅ MATCH_UPCOMING handled+marked match=%s [%s vs %s]", matchID, home, away)
		}(event)
	}

	if len(prematchList) > 0 {
		p.logger.Printf("🗓️  [PREMATCH POLLER] Prematch matches: %s", strings.Join(prematchList, ", "))

	}
	if len(liveList) > 0 {
		p.logger.Printf("🗓️  [PREMATCH POLLER] Live matches: %s", strings.Join(liveList, ", "))
		p.logger.Printf("WAITING FOR KAFKA MESSAGE FROM LIVE MATCHES !")
	}

	p.logger.Printf("🗓️  [PREMATCH POLLER] 📊 Phase summary — waiting=%d | prematch=%d | live=%d | ended=%d",
		countPhaseWaiting, countPhasePrematch, countPhaseLive, countPhaseEnded)
}

// matchEndedThreshold là thời gian tối đa tính từ kick-off để coi trận đã kết thúc
// (90 phút + 20 phút bù giờ/penalty = 110 phút).
const matchEndedThreshold = 110 * 60

func ComputePhase(matchTime int64, now time.Time) model.MatchPhase {
	diff := now.Unix() - matchTime // âm = chưa đến giờ đá
	switch {
	case diff >= matchEndedThreshold:
		return model.PhaseEnded // quá 110 phút sau kick-off → kết thúc
	case diff >= 0:
		return model.PhaseLive
	case diff >= -30*60:
		return model.PhasePrematch // trong cửa sổ 30 phút trước kick-off , chỉnh trước trận đấu ở đây
	default:
		return model.PhaseWaiting
	}
}
