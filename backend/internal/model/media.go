package model

// MediaType is intentionally a small domain type.
// New media types can be introduced without changing handlers.
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
	MediaTypeBlank MediaType = "blank"
)

type MediaItem struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Type     MediaType `json:"type"`
	URL      string    `json:"url,omitempty"`
	Duration int       `json:"duration"` // seconds
}
