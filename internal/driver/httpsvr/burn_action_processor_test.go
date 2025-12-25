package httpsvr

import (
	"testing"

	"github.com/daominah/turn_based_game/internal/core/card_game_burn"
	"github.com/daominah/turn_based_game/internal/core/turnbased"
	"github.com/daominah/turn_based_game/internal/model"
)

func TestBurnActionProcessor_CreateDuel(t *testing.T) {
	duelsManager := turnbased.NewInMemoryDuelsManager()
	processor := NewBurnActionProcessor(duelsManager)

	players := []turnbased.PlayerID{"player1", "player2"}
	duel, err := processor.CreateDuel(card_game_burn.GameName, players)
	if err != nil {
		t.Fatalf("CreateDuel failed: %v", err)
	}

	if duel == nil {
		t.Fatal("CreateDuel returned nil")
	}

	if duel.ID == "" {
		t.Error("Duel ID should be set")
	}

	if len(duel.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(duel.Players))
	}

	// Verify duel is persisted
	retrieved := duelsManager.GetDuel(duel.ID)
	if retrieved == nil {
		t.Error("Duel should be retrievable from manager")
	}

	// Verify it's a Burn duel
	burnDuel, ok := duel.Game.(*card_game_burn.BurnDuel)
	if !ok {
		t.Fatal("Duel.Game should be *BurnDuel")
	}

	if len(burnDuel.Players) != 2 {
		t.Errorf("Expected 2 players in BurnDuel, got %d", len(burnDuel.Players))
	}
}

func TestBurnActionProcessor_ParseAction(t *testing.T) {
	duelsManager := turnbased.NewInMemoryDuelsManager()
	processor := NewBurnActionProcessor(duelsManager)

	// Test parsing ActionPlayCard
	cardID := "card123"
	option := "GAIN"
	actionData := model.ActionData{
		CardID: &cardID,
		Option: &option,
	}

	action, err := processor.parseAction(actionData)
	if err != nil {
		t.Fatalf("parseAction failed: %v", err)
	}

	playCardAction, ok := action.(card_game_burn.ActionPlayCard)
	if !ok {
		t.Fatalf("Expected ActionPlayCard, got %T", action)
	}

	if playCardAction.CardID != "card123" {
		t.Errorf("Expected card_id card123, got %s", playCardAction.CardID)
	}

	if playCardAction.Option != card_game_burn.PlayCardOptionGain {
		t.Errorf("Expected option GAIN, got %s", playCardAction.Option)
	}

	// Test parsing ActionEndTurn
	endTurn := true
	actionData2 := model.ActionData{
		EndTurn: &endTurn,
	}

	action2, err := processor.parseAction(actionData2)
	if err != nil {
		t.Fatalf("parseAction for EndTurn failed: %v", err)
	}

	_, ok = action2.(card_game_burn.ActionEndTurn)
	if !ok {
		t.Fatalf("Expected ActionEndTurn, got %T", action2)
	}

	// Test empty object as EndTurn (should fail or handle gracefully)
	actionData3 := model.ActionData{}
	action3, err := processor.parseAction(actionData3)
	if err != nil {
		t.Fatalf("parseAction for empty EndTurn failed: %v", err)
	}

	_, ok = action3.(card_game_burn.ActionEndTurn)
	if !ok {
		t.Fatalf("Expected ActionEndTurn for empty object, got %T", action3)
	}
}

func TestBurnActionProcessor_ProcessAction(t *testing.T) {
	duelsManager := turnbased.NewInMemoryDuelsManager()
	processor := NewBurnActionProcessor(duelsManager)
	connectionMgr := NewConnectionManager()
	processor.SetConnectionManager(connectionMgr)

	// Create a duel first
	players := []turnbased.PlayerID{"player1", "player2"}
	duel, err := processor.CreateDuel(card_game_burn.GameName, players)
	if err != nil {
		t.Fatalf("CreateDuel failed: %v", err)
	}

	// Get the turn player and a card
	burnDuel := duel.Game.(*card_game_burn.BurnDuel)
	turnPlayer := duel.TurnPlayer
	ps := burnDuel.Players[turnPlayer]

	if len(ps.Hand) == 0 {
		t.Fatal("Player should have cards")
	}

	card := ps.Hand[0]

	// Test ProcessAction with PlayCard
	cardID := string(card.UniqueCardID)
	option := "GAIN"
	actionData := model.ActionData{
		CardID: &cardID,
		Option: &option,
	}

	err = processor.ProcessAction(duel.ID, turnPlayer, actionData)
	if err != nil {
		t.Fatalf("ProcessAction failed: %v", err)
	}

	// Verify the action was applied
	updatedDuel := duelsManager.GetDuel(duel.ID)
	if updatedDuel == nil {
		t.Fatal("Duel should still exist")
	}

	updatedBurnDuel := updatedDuel.Game.(*card_game_burn.BurnDuel)
	updatedPS := updatedBurnDuel.Players[turnPlayer]

	// Card should be removed from hand
	found := false
	for _, c := range updatedPS.Hand {
		if c.UniqueCardID == card.UniqueCardID {
			found = true
			break
		}
	}
	if found {
		t.Error("Card should be removed from hand after playing")
	}

	// Test ProcessAction with EndTurn
	endTurn := true
	actionData2 := model.ActionData{
		EndTurn: &endTurn,
	}

	// Get current turn before EndTurn
	currentTurn := updatedDuel.Turn

	err = processor.ProcessAction(duel.ID, turnPlayer, actionData2)
	if err != nil {
		t.Fatalf("ProcessAction for EndTurn failed: %v", err)
	}

	// Verify turn advanced
	updatedDuel2 := duelsManager.GetDuel(duel.ID)
	if updatedDuel2.Turn != currentTurn+1 {
		t.Errorf("Expected turn %d, got %d", currentTurn+1, updatedDuel2.Turn)
	}
}

func TestBurnActionProcessor_ProcessAction_Invalid(t *testing.T) {
	duelsManager := turnbased.NewInMemoryDuelsManager()
	processor := NewBurnActionProcessor(duelsManager)
	connectionMgr := NewConnectionManager()
	processor.SetConnectionManager(connectionMgr)

	// Try to process action for non-existent duel
	err := processor.ProcessAction("nonexistent", "player1", model.ActionData{})
	if err == nil {
		t.Error("ProcessAction should fail for non-existent duel")
	}

	// Create a duel
	players := []turnbased.PlayerID{"player1", "player2"}
	duel, err := processor.CreateDuel(card_game_burn.GameName, players)
	if err != nil {
		t.Fatalf("CreateDuel failed: %v", err)
	}

	// Try to process action as wrong player
	burnDuel := duel.Game.(*card_game_burn.BurnDuel)
	turnPlayer := duel.TurnPlayer
	var otherPlayer turnbased.PlayerID
	for _, pid := range players {
		if pid != turnPlayer {
			otherPlayer = pid
			break
		}
	}

	ps := burnDuel.Players[turnPlayer]
	if len(ps.Hand) == 0 {
		t.Fatal("Player should have cards")
	}

	card := ps.Hand[0]
	cardID := string(card.UniqueCardID)
	option := "GAIN"
	actionData := model.ActionData{
		CardID: &cardID,
		Option: &option,
	}

	err = processor.ProcessAction(duel.ID, otherPlayer, actionData)
	if err == nil {
		t.Error("ProcessAction should fail for non-turn player")
	}
}
