package httpsvr

import (
	"encoding/json"
	"fmt"

	"github.com/daominah/turn_based_game/internal/core/card_game_burn"
	"github.com/daominah/turn_based_game/internal/core/turnbased"
	"github.com/daominah/turn_based_game/internal/model"
)

// BurnActionProcessor processes actions for the Burn card game
type BurnActionProcessor struct {
	duelsManager  turnbased.DuelsManager
	connectionMgr *ConnectionManager
}

// NewBurnActionProcessor creates a new Burn action processor
func NewBurnActionProcessor(duelsManager turnbased.DuelsManager) *BurnActionProcessor {
	// Note: connectionMgr will be set by WebSocketHandler after creation
	return &BurnActionProcessor{
		duelsManager: duelsManager,
	}
}

// SetConnectionManager sets the connection manager (called by WebSocketHandler)
func (p *BurnActionProcessor) SetConnectionManager(cm *ConnectionManager) {
	p.connectionMgr = cm
}

// CreateDuel creates a new Burn duel
func (p *BurnActionProcessor) CreateDuel(game string, players []turnbased.PlayerID) (*turnbased.Duel, error) {
	// Create BurnDuel
	burnDuel := card_game_burn.NewBurnDuel(players)

	// Wrap in generic Duel
	duel := burnDuel.Duel
	duel.Game = burnDuel

	// Persist via DuelsManager
	createdDuel := p.duelsManager.CreateDuel(duel)

	// Fanout: send state to all connections (will be done by handler after connection registration)
	return createdDuel, nil
}

// ProcessAction implements the three-stage flow: Message In → Persist → Fanout
func (p *BurnActionProcessor) ProcessAction(duelID turnbased.DuelID, playerID turnbased.PlayerID, actionData model.ActionData) error {
	// Stage 1: Message In - action is already received, now parse it
	action, err := p.parseAction(actionData)
	if err != nil {
		return fmt.Errorf("failed to parse action: %w", err)
	}

	// Get the duel
	duel := p.duelsManager.GetDuel(duelID)
	if duel == nil {
		return fmt.Errorf("duel not found: %s", duelID)
	}

	// Verify it's a Burn duel
	burnDuel, ok := duel.Game.(*card_game_burn.BurnDuel)
	if !ok {
		return fmt.Errorf("duel is not a Burn duel")
	}

	// Process the action
	if err := burnDuel.HandleActionWithPlayer(action, playerID); err != nil {
		return err
	}

	// Stage 2: Persist - update the duel in storage
	updatedDuel, err := p.duelsManager.UpdateDuel(duel)
	if err != nil {
		return fmt.Errorf("failed to persist duel: %w", err)
	}

	// Stage 3: Fanout - broadcast updated state to all connected clients
	return p.fanoutState(updatedDuel)
}

func (p *BurnActionProcessor) parseAction(actionData model.ActionData) (any, error) {
	// Check for PlayCard action
	if actionData.CardID != nil && *actionData.CardID != "" {
		option := card_game_burn.PlayCardOptionPending
		if actionData.Option != nil {
			option = card_game_burn.PlayCardOption(*actionData.Option)
		}
		return card_game_burn.ActionPlayCard{
			CardID: card_game_burn.UniqueCardID(*actionData.CardID),
			Option: option,
		}, nil
	}

	// Check for EndTurn action
	if actionData.EndTurn != nil && *actionData.EndTurn {
		return card_game_burn.ActionEndTurn{}, nil
	}

	// Fallback: try JSON unmarshaling for backward compatibility
	data, err := json.Marshal(actionData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal action data: %w", err)
	}

	// Try to parse as ActionPlayCard
	var playCardAction card_game_burn.ActionPlayCard
	if err := json.Unmarshal(data, &playCardAction); err == nil && playCardAction.CardID != "" {
		return playCardAction, nil
	}

	// Try to parse as ActionEndTurn
	var endTurnAction card_game_burn.ActionEndTurn
	if err := json.Unmarshal(data, &endTurnAction); err == nil {
		return endTurnAction, nil
	}

	return nil, fmt.Errorf("unknown action format")
}

func (p *BurnActionProcessor) fanoutState(duel *turnbased.Duel) error {
	if p.connectionMgr == nil {
		return fmt.Errorf("connection manager not set")
	}

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

	return p.connectionMgr.BroadcastToDuel(duel.ID, msg)
}
