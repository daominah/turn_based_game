// Package httpsvr defines WebSocket message types and protocol
package httpsvr

import (
	"github.com/daominah/turn_based_game/internal/model"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// MessageTypeAction is sent from client to server to perform a game action
	MessageTypeAction MessageType = "action"
	// MessageTypeCreateDuel is sent from client to server to create a new duel
	MessageTypeCreateDuel MessageType = "create_duel"
	// MessageTypeStateUpdate is sent from server to client with updated game state
	MessageTypeStateUpdate MessageType = "state_update"
	// MessageTypeError is sent from server to client when an error occurs
	MessageTypeError MessageType = "error"
	// MessageTypeJoinDuel is sent from client to server to join an existing duel
	MessageTypeJoinDuel MessageType = "join_duel"
)

// ClientMessage represents a message sent from client to server
type ClientMessage struct {
	Type     MessageType      `json:"type"`
	DuelID   string           `json:"duel_id,omitempty"`
	PlayerID string           `json:"player_id,omitempty"`
	Game     string           `json:"game,omitempty"`
	Players  []string         `json:"players,omitempty"`
	Action   model.ActionData `json:"action,omitempty"`
}

// ServerMessage represents a message sent from server to client
// GameState uses any because different games have different state structures
type ServerMessage struct {
	Type      MessageType             `json:"type"`
	Duel      *model.SerializableDuel `json:"duel,omitempty"`
	GameState any                     `json:"game_state,omitempty"` // Game-specific state (e.g., model.BurnGameState)
	Error     string                  `json:"error,omitempty"`
	Message   string                  `json:"message,omitempty"`
}
