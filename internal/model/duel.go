// Package model defines shared data models used across packages
package model

import (
	"github.com/daominah/turn_based_game/internal/core/turnbased"
)

// SerializableActionLogEntry represents an action log entry in JSON format
type SerializableActionLogEntry struct {
	Seq       int                    `json:"seq"`
	Timestamp string                 `json:"timestamp"` // ISO 8601 format: 2006-01-02T15:04:05.999
	PlayerID  string                 `json:"player_id"`
	Action    string                 `json:"action"`
	Data      map[string]interface{} `json:"data"`
}

// SerializableDuel represents a Duel in a JSON-serializable format
// This is used for WebSocket messages and API responses
type SerializableDuel struct {
	ID           string                       `json:"id"`
	Players      []string                     `json:"players"`
	Turn         int                          `json:"turn"`
	TurnPlayer   string                       `json:"turn_player"`
	Winner       string                       `json:"winner"`
	State        string                       `json:"state"`
	ActionLog    []SerializableActionLogEntry `json:"action_log"`
	PlayerColors map[string]string            `json:"player_colors"` // Player ID -> color hex code
}

// FromDuel converts a turnbased.Duel to SerializableDuel
func FromDuel(duel *turnbased.Duel) SerializableDuel {
	players := make([]string, len(duel.Players))
	for i, pid := range duel.Players {
		players[i] = string(pid)
	}

	// Convert action log
	actionLog := make([]SerializableActionLogEntry, len(duel.ActionLog))
	for i, entry := range duel.ActionLog {
		// Format timestamp as ISO 8601: 2006-01-02T15:04:05.999
		timestamp := entry.Timestamp.Format("2006-01-02T15:04:05.000")
		actionLog[i] = SerializableActionLogEntry{
			Seq:       entry.Seq,
			Timestamp: timestamp,
			PlayerID:  string(entry.PlayerID),
			Action:    entry.Action,
			Data:      entry.Data,
		}
	}

	// Assign player colors consistently: first player (by order in duel) = Blue, second = Purple
	playerColors := make(map[string]string)
	if len(players) > 0 {
		playerColors[players[0]] = "#007bff" // Blue
	}
	if len(players) > 1 {
		playerColors[players[1]] = "#6f42c1" // Purple
	}

	return SerializableDuel{
		ID:           string(duel.ID),
		Players:      players,
		Turn:         duel.Turn,
		TurnPlayer:   string(duel.TurnPlayer),
		Winner:       string(duel.Winner),
		State:        string(duel.State),
		ActionLog:    actionLog,
		PlayerColors: playerColors,
	}
}
