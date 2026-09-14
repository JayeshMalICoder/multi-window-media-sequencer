# Backend Architecture

The backend follows a layered architecture designed for future scaling.

```text
HTTP Handler
     ↓
Service / Use Case
     ↓
Repository Interface
     ↓
JSON Repository
     ↓
Persistent Storage
```

Synchronization is isolated behind a manager so it can later be replaced with Redis Pub/Sub, NATS or Kafka.

## SOLID application

- **S — Single Responsibility:** handlers handle HTTP, services handle business rules, repositories handle persistence, sync manager handles event distribution.
- **O — Open/Closed:** add a PostgreSQL repository without changing service logic.
- **L — Liskov Substitution:** any `PlaylistRepository` implementation can be injected.
- **I — Interface Segregation:** the repository exposes only persistence operations needed by the playlist service.
- **D — Dependency Inversion:** services depend on the repository abstraction rather than JSON/files.

## Run

```bash
go run ./cmd/server
```

Default: `http://localhost:8080`
