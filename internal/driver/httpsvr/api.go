// Package httpsvr is an HTTP server that allows users to interact with the app
package httpsvr

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/daominah/turn_based_game/internal/core/turnbased"
)

// allowCORS is a middleware that sets CORS headers to allow requests from specified origins.
func allowCORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if strings.HasPrefix(origin, "http://localhost:") ||
				strings.HasPrefix(origin, "http://127.0.0.1:") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// NewHandlerAPI creates an API handler that routes actions to the correct DuelsManager by game name.
func NewHandlerAPI(duelsManagers map[string]turnbased.DuelsManager) http.Handler {
	handler := http.NewServeMux()

	handler.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprintf("Hello from Go backend internal/driver/httpsvr/api.go at %v",
			time.Now().Format(time.RFC3339Nano))
		_, _ = w.Write([]byte(response))
	})

	// Example: POST /api/duel?game=GAME_NAME
	handler.HandleFunc("/api/duel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			game := r.URL.Query().Get("game")
			if game == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("missing game name"))
				return
			}
			duelsManager, ok := duelsManagers[game]
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("unknown game: " + game))
				return
			}
			_ = duelsManager
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Example: POST /api/duel/{duelID}/action?game=GAME_NAME
	handler.HandleFunc("/api/duel/{duelID}/action", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			game := r.URL.Query().Get("game")
			if game == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("missing game name"))
				return
			}
			duelsManager, ok := duelsManagers[game]
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("unknown game: " + game))
				return
			}
			_ = duelsManager
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	return allowCORS()(handler)
}
