package card_game_burn

import (
	"testing"

	"github.com/daominah/turn_based_game/internal/core/turnbased"
	"github.com/daominah/turn_based_game/internal/model"
)

func TestNewBurnDuel(t *testing.T) {
	players := []turnbased.PlayerID{"player1", "player2"}
	duel := NewBurnDuel(players)

	if duel == nil {
		t.Fatal("NewBurnDuel returned nil")
	}

	if duel.Duel == nil {
		t.Fatal("Duel is nil")
	}

	if len(duel.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(duel.Players))
	}

	// Check that each player has 8000 LP
	for _, ps := range duel.Players {
		if ps.LifePoint != 8000 {
			t.Errorf("Expected 8000 LP, got %f", ps.LifePoint)
		}
		if len(ps.Hand) != 5 {
			t.Errorf("Expected 5 cards in hand, got %d", len(ps.Hand))
		}
		if len(ps.Deck) != 15 {
			t.Errorf("Expected 15 cards in deck (20-5), got %d", len(ps.Deck))
		}
	}

	// Check that duel state is RUNNING
	if duel.Duel.State != turnbased.DuelStateRunning {
		t.Errorf("Expected state RUNNING, got %s", duel.Duel.State)
	}

	// Check that a turn player is set
	if duel.Duel.TurnPlayer == "" {
		t.Error("TurnPlayer should be set")
	}

	// Check that turn is 1
	if duel.Duel.Turn != 1 {
		t.Errorf("Expected turn 1, got %d", duel.Duel.Turn)
	}
}

func TestPlayCard(t *testing.T) {
	players := []turnbased.PlayerID{"player1", "player2"}
	duel := NewBurnDuel(players)

	turnPlayer := duel.Duel.TurnPlayer
	ps := duel.Players[turnPlayer]

	if len(ps.Hand) == 0 {
		t.Fatal("Player should have cards in hand")
	}

	card := ps.Hand[0]
	initialLP := ps.LifePoint
	opponentLP := 8000.0
	for pid, opp := range duel.Players {
		if pid != turnPlayer {
			opponentLP = opp.LifePoint
		}
	}

	// Play card with GAIN option
	success := duel.PlayCard(turnPlayer, card.UniqueCardID, PlayCardOptionGain)
	if !success {
		t.Error("PlayCard should succeed")
	}

	// Check that LP increased
	if ps.LifePoint != initialLP+card.Gain {
		t.Errorf("Expected LP %f, got %f", initialLP+card.Gain, ps.LifePoint)
	}

	// Check that card is removed from hand
	found := false
	for _, c := range ps.Hand {
		if c.UniqueCardID == card.UniqueCardID {
			found = true
			break
		}
	}
	if found {
		t.Error("Card should be removed from hand")
	}

	// Check that card is in graveyard
	found = false
	for _, c := range ps.Graveyard {
		if c.UniqueCardID == card.UniqueCardID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Card should be in graveyard")
	}

	// Test INFLICT option
	if len(ps.Hand) == 0 {
		t.Skip("No more cards to test INFLICT")
		return
	}

	card2 := ps.Hand[0]
	success = duel.PlayCard(turnPlayer, card2.UniqueCardID, PlayCardOptionInflict)
	if !success {
		t.Error("PlayCard with INFLICT should succeed")
	}

	// Check that opponent's LP decreased
	for pid, opp := range duel.Players {
		if pid != turnPlayer {
			if opp.LifePoint != opponentLP-card2.Inflict {
				t.Errorf("Expected opponent LP %f, got %f", opponentLP-card2.Inflict, opp.LifePoint)
			}
		}
	}
}

func TestPlayCard_NotTurnPlayer(t *testing.T) {
	players := []turnbased.PlayerID{"player1", "player2"}
	duel := NewBurnDuel(players)

	turnPlayer := duel.Duel.TurnPlayer
	var otherPlayer turnbased.PlayerID
	for _, pid := range players {
		if pid != turnPlayer {
			otherPlayer = pid
			break
		}
	}

	ps := duel.Players[turnPlayer]
	if len(ps.Hand) == 0 {
		t.Fatal("Player should have cards")
	}

	card := ps.Hand[0]

	// Try to play card as wrong player
	success := duel.PlayCard(otherPlayer, card.UniqueCardID, PlayCardOptionGain)
	if success {
		t.Error("PlayCard should fail for non-turn player")
	}
}

func TestEndTurn(t *testing.T) {
	players := []turnbased.PlayerID{"player1", "player2"}
	duel := NewBurnDuel(players)

	initialTurn := duel.Duel.Turn
	initialTurnPlayer := duel.Duel.TurnPlayer

	duel.EndTurn()

	// Check that turn increased
	if duel.Duel.Turn != initialTurn+1 {
		t.Errorf("Expected turn %d, got %d", initialTurn+1, duel.Duel.Turn)
	}

	// Check that turn player changed
	if duel.Duel.TurnPlayer == initialTurnPlayer {
		t.Error("Turn player should have changed")
	}

	// Check that new turn player drew a card
	newTurnPlayer := duel.Duel.TurnPlayer
	newPS := duel.Players[newTurnPlayer]
	// They should have 6 cards now (5 initial + 1 drawn)
	if len(newPS.Hand) != 6 {
		t.Errorf("Expected 6 cards in hand after drawing, got %d", len(newPS.Hand))
	}
}

func TestEndTurn_EmptyDeck(t *testing.T) {
	players := []turnbased.PlayerID{"player1", "player2"}
	duel := NewBurnDuel(players)

	// Empty the NEXT player's deck (the one who will draw after turn ends)
	turnPlayer := duel.Duel.TurnPlayer
	var nextPlayer turnbased.PlayerID
	for _, pid := range players {
		if pid != turnPlayer {
			nextPlayer = pid
			break
		}
	}
	nextPS := duel.Players[nextPlayer]
	nextPS.Deck = []Card{}

	// End turn - next player tries to draw but can't, so they lose and current player wins
	duel.EndTurn()

	// Check that duel ended
	if duel.Duel.State != turnbased.DuelStateEnd {
		t.Error("Duel should have ended due to empty deck")
	}

	// Check that the player who ended their turn won (because next player couldn't draw)
	if duel.Duel.Winner != turnPlayer {
		t.Errorf("Expected winner %s (the one who ended turn), got %s", turnPlayer, duel.Duel.Winner)
	}
}

func TestGetState(t *testing.T) {
	players := []turnbased.PlayerID{"player1", "player2"}
	duel := NewBurnDuel(players)

	state := duel.GetState()
	if state == nil {
		t.Fatal("GetState returned nil")
	}

	// GetState now returns model.BurnGameState
	gameState, ok := state.(model.BurnGameState)
	if !ok {
		t.Fatalf("GetState should return model.BurnGameState, got %T", state)
	}

	if len(gameState.Players) != 2 {
		t.Errorf("Expected 2 players in state, got %d", len(gameState.Players))
	}
}

func TestHandleActionWithPlayer(t *testing.T) {
	players := []turnbased.PlayerID{"player1", "player2"}
	duel := NewBurnDuel(players)

	turnPlayer := duel.Duel.TurnPlayer
	ps := duel.Players[turnPlayer]

	if len(ps.Hand) == 0 {
		t.Fatal("Player should have cards")
	}

	card := ps.Hand[0]

	// Test PlayCard action
	action := ActionPlayCard{
		CardID: card.UniqueCardID,
		Option: PlayCardOptionGain,
	}

	err := duel.HandleActionWithPlayer(action, turnPlayer)
	if err != nil {
		t.Errorf("HandleActionWithPlayer should succeed: %v", err)
	}

	// Test EndTurn action
	actionEnd := ActionEndTurn{}
	err = duel.HandleActionWithPlayer(actionEnd, turnPlayer)
	if err != nil {
		t.Errorf("HandleActionWithPlayer for EndTurn should succeed: %v", err)
	}

	// After EndTurn, the turn should have advanced
	// Now test that the PREVIOUS turn player (who is no longer the turn player) cannot end turn again
	// The current turn player should be the other player now
	currentTurnPlayer := duel.Duel.TurnPlayer
	if currentTurnPlayer == turnPlayer {
		t.Error("Turn should have advanced after EndTurn")
	}

	// Try to end turn as the previous turn player (should fail)
	err = duel.HandleActionWithPlayer(actionEnd, turnPlayer)
	if err == nil {
		t.Error("HandleActionWithPlayer should fail for non-turn player")
	}
}
