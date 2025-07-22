package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// This script requires the server (main.go) to be running.
// It calls the API to play a simple generic turn-based game.
func main() {

	baseURL := "http://localhost:11995/api"

	// 1. Create a new duel
	createReq := map[string]interface{}{
		"players": []string{"alice", "bob"},
	}
	duelID, err := apiPost(baseURL+"/duel", createReq)
	if err != nil {
		fmt.Println("Error creating duel:", err)
		return
	}
	fmt.Println("Created duel with ID:", duelID)

	// 2. Simulate a few turns (this is a placeholder, adjust to your API)
	for i := 0; i < 3; i++ {
		turnReq := map[string]interface{}{
			"duel_id": duelID,
			"action":  "next_turn",
		}
		resp, err := apiPost(baseURL+"/duel/"+duelID+"/action", turnReq)
		if err != nil {
			fmt.Println("Error performing turn:", err)
			return
		}
		fmt.Printf("Turn %d response: %v\n", i+1, resp)
	}

	// 3. End the duel (placeholder)
	// Removed: duel ends automatically based on state, no manual end request needed
}

// apiPost sends a POST request with JSON and returns the response as a string.
func apiPost(url string, data map[string]interface{}) (string, error) {
	body, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return string(respBody), nil
}
