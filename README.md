# SentinelDB

SentinelDB is a self-hosted monitoring service for internet-facing assets. It exposes an HTTP API to register assets and trigger scans, and processes those scans asynchronously through a PostgreSQL-backed transactional outbox.

## Current Features

- Asset CRUD for `ip`, `domain`, and `email`
- Manual trigger of scans for all active assets or a single asset
- Asynchronous worker that consumes jobs with `SELECT FOR UPDATE SKIP LOCKED`
- InternetDB integration for IP-based lookups
- Snapshot persistence and diff-based finding creation
- Telegram notification delivery for newly detected findings
- Run tracking with automatic completion when all jobs finish
- Retry with incremental backoff for failed jobs
- Graceful shutdown for both API and worker processes

## Architecture

SentinelDB uses PostgreSQL as both the primary datastore and the job queue.

```text
POST /api/v1/trigger
      |
      v
API creates a run + outbox jobs in the same DB transaction
      |
      v
Worker dequeues pending jobs with SKIP LOCKED
      |
      v
InternetDB lookup runs for supported assets
      |
      v
Snapshot is stored and compared with the previous snapshot
      |
      v
New findings are persisted and optionally sent to Telegram
```

### Key Decisions

- **Transactional outbox:** run creation and job enqueueing are atomic
- **PostgreSQL queue:** no external broker is required
- **`SKIP LOCKED`:** concurrent workers do not process the same job twice
- **Snapshot diffing:** findings are created only when data changes
- **Retry/backoff:** failed jobs return to `pending` with a future `scheduled_at`

## Implemented API

| Method | Route | Description |
| --- | --- | --- |
| POST | `/api/v1/assets` | Create an asset |
| GET | `/api/v1/assets` | List assets |
| GET | `/api/v1/assets/:id` | Get one asset |
| PUT | `/api/v1/assets/:id` | Update label/active |
| DELETE | `/api/v1/assets/:id` | Soft-delete an asset |
| POST | `/api/v1/trigger` | Trigger jobs for all active supported assets |
| POST | `/api/v1/trigger/:id` | Trigger jobs for one supported asset |
| GET | `/api/v1/runs` | List runs |
| GET | `/api/v1/runs/:id` | Get one run |
| GET | `/api/v1/findings` | List findings |
| GET | `/api/v1/findings/:id` | Get one finding |
| PATCH | `/api/v1/findings/:id/resolve` | Mark a finding as closed |

## Data Model

| Table | Purpose |
| --- | --- |
| `assets` | Registered assets to monitor |
| `runs` | Scan execution records |
| `outboxes` | Pending/processing/completed jobs |
| `asset_snapshots` | Raw InternetDB snapshots |
| `findings` | Changes detected between snapshots |

## Project Structure

```text
sentineldb/
├── cmd/
│   ├── api/                 # HTTP API entrypoint
│   └── worker/              # Background worker entrypoint
├── internal/
│   ├── job/
│   │   ├── domain/          # Repositories and business rules
│   │   ├── handlers/        # Echo handlers
│   │   ├── models/          # GORM models
│   │   └── routes/          # Route wiring
│   ├── services/            # InternetDB and Telegram integrations
│   ├── storage/             # PostgreSQL connection setup
│   └── worker/              # Dequeue and job processing logic
├── pkg/
│   └── logger/              # Application logger
├── tests/
│   └── integration/         # PostgreSQL-backed integration tests
├── instructions/            # Project requirements and implementation notes
├── Makefile
└── README.md
```

## Running Locally

### Requirements

- Go 1.24+
- PostgreSQL

### Environment

Create a `.env` file in the repository root or export the variables in your shell.

| Variable | Required | Description |
| --- | --- | --- |
| `SERVER_PORT` | API only | Port used by `cmd/api` |
| `DATABASE_URL` | yes | PostgreSQL connection string |
| `DB_SCHEMA` | no | Optional PostgreSQL schema via `search_path` |
| `JWT_SECRET_KEY` | API only | API startup requirement |
| `TELEGRAM_BOT_TOKEN` | no | Enables Telegram delivery |
| `TELEGRAM_CHAT_ID` | no | Enables Telegram delivery |
| `INTEGRATION_TEST` | no | Set to `1` to run integration tests |

### Start the API

```bash
go run cmd/api/main.go
```

### Start the Worker

```bash
go run cmd/worker/main.go
```

## Testing

### Unit Tests

```bash
go test ./...
```

### Integration Tests

Integration tests use a real PostgreSQL instance and only run when `INTEGRATION_TEST=1`.

```bash
INTEGRATION_TEST=1 go test ./tests/integration/...
```

### Makefile Helpers

```bash
make test
make test-verbose
make test-race
make test-cover
make test-integration
```

## Test Coverage Areas

The repository currently includes:

- Handler-level unit tests with mocked repositories
- Service and storage tests
- Integration tests for:
  - transactional outbox atomicity
  - concurrent dequeue with `SKIP LOCKED`
  - full job lifecycle completion
  - retry and backoff behavior

## Current Scope vs Roadmap

### Implemented

- [x] Asset CRUD
- [x] Worker processing with transactional outbox
- [x] InternetDB-based snapshot ingestion
- [x] Findings generation from snapshot diffs
- [x] Telegram notification integration
- [x] Unit and integration test suites

### Planned / Not Yet Implemented

- [ ] HaveIBeenPwned integration
- [ ] Cross-source correlation
- [ ] Metrics endpoint
- [ ] OpenTelemetry tracing
- [ ] Grafana dashboard
