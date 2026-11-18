package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"codecrashreview/backend/internal/auth"
	"codecrashreview/backend/internal/database"
)

type Hub struct {
	redis     *redis.Client
	logger    zerolog.Logger
	clients   map[uuid.UUID]*Client
	rooms     map[uuid.UUID]map[uuid.UUID]*Client // matchID -> userID -> client
	mu        sync.RWMutex
	authSvc   *auth.Service
	matchRepo *database.MatchRepository
}

type Client struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	Username string
	MatchID  uuid.UUID
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *Hub
	IsSpectator bool
}

type WSMessage struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp string          `json:"timestamp"`
	Sequence  int64           `json:"sequence"`
}

const (
	// Message types
	MsgAuth              = "auth"
	MsgAuthResponse      = "auth_response"
	MsgJoinMatch         = "join_match"
	MsgLeaveMatch        = "leave_match"
	MsgMatchState        = "match_state"
	MsgMatchUpdate       = "match_update"
	MsgAgentAction       = "agent_action"
	MsgAgentStatus       = "agent_status"
	MsgCodeChange        = "code_change"
	MsgSubmission        = "submission"
	MsgSubmissionResult  = "submission_result"
	MsgTestResult        = "test_result"
	MsgExecutionResult   = "execution_result"
	MsgTurnChange        = "turn_change"
	MsgTimerUpdate       = "timer_update"
	MsgMatchEnd          = "match_end"
	MsgLeaderboardUpdate = "leaderboard_update"
	MsgQueueStatus       = "queue_status"
	MsgError             = "error"
	MsgPing              = "ping"
	MsgPong              = "pong"
)

func NewHub(redis *redis.Client, logger zerolog.Logger) *Hub {
	return &Hub{
		redis:   redis,
		logger:  logger,
		clients: make(map[uuid.UUID]*Client),
		rooms:   make(map[uuid.UUID]map[uuid.UUID]*Client),
	}
}

func (h *Hub) SetAuthService(svc *auth.Service) {
	h.authSvc = svc
}

func (h *Hub) SetMatchRepository(repo *database.MatchRepository) {
	h.matchRepo = repo
}

func (h *Hub) Run() {
	// Subscribe to Redis channels for cross-instance messaging
	pubsub := h.redis.Subscribe(context.Background(), "ws:broadcast", "ws:match:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		h.handleRedisMessage(msg.Channel, msg.Payload)
	}
}

func (h *Hub) handleRedisMessage(channel, payload string) {
	// Handle messages from other instances
	var msg WSMessage
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		h.logger.Error().Err(err).Msg("Failed to unmarshal Redis message")
		return
	}

	// Route to appropriate match room
	if matchID, err := uuid.Parse(msg.Payload); err == nil {
		h.broadcastToMatch(matchID, msg, nil)
	}
}

func (h *Hub) HandleWebSocket(conn *websocket.Conn, userID uuid.UUID, username string, matchID uuid.UUID, isSpectator bool) {
	client := &Client{
		ID:            uuid.New(),
		UserID:        userID,
		Username:      username,
		MatchID:       matchID,
		Conn:          conn,
		Send:          make(chan []byte, 256),
		Hub:           h,
		IsSpectator:   isSpectator,
	}

	h.registerClient(client)

	go client.writePump()
	go client.readPump()
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client

	if _, ok := h.rooms[client.MatchID]; !ok {
		h.rooms[client.MatchID] = make(map[uuid.UUID]*Client)
	}
	h.rooms[client.MatchID][client.UserID] = client

	h.logger.Info().
		Str("client_id", client.ID.String()).
		Str("user_id", client.UserID.String()).
		Str("match_id", client.MatchID.String()).
		Bool("spectator", client.IsSpectator).
		Msg("Client connected")
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients, client.ID)

	if room, ok := h.rooms[client.MatchID]; ok {
		delete(room, client.UserID)
		if len(room) == 0 {
			delete(h.rooms, client.MatchID)
		}
	}

	close(client.Send)

	h.logger.Info().
		Str("client_id", client.ID.String()).
		Str("user_id", client.UserID.String()).
		Msg("Client disconnected")
}

func (h *Hub) broadcastToMatch(matchID uuid.UUID, msg WSMessage, exclude *Client) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, ok := h.rooms[matchID]
	if !ok {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to marshal message")
		return
	}

	for _, client := range room {
		if client != exclude {
			select {
			case client.Send <- data:
			default:
				// Client buffer full, disconnect
				go h.unregisterClient(client)
			}
		}
	}
}

func (h *Hub) BroadcastToAll(msg WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for _, client := range h.clients {
		select {
		case client.Send <- data:
		default:
			go h.unregisterClient(client)
		}
	}
}

func (h *Hub) SendToUser(userID uuid.UUID, msg WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.UserID == userID {
			data, _ := json.Marshal(msg)
			select {
			case client.Send <- data:
			default:
				go h.unregisterClient(client)
			}
		}
	}
}

// Helper functions for broadcasting match events
func (h *Hub) BroadcastMatchUpdate(matchID uuid.UUID, update map[string]interface{}) {
	msg := WSMessage{
		Type:      MsgMatchUpdate,
		Payload:   mustMarshal(update),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	h.broadcastToMatch(matchID, msg, nil)
}

func (h *Hub) BroadcastAgentStatus(matchID uuid.UUID, persona string, status map[string]interface{}) {
	payload := map[string]interface{}{
		"agent_persona": persona,
		"status":        status,
	}
	msg := WSMessage{
		Type:      MsgAgentStatus,
		Payload:   mustMarshal(payload),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	h.broadcastToMatch(matchID, msg, nil)
}

func (h *Hub) BroadcastCodeChange(matchID uuid.UUID, change map[string]interface{}) {
	msg := WSMessage{
		Type:      MsgCodeChange,
		Payload:   mustMarshal(change),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	h.broadcastToMatch(matchID, msg, nil)
}

func (h *Hub) BroadcastTurnChange(matchID uuid.UUID, turn int, timeRemaining int) {
	payload := map[string]interface{}{
		"turn":                   turn,
		"time_remaining_seconds": timeRemaining,
	}
	msg := WSMessage{
		Type:      MsgTurnChange,
		Payload:   mustMarshal(payload),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	h.broadcastToMatch(matchID, msg, nil)
}

func (h *Hub) BroadcastTimerUpdate(matchID uuid.UUID, seconds int, warning bool) {
	payload := map[string]interface{}{
		"time_remaining_seconds": seconds,
		"warning":                warning,
	}
	msg := WSMessage{
		Type:      MsgTimerUpdate,
		Payload:   mustMarshal(payload),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	h.broadcastToMatch(matchID, msg, nil)
}

func (h *Hub) BroadcastMatchEnd(matchID uuid.UUID, result map[string]interface{}) {
	msg := WSMessage{
		Type:      MsgMatchEnd,
		Payload:   mustMarshal(result),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	h.broadcastToMatch(matchID, msg, nil)
}

func (h *Hub) BroadcastMatchState(matchID uuid.UUID, match *database.Match, participants []*database.MatchParticipant, yourParticipantID *uuid.UUID) {
	payload := map[string]interface{}{
		"match":      match,
		"participants": participants,
	}
	if yourParticipantID != nil {
		payload["your_participant_id"] = *yourParticipantID
	}

	msg := WSMessage{
		Type:      MsgMatchState,
		Payload:   mustMarshal(payload),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	h.broadcastToMatch(matchID, msg, nil)
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

// Client pumps
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024) // 512KB
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	sequence := int64(0)

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.logger.Error().Err(err).Msg("WebSocket read error")
			}
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			c.Hub.logger.Error().Err(err).Msg("Failed to unmarshal client message")
			continue
		}

		msg.Sequence = sequence
		sequence++

		c.handleMessage(msg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case data, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(data)

			// Add queued messages
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg WSMessage) {
	switch msg.Type {
	case MsgPing:
		c.sendPong(msg.Sequence)

	case MsgJoinMatch:
		c.handleJoinMatch(msg)

	case MsgAgentAction:
		c.handleAgentAction(msg)

	case MsgSubmission:
		c.handleSubmission(msg)

	default:
		c.Hub.logger.Warn().Str("type", msg.Type).Msg("Unhandled message type")
	}
}

func (c *Client) sendPong(sequence int64) {
	pong := WSMessage{
		Type:      MsgPong,
		Payload:   mustMarshal(fmt.Sprintf(`{"client_sequence":%d,"server_time":"%s"}`, sequence, time.Now().Format(time.RFC3339))),
		Timestamp: time.Now().Format(time.RFC3339),
		Sequence:  sequence,
	}
	data, _ := json.Marshal(pong)
	select {
	case c.Send <- data:
	default:
	}
}

func (c *Client) handleJoinMatch(msg WSMessage) {
	var payload struct {
		MatchID      string `json:"match_id"`
		AsSpectator  bool   `json:"as_spectator"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return
	}

	matchID, err := uuid.Parse(payload.MatchID)
	if err != nil {
		return
	}

	c.MatchID = matchID
	c.IsSpectator = payload.AsSpectator

	// Re-register in new room
	c.Hub.unregisterClient(c)
	c.Hub.registerClient(c)

	// Send current match state
	if c.Hub.matchRepo != nil {
		if match, err := c.Hub.matchRepo.GetByID(context.Background(), matchID); err == nil {
			participants, _ := c.Hub.matchRepo.GetParticipants(context.Background(), matchID)
			c.sendMatchState(match, participants)
		}
	}
}

func (c *Client) handleAgentAction(msg WSMessage) {
	// Forward to match room
	c.Hub.broadcastToMatch(c.MatchID, msg, c)
}

func (c *Client) handleSubmission(msg WSMessage) {
	// Forward to match room
	c.Hub.broadcastToMatch(c.MatchID, msg, c)
}

func (c *Client) sendMatchState(match *database.Match, participants []*database.MatchParticipant) {
	payload := map[string]interface{}{
		"match":      match,
		"participants": participants,
	}
	payloadBytes, _ := json.Marshal(payload)

	msg := WSMessage{
		Type:      MsgMatchState,
		Payload:   payloadBytes,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	data, _ := json.Marshal(msg)
	select {
	case c.Send <- data:
	default:
	}
}