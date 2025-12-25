package httpsvr

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/coder/websocket"
	"github.com/daominah/turn_based_game/internal/core/turnbased"
)

// ConnectionManager manages WebSocket connections for duels and players
type ConnectionManager struct {
	// duelID -> []*websocket.Conn
	duelConnections map[turnbased.DuelID][]*websocket.Conn
	// playerID -> *websocket.Conn (one connection per player)
	playerConnections map[turnbased.PlayerID]*websocket.Conn
	// conn -> playerID (reverse mapping for cleanup)
	connToPlayer map[*websocket.Conn]turnbased.PlayerID
	// conn -> duelID (track which duel a connection is watching)
	connToDuel map[*websocket.Conn]turnbased.DuelID
	mu         sync.RWMutex
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		duelConnections:   make(map[turnbased.DuelID][]*websocket.Conn),
		playerConnections: make(map[turnbased.PlayerID]*websocket.Conn),
		connToPlayer:      make(map[*websocket.Conn]turnbased.PlayerID),
		connToDuel:        make(map[*websocket.Conn]turnbased.DuelID),
	}
}

// AddConnection adds a WebSocket connection for a player in a duel
func (cm *ConnectionManager) AddConnection(conn *websocket.Conn, playerID turnbased.PlayerID, duelID turnbased.DuelID) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Remove old connection if player already has one
	if oldConn, exists := cm.playerConnections[playerID]; exists {
		cm.removeConnectionLocked(oldConn)
	}

	// Add new connection
	cm.playerConnections[playerID] = conn
	cm.connToPlayer[conn] = playerID
	cm.connToDuel[conn] = duelID

	// Add to duel connections
	cm.duelConnections[duelID] = append(cm.duelConnections[duelID], conn)

	log.Printf("Connection added: player=%s, duel=%s, total connections for duel=%d",
		playerID, duelID, len(cm.duelConnections[duelID]))
}

// RemoveConnection removes a WebSocket connection
func (cm *ConnectionManager) RemoveConnection(conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.removeConnectionLocked(conn)
}

func (cm *ConnectionManager) removeConnectionLocked(conn *websocket.Conn) {
	playerID, hasPlayer := cm.connToPlayer[conn]
	duelID, hasDuel := cm.connToDuel[conn]

	if hasPlayer {
		delete(cm.playerConnections, playerID)
		delete(cm.connToPlayer, conn)
	}

	if hasDuel {
		// Remove from duel connections slice
		conns := cm.duelConnections[duelID]
		for i, c := range conns {
			if c == conn {
				cm.duelConnections[duelID] = append(conns[:i], conns[i+1:]...)
				break
			}
		}
		// Clean up empty duel entry
		if len(cm.duelConnections[duelID]) == 0 {
			delete(cm.duelConnections, duelID)
		}
		delete(cm.connToDuel, conn)
	}

	if hasPlayer || hasDuel {
		log.Printf("Connection removed: player=%s, duel=%s", playerID, duelID)
	}
}

// BroadcastToDuel sends a message to all connections watching a duel
func (cm *ConnectionManager) BroadcastToDuel(duelID turnbased.DuelID, message ServerMessage) error {
	cm.mu.RLock()
	conns := make([]*websocket.Conn, len(cm.duelConnections[duelID]))
	copy(conns, cm.duelConnections[duelID])
	cm.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, conn := range conns {
		wg.Add(1)
		go func(c *websocket.Conn) {
			defer wg.Done()
			if err := c.Write(context.Background(), websocket.MessageText, data); err != nil {
				log.Printf("Error broadcasting to connection: %v", err)
				cm.RemoveConnection(c)
			}
		}(conn)
	}
	wg.Wait()

	return nil
}

// SendToPlayer sends a message to a specific player's connection
func (cm *ConnectionManager) SendToPlayer(playerID turnbased.PlayerID, message ServerMessage) error {
	cm.mu.RLock()
	conn, exists := cm.playerConnections[playerID]
	cm.mu.RUnlock()

	if !exists {
		return nil // Player not connected, silently ignore
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return conn.Write(context.Background(), websocket.MessageText, data)
}
