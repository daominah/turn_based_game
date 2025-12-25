package card_game_burn

import (
	"github.com/daominah/turn_based_game/internal/model"
)

// ToModelBurnGameState converts a BurnDuel to model.BurnGameState
func (cgb *BurnDuel) ToModelBurnGameState() model.BurnGameState {
	playersState := make(map[string]model.BurnPlayerState)

	for pid, ps := range cgb.Players {
		// Convert cards
		hand := make([]model.BurnCard, len(ps.Hand))
		for i, c := range ps.Hand {
			hand[i] = model.BurnCard{
				UniqueCardID: string(c.UniqueCardID),
				Gain:         c.Gain,
				Inflict:      c.Inflict,
				PlayedOption: string(c.PlayedOption),
			}
		}

		field := make([]model.BurnCard, len(ps.Field))
		for i, c := range ps.Field {
			field[i] = model.BurnCard{
				UniqueCardID: string(c.UniqueCardID),
				Gain:         c.Gain,
				Inflict:      c.Inflict,
				PlayedOption: string(c.PlayedOption),
			}
		}

		graveyard := make([]model.BurnCard, len(ps.Graveyard))
		for i, c := range ps.Graveyard {
			graveyard[i] = model.BurnCard{
				UniqueCardID: string(c.UniqueCardID),
				Gain:         c.Gain,
				Inflict:      c.Inflict,
				PlayedOption: string(c.PlayedOption),
			}
		}

		playersState[string(pid)] = model.BurnPlayerState{
			ID:        string(ps.ID),
			LifePoint: ps.LifePoint,
			Hand:      hand,
			DeckSize:  len(ps.Deck),
			Field:     field,
			Graveyard: graveyard,
		}
	}

	return model.BurnGameState{
		Players: playersState,
	}
}
