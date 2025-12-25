// Package card_game_burn demo for a simple turn-based card game,
// this demonstrates how to use the generic turn-based engine package
package card_game_burn

import (
	randc "crypto/rand"
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/daominah/turn_based_game/internal/core/turnbased"
)

const GameName = "CARD_GAME_BURN"

type Card struct {
	UniqueCardID UniqueCardID
	Gain         float64        // amount of LP gained if player chooses option to gain LP
	Inflict      float64        // amount of LP inflicted to opponent if player chooses option to burn
	PlayedOption PlayCardOption // empty at first, will be set by player when playing card
}

// UniqueCardID unique everywhere, so it easier to connect action to card,
// even the same copies with same effects still have different UniqueCardID
type UniqueCardID string

type PlayCardOption string // PlayCardOption can be Gain or Inflict

// PlayCardOption enum
const (
	PlayCardOptionPending PlayCardOption = "" // card is not played yet
	PlayCardOptionGain    PlayCardOption = "GAIN"
	PlayCardOptionInflict PlayCardOption = "INFLICT"
)

type PlayerState struct {
	ID        turnbased.PlayerID
	LifePoint float64
	Hand      []Card
	Deck      []Card
	// as current rule, field only needs 1 zone to play card,
	// and the card probably only lasts a second on the field to choose option,
	// then it resolves the Gain or Inflict effect, then send to the Graveyard
	Field     []Card
	Graveyard []Card
}

type BurnDuel struct {
	Duel    *turnbased.Duel
	Players map[turnbased.PlayerID]*PlayerState
}

func NewBurnDuel(players []turnbased.PlayerID) *BurnDuel {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	genericDuel := turnbased.NewDuel("", players)
	duel := &BurnDuel{
		Duel:    genericDuel,
		Players: make(map[turnbased.PlayerID]*PlayerState),
	}
	for _, pid := range players {
		// just default deck size of 20 cards, generated on the fly,
		// usually in a real game, the deck is predefined by players, should be arg for init duel func
		deck := make([]Card, 20)
		for i := range deck {
			deck[i] = Card{
				UniqueCardID: UUIDGen(),
				Gain:         float64((random.Intn(10) + 1) * 100),
				Inflict:      float64((random.Intn(30) + 1) * 100),
			}
		}
		duel.Players[pid] = &PlayerState{
			ID:        pid,
			LifePoint: 8000,
			Deck:      deck,
			Hand:      []Card{},
		}
	}
	// Toss coin for first turn
	first := players[random.Intn(len(players))]
	duel.Duel.TurnPlayer = first
	duel.Duel.Turn = 1
	// Draw 5 cards for each player
	for _, ps := range duel.Players {
		for i := 0; i < 5; i++ {
			ps.drawCard()
		}
	}
	duel.Duel.State = turnbased.DuelStateRunning
	// from now, wait for players to send actions,
	// check if the action is valid, update Duel state, and resolve actions
	// until one player wins (or draw)
	return duel
}

// drawCard draws a card from the player's deck to their hand,
// returns the drawn card or nil if deck is empty
func (ps *PlayerState) drawCard() *Card {
	if len(ps.Deck) == 0 {
		return nil
	}
	card := ps.Deck[0]
	ps.Hand = append(ps.Hand, card)
	ps.Deck = ps.Deck[1:]
	return &card
}

// PlayCard plays a card from hand (by UniqueCardID), applying its effect.
func (cgb *BurnDuel) PlayCard(
	player turnbased.PlayerID, cardID UniqueCardID, option PlayCardOption) bool {
	// Only current turn player can play card
	if cgb.Duel.TurnPlayer != player {
		return false
	}
	ps := cgb.Players[player]
	handIdx := -1
	for i, c := range ps.Hand {
		if c.UniqueCardID == cardID {
			handIdx = i
			break
		}
	}
	if handIdx == -1 {
		return false
	}
	card := ps.Hand[handIdx]
	// Set PlayedOption
	card.PlayedOption = option
	// Remove from hand, put to field
	ps.Hand = append(ps.Hand[:handIdx], ps.Hand[handIdx+1:]...)
	ps.Field = append(ps.Field, card)
	// Resolve effect
	if option == PlayCardOptionInflict {
		for pid, opp := range cgb.Players {
			if pid != player {
				opp.LifePoint -= card.Inflict
				if opp.LifePoint <= 0 {
					cgb.Duel.SetWinner(player)
				}
			}
		}
	} else if option == PlayCardOptionGain {
		ps.LifePoint += card.Gain
	}
	// Move card from field to graveyard
	ps.Graveyard = append(ps.Graveyard, card)
	ps.Field = ps.Field[:len(ps.Field)-1]

	// Log the action in the generic duel log
	cgb.Duel.LogAction(player, "PLAY_CARD", map[string]interface{}{
		"option":  string(option),
		"gain":    card.Gain,
		"inflict": card.Inflict,
	})

	return true
}

// EndTurn ends the current player's turn, draws a card for next player, advances turn.
func (cgb *BurnDuel) EndTurn() {
	// Save the player who is ending their turn (they will win if next player can't draw)
	endingPlayer := cgb.Duel.TurnPlayer

	// Log the end turn action
	cgb.Duel.LogAction(endingPlayer, "END_TURN", map[string]interface{}{})

	cgb.Duel.NextTurn()
	// Now the next player is the turn player, they draw at start of turn
	ps := cgb.Players[cgb.Duel.TurnPlayer]
	if ps.drawCard() == nil {
		// Deck is empty for the new turn player, so they lose
		// The player who just ended their turn wins
		cgb.Duel.SetWinner(endingPlayer)
	}
}

func UUIDGen() UniqueCardID {
	b := make([]byte, 16)
	_, _ = randc.Read(b)
	r := hex.EncodeToString(b)
	return UniqueCardID(r)
}
