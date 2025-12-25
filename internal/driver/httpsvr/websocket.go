package httpsvr

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/daominah/turn_based_game/internal/core/card_game_burn"
	"github.com/daominah/turn_based_game/internal/core/turnbased"
	"github.com/daominah/turn_based_game/internal/model"
)

// WebSocketHandler handles WebSocket connections and game actions
type WebSocketHandler struct {
	duelsManagers    map[string]turnbased.DuelsManager
	connectionMgr    *ConnectionManager
	actionProcessors map[string]ActionProcessor
}

// ActionProcessor processes actions for a specific game
type ActionProcessor interface {
	ProcessAction(duelID turnbased.DuelID, playerID turnbased.PlayerID, action model.ActionData) error
	CreateDuel(game string, players []turnbased.PlayerID) (*turnbased.Duel, error)
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(
	duelsManagers map[string]turnbased.DuelsManager,
	connectionMgr *ConnectionManager,
) *WebSocketHandler {
	handler := &WebSocketHandler{
		duelsManagers:    duelsManagers,
		connectionMgr:    connectionMgr,
		actionProcessors: make(map[string]ActionProcessor),
	}

	// Register action processors for each game
	if _, ok := duelsManagers[card_game_burn.GameName]; ok {
		processor := NewBurnActionProcessor(duelsManagers[card_game_burn.GameName])
		processor.SetConnectionManager(connectionMgr)
		handler.actionProcessors[card_game_burn.GameName] = processor
	}

	return handler
}

// HandleWebSocket handles WebSocket connections
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	connectStartTime := time.Now()
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:*", "127.0.0.1:*"},
	})
	if err != nil {
		log.Printf("WebSocket accept error: %v", err)
		return
	}
	defer conn.Close(websocket.StatusInternalError, "connection closed")

	connectDuration := time.Since(connectStartTime)
	log.Printf("WebSocket connection established from %s in %v", r.RemoteAddr, connectDuration)

	// Read messages from client
	for {
		_, data, err := conn.Read(context.Background())
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		var clientMsg ClientMessage
		if err := json.Unmarshal(data, &clientMsg); err != nil {
			h.sendError(conn, fmt.Sprintf("invalid message format: %v", err))
			continue
		}

		if err := h.handleMessage(conn, &clientMsg); err != nil {
			log.Printf("Error handling message: %v", err)
			h.sendError(conn, err.Error())
		}
	}
}

func (h *WebSocketHandler) handleMessage(conn *websocket.Conn, msg *ClientMessage) error {
	switch msg.Type {
	case MessageTypeCreateDuel:
		return h.handleCreateDuel(conn, msg)
	case MessageTypeJoinDuel:
		return h.handleJoinDuel(conn, msg)
	case MessageTypeAction:
		return h.handleAction(conn, msg)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

func (h *WebSocketHandler) handleCreateDuel(conn *websocket.Conn, msg *ClientMessage) error {
	if msg.Game == "" {
		return fmt.Errorf("game name required")
	}
	if len(msg.Players) == 0 {
		return fmt.Errorf("at least one player required")
	}

	processor, ok := h.actionProcessors[msg.Game]
	if !ok {
		return fmt.Errorf("unknown game: %s", msg.Game)
	}

	playerIDs := make([]turnbased.PlayerID, len(msg.Players))
	for i, p := range msg.Players {
		playerIDs[i] = turnbased.PlayerID(p)
	}

	duel, err := processor.CreateDuel(msg.Game, playerIDs)
	if err != nil {
		return err
	}

	// Register connection for all players in the duel
	for _, playerID := range playerIDs {
		h.connectionMgr.AddConnection(conn, playerID, duel.ID)
	}

	// Send initial state to client
	return h.sendStateUpdate(conn, duel)
}

func (h *WebSocketHandler) handleJoinDuel(conn *websocket.Conn, msg *ClientMessage) error {
	if msg.DuelID == "" {
		return fmt.Errorf("duel_id required")
	}
	if msg.PlayerID == "" {
		return fmt.Errorf("player_id required")
	}

	duelID := turnbased.DuelID(msg.DuelID)
	playerID := turnbased.PlayerID(msg.PlayerID)

	// Find which game this duel belongs to
	var duel *turnbased.Duel
	for _, manager := range h.duelsManagers {
		d := manager.GetDuel(duelID)
		if d != nil {
			duel = d
			break
		}
	}

	if duel == nil {
		return fmt.Errorf("duel not found: %s", msg.DuelID)
	}

	// Verify player is in the duel
	found := false
	for _, pid := range duel.Players {
		if pid == playerID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("player %s is not in duel %s", msg.PlayerID, msg.DuelID)
	}

	// Register connection
	h.connectionMgr.AddConnection(conn, playerID, duelID)

	// Send current state
	return h.sendStateUpdate(conn, duel)
}

func (h *WebSocketHandler) handleAction(conn *websocket.Conn, msg *ClientMessage) error {
	if msg.DuelID == "" {
		return fmt.Errorf("duel_id required")
	}
	if msg.PlayerID == "" {
		return fmt.Errorf("player_id required")
	}
	if msg.Game == "" {
		return fmt.Errorf("game required")
	}

	duelID := turnbased.DuelID(msg.DuelID)
	playerID := turnbased.PlayerID(msg.PlayerID)

	// Find the duel
	manager, ok := h.duelsManagers[msg.Game]
	if !ok {
		return fmt.Errorf("unknown game: %s", msg.Game)
	}

	duel := manager.GetDuel(duelID)
	if duel == nil {
		return fmt.Errorf("duel not found: %s", msg.DuelID)
	}

	// Parse action based on game type
	processor, ok := h.actionProcessors[msg.Game]
	if !ok {
		return fmt.Errorf("no processor for game: %s", msg.Game)
	}

	// Process action (Message In → Persist → Fanout happens in processor)
	if err := processor.ProcessAction(duelID, playerID, msg.Action); err != nil {
		return err
	}

	// State update is sent by the processor via fanout
	return nil
}

func (h *WebSocketHandler) sendStateUpdate(conn *websocket.Conn, duel *turnbased.Duel) error {
	// Get game-specific state
	gameStateRaw := duel.Game.GetState()

	// Convert to typed game state if it's a Burn duel
	var gameState any
	if burnDuel, ok := duel.Game.(*card_game_burn.BurnDuel); ok {
		gameState = burnDuel.ToModelBurnGameState()
	} else {
		// Fallback to raw state for other games
		gameState = gameStateRaw
	}

	// Create serializable duel
	serializableDuel := model.FromDuel(duel)

	msg := ServerMessage{
		Type:      MessageTypeStateUpdate,
		Duel:      &serializableDuel,
		GameState: gameState,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return conn.Write(context.Background(), websocket.MessageText, data)
}

func (h *WebSocketHandler) sendError(conn *websocket.Conn, errorMsg string) {
	msg := ServerMessage{
		Type:    MessageTypeError,
		Error:   errorMsg,
		Message: errorMsg,
	}
	data, _ := json.Marshal(msg)
	_ = conn.Write(context.Background(), websocket.MessageText, data)
}
