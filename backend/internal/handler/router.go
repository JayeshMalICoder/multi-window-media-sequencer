package handler

import (
	"net/http"

	"media-sequencer/backend/internal/service"
)

func NewRouter(playlistService *service.PlaylistService, syncService *service.SyncService) http.Handler {
	mux := http.NewServeMux()
	h := NewHandler(playlistService, syncService)

	mux.HandleFunc("/api/health", h.Health)
	mux.HandleFunc("/api/windows", h.Windows)
	mux.HandleFunc("/api/windows/", h.Window)
	mux.HandleFunc("/api/sync", h.Sync)
	mux.HandleFunc("/api/events", h.Events)

	return mux
}
