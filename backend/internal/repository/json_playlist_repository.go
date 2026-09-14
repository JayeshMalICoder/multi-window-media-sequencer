package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"media-sequencer/backend/internal/model"
)

type JSONPlaylistRepository struct {
	mu       sync.RWMutex
	filePath string
	windows  []model.DisplayWindow
}

func NewJSONPlaylistRepository(filePath string) *JSONPlaylistRepository {
	repo := &JSONPlaylistRepository{filePath: filePath}
	if err := repo.load(); err != nil {
		// If no file exists, seed data is created. For malformed existing data,
		// failing fast is safer than silently overwriting user configuration.
		if os.IsNotExist(err) {
			repo.windows = seedWindows()
			if saveErr := repo.saveLocked(); saveErr != nil {
				panic(saveErr)
			}
		} else {
			panic(err)
		}
	}
	return repo
}

func (r *JSONPlaylistRepository) load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	raw, err := os.ReadFile(r.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, &r.windows)
}

func (r *JSONPlaylistRepository) GetAll() ([]model.DisplayWindow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]model.DisplayWindow, len(r.windows))
	copy(result, r.windows)
	for i := range result {
		result[i].Playlist = append([]model.MediaItem(nil), result[i].Playlist...)
	}
	return result, nil
}

func (r *JSONPlaylistRepository) AppendMedia(windowID string, media model.MediaItem) (model.DisplayWindow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.windows {
		if r.windows[i].ID == windowID {
			r.windows[i].Playlist = append(r.windows[i].Playlist, media)
			r.windows[i].UpdatedAt = time.Now().UTC()

			if err := r.saveLocked(); err != nil {
				return model.DisplayWindow{}, err
			}
			return r.windows[i], nil
		}
	}
	return model.DisplayWindow{}, ErrWindowNotFound
}

func (r *JSONPlaylistRepository) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(r.filePath), 0755); err != nil {
		return err
	}

	raw, err := json.MarshalIndent(r.windows, "", "  ")
	if err != nil {
		return err
	}

	temp := r.filePath + ".tmp"
	if err := os.WriteFile(temp, raw, 0644); err != nil {
		return err
	}
	return os.Rename(temp, r.filePath)
}

func seedWindows() []model.DisplayWindow {
	now := time.Now().UTC()
	return []model.DisplayWindow{
		{
			ID: "window-1", Name: "Display Window 1", UpdatedAt: now,
			Playlist: []model.MediaItem{
				{ID: "M1", Name: "Mountain", Type: model.MediaTypeImage, URL: "https://images.unsplash.com/photo-1500534623283-312aade485b7?auto=format&fit=crop&w=1200&q=80", Duration: 8},
				{ID: "M2", Name: "Ocean", Type: model.MediaTypeImage, URL: "https://images.unsplash.com/photo-1507525428034-b723cf961d3e?auto=format&fit=crop&w=1200&q=80", Duration: 8},
				{ID: "M3", Name: "Sample Video", Type: model.MediaTypeVideo, URL: "https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4", Duration: 12},
			},
		},
		{
			ID: "window-2", Name: "Display Window 2", UpdatedAt: now,
			Playlist: []model.MediaItem{
				{ID: "M4", Name: "City", Type: model.MediaTypeImage, URL: "https://images.unsplash.com/photo-1477959858617-67f85cf4f1df?auto=format&fit=crop&w=1200&q=80", Duration: 7},
				{ID: "M2", Name: "Ocean", Type: model.MediaTypeImage, URL: "https://images.unsplash.com/photo-1507525428034-b723cf961d3e?auto=format&fit=crop&w=1200&q=80", Duration: 8},
				{ID: "M5", Name: "Blank", Type: model.MediaTypeBlank, Duration: 5},
			},
		},
		{
			ID: "window-3", Name: "Display Window 3", UpdatedAt: now,
			Playlist: []model.MediaItem{
				{ID: "M6", Name: "Forest", Type: model.MediaTypeImage, URL: "https://images.unsplash.com/photo-1441974231531-c6227db76b6e?auto=format&fit=crop&w=1200&q=80", Duration: 9},
				{ID: "M3", Name: "Sample Video", Type: model.MediaTypeVideo, URL: "https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4", Duration: 12},
			},
		},
	}
}
