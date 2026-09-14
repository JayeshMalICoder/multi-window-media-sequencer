package repository

import (
	"errors"

	"media-sequencer/backend/internal/model"
)

var ErrWindowNotFound = errors.New("window not found")

// PlaylistRepository describes what the service needs from persistence.
// The service does not know whether the implementation is JSON, PostgreSQL,
// MongoDB, etc. This is the Dependency Inversion Principle in practice.
type PlaylistRepository interface {
	GetAll() ([]model.DisplayWindow, error)
	AppendMedia(windowID string, media model.MediaItem) (model.DisplayWindow, error)
}
