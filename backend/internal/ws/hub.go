package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

// Hub maintains the set of active clients grouped by rooms.
// Supports Redis PubSub for horizontal scaling across multiple instances.
type Hub struct {
	rooms map[string]map[*Client]bool
	mu    sync.RWMutex

	registerCh   chan *Client
	unregisterCh chan *Client
	broadcastCh  chan *BroadcastMessage

	// Redis PubSub for cross-instance broadcast
	rdb    *redis.Client
	ctx    context.Context
	cancel context.CancelFunc
}

// BroadcastMessage wraps a message with routing metadata.
type BroadcastMessage struct {
	RoomID  string `json:"room_id"`
	Message Message `json:"message"`
	Exclude uint   `json:"exclude"` // exclude this user ID from receiving
}

// NewHub creates a Hub. If redisURL is non-empty, enables Redis PubSub.
func NewHub(redisURL string) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	h := &Hub{
		rooms:        make(map[string]map[*Client]bool),
		registerCh:   make(chan *Client),
		unregisterCh: make(chan *Client),
		broadcastCh:  make(chan *BroadcastMessage, 256),
		ctx:          ctx,
		cancel:       cancel,
	}

	if redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Printf("⚠️ Redis URL invalid, running in-memory only: %v", err)
		} else {
			h.rdb = redis.NewClient(opt)
			if err := h.rdb.Ping(ctx).Err(); err != nil {
				log.Printf("⚠️ Redis not reachable, running in-memory only: %v", err)
				h.rdb = nil
			} else {
				log.Println("📡 Redis PubSub connected — WebSocket horizontal scaling enabled")
			}
		}
	} else {
		log.Println("📡 WebSocket Hub running in-memory mode (single instance)")
	}

	return h
}

// redisPubSubChannel is the Redis channel for cross-instance WS messages.
const redisPubSubChannel = "radimaster:ws:broadcast"

// Run starts the Hub's main event loop. Must be called in a goroutine.
func (h *Hub) Run() {
	// Start Redis subscriber if available
	if h.rdb != nil {
		go h.subscribeRedis()
	}

	for {
		select {
		case client := <-h.registerCh:
			h.mu.Lock()
			if _, ok := h.rooms[client.roomID]; !ok {
				h.rooms[client.roomID] = make(map[*Client]bool)
			}
			h.rooms[client.roomID][client] = true
			roomSize := len(h.rooms[client.roomID])
			h.mu.Unlock()

			log.Printf("👤 User %d (%s) joined room %s (online: %d)",
				client.userID, client.userName, client.roomID, roomSize)
			h.broadcastPresence(client.roomID)

		case client := <-h.unregisterCh:
			h.mu.Lock()
			if room, ok := h.rooms[client.roomID]; ok {
				if _, ok := room[client]; ok {
					delete(room, client)
					close(client.send)
					if len(room) == 0 {
						delete(h.rooms, client.roomID)
					}
				}
			}
			h.mu.Unlock()

			log.Printf("👤 User %d (%s) left room %s", client.userID, client.userName, client.roomID)
			h.broadcastPresence(client.roomID)

		case bMsg := <-h.broadcastCh:
			// If Redis is connected, publish to Redis for cross-instance delivery
			if h.rdb != nil {
				h.publishToRedis(bMsg)
			} else {
				// In-memory only: deliver directly to local clients
				h.deliverToLocal(bMsg)
			}
		}
	}
}

// publishToRedis publishes a broadcast message to Redis PubSub channel.
func (h *Hub) publishToRedis(bMsg *BroadcastMessage) {
	data, err := json.Marshal(bMsg)
	if err != nil {
		log.Printf("❌ Failed to marshal for Redis: %v", err)
		return
	}
	if err := h.rdb.Publish(h.ctx, redisPubSubChannel, data).Err(); err != nil {
		log.Printf("⚠️ Redis publish failed, delivering locally: %v", err)
		h.deliverToLocal(bMsg)
	}
}

// subscribeRedis listens for broadcast messages from other instances via Redis.
func (h *Hub) subscribeRedis() {
	pubsub := h.rdb.Subscribe(h.ctx, redisPubSubChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	log.Printf("📡 Redis PubSub subscriber listening on channel: %s", redisPubSubChannel)

	for {
		select {
		case <-h.ctx.Done():
			return
		case redisMsg, ok := <-ch:
			if !ok {
				return
			}
			var bMsg BroadcastMessage
			if err := json.Unmarshal([]byte(redisMsg.Payload), &bMsg); err != nil {
				log.Printf("⚠️ Redis message unmarshal error: %v", err)
				continue
			}
			// Deliver to local clients only (each instance handles its own clients)
			h.deliverToLocal(&bMsg)
		}
	}
}

// deliverToLocal sends a message to clients connected to this instance.
func (h *Hub) deliverToLocal(bMsg *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, ok := h.rooms[bMsg.RoomID]
	if !ok {
		return
	}

	data, err := json.Marshal(bMsg.Message)
	if err != nil {
		log.Printf("❌ Failed to marshal broadcast message: %v", err)
		return
	}

	for client := range room {
		if bMsg.Exclude > 0 && client.userID == bMsg.Exclude {
			continue
		}
		select {
		case client.send <- data:
		default:
			close(client.send)
			delete(room, client)
		}
	}
}

// Register adds a client to its room.
func (h *Hub) Register(client *Client) {
	h.registerCh <- client
}

// Unregister removes a client from its room.
func (h *Hub) Unregister(client *Client) {
	h.unregisterCh <- client
}

// BroadcastToRoom sends a message to all clients in a room.
func (h *Hub) BroadcastToRoom(roomID string, msg Message) {
	h.broadcastCh <- &BroadcastMessage{
		RoomID:  roomID,
		Message: msg,
	}
}

// BroadcastToRoomExclude sends a message to all clients in a room except the specified user.
func (h *Hub) BroadcastToRoomExclude(roomID string, msg Message, excludeUserID uint) {
	h.broadcastCh <- &BroadcastMessage{
		RoomID:  roomID,
		Message: msg,
		Exclude: excludeUserID,
	}
}

// broadcastPresence sends the current online user list to everyone in a room.
func (h *Hub) broadcastPresence(roomID string) {
	h.mu.RLock()
	users := make([]map[string]interface{}, 0)
	if room, ok := h.rooms[roomID]; ok {
		seen := make(map[uint]bool)
		for client := range room {
			if !seen[client.userID] {
				seen[client.userID] = true
				users = append(users, map[string]interface{}{
					"id":   client.userID,
					"name": client.userName,
				})
			}
		}
	}
	h.mu.RUnlock()

	h.broadcastCh <- &BroadcastMessage{
		RoomID: roomID,
		Message: Message{
			Type:    PresenceUpdate,
			Payload: users,
		},
	}
}

// GetOnlineUsers returns the list of unique online users in a room.
func (h *Hub) GetOnlineUsers(roomID string) []map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]map[string]interface{}, 0)
	if room, ok := h.rooms[roomID]; ok {
		seen := make(map[uint]bool)
		for client := range room {
			if !seen[client.userID] {
				seen[client.userID] = true
				users = append(users, map[string]interface{}{
					"id":   client.userID,
					"name": client.userName,
				})
			}
		}
	}
	return users
}

// Stats returns hub-level statistics for monitoring.
func (h *Hub) Stats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	totalClients := 0
	for _, room := range h.rooms {
		totalClients += len(room)
	}

	redisStatus := "disconnected"
	if h.rdb != nil {
		if err := h.rdb.Ping(h.ctx).Err(); err == nil {
			redisStatus = "connected"
		}
	}

	return map[string]interface{}{
		"active_rooms":       len(h.rooms),
		"active_connections": totalClients,
		"redis_status":       redisStatus,
	}
}
