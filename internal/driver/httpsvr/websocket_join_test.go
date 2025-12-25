package httpsvr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coder/websocket"
	"github.com/daominah/turn_based_game/internal/core/card_game_burn"
	"github.com/daominah/turn_based_game/internal/core/turnbased"
)

// TestJoinDuelViaURL tests that joining a duel via URL parameters works correctly
// This simulates the join URL feature where users can join with ?duelId=X&playerId=Y
func TestJoinDuelViaURL(t *testing.T) {
	// Setup
	manager := turnbased.NewInMemoryDuelsManager()
	connectionMgr := NewConnectionManager()
	handler := NewWebSocketHandler(
		map[string]turnbased.DuelsManager{
			card_game_burn.GameName: manager,
		},
		connectionMgr,
	)

	// Create a test duel first
	processor := NewBurnActionProcessor(manager)
	processor.SetConnectionManager(connectionMgr)
	duel, err := processor.CreateDuel(card_game_burn.GameName, []turnbased.PlayerID{"Alice_123456", "Bob_789012"})
	if err != nil {
		t.Fatalf("Failed to create duel: %v", err)
	}

	// Create WebSocket server
	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + server.URL[4:]

	// Test joining as first player
	t.Run("Join as first player", func(t *testing.T) {
		conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		// Send join_duel message
		joinMsg := ClientMessage{
			Type:     MessageTypeJoinDuel,
			DuelID:   string(duel.ID),
			PlayerID: "Alice_123456",
		}

		data, err := json.Marshal(joinMsg)
		if err != nil {
			t.Fatalf("Failed to marshal message: %v", err)
		}

		if err := conn.Write(context.Background(), websocket.MessageText, data); err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}

		// Read response
		_, responseData, err := conn.Read(context.Background())
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		var serverMsg ServerMessage
		if err := json.Unmarshal(responseData, &serverMsg); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Verify we got a state update
		if serverMsg.Type != MessageTypeStateUpdate {
			t.Errorf("Expected state_update, got %s", serverMsg.Type)
		}

		// Verify duel ID matches
		if serverMsg.Duel == nil {
			t.Fatal("Expected duel in response")
		}
		if serverMsg.Duel.ID != string(duel.ID) {
			t.Errorf("Expected duel ID %s, got %s", duel.ID, serverMsg.Duel.ID)
		}

		// Verify player is in the duel
		found := false
		for _, pid := range serverMsg.Duel.Players {
			if pid == "Alice_123456" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Player Alice_123456 not found in duel players")
		}

		// Verify game state is present
		if serverMsg.GameState == nil {
			t.Error("Expected game state in response")
		}
	})

	// Test joining as second player
	t.Run("Join as second player", func(t *testing.T) {
		conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		// Send join_duel message
		joinMsg := ClientMessage{
			Type:     MessageTypeJoinDuel,
			DuelID:   string(duel.ID),
			PlayerID: "Bob_789012",
		}

		data, err := json.Marshal(joinMsg)
		if err != nil {
			t.Fatalf("Failed to marshal message: %v", err)
		}

		if err := conn.Write(context.Background(), websocket.MessageText, data); err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}

		// Read response
		_, responseData, err := conn.Read(context.Background())
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		var serverMsg ServerMessage
		if err := json.Unmarshal(responseData, &serverMsg); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Verify we got a state update
		if serverMsg.Type != MessageTypeStateUpdate {
			t.Errorf("Expected state_update, got %s", serverMsg.Type)
		}

		// Verify player is in the duel
		found := false
		for _, pid := range serverMsg.Duel.Players {
			if pid == "Bob_789012" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Player Bob_789012 not found in duel players")
		}
	})

	// Test joining with invalid player
	t.Run("Join with invalid player", func(t *testing.T) {
		conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		// Send join_duel message with invalid player
		joinMsg := ClientMessage{
			Type:     MessageTypeJoinDuel,
			DuelID:   string(duel.ID),
			PlayerID: "InvalidPlayer",
		}

		data, err := json.Marshal(joinMsg)
		if err != nil {
			t.Fatalf("Failed to marshal message: %v", err)
		}

		if err := conn.Write(context.Background(), websocket.MessageText, data); err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}

		// Read response
		_, responseData, err := conn.Read(context.Background())
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		var serverMsg ServerMessage
		if err := json.Unmarshal(responseData, &serverMsg); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Verify we got an error
		if serverMsg.Type != MessageTypeError {
			t.Errorf("Expected error message, got %s", serverMsg.Type)
		}
	})

	// Test joining with invalid duel ID
	t.Run("Join with invalid duel ID", func(t *testing.T) {
		conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		// Send join_duel message with invalid duel ID
		joinMsg := ClientMessage{
			Type:     MessageTypeJoinDuel,
			DuelID:   "invalid_duel_id",
			PlayerID: "Alice_123456",
		}

		data, err := json.Marshal(joinMsg)
		if err != nil {
			t.Fatalf("Failed to marshal message: %v", err)
		}

		if err := conn.Write(context.Background(), websocket.MessageText, data); err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}

		// Read response
		_, responseData, err := conn.Read(context.Background())
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		var serverMsg ServerMessage
		if err := json.Unmarshal(responseData, &serverMsg); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Verify we got an error
		if serverMsg.Type != MessageTypeError {
			t.Errorf("Expected error message, got %s", serverMsg.Type)
		}
	})
}
