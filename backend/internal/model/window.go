package model

import "time"

type DisplayWindow struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Playlist  []MediaItem `json:"playlist"`
	UpdatedAt time.Time   `json:"updatedAt"`
}
