package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 30 * time.Second // Increase write timeout
	pongWait       = 90 * time.Second // Increase pong wait
	pingPeriod     = 45 * time.Second // Send ping every 45s
	maxMessageSize = 512 * 1024       // 512KB for large messages
)

const messageBufferSize = 500 // Keep last 500 messages in memory

type outboundMessage struct {
	roomID string
	data   []byte
}

var (
	// Store all active WebSocket connections and room subscriptions.
	// `*` means receive all rooms.
	clients     = make(map[*websocket.Conn]map[string]bool)
	clientsLock sync.RWMutex

	// Channel to broadcast messages to clients, optionally constrained by room.
	broadcast = make(chan outboundMessage, 200)

	userMessageHandler func(IncomingUserMessage)

	// Message buffer for replay on reconnect
	messageBuffer     []outboundMessage
	messageBufferLock sync.RWMutex

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for testing
		},
	}
)

type IncomingUserMessage struct {
	MatchID string `json:"match_id"`
	RoomID  string `json:"room_id"`
	Content string `json:"content"`
}

func SetUserMessageHandler(fn func(IncomingUserMessage)) {
	userMessageHandler = fn
}

// Start the broadcaster goroutine
func init() {
	messageBuffer = make([]outboundMessage, 0, messageBufferSize)
	go handleBroadcast()
}

// WebSocketHandler handles WebSocket connections
func WebSocketHandler(c *gin.Context) {
	roomID := strings.TrimSpace(c.Query("room_id"))

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ Failed to upgrade connection: %v", err)
		return
	}

	subs := map[string]bool{}
	if roomID == "" {
		subs["*"] = true
	} else {
		subs[roomID] = true
	}

	// Register new client alongside existing ones (supports multiple concurrent viewers)
	clientsLock.Lock()
	clients[conn] = subs
	clientsLock.Unlock()

	log.Printf("✅ New WebSocket client connected. Total clients: %d", len(clients))

	// Replay buffered messages so client doesn't miss anything.
	// If client subscribes to one room, replay only messages from that room.
	messageBufferLock.RLock()
	buffered := make([]outboundMessage, len(messageBuffer))
	copy(buffered, messageBuffer)
	messageBufferLock.RUnlock()

	for _, msg := range buffered {
		if !isSubscribed(subs, msg.roomID) {
			continue
		}
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		_ = conn.WriteMessage(websocket.TextMessage, msg.data)
	}
	if len(buffered) > 0 {
		log.Printf("📼 Replayed %d buffered messages to new client", len(buffered))
	}

	// Setup ping/pong handlers
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Start goroutines for reading and writing
	done := make(chan struct{})

	// Read pump - handle incoming messages
	go func() {
		defer func() {
			conn.Close()
			clientsLock.Lock()
			delete(clients, conn)
			clientsLock.Unlock()
			log.Printf("👋 Client removed. Total clients: %d", len(clients))
			close(done)
		}()

		conn.SetReadLimit(maxMessageSize)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("⚠️ Unexpected close: %v", err)
				}
				break
			}

			if userMessageHandler != nil {
				var incoming IncomingUserMessage
				if err := json.Unmarshal(data, &incoming); err == nil {
					incoming.MatchID = strings.TrimSpace(incoming.MatchID)
					incoming.RoomID = strings.TrimSpace(incoming.RoomID)
					incoming.Content = strings.TrimSpace(incoming.Content)
					if incoming.RoomID != "" {
						clientsLock.Lock()
						if existing, ok := clients[conn]; ok {
							existing[incoming.RoomID] = true
						}
						clientsLock.Unlock()
					}
					if incoming.Content != "" {
						go userMessageHandler(incoming)
					}
				}
			}
		}
	}()

	// Write pump - send ping periodically
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}

// BroadcastMessage sends a message to all connected WebSocket clients
func BroadcastMessage(data []byte) {
	roomID := extractRoomID(data)
	queueBroadcast(outboundMessage{roomID: roomID, data: data})
}

// BroadcastMessageToRoom sends a message to connected clients subscribed to this room.
func BroadcastMessageToRoom(roomID string, data []byte) {
	queueBroadcast(outboundMessage{roomID: strings.TrimSpace(roomID), data: data})
}

func queueBroadcast(msg outboundMessage) {
	messageBufferLock.Lock()
	messageBuffer = append(messageBuffer, msg)
	if len(messageBuffer) > messageBufferSize {
		messageBuffer = messageBuffer[len(messageBuffer)-messageBufferSize:] // keep last N
	}
	messageBufferLock.Unlock()

	select {
	case broadcast <- msg:
	default:
		log.Printf("⚠️ Broadcast channel full, message dropped")
	}
}

// handleBroadcast runs in a goroutine and broadcasts messages to all clients
func handleBroadcast() {
	for {
		message := <-broadcast

		clientsLock.RLock()
		deadClients := [](*websocket.Conn){}
		for client, subs := range clients {
			if !isSubscribed(subs, message.roomID) {
				continue
			}
			client.SetWriteDeadline(time.Now().Add(writeWait))
			err := client.WriteMessage(websocket.TextMessage, message.data)
			if err != nil {
				log.Printf("❌ Failed to send to client: %v", err)
				deadClients = append(deadClients, client)
			}
		}
		clientsLock.RUnlock()

		// Clean up dead clients outside the read lock
		if len(deadClients) > 0 {
			clientsLock.Lock()
			for _, dc := range deadClients {
				dc.Close()
				delete(clients, dc)
			}
			clientsLock.Unlock()
			log.Printf("🗑️ Removed %d dead clients. Remaining: %d", len(deadClients), len(clients))
		}
	}
}

func isSubscribed(subs map[string]bool, roomID string) bool {
	if len(subs) == 0 {
		return true
	}
	if subs["*"] {
		return true
	}
	if roomID == "" {
		return false
	}
	return subs[roomID]
}

func extractRoomID(data []byte) string {
	var payload struct {
		RoomID string `json:"room_id"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.RoomID)
}

// GetConnectedClients returns the number of connected clients
func GetConnectedClients() int {
	clientsLock.RLock()
	defer clientsLock.RUnlock()
	return len(clients)
}
