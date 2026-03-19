package handler

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"
	"uniscore-seeding-bot/internal/domain/model"
	"uniscore-seeding-bot/internal/domain/service"
	"uniscore-seeding-bot/internal/worker"

	"github.com/gin-gonic/gin"
)

type ActiveMatchesHandler struct {
	contextStore service.ContextStore
}

type ActiveMatchItem struct {
	MatchID    string           `json:"match_id"`
	RoomID     string           `json:"room_id"`
	Phase      model.MatchPhase `json:"phase"`
	KickoffAt  int64            `json:"kickoff_at"`
	HomeTeam   string           `json:"home_team"`
	AwayTeam   string           `json:"away_team"`
	LeagueName string           `json:"league_name"`
}

func NewActiveMatchesHandler(contextStore service.ContextStore) *ActiveMatchesHandler {
	return &ActiveMatchesHandler{contextStore: contextStore}
}

func (h *ActiveMatchesHandler) List(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	if c.Request.Method == http.MethodOptions {
		c.Status(http.StatusNoContent)
		return
	}

	matches, err := h.contextStore.GetAllTodayMatches(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load matches"})
		return
	}

	includeWaiting := true
	if raw := c.Query("include_waiting"); raw != "" {
		parsed, parseErr := strconv.ParseBool(raw)
		if parseErr == nil {
			includeWaiting = parsed
		}
	}

	now := time.Now()
	items := make([]ActiveMatchItem, 0, len(matches))
	phaseCounts := map[model.MatchPhase]int{
		model.PhaseWaiting:  0,
		model.PhasePrematch: 0,
		model.PhaseLive:     0,
	}

	for _, m := range matches {
		phase := worker.ComputePhase(int64(m.MatchTime), now)
		if phase == model.PhaseEnded {
			continue
		}
		if phase == model.PhaseWaiting && !includeWaiting {
			continue
		}
		phaseCounts[phase]++

		home := m.HomeTeam.ShortName
		if home == "" {
			home = m.HomeTeam.Name
		}
		away := m.AwayTeam.ShortName
		if away == "" {
			away = m.AwayTeam.Name
		}

		items = append(items, ActiveMatchItem{
			MatchID:    m.MatchID,
			RoomID:     fmt.Sprintf("room-%s", m.MatchID),
			Phase:      phase,
			KickoffAt:  int64(m.MatchTime),
			HomeTeam:   home,
			AwayTeam:   away,
			LeagueName: m.Competition.Name,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		phaseRank := func(p model.MatchPhase) int {
			switch p {
			case model.PhaseLive:
				return 0
			case model.PhasePrematch:
				return 1
			default:
				return 2
			}
		}
		ri := phaseRank(items[i].Phase)
		rj := phaseRank(items[j].Phase)
		if ri != rj {
			return ri < rj
		}
		return items[i].KickoffAt < items[j].KickoffAt
	})

	c.JSON(http.StatusOK, gin.H{
		"count":   len(items),
		"summary": gin.H{
			"live":     phaseCounts[model.PhaseLive],
			"prematch": phaseCounts[model.PhasePrematch],
			"waiting":  phaseCounts[model.PhaseWaiting],
		},
		"matches": items,
	})
}
