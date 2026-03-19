package seeding

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"
	"uniscore-seeding-bot/internal/adapter/kafka"
	"uniscore-seeding-bot/internal/adapter/mqtt"
	"uniscore-seeding-bot/internal/domain/model"
	domainrepo "uniscore-seeding-bot/internal/domain/repository"
	"uniscore-seeding-bot/internal/domain/service"
	"uniscore-seeding-bot/internal/observability/metrics"
	"uniscore-seeding-bot/internal/pkg/retry"
	"uniscore-seeding-bot/internal/usecase/safety"

	"github.com/IBM/sarama"
)

type EventHandler struct {
	policyChecker    PolicyChecker
	personaSelector  *PersonaSelector
	logger           *log.Logger
	contextStore     service.ContextStore
	contextBuilder   *ContextBuilder
	messageGenerator *MessageGenerator
	qualityFilter    *QualityFilter
	publisher        service.PublisherService
	mqttPublisher    *mqtt.Publisher
	roomManager      *RoomManager
	redisAuth        service.DedupService
	messageRepo      domainrepo.MessageRepository
	shadowBanService *safety.ShadowBanService
}

// Init event handler with dependencies
func NewEventHandler(policyChecker PolicyChecker, personaSelector *PersonaSelector, contextBuilder *ContextBuilder, contextStore service.ContextStore, messageGenerator *MessageGenerator, qualityFilter *QualityFilter, logger *log.Logger, publisher service.PublisherService, mqttPublisher *mqtt.Publisher, roomManager *RoomManager, redisAuth service.DedupService, messageRepo domainrepo.MessageRepository) *EventHandler {
	return &EventHandler{
		policyChecker:    policyChecker,
		personaSelector:  personaSelector,
		contextBuilder:   contextBuilder,
		contextStore:     contextStore,
		messageGenerator: messageGenerator,
		qualityFilter:    qualityFilter,
		logger:           logger,
		publisher:        publisher,
		mqttPublisher:    mqttPublisher,
		roomManager:      roomManager,
		redisAuth:        redisAuth,
		messageRepo:      messageRepo,
	}
}

func (h *EventHandler) SetShadowBanService(shadowBanService *safety.ShadowBanService) {
	h.shadowBanService = shadowBanService
}

func (s *EventHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *EventHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.logger.Println("👉 Consumer group session cleanup")
	return nil
}

func (h *EventHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()
	for message := range claim.Messages() {
		h.logger.Printf("┌────────────────────────────────── 🔍 KAFKA MSG offset=%d partition=%d ───────────────────",
			message.Offset, message.Partition)

		bundle, err := kafka.MapBytesToEventsBundle(message.Value)
		if err != nil {
			h.logger.Printf("│ ❌ PARSE ERROR  %v", err)
			h.logger.Printf("└──────────────────────────────────────────────────── SKIP ────────")
			session.MarkMessage(message, "")
			continue
		}
		h.logger.Printf("✅✅✅ PARSED  %d events", len(bundle.MatchEvents))

		total := len(bundle.MatchEvents)
		for i, event := range bundle.MatchEvents {
			if err := h.handleEvent(ctx, i+1, total, event, "kafka"); err != nil {
				h.logger.Printf("│ │  ⚠️  HANDLE   event failed: %v", err)
			}
		}

		h.logger.Printf("│")
		session.MarkMessage(message, "")
		h.logger.Printf("└──────────────────────────────── ✅ COMMITTED offset=%d ────────────────────────────────",
			message.Offset)
	}
	return nil
}

func (h *EventHandler) updateMatchState(event model.MatchEvent) error {
	ctx := context.Background()

	phase := model.PhaseLive
	homeTeam := model.Team{Name: event.HomeTeam, ShortName: event.HomeTeam}
	awayTeam := model.Team{Name: event.AwayTeam, ShortName: event.AwayTeam}

	if daily, err := h.contextStore.GetMatchByID(ctx, event.MatchID); err == nil && daily != nil {
		homeTeam = daily.HomeTeam
		awayTeam = daily.AwayTeam
		league := daily.Competition.Name
		diff := time.Now().Unix() - int64(daily.MatchTime)
		switch {
		case diff >= 0:
			phase = model.PhaseLive
		case diff >= -30*60:
			phase = model.PhasePrematch
		default:
			phase = model.PhaseWaiting
		}
		h.logger.Printf("│ │  🔍 ENRICH   [%s vs %s] phase=%s league=%s",
			homeTeam.ShortName, awayTeam.ShortName, phase, league)
	}

	state := model.MatchState{
		Competition: model.Competition{
			ID:   event.LeagueID,
			Name: "",
		},
		MatchID:   event.MatchID,
		Phase:     phase,
		Minute:    event.Minute,
		HomeScore: event.HomeScore,
		AwayScore: event.AwayScore,
		HomeTeam:  homeTeam,
		AwayTeam:  awayTeam,
		UpdatedAt: time.Now().Unix(),
	}
	return h.contextStore.SetMatchState(event.MatchID, state)
}

func (h *EventHandler) saveBotMessageToDB(msg model.ChatMessage) error {
	if h.messageRepo == nil {
		return fmt.Errorf("message repository is not initialized")
	}
	return h.messageRepo.SaveMessage(msg)
}

func (h *EventHandler) enrichEventFromRedis(ctx context.Context, event *model.MatchEvent) error {
	daily, err := h.contextStore.GetMatchByID(ctx, event.MatchID)
	if err != nil {
		return err
	}
	if daily == nil {
		return nil
	}

	if event.LeagueID == "" || event.LeagueID == "UNKNOWN" {
		event.LeagueID = daily.Competition.Name
		if event.LeagueID == "" {
			event.LeagueID = daily.Competition.ID
		}
	}

	if event.HomeTeam == "" {
		event.HomeTeam = daily.HomeTeam.ShortName
		if event.HomeTeam == "" {
			event.HomeTeam = daily.HomeTeam.Name
		}
	}

	if event.AwayTeam == "" {
		event.AwayTeam = daily.AwayTeam.ShortName
		if event.AwayTeam == "" {
			event.AwayTeam = daily.AwayTeam.Name
		}
	}

	return nil
}

func (h *EventHandler) Handle(ctx context.Context, event model.MatchEvent) error {
	return h.handleEvent(ctx, 1, 1, event, "prematch")
}

func (h *EventHandler) handleEvent(ctx context.Context, index int, total int, event model.MatchEvent, source string) error {
	metrics.Inc("seeding_events_total", map[string]string{"event_type": event.Type})
	h.logger.Printf("│")
	h.logger.Printf("│ ┌─ EVENT [%d/%d] %-12s │ match=%-8s │ player=%-15s │ min=%d │ source=%s",
		index, total, event.Type, event.MatchID, event.PlayerName, event.Minute, source)

	if err := h.enrichEventFromRedis(ctx, &event); err != nil {
		h.logger.Printf("│ │  ⚠️  ENRICH   redis lookup failed: %v", err)
	}

	if event.Type != "MATCH_UPCOMING" {
		pass, err := h.policyChecker.CheckPolicy(ctx, event)
		if err != nil {
			h.logger.Printf("│ │  ❌ POLICY   ERROR: %v", err)
			h.logger.Printf("│ └──────────────────────────────────────────────────────────────")
			return fmt.Errorf("policy check failed: %w", err)
		}
		if !pass {
			h.logger.Printf("│ │⏭️  POLICY   SKIP")
			h.logger.Printf("│ └──────────────────────────────────────────────────────────────")
			return nil
		}
		h.logger.Printf("│ │  ✅ POLICY   PASS")
	} else {
		h.logger.Printf("│ │  ✅ POLICY   BYPASS (prematch)")
		if event.Minute == 0 {
			event.Minute = -1
		}
	}

	if err := h.updateMatchState(event); err != nil {
		h.logger.Printf("│ │  ⚠️  REDIS    hset failed (non-critical)")
	}
	h.contextStore.PushEvent(event.MatchID, event)

	roomID := fmt.Sprintf("room-%s", event.MatchID)
	if h.roomManager != nil {
		if resolvedRoomID, roomErr := h.roomManager.GetOrCreate(ctx, event.MatchID); roomErr != nil {
			h.logger.Printf("│ │  ⚠️  ROOM     get/create failed: %v", roomErr)
		} else {
			roomID = resolvedRoomID
		}
	}
	ctxBundle, err := h.contextBuilder.BuildBundle(ctx, event.MatchID, roomID)
	if err != nil {
		h.logger.Printf("│ │  ⚠️  BUNDLE   build failed: %v", err)
	}
	if ctxBundle == nil {
		ctxBundle = &model.ContextBundle{
			Match: model.MatchState{MatchID: event.MatchID},
		}
	}
	ctxBundle.CurrentEvent = event
	ctxBundle.Match.Events = []model.MatchEvent{event}

	persona, err := h.personaSelector.SelectPersona(ctx, *ctxBundle)
	if err != nil || persona == nil {
		persona = h.personaSelector.SelectPersonaAllowReuse(*ctxBundle)
		if persona != nil {
			h.logger.Printf("│ │  ♻️  PERSONA  fallback allow-reuse=%s event=%s", persona.ID, event.Type)
			err = nil
		}
	}
	if err != nil || persona == nil {
		h.logger.Printf("│ │  ❌ PERSONA  ERROR: %v", err)
		h.logger.Printf("│ └──────────────────────────────────────────────────────────────")
		if err != nil {
			return fmt.Errorf("persona selection failed: %w", err)
		}
		return fmt.Errorf("persona selection failed: no eligible persona")
	}
	h.logger.Printf("│ │  🎭 PERSONA  %s (%s/%v)", persona.ID, persona.Profile.Tone, persona.Profile.Language)

	draftMsg, err := h.messageGenerator.GenerateMessage(ctxBundle, persona)
	if err != nil {
		h.logger.Printf("│ │  ❌ GENERATE %v", err)
		h.logger.Printf("│ └──────────────────────────────────────────────────────────────")
		return fmt.Errorf("message generation failed: %w", err)
	}
	msgSource := draftMsg.Meta["source"]
	if msgSource == "" {
		msgSource = "unknown"
	}
	templateID := draftMsg.Meta["template_id"]
	if templateID == "" {
		templateID = "-"
	}
	h.logger.Printf("│ │  📝 SOURCE   %s (template=%s)", msgSource, templateID)
	h.logger.Printf("│ │  💬 MESSAGE  %s", draftMsg.Text)

	qualityResult, err := h.qualityFilter.Check(ctx, event, draftMsg, ctxBundle)
	if err != nil || !qualityResult.IsPass {
		metrics.Inc("seeding_quality_filter_rejected_total", map[string]string{"event_type": event.Type})
		h.logger.Printf("│ │  ❌ QUALITY  FAIL %v", err)
		h.logger.Printf("│ └──────────────────────────────────────────────────────────────")
		if err != nil {
			return fmt.Errorf("quality check failed: %w", err)
		}
		return fmt.Errorf("quality check rejected message")
	}
	h.logger.Printf("│ │  ✅ QUALITY  PASS")

	msg := model.ChatMessage{
		ID:        fmt.Sprintf("%s-%d", event.MatchID, time.Now().UnixNano()),
		Content:   draftMsg.Text,
		Timestamp: time.Now().Unix(),
		IsBot:     true,
		Persona:   persona.ID,
		MatchID:   event.MatchID,
		RoomID:    fmt.Sprintf("room-%s", event.MatchID),
		EventType: event.Type,
		CreatedAt: time.Now(),
	}

	delaySeconds := 1 + rand.Intn(13)
	h.logger.Printf("│ │  ⏰ DELAY    %ds → publishing...", delaySeconds)
	time.Sleep(time.Duration(delaySeconds) * time.Second)

	isShadowed := false
	if h.shadowBanService != nil {
		shadowed, shadowErr := h.shadowBanService.IsShadowed(ctx, event.MatchID, event.LeagueID)
		if shadowErr != nil {
			h.logger.Printf("│ │  ⚠️  SHADOW   check failed: %v", shadowErr)
		} else {
			isShadowed = shadowed
		}
	}

	if isShadowed {
		metrics.Inc("seeding_shadowban_skipped_total", map[string]string{"event_type": event.Type})
		h.logger.Printf("│ │  👻 SHADOW   enabled, skip external publish")
		h.contextStore.PushChatMessage(roomID, msg)
		if err := h.saveBotMessageToDB(msg); err != nil {
			h.logger.Printf("│ │  ⚠️  DB SAVE  failed: %v", err)
		}
	} else {
		if h.mqttPublisher != nil {
			mqttMsg := mqtt.ChatMessage{
				ID:        msg.ID,
				MatchID:   msg.MatchID,
				RoomID:    roomID,
				UserID:    persona.ID,
				Content:   msg.Content,
				Timestamp: msg.Timestamp,
				IsBot:     true,
				PersonaID: persona.ID,
				EventType: event.Type,
			}
			err = h.mqttPublisher.PublishBotMessage(ctx, roomID, mqttMsg)
		} else {
			err = retry.Do(func() error {
				return h.publisher.Publish(msg)
			}, 3, time.Second)
		}

		if err != nil {
			h.logger.Printf("│ │  ❌ PUBLISH  FAILED after 3 retries: %v", err)
			return fmt.Errorf("publish failed: %w", err)
		} else {
			metrics.Inc("seeding_messages_published_total", map[string]string{"event_type": event.Type})
			h.logger.Printf("│ │  🚀 PUBLISH  ✅ sent")
			h.contextStore.PushChatMessage(roomID, msg)
			_ = h.policyChecker.rateLimitService.IncrTotalMessages(ctx, msg.MatchID)

			if err := h.saveBotMessageToDB(msg); err != nil {
				h.logger.Printf("│ │  ⚠️  DB SAVE  failed: %v", err)
			} else {
				h.logger.Printf("│ │  💾 DB SAVE  ✅ %s", msg.ID)
			}
		}
	}
	h.logger.Printf("│ └──────────────────────────────────────────────────────────────")
	return nil
}
