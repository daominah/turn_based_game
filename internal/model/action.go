// Package model defines shared data models used across packages
package model

// ActionData represents action data sent from client
// This is a union type - the actual structure depends on the game and action type
type ActionData struct {
	// For Burn game PlayCard action
	CardID *string `json:"card_id,omitempty"`
	Option *string `json:"option,omitempty"`

	// For EndTurn action
	EndTurn *bool `json:"end_turn,omitempty"`
}
