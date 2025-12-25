package card_game_burn

import (
	"encoding/json"
	"fmt"

	"github.com/daominah/turn_based_game/internal/core/turnbased"
	"github.com/daominah/turn_based_game/internal/model"
)

// Ensure BurnDuel implements GameLogic interface
var _ turnbased.GameLogic = (*BurnDuel)(nil)

// GetState returns the game-specific state as JSON-serializable data
// Returns model.BurnGameState for type safety
func (cgb *BurnDuel) GetState() any {
	return cgb.ToModelBurnGameState()
}

// HandleAction processes a game action
func (cgb *BurnDuel) HandleAction(action any) error {
	switch action.(type) {
	case ActionPlayCard:
		// PlayerID and DuelID are set by the handler before calling this
		// We need to get them from the action context, but for now we'll pass them separately
		// Actually, we need to modify the interface or pass playerID separately
		// Let's create a wrapper that includes playerID
		return fmt.Errorf("ActionPlayCard needs playerID context - use HandleActionWithPlayer instead")
	case ActionEndTurn:
		cgb.EndTurn()
		return nil
	default:
		return fmt.Errorf("unknown action type: %T", action)
	}
}

// HandleActionWithPlayer processes a game action with player context
func (cgb *BurnDuel) HandleActionWithPlayer(action any, playerID turnbased.PlayerID) error {
	switch a := action.(type) {
	case ActionPlayCard:
		success := cgb.PlayCard(playerID, a.CardID, a.Option)
		if !success {
			return fmt.Errorf("failed to play card: invalid action or not player's turn")
		}
		return nil
	case ActionEndTurn:
		if cgb.Duel.TurnPlayer != playerID {
			return fmt.Errorf("not player's turn")
		}
		cgb.EndTurn()
		return nil
	default:
		return fmt.Errorf("unknown action type: %T", action)
	}
}

// SerializeState returns the full game state as JSON bytes
func (cgb *BurnDuel) SerializeState() ([]byte, error) {
	state := struct {
		Duel      model.SerializableDuel `json:"duel"`
		GameState model.BurnGameState    `json:"game_state"`
	}{
		Duel:      model.FromDuel(cgb.Duel),
		GameState: cgb.ToModelBurnGameState(),
	}
	return json.Marshal(state)
}
