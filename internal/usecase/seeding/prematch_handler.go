package seeding

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
	"uniscore-seeding-bot/internal/adapter/mqtt"
	"uniscore-seeding-bot/internal/domain/model"
	"uniscore-seeding-bot/internal/domain/service"
	"uniscore-seeding-bot/internal/pkg/retry"
)

type PrematchHandler struct {
	contextStore     service.ContextStore
	contextBuilder   *ContextBuilder
	messageGenerator *MessageGenerator
	personaSelector  *PersonaSelector
	publisher        service.PublisherService
	mqttPublisher    *mqtt.Publisher
	roomManager      *RoomManager
	messageRepo      interface{ SaveMessage(model.ChatMessage) error }
	logger           *log.Logger
	autoScaler       *AutoScalerLogic
}

func NewPrematchHandler(
	contextStore service.ContextStore,
	contextBuilder *ContextBuilder,
	messageGenerator *MessageGenerator,
	personaSelector *PersonaSelector,
	publisher service.PublisherService,
	mqttPublisher *mqtt.Publisher,
	roomManager *RoomManager,
	messageRepo interface{ SaveMessage(model.ChatMessage) error },
	logger *log.Logger,
	autoScaler *AutoScalerLogic,
) *PrematchHandler {
	return &PrematchHandler{
		contextStore:     contextStore,
		contextBuilder:   contextBuilder,
		messageGenerator: messageGenerator,
		personaSelector:  personaSelector,
		publisher:        publisher,
		mqttPublisher:    mqttPublisher,
		roomManager:      roomManager,
		messageRepo:      messageRepo,
		logger:           logger,
		autoScaler:       autoScaler,
	}
}

func (h *PrematchHandler) Handle(ctx context.Context, event model.MatchEvent) error {
	h.logger.Printf("│ │  🗓️  PREMATCH  handling match=%s", event.MatchID)

	// Lấy thông tin đội từ matches_daily
	daily, err := h.contextStore.GetMatchByID(ctx, event.MatchID)
	if err != nil || daily == nil {
		h.logger.Printf("│ │  ⚠️  PREMATCH  match not found in daily store: %v", err)
		if err != nil {
			return err
		}
		return fmt.Errorf("match not found in daily store: %s", event.MatchID)
	}

	// Khởi tạo match state ở phase prematch
	state := model.MatchState{
		MatchID:     event.MatchID,
		Phase:       model.PhasePrematch,
		MatchTime:   int64(daily.MatchTime),
		Date:        daily.Date,
		RoomID:      fmt.Sprintf("room-%s", event.MatchID),
		Minute:      event.Minute,
		HomeTeam:    daily.HomeTeam,
		AwayTeam:    daily.AwayTeam,
		Competition: daily.Competition,
		UpdatedAt:   time.Now().Unix(),
	}
	if err := h.contextStore.SetMatchState(event.MatchID, state); err != nil {
		h.logger.Printf("│ │  ⚠️  PREMATCH  set state failed (non-critical): %v", err)
	}

	roomID := fmt.Sprintf("room-%s", event.MatchID)
	if h.roomManager != nil {
		if resolvedRoomID, roomErr := h.roomManager.GetOrCreate(ctx, event.MatchID); roomErr != nil {
			h.logger.Printf("│ │  ⚠️  PREMATCH  room get/create failed: %v", roomErr)
		} else {
			roomID = resolvedRoomID
		}
	}
	usedTexts := map[string]bool{}

	// Generate 5 prematch messages; strict 10-second interval between messages.
	for i := range 5 {
		kickoffAt := time.Unix(int64(daily.MatchTime), 0)
		minutesToKickoff := int(time.Until(kickoffAt).Minutes())
		if minutesToKickoff < 0 {
			h.logger.Printf("│ │  ⏭️  PREMATCH  stop burst early (kickoff passed) match=%s", event.MatchID)
			break
		}

		ctxBundle, err := h.contextBuilder.BuildBundle(ctx, event.MatchID, roomID)

		if err != nil || ctxBundle == nil {
			h.logger.Printf("│ │  ⚠️  PREMATCH  bundle build failed [%d/5]: %v", i+1, err)
			if i < 4 {
				time.Sleep(10 * time.Second)
			}
			continue
		}

		ctxBundle.CurrentEvent = event
		ctxBundle.Match.Phase = model.PhasePrematch
		ctxBundle.Match.HomeTeam = daily.HomeTeam
		ctxBundle.Match.AwayTeam = daily.AwayTeam
		ctxBundle.Match.MatchID = event.MatchID
		ctxBundle.Match.Minute = -minutesToKickoff

		h.logger.Printf("│ │  ✅ PREMATCH  context ready [%d/5] match=%s phase=%s chat=%d",
			i+1,
			ctxBundle.Match.MatchID,
			ctxBundle.Match.Phase,
			len(ctxBundle.Chat.RawMessages),
		)

		persona, err := h.personaSelector.SelectPersona(ctx, *ctxBundle)
		if err != nil || persona == nil {
			persona = h.personaSelector.SelectPersonaAllowReuse(*ctxBundle)
			if persona == nil {
				h.logger.Printf("│ │  ❌ PREMATCH  persona error [%d/5]: %v", i+1, err)
				continue
			}
			h.logger.Printf("│ │  ♻️  PREMATCH  fallback persona=%s [%d/5]", persona.ID, i+1)
		}
		h.logger.Printf("│ │  🎭 PREMATCH  persona=%s [%d/5]", persona.ID, i+1)

		draftMsg, err := h.messageGenerator.GenerateMessage(ctxBundle, persona)
		if err != nil {
			h.logger.Printf("│ │  ❌ PREMATCH  generate failed [%d/5]: %v", i+1, err)
			if i < 4 {
				time.Sleep(10 * time.Second)
			}
			continue
		}

		draftMsg.Text = diversifyPrematchText(draftMsg.Text, persona, i, minutesToKickoff, usedTexts)
		h.logger.Printf("│ │  💬 PREMATCH  [%d/5] %s", i+1, draftMsg.Text)

		msg := model.ChatMessage{
			ID:        fmt.Sprintf("%s-pre-%d-%d", event.MatchID, i+1, time.Now().UnixNano()),
			Content:   draftMsg.Text,
			Timestamp: time.Now().Unix(),
			IsBot:     true,
			Persona:   persona.ID,
			MatchID:   event.MatchID,
			RoomID:    roomID,
			EventType: "MATCH_UPCOMING",
			CreatedAt: time.Now(),
		}

		h.logger.Printf("│ │  ⏰ PREMATCH  send [%d/5] now", i+1)

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
				EventType: msg.EventType,
			}
			err = h.mqttPublisher.PublishBotMessage(ctx, roomID, mqttMsg)
		} else {
			err = retry.Do(func() error {
				return h.publisher.Publish(msg)
			}, 3, time.Second)
		}
		if err != nil {
			h.logger.Printf("│ │  ❌ PREMATCH  publish failed [%d/5]: %v", i+1, err)
			if i < 4 {
				time.Sleep(10 * time.Second)
			}
			continue
		}

		h.logger.Printf("│ │  🚀 PREMATCH  published [%d/5] matchID=%s", i+1, event.MatchID)
		h.contextStore.PushChatMessage(roomID, msg)
		if h.messageRepo != nil {
			if err := h.messageRepo.SaveMessage(msg); err != nil {
				h.logger.Printf("│ │  ⚠️  PREMATCH  DB save failed [%d/5]: %v", i+1, err)
			}
		}

		if i < 4 {
			time.Sleep(10 * time.Second)
		}
	}

	h.logger.Printf("│ │  ✅ PREMATCH  done match=%s [%s vs %s]",
		event.MatchID, daily.HomeTeam.ShortName, daily.AwayTeam.ShortName)
	return nil
}

func diversifyPrematchText(base string, persona *model.Persona, idx int, minutesToKickoff int, used map[string]bool) string {
	text := strings.TrimSpace(base)
	if text == "" {
		text = "Cho xem trước trận nào!"
	}

	seed := ""
	if persona != nil && len(persona.Profile.SeedPhrases) > 0 {
		seed = strings.TrimSpace(persona.Profile.SeedPhrases[rand.Intn(len(persona.Profile.SeedPhrases))])
	}

	countdown := ""
	if minutesToKickoff > 0 {
		countdown = fmt.Sprintf("Con %d phut nua bong lan.", minutesToKickoff)
	} else {
		countdown = "Sap den gio bong lan!"
	}

	variants := []string{text}
	if seed != "" {
		variants = append(variants, fmt.Sprintf("%s %s", seed, text))
	}
	variants = append(variants,
		fmt.Sprintf("%s %s", text, countdown),
		fmt.Sprintf("%s %s", countdown, text),
	)

	for _, candidate := range variants {
		norm := normalizePrematchText(candidate)
		if !used[norm] {
			used[norm] = true
			return candidate
		}
	}

	fallback := fmt.Sprintf("%s (%d)", text, idx+1)
	used[normalizePrematchText(fallback)] = true
	return fallback
}

func normalizePrematchText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Join(strings.Fields(s), " ")
	return s
}
