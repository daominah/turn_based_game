package turnbased

import (
	"time"
)

// ActionLogEntry represents a single action in the duel log.
// This is generic and can be used by any game.
type ActionLogEntry struct {
	Seq       int                    // Sequence number (1, 2, 3, ...) for replay ordering
	Timestamp time.Time              // When the action occurred
	PlayerID  PlayerID               // Player who performed the action
	Action    string                 // Action type (game-specific, e.g., "PLAY_CARD", "END_TURN")
	Data      map[string]interface{} // Action-specific data (game-specific)
}

// Duel represents a generic turn-based game duel.
type Duel struct {
	ID         DuelID     // Unique identifier for this duel
	Players    []PlayerID // Player IDs (supports any number of players)
	Turn       int        // Current turn number (starts from 1)
	TurnPlayer PlayerID   // Player ID whose turn it is
	Winner     PlayerID   // Player ID if someone has won, empty if ongoing, "DRAW" for draw
	State      DuelState  // BEGIN, RUNNING, END
	Game       GameLogic
	ActionLog  []ActionLogEntry // Log of all actions for replay
}

// GameLogic is implemented differently for each game,
// but all implementations share some common methods
// so that the centralized router can interact with them.
type GameLogic interface {
	GetState() any
	HandleAction(action any) error
	// Add more methods as needed for your engine
}

// NewDuel creates a new Duel with the given players.
func NewDuel(id DuelID, players []PlayerID) *Duel {
	return &Duel{
		ID:         id,
		Players:    players,
		Turn:       0,
		TurnPlayer: "",
		Winner:     "",
		State:      DuelStateBegin,
		ActionLog:  []ActionLogEntry{},
	}
}

// LogAction adds an action to the duel log.
// This is called by game logic implementations when actions are performed.
func (d *Duel) LogAction(playerID PlayerID, action string, data map[string]interface{}) {
	seq := len(d.ActionLog) + 1
	d.ActionLog = append(d.ActionLog, ActionLogEntry{
		Seq:       seq,
		Timestamp: time.Now(),
		PlayerID:  playerID,
		Action:    action,
		Data:      data,
	})
}

// NextTurn advances the duel to the next turn and updates the turn player.
func (d *Duel) NextTurn() {
	if len(d.Players) == 0 {
		return
	}
	d.Turn++
	idx := 0
	for i, id := range d.Players {
		if id == d.TurnPlayer {
			idx = i
			break
		}
	}
	nextIdx := (idx + 1) % len(d.Players)
	d.TurnPlayer = d.Players[nextIdx]
}

// IsOver returns true if the duel has ended.
func (d *Duel) IsOver() bool {
	return d.State == DuelStateEnd && d.Winner != ""
}

// SetWinner sets the winner and ends the duel.
func (d *Duel) SetWinner(winnerID PlayerID) {
	d.Winner = winnerID
	d.State = DuelStateEnd
}

// SetDraw ends the duel as a draw.
func (d *Duel) SetDraw() {
	d.Winner = "DRAW"
	d.State = DuelStateEnd
}

// PlayerID is unique identifier for a player
type PlayerID string

// DuelID is a unique identifier for a duel
type DuelID string

// DuelState represents the main state of a duel.
//
// Possible values:
//   - "BEGIN": The duel has just been initialized. Some automatic actions are performed,
//     such as tossing a coin to determine who plays first, drawing cards, or placing
//     chess pieces in their starting positions. No player can perform actions in this state.
//   - "RUNNING": The duel is in progress. Only one player can perform valid actions at a time.
//     After each action, the game state changes, and the engine determines which player
//     can act next and what actions are available, according to the game logic.
//   - "END": The duel has ended, either because someone has won or, rarely, due to a draw.
//     After each action, the engine checks if the duel is in the END state and determines
//     the winner if applicable.
//
// See README for more details on the generic turn-based game engine states.
type DuelState string

const (
	// DuelStateBegin means the duel has just been initialized. Some automatic actions are performed,
	// such as tossing a coin to determine who plays first, drawing cards, or placing
	// chess pieces in their starting positions. No player can perform actions in this state.
	DuelStateBegin DuelState = "BEGIN"
	// DuelStateRunning means the duel is in progress. Only one player can perform valid actions at a time.
	// After each action, the game state changes, and the engine determines which player
	// can act next and what actions are available, according to the game logic.
	DuelStateRunning DuelState = "RUNNING"
	// DuelStateEnd means the duel has ended, either because someone has won or, rarely, due to a draw.
	DuelStateEnd DuelState = "END"
)
