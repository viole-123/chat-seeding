// internal/adapter/mqtt/consumer_bridge.go
package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

type ConsumerBridge struct {
	mqttClient pahomqtt.Client
	route      func(roomID string, payload []byte)
}

func NewConsumerBridge(brokerURL, clientID string, route func(roomID string, payload []byte)) (*ConsumerBridge, error) {
	opts := pahomqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(clientID). // khác với publisher clientID, ví dụ "ws-bridge-001"
		SetAutoReconnect(true)

	client := pahomqtt.NewClient(opts)
	token := client.Connect()
	token.WaitTimeout(10 * time.Second)
	if err := token.Error(); err != nil {
		return nil, err
	}

	return &ConsumerBridge{mqttClient: client, route: route}, nil
}

// Start — subscribe "room/+" để nhận TẤT CẢ message từ mọi room
// Wildcard "+" nghĩa là: room/match001, room/match002... đều nhận được
func (b *ConsumerBridge) Start(ctx context.Context) error {
	// "room/+" là MQTT single-level wildcard
	token := b.mqttClient.Subscribe("room/+", 1, b.handleMessage)
	token.WaitTimeout(5 * time.Second)
	if err := token.Error(); err != nil {
		return fmt.Errorf("MQTT subscribe failed: %w", err)
	}

	log.Printf("🔌 [Bridge] Subscribed to room/+ — bridging MQTT → WebSocket")

	// Block cho đến khi context cancel (server shutdown)
	<-ctx.Done()
	b.mqttClient.Disconnect(250)
	return nil
}

// handleMessage — xử lý mỗi tin nhắn đến từ MQTT
func (b *ConsumerBridge) handleMessage(_ pahomqtt.Client, msg pahomqtt.Message) {
	// Parse topic để lấy roomID: "room/match001" → "match001"
	// msg.Topic() trả về "room/room-match001"
	topic := msg.Topic()
	roomID := strings.TrimPrefix(topic, "room/")
	if roomID == "" {
		log.Printf("⚠️  [Bridge] empty room from topic=%s", topic)
		return
	}

	var chatMsg ChatMessage
	if err := json.Unmarshal(msg.Payload(), &chatMsg); err != nil {
		log.Printf("⚠️  [Bridge] Unmarshal failed topic=%s: %v", topic, err)
		return
	}
	if chatMsg.RoomID == "" {
		chatMsg.RoomID = roomID
	}
	if chatMsg.MatchID == "" {
		chatMsg.MatchID = strings.TrimPrefix(roomID, "room-")
	}
	normalized, err := json.Marshal(chatMsg)
	if err != nil {
		log.Printf("⚠️  [Bridge] Marshal normalized failed room=%s: %v", roomID, err)
		return
	}

	// Broadcast tới tất cả WS clients đang ở room này
	if b.route != nil {
		b.route(roomID, normalized)
	}

	log.Printf("🔁 [Bridge] Routed to room=%s is_bot=%v", roomID, chatMsg.IsBot)
}
