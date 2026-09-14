package main

import (
	"log"
	"net/http"
	"os"

	"media-sequencer/backend/internal/config"
	"media-sequencer/backend/internal/handler"
	"media-sequencer/backend/internal/repository"
	"media-sequencer/backend/internal/service"
	"media-sequencer/backend/internal/sync"
)

func main() {
	cfg := config.Load()

	// Composition root:
	// dependencies are created here and injected into handlers/services.
	// This keeps business logic independent from HTTP and storage details.
	repo := repository.NewJSONPlaylistRepository(cfg.DataFile)
	syncManager := sync.NewManager()
	playlistService := service.NewPlaylistService(repo, syncManager)
	syncService := service.NewSyncService(syncManager)
	router := handler.NewRouter(playlistService, syncService)

	log.Printf("Media Sequencer API running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, withCORS(router, cfg.FrontendOrigin)))
}

func withCORS(next http.Handler, origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var _ = os.Getenv
