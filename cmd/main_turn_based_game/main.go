package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/daominah/turn_based_game/internal/core/card_game_burn"
	"github.com/daominah/turn_based_game/internal/core/turnbased"
	"github.com/daominah/turn_based_game/internal/driver/httpsvr"
)

func main() {
	log.SetFlags(log.Lshortfile)
	log.SetOutput(customLogger{})

	listenPort := ":11995"

	// init DuelsManager for each game,
	// the centralized duelsManagers is read-only after this point
	duelsManagers := map[string]turnbased.DuelsManager{
		card_game_burn.GameName: turnbased.NewInMemoryDuelsManager(),
		// Add more games here as needed
	}

	guiHandler, err := httpsvr.NewHandlerGUI("")
	if err != nil {
		log.Fatalf("error NewHandlerGUI: %v", err)
	}
	apiHandler := httpsvr.NewHandlerAPI(duelsManagers)

	// Setup WebSocket handler
	connectionMgr := httpsvr.NewConnectionManager()
	wsHandler := httpsvr.NewWebSocketHandler(duelsManagers, connectionMgr)

	mux := http.NewServeMux()
	mux.Handle("/api/", apiHandler)
	mux.HandleFunc("/ws", wsHandler.HandleWebSocket)
	mux.Handle("/", guiHandler)

	log.Printf("serving API and user interface on http://localhost%v", listenPort)
	err = http.ListenAndServe(listenPort, mux)
	if err != nil {
		log.Fatalf("error ListenAndServe: %v", err)
	}
}

// customLogger adds time to the beginning of each log line, write to stdout
type customLogger struct{}

func (writer customLogger) Write(bytes []byte) (int, error) {
	return fmt.Printf("%v %s", time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00"), bytes)
}
