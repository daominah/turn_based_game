package generic_turn_based

import (
	"fmt"
	"sync"
	"time"
)

// DuelsManager manages all duels (open/ongoing and ended).
// It is defined as an interface to allow different storage backends (in-memory, DB, etc).
type DuelsManager interface {
	// CreateDuel adds a new duel and returns its ID.
	CreateDuel(duel *Duel) DuelID
	// GetDuel returns the duel by ID, or nil if not found.
	GetDuel(id DuelID) *Duel
	// ListDuels returns all duels (optionally filter by open/ended).
	ListDuels(openOnly bool) []*Duel
	// UpdateDuel updates the duel state by ID.
	UpdateDuel(duel *Duel) bool
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

func (m *InMemoryDuelsManager) CreateDuel(duel *Duel) DuelID {
	id := DuelID(fmt.Sprintf("duel_%d", time.Now().UnixNano()))
	duel.ID = id
	m.mu.Lock()
	m.duels[id] = duel
	m.mu.Unlock()
	return id
}

func (m *InMemoryDuelsManager) GetDuel(id DuelID) *Duel {
	m.mu.RLock()
	duel := m.duels[id]
	m.mu.RUnlock()
	return duel
}

func (m *InMemoryDuelsManager) ListDuels(openOnly bool) []*Duel {
	m.mu.RLock()
	var result []*Duel
	for _, d := range m.duels {
		if !openOnly || (d.State != DuelStateEnd) {
			result = append(result, d)
		}
	}
	m.mu.RUnlock()
	return result
}

func (m *InMemoryDuelsManager) UpdateDuel(duel *Duel) bool {
	if duel == nil || duel.ID == "" {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.duels[duel.ID]; !ok {
		return false
	}
	m.duels[duel.ID] = duel
	return true
}
