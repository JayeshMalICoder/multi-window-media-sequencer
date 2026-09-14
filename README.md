# Multi-Window Media Sequencer — Scalable Assignment

This version intentionally uses a proper backend architecture rather than placing all Go code in one file.

The assignment requires React on the frontend, Golang on the backend, persistent storage, dynamic playlist changes, sync playback and deployment. fileciteturn0file0L16-L35

## Project structure

```text
backend/
├── cmd/
│   └── server/
│       └── main.go                 # application entry point / DI
├── internal/
│   ├── config/
│   │   └── config.go               # environment configuration
│   ├── model/
│   │   ├── media.go                # domain entity
│   │   ├── window.go
│   │   └── sync.go
│   ├── repository/
│   │   ├── playlist_repository.go  # abstraction
│   │   └── json_playlist_repository.go
│   ├── service/
│   │   ├── playlist_service.go     # business/use-case logic
│   │   └── sync_service.go
│   ├── sync/
│   │   └── manager.go              # real-time event distribution
│   └── handler/
│       ├── router.go
│       └── handler.go               # HTTP adapter
├── data/
│   └── playlists.json
├── Dockerfile
└── go.mod

frontend/
├── src/
│   ├── main.jsx
│   └── styles.css
├── Dockerfile
├── package.json
└── index.html
```

## Architecture

```text
                 React
                   │
              REST + SSE
                   │
                   ▼
          ┌─────────────────┐
          │ HTTP Handlers   │
          │ Presentation    │
          └────────┬────────┘
                   ▼
          ┌─────────────────┐
          │ Services        │
          │ Business Rules  │
          └────────┬────────┘
                   ▼
          ┌─────────────────┐
          │ Repository      │
          │ Interface       │
          └────────┬────────┘
                   ▼
          ┌─────────────────┐
          │ JSON Repository │
          └────────┬────────┘
                   ▼
             Persistent DB
```

Sync is a separate cross-cutting component:

```text
Sync Service
     │
     ▼
Sync Manager
     │
     ├── Display 1
     ├── Display 2
     └── Display 3
```

## SOLID principles

### Single Responsibility Principle

Each component has one reason to change:

- `handler` changes when the HTTP contract changes.
- `service` changes when business rules change.
- `repository` changes when persistence changes.
- `sync.Manager` changes when event distribution changes.
- `model` contains domain data.

### Open/Closed Principle

`PlaylistRepository` is an interface. A future implementation can be:

```go
type PostgresPlaylistRepository struct {
    // ...
}
```

without rewriting `PlaylistService`.

### Liskov Substitution Principle

`PlaylistService` works with the `PlaylistRepository` abstraction, so JSON and PostgreSQL implementations can be substituted as long as they honor the interface contract.

### Interface Segregation Principle

The repository interface contains only operations needed by the playlist use case rather than exposing storage internals.

### Dependency Inversion Principle

The service depends on:

```go
repository.PlaylistRepository
```

instead of:

```text
JSON file
```

The concrete dependency is assembled in `cmd/server/main.go`.

## Why this structure is better for scaling

If the assignment later requires PostgreSQL:

```text
service
   ↓
PlaylistRepository interface
   ↓
PostgresPlaylistRepository
   ↓
PostgreSQL
```

Only the repository implementation needs to change.

If the system needs multiple backend instances:

```text
Display clients
      ↓
Load Balancer
   ↙       ↘
Go API     Go API
   ↓         ↓
PostgreSQL + Redis Pub/Sub
```

The current `sync.Manager` is deliberately isolated so the in-memory broadcaster can later be replaced with Redis Pub/Sub.

## 5-hour playback

The frontend treats 5 hours as the outer cycle boundary and repeats each configured playlist inside that cycle.

Blank is not automatically inserted into unused time. A blank screen is displayed only when a configured playlist item has `type: "blank"`. This follows the assignment's explicit blank-playback requirement. fileciteturn0file0L8-L15

## Sync behavior

When M2 is synchronized:

1. Backend creates a `SyncState`.
2. Backend records one server `startedAt`.
3. Backend broadcasts that state through SSE.
4. Every browser calculates the same end timestamp.
5. M2 overrides normal playback temporarily.
6. The playlist itself is never changed.
7. After the sync interval, each display returns to its own normal sequence.

This satisfies the requirement that the windows resume their configured sequences without losing playlist configuration. fileciteturn0file0L12-L15

## Local setup

### Backend

```bash
cd backend
go run ./cmd/server
```

Expected:

```text
Media Sequencer API running on :8080
```

### Frontend

Open another terminal:

```bash
cd frontend
npm install
npm run dev
```

Open:

```text
http://localhost:5173
```

### API

```text
GET  /api/health
GET  /api/windows
POST /api/windows/{windowId}
POST /api/sync
GET  /api/events
```

## Deployment

### Backend Docker

```bash
docker build -t media-sequencer-backend ./backend
docker run -p 8080:8080 media-sequencer-backend
```

Environment:

```text
PORT=8080
FRONTEND_ORIGIN=https://your-frontend-domain
DATA_FILE=data/playlists.json
```

For a real multi-instance production deployment, replace JSON storage with PostgreSQL and the in-memory sync manager with a shared pub/sub system.

### Frontend Docker

```bash
docker build --build-arg VITE_API_URL=https://your-backend-domain -t media-sequencer-frontend ./frontend
docker run -p 3000:80 media-sequencer-frontend
```

The assignment asks for deployment and clear access/configuration instructions. fileciteturn0file0L31-L35

## Demo checklist

1. Start Go backend.
2. Start React frontend.
3. Verify three windows play their own sequences.
4. Select M2.
5. Click Sync all windows.
6. Verify M2 appears in all displays.
7. Wait for sync duration.
8. Verify each window resumes its own playlist.
9. Add media to Window 1.
10. Verify the new media appears without refreshing the browser.
11. Restart backend.
12. Verify the playlist remains because data is persisted.

## Deliverables

The repository contains the React application, Go application, persistent storage, seed data and README documentation required by the assignment. fileciteturn0file0L36-L44
