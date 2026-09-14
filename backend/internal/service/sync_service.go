package service

import (
	"errors"
	"time"

	"media-sequencer/backend/internal/model"
	"media-sequencer/backend/internal/sync"
)

var ErrInvalidSyncMedia = errors.New("sync media is invalid")

type SyncService struct {
	manager *sync.Manager
}

func NewSyncService(manager *sync.Manager) *SyncService {
	return &SyncService{manager: manager}
}

func (s *SyncService) Start(media model.MediaItem, duration int) (model.SyncState, error) {
	if media.ID == "" {
		return model.SyncState{}, ErrInvalidSyncMedia
	}

	duration = SyncDuration(duration)
	state := model.SyncState{
		Active:    true,
		Media:     media,
		StartedAt: time.Now().UnixMilli(),
		Duration:  duration,
	}
	s.manager.Start(state)

	go func(startedAt int64, seconds int) {
		timer := time.NewTimer(time.Duration(seconds)*time.Second + 150*time.Millisecond)
		defer timer.Stop()
		<-timer.C
		s.manager.End(startedAt)
	}(state.StartedAt, state.Duration)

	return state, nil
}

func (s *SyncService) Subscribe() (chan []byte, func()) {
	return s.manager.Subscribe()
}
