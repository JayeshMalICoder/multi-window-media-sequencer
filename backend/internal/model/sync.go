package model

type SyncState struct {
	Active    bool      `json:"active"`
	Media     MediaItem `json:"media"`
	StartedAt int64     `json:"startedAt"` // Unix milliseconds
	Duration  int       `json:"duration"`  // seconds
}
