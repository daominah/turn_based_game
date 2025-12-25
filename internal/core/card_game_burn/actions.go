package card_game_burn

import (
	"github.com/daominah/turn_based_game/internal/core/turnbased"
)

// ActionPlayCard represents an action to play a card
type ActionPlayCard struct {
	CardID UniqueCardID
	Option PlayCardOption
}

// ActionEndTurn represents an action to end the current turn
type ActionEndTurn struct{}

// CreateDuelAction represents an action to create a new duel
type CreateDuelAction struct {
	Players []turnbased.PlayerID
}

// Implement turnbased.Action interface for ActionPlayCard
func (a ActionPlayCard) GameName() string {
	return GameName
}

func (a ActionPlayCard) DuelID() turnbased.DuelID {
	// DuelID will be set by the handler from the message context
	return ""
}

func (a ActionPlayCard) PlayerID() turnbased.PlayerID {
	// PlayerID will be set by the handler from the message context
	return ""
}

// Implement turnbased.Action interface for ActionEndTurn
func (a ActionEndTurn) GameName() string {
	return GameName
}

func (a ActionEndTurn) DuelID() turnbased.DuelID {
	return ""
}

func (a ActionEndTurn) PlayerID() turnbased.PlayerID {
	return ""
}

// Implement turnbased.Action interface for CreateDuelAction
func (a CreateDuelAction) GameName() string {
	return GameName
}

func (a CreateDuelAction) DuelID() turnbased.DuelID {
	return ""
}

func (a CreateDuelAction) PlayerID() turnbased.PlayerID {
	return ""
}
