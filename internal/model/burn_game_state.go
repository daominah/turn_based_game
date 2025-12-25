// Package model defines shared data models used across packages
package model

// BurnCard represents a card in the Burn card game for serialization
type BurnCard struct {
	UniqueCardID string  `json:"unique_card_id"`
	Gain         float64 `json:"gain"`
	Inflict      float64 `json:"inflict"`
	PlayedOption string  `json:"played_option,omitempty"`
}

// BurnPlayerState represents a player's state in the Burn card game
type BurnPlayerState struct {
	ID        string     `json:"id"`
	LifePoint float64    `json:"life_point"`
	Hand      []BurnCard `json:"hand"`
	DeckSize  int        `json:"deck_size"`
	Field     []BurnCard `json:"field"`
	Graveyard []BurnCard `json:"graveyard"`
}

// BurnGameState represents the complete game state for the Burn card game
type BurnGameState struct {
	Players map[string]BurnPlayerState `json:"players"`
}
