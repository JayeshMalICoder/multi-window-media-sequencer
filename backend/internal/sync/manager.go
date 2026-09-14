package sync

import (
	"encoding/json"
	"sync"

	"media-sequencer/backend/internal/model"
)

// Manager owns synchronization state and subscriber delivery.
// Keeping this concern separate makes it replaceable later with Redis Pub/Sub,
// NATS, Kafka, etc. when the application is deployed as multiple replicas.
type Manager struct {
	mu       sync.RWMutex
	current  *model.SyncState
	clients  map[chan []byte]struct{}
}

func NewManager() *Manager {
	return &Manager{clients: make(map[chan []byte]struct{})}
}

func (m *Manager) Subscribe() (chan []byte, func()) {
	ch := make(chan []byte, 10)
	m.mu.Lock()
	m.clients[ch] = struct{}{}
	m.mu.Unlock()

	return ch, func() {
		m.mu.Lock()
		if _, ok := m.clients[ch]; ok {
			delete(m.clients, ch)
			close(ch)
		}
		m.mu.Unlock()
	}
}

func (m *Manager) Start(state model.SyncState) {
	m.mu.Lock()
	m.current = &state
	m.mu.Unlock()
	m.publish("sync_started", state)
}

func (m *Manager) End(startedAt int64) {
	m.mu.Lock()
	if m.current == nil || m.current.StartedAt != startedAt {
		m.mu.Unlock()
		return
	}
	m.current = nil
	m.mu.Unlock()
	m.publish("sync_ended", map[string]any{"endedAt": startedAt})
}

func (m *Manager) PublishPlaylist(window model.DisplayWindow) {
	m.publish("playlist_updated", window)
}

func (m *Manager) publish(event string, payload any) {
	data, _ := json.Marshal(map[string]any{"event": event, "data": payload})

	m.mu.RLock()
	defer m.mu.RUnlock()
	for client := range m.clients {
		select {
		case client <- data:
		default:
			// Do not let one slow display block all other displays.
		}
	}
}
