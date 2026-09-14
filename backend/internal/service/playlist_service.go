package service

import (
	"errors"
	"strings"
	"time"

	"media-sequencer/backend/internal/model"
	"media-sequencer/backend/internal/repository"
	"media-sequencer/backend/internal/sync"
)

var (
	ErrInvalidMedia = errors.New("invalid media")
	ErrInvalidWindow = errors.New("invalid window")
)

type PlaylistService struct {
	repo repository.PlaylistRepository
	sync *sync.Manager
}

func NewPlaylistService(repo repository.PlaylistRepository, syncManager *sync.Manager) *PlaylistService {
	return &PlaylistService{repo: repo, sync: syncManager}
}

func (s *PlaylistService) GetWindows() ([]model.DisplayWindow, error) {
	return s.repo.GetAll()
}

func (s *PlaylistService) AddMedia(windowID string, media model.MediaItem) (model.DisplayWindow, error) {
	windowID = strings.TrimSpace(windowID)
	media.ID = strings.TrimSpace(media.ID)
	media.Name = strings.TrimSpace(media.Name)

	if windowID == "" {
		return model.DisplayWindow{}, ErrInvalidWindow
	}
	if media.ID == "" || media.Name == "" {
		return model.DisplayWindow{}, ErrInvalidMedia
	}
	if media.Type != model.MediaTypeImage && media.Type != model.MediaTypeVideo && media.Type != model.MediaTypeBlank {
		return model.DisplayWindow{}, ErrInvalidMedia
	}
	if media.Duration <= 0 {
		media.Duration = 10
	}
	if media.Type == model.MediaTypeBlank {
		media.URL = ""
	}

	updated, err := s.repo.AppendMedia(windowID, media)
	if err != nil {
		return model.DisplayWindow{}, err
	}

	s.sync.PublishPlaylist(updated)
	return updated, nil
}

func SyncDuration(duration int) int {
	if duration <= 0 {
		return 10
	}
	return duration
}

func IsExpired(state model.SyncState, now time.Time) bool {
	return now.UnixMilli() >= state.StartedAt+int64(state.Duration)*1000
}
