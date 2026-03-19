package publisher

import (
	"encoding/json"
	"log"
	"uniscore-seeding-bot/internal/domain/model"
)

type WebSocketPublisher struct {
	gatewayURL  string
	broadcaster func([]byte)
}

// init new websocket publish mới
func NewWebSocketPublisher(gatewayURL string) *WebSocketPublisher {
	return &WebSocketPublisher{
		gatewayURL: gatewayURL,
	}
}

// SetBroadcaster sets the broadcast function (called from app initialization)
func (w *WebSocketPublisher) SetBroadcaster(broadcaster func([]byte)) {
	w.broadcaster = broadcaster
}

func (w *WebSocketPublisher) Publish(msg model.ChatMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// If broadcaster is set, use it
	if w.broadcaster != nil {
		w.broadcaster(payload)
	} else {
		log.Printf("⚠️ [WEBSOCKET] No broadcaster set, message not sent: %d bytes", len(payload))
	}

	return nil
}
