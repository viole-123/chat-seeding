// internal/adapter/mqtt/publisher.go
package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

// ChatMessage là format chuẩn — dùng chung giữa chat-api và seeding bot
// Client phân biệt bot/user qua field is_bot
type ChatMessage struct {
	ID        string `json:"id,omitempty"`
	MatchID   string `json:"match_id,omitempty"`
	RoomID    string `json:"room_id"`
	UserID    string `json:"user_id,omitempty"` // personaID nếu is_bot=true
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
	IsBot     bool   `json:"is_bot"`
	PersonaID string `json:"persona_id,omitempty"`
	EventType string `json:"event_type,omitempty"`
}

type Publisher struct {
	client pahomqtt.Client
}

func NewPublisher(brokerURL, clientID string) (*Publisher, error) {
	opts := pahomqtt.NewClientOptions().
		AddBroker(brokerURL). // ví dụ: "tcp://localhost:1883"
		SetClientID(clientID).
		SetKeepAlive(30 * time.Second).
		SetAutoReconnect(true). // tự reconnect khi broker restart
		SetConnectionLostHandler(func(_ pahomqtt.Client, err error) {
			log.Printf("⚠️  [MQTT] Connection lost: %v", err)
		})

	client := pahomqtt.NewClient(opts)
	token := client.Connect()
	token.WaitTimeout(10 * time.Second)
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("MQTT connect failed: %w", err)
	}

	log.Printf("✅ [MQTT] Connected to %s as %s", brokerURL, clientID)
	return &Publisher{client: client}, nil
}

func (p *Publisher) Publish(ctx context.Context, roomID string, msg ChatMessage) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if roomID == "" {
		return fmt.Errorf("roomID is required")
	}

	topic := fmt.Sprintf("room/%s", roomID) // khớp với chat-api topic
	msg.RoomID = roomID
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().Unix()
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	// QoS=1: đảm bảo ít nhất 1 lần deliver, không mất tin khi network glitch
	token := p.client.Publish(topic, 1, false, data)
	token.WaitTimeout(5 * time.Second)
	if err := token.Error(); err != nil {
		return fmt.Errorf("MQTT publish to %s: %w", topic, err)
	}

	log.Printf("📤 [MQTT] Published to %s: %.80s", topic, string(data))
	return nil
}

// PublishBotMessage — seeding bot gọi cái này sau khi quality filter pass
func (p *Publisher) PublishBotMessage(ctx context.Context, roomID string, msg ChatMessage) error {
	msg.IsBot = true
	return p.Publish(ctx, roomID, msg)
}

func (p *Publisher) PublishUserMessage(ctx context.Context, roomID string, msg ChatMessage) error {
	msg.IsBot = false
	return p.Publish(ctx, roomID, msg)
}
