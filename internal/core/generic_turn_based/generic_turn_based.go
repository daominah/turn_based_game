package generic_turn_based

// Duel represents a generic turn-based game duel.
type Duel struct {
	ID         DuelID     // Unique identifier for this duel
	Players    []PlayerID // Player IDs (supports any number of players)
	Turn       int        // Current turn number (starts from 1)
	TurnPlayer PlayerID   // Player ID whose turn it is
	Winner     PlayerID   // Player ID if someone has won, empty if ongoing, "DRAW" for draw
	State      DuelState  // BEGIN, OPEN, END
}

// NewDuel creates a new Duel with the given players.
func NewDuel(id DuelID, players []PlayerID) *Duel {
	return &Duel{
		ID:         id,
		Players:    players,
		Turn:       1,
		TurnPlayer: players[0],
		Winner:     "",
		State:      DuelStateBegin,
	}
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
//   - "BEGIN":
//   - "OPEN": The duel is in progress. Only one player can perform valid actions at a time.
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
	// DuelStateOpen means the duel is in progress. Only one player can perform valid actions at a time.
	// After each action, the game state changes, and the engine determines which player
	// can act next and what actions are available, according to the game logic.
	DuelStateOpen DuelState = "OPEN"
	// DuelStateEnd means the duel has ended, either because someone has won or, rarely, due to a draw.
	DuelStateEnd DuelState = "END"
)
