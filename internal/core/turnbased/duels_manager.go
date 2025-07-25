package turnbased

import (
	"fmt"
	"sync"
	"time"
)

// DuelsManager manages all duels (running and ended).
// It is defined as an interface to allow different storage backends (in-memory, DB, etc).
type DuelsManager interface {
	// CreateDuel adds a new duel and returns the created duel.
	CreateDuel(duel *Duel) *Duel
	// GetDuel returns the duel by ID, or nil if not found.
	GetDuel(id DuelID) *Duel
	// UpdateDuel updates the duel state by ID.
	UpdateDuel(duel *Duel) (*Duel, error)
}

var _ DuelsManager = (*InMemoryDuelsManager)(nil) // ensure interface compliance

// Action represents a generic action sent by a player to interact with a duel.
type Action interface {
	// GameName used to route the action to the DuelsManager corresponding to the game.
	GameName() string
	// DuelID returns the ID of the duel this action targets.
	DuelID() DuelID
	// PlayerID returns the ID of the player performing the action.
	PlayerID() PlayerID
}

// InMemoryDuelsManager is an in-memory implementation of DuelsManager using a Go map.
type InMemoryDuelsManager struct {
	duels map[DuelID]*Duel
	mu    sync.RWMutex // protects duels
}

// NewInMemoryDuelsManager creates a new in-memory duels manager.
func NewInMemoryDuelsManager() *InMemoryDuelsManager {
	return &InMemoryDuelsManager{
		duels: make(map[DuelID]*Duel),
	}
}

func (m *InMemoryDuelsManager) CreateDuel(duel *Duel) *Duel {
	id := DuelID(fmt.Sprintf("duel_%d", time.Now().UnixNano()))
	duel.ID = id
	m.mu.Lock()
	m.duels[id] = duel
	m.mu.Unlock()
	return duel
}

func (m *InMemoryDuelsManager) GetDuel(id DuelID) *Duel {
	m.mu.RLock()
	duel := m.duels[id]
	m.mu.RUnlock()
	return duel
}

func (m *InMemoryDuelsManager) UpdateDuel(duel *Duel) (*Duel, error) {
	if duel == nil || duel.ID == "" {
		return nil, fmt.Errorf("empty duel ID")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.duels[duel.ID]
	if !ok {
		return nil, fmt.Errorf("duel not found")
	}
	if existing.State == DuelStateEnd {
		return existing, fmt.Errorf("duel already ended")
	}
	m.duels[duel.ID] = duel
	return duel, nil
}
