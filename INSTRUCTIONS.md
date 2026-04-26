# SentinelDB — Architecture Guide

This document describes the architectural decisions, project structure, and development conventions adopted in SentinelDB. It serves as a reference for understanding how the pieces fit together and why each decision was made.

---

## Project Structure

SentinelDB is organized as a **monorepo** — all components live in a single repository and share packages. This decision was made because the API, worker, and scheduler share the same domain types, interfaces, and infrastructure code. Splitting into multiple repositories would require publishing shared modules or using git submodules, adding unnecessary complexity for a solo project.

```
sentineldb/
├── cmd/
│   ├── api/                  # API binary entrypoint
│   │   ├── main.go
│   │   └── main_test.go
│   └── worker/               # Worker binary entrypoint
│       ├── main.go
│       └── main_test.go
├── internal/
│   ├── job/
│   │   ├── domain/           # Business logic, validation, repository interfaces
│   │   │   ├── asset.go
│   │   │   └── asset_test.go
│   │   ├── handlers/         # HTTP handlers and DTOs
│   │   │   ├── asset.go
│   │   │   ├── asset_dto.go
│   │   │   ├── asset_test.go
│   │   │   ├── run.go
│   │   │   ├── run_dto.go
│   │   │   └── run_test.go
│   │   ├── models/           # GORM models (database representation)
│   │   │   └── asset.go
│   │   └── routes/           # Route registration
│   │       └── routes.go
│   ├── middlewares/          # Echo middlewares
│   │   └── security.go
│   └── storage/              # Database connection
│       ├── postgresql.go
│       └── postgresql_test.go
├── pkg/
│   ├── logger/               # Structured logger (reusable)
│   └── retry/                # Exponential backoff and circuit breaker (reusable)
├── tests/
│   └── integration/          # Integration tests requiring real infrastructure
├── docker-compose.yml
├── .env.example
├── go.mod
├── go.sum
├── README.md
└── ARCHITECTURE.md
```

---

## Component Separation

Each `cmd/` directory produces an independent binary. They share internal packages but are deployed and run separately.

```
cmd/api     → serves HTTP requests, writes to outboxes
cmd/worker  → consumes outboxes, calls external APIs, persists findings
```

This separation means the API never blocks waiting for external calls (Shodan, HIBP). It enqueues work and returns immediately. The worker processes jobs asynchronously in the background.

---

## Layered Architecture

Each domain follows a consistent four-layer structure:

```
routes      → registers endpoints, wires dependencies
handlers    → receives HTTP request, validates input, calls domain
domain      → business logic, validation, repository interfaces
models      → database representation (GORM structs)
```

### The three representations of the same entity

An `Asset` exists in three different forms intentionally:

```
models.Asset         → how the database sees it (GORM tags)
handlers.AssetDTO    → how the API sees it (JSON tags, no internal fields)
domain.Asset         → what it means to the business (if needed)
```

Changing the database schema does not break the API contract, and changing the API contract does not require a database migration. Each layer is isolated.

---

## Dependency Injection

SentinelDB uses constructor-based dependency injection throughout. No global state, no service locators.

The flow of dependencies:

```
main.go
└── creates db connection
        ↓
routes.go
└── creates repository (injects db)
└── creates handler (injects repository)
        ↓
handler
└── calls repository via interface
└── does not know if it is real or mock
```

### Why interfaces matter

Repository interfaces allow handlers to be tested without a real database:

```go
// domain/asset.go — the contract
type AssetRepositoryInterface interface {
    RegisterAsset(asset *models.Asset) error
    ListAssets() ([]models.Asset, error)
    GetAssetByID(id string) (*models.Asset, error)
    UpdateAsset(id string, label *string, active *bool) error
    SoftDeleteAsset(id string) error
}

// domain/asset.go — real implementation
type AssetRepository struct {
    DB     *gorm.DB
    Logger *logger.Logger
}

// handlers/asset_test.go — mock for tests
type MockAssetRepository struct {
    ShouldFail bool
}
```

The handler depends on the interface — not on the concrete implementation. This is the Dependency Inversion Principle (D in SOLID) applied in practice.

---

## Asynchronous Processing — Transactional Outbox Pattern

SentinelDB implements the **Transactional Outbox Pattern** using PostgreSQL as the job queue — without any external message broker (RabbitMQ, Pub/Sub, SQS).

### Why PostgreSQL instead of a broker

The core problem this solves is atomicity. Without outbox:

```
INSERT run ✓
application crashes here
publish job to broker ✗ — never happens
→ run exists but no job was ever created
→ silent inconsistency
```

With outbox (same transaction):

```
BEGIN
INSERT run
INSERT outboxes jobs
COMMIT ✓ — both persist or neither does
→ impossible to have a run without jobs
```

### Trade-offs

PostgreSQL as a queue is the right choice when:

- Consistency between data and job creation is critical
- Volume is moderate (under ~10k jobs/minute)
- Operational simplicity matters (no extra infrastructure)

An external broker (RabbitMQ, Kafka) is better when:

- Volume is very high
- Sub-millisecond latency is required
- Multiple services consume the same queue
- Complex routing topologies are needed

---

## Worker — Concurrent Job Processing

The worker uses `SELECT FOR UPDATE SKIP LOCKED` to safely dequeue jobs across multiple concurrent goroutines without conflicts or duplicate processing.

```sql
SELECT * FROM outboxes
WHERE status = 'pending'
AND scheduled_at <= NOW()
ORDER BY scheduled_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;
```

`FOR UPDATE` locks the row for the current worker. `SKIP LOCKED` skips rows already locked by other workers instead of waiting — enabling true parallelism.

### Worker pool structure

```
Worker starts
└── launches N goroutines (configurable pool size)
    └── each goroutine runs a loop:
        ├── dequeue one job via SELECT FOR UPDATE SKIP LOCKED
        ├── if no job → sleep and retry
        └── if job found → process → update status
```

### Job chaining

Some jobs depend on others completing first. Correlation jobs (cross-source analysis) are scheduled with a future `scheduled_at` to ensure scan jobs finish first:

```
scan_shodan  (scheduled_at = now)
scan_hibp    (scheduled_at = now)
correlate    (scheduled_at = now + 5 minutes)
```

The worker only picks up jobs where `scheduled_at <= NOW()` — so correlation naturally runs after scans complete.

---

## Retry and Resilience

### Exponential backoff with jitter

Failed jobs are rescheduled with increasing delay to avoid overwhelming external APIs:

```
attempt 1 failed → reschedule in ~1s  (+ random jitter)
attempt 2 failed → reschedule in ~2s  (+ random jitter)
attempt 3 failed → reschedule in ~4s  (+ random jitter)
max attempts reached → status = failed
```

Jitter (random noise) prevents the thundering herd problem — multiple workers retrying at exactly the same moment after a failure.

### Circuit breaker

Each external source (Shodan, HIBP) has an independent circuit breaker. If a source fails repeatedly, the circuit opens and that source is temporarily skipped — other sources continue processing normally.

---

## Database Schema

### Tables and responsibilities

| Table             | Responsibility                                   |
| ----------------- | ------------------------------------------------ |
| `assets`          | Digital assets registered for monitoring         |
| `runs`            | One record per trigger execution                 |
| `outboxes`        | Job queue — one job per asset per source per run |
| `asset_snapshots` | Raw API responses stored after each scan         |
| `findings`        | Actionable results — only what changed or is new |

### Key relationships

```
assets (1) ←→ (N) outboxes (N) ←→ (1) runs
assets (1) ←→ (N) asset_snapshots
assets (1) ←→ (N) findings
runs   (1) ←→ (N) outboxes
runs   (1) ←→ (N) asset_snapshots
runs   (1) ←→ (N) findings
```

### Why outbox and asset_snapshots are separate tables

`outboxes` exists **before** processing — it tells the worker what to do. `asset_snapshots` exists **after** processing — it stores what was found. They represent opposite moments in the job lifecycle and cannot be merged.

### Why findings and snapshots are separate

`asset_snapshots` stores everything the API returned (raw data). `findings` stores only what changed or is new (actionable data). The worker compares the new snapshot against the previous one — only differences become findings.

```
Shodan returns 10 open ports
→ snapshot stores all 10
→ finding is only created for port 5432 that opened since last scan
   (the other 9 were already known)
```

---

## Testing Strategy

SentinelDB uses two categories of tests with different scopes and infrastructure requirements.

### Unit tests — run anywhere, no infrastructure

Tests that mock all external dependencies. These run on every `go test ./...` call.

Located alongside the file they test:

```
internal/job/handlers/asset.go
internal/job/handlers/asset_test.go   ← unit test, uses MockAssetRepository
internal/job/domain/asset.go
internal/job/domain/asset_test.go     ← unit test, pure function
```

### Integration tests — require real PostgreSQL

Tests that use a real database connection. Skipped by default, run explicitly with an environment variable:

```
tests/integration/
└── storage_test.go   ← requires INTEGRATION_TEST=1
```

Run with:

```bash
# unit tests only (default)
go test ./...

# including integration tests
INTEGRATION_TEST=1 go test ./...

# with race detector (always recommended)
go test -race ./...
```

### Table-driven tests

Go idiomatic pattern for testing multiple cases of the same function:

```go
func TestValidateAsset(t *testing.T) {
    tests := []struct {
        name      string
        assetType string
        value     string
        wantErr   bool
    }{
        {"valid ip",   "ip",    "192.168.1.1",   false},
        {"invalid ip", "ip",    "999.999.999.999", true},
        {"valid email","email", "user@test.com", false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateAsset(tt.assetType, tt.value)
            if tt.wantErr && err == nil {
                t.Errorf("expected error, got nil")
            }
            if !tt.wantErr && err != nil {
                t.Errorf("expected no error, got %v", err)
            }
        })
    }
}
```

---

## API Design

### Conventions

- All routes are versioned under `/api/v1`
- `POST /trigger` returns `202 Accepted` — not `200 OK` — because processing is asynchronous
- `DELETE /assets/:id` is a soft delete — sets `active = false`, preserves history
- `PATCH /findings/:id/resolve` uses `PATCH` because only one field is updated

### HTTP status codes

| Scenario                   | Status                    |
| -------------------------- | ------------------------- |
| Resource created           | 201 Created               |
| Async job enqueued         | 202 Accepted              |
| Successful read/update     | 200 OK                    |
| Invalid request body       | 400 Bad Request           |
| Resource not found         | 404 Not Found             |
| Database or internal error | 500 Internal Server Error |

---

## Graceful Shutdown

The API waits for in-flight requests to complete before exiting. The worker waits for in-progress jobs to finish before exiting. Neither cuts work in the middle on SIGTERM.

```
SIGTERM received
└── API stops accepting new requests
└── waits up to 10 seconds for in-flight requests
└── worker finishes current job
└── process exits cleanly
```

---

## Observability

### Metrics — Prometheus

Exposed at `GET /metrics`. Key metrics:

```
assets_monitored_total
runs_total by status (completed, failed)
jobs_total by type and status
findings_total by severity and source
external_api_duration_seconds by source
external_api_errors_total by source
```

### Tracing — OpenTelemetry

Each request generates a trace propagated through the worker via job payload. This allows seeing the full lifecycle of a trigger in Grafana:

```
POST /trigger (12ms)
└── insert outboxes (3ms)
    └── worker dequeue (2ms)
        └── scan_shodan (230ms)
        └── scan_hibp (180ms)
        └── correlate (45ms)
        └── notify_telegram (95ms)
```

---

## Environment Variables

| Variable             | Required | Description                              |
| -------------------- | -------- | ---------------------------------------- |
| `SERVER_PORT`        | yes      | API port                                 |
| `DATABASE_URL`       | yes      | PostgreSQL connection string             |
| `JWT_SECRET_KEY`     | yes      | API authentication key                   |
| `SHODAN_API_KEY`     | yes      | Shodan API key                           |
| `HIBP_API_KEY`       | yes      | HaveIBeenPwned API key                   |
| `TELEGRAM_BOT_TOKEN` | yes      | Telegram bot token                       |
| `TELEGRAM_CHAT_ID`   | yes      | Telegram chat ID                         |
| `WORKER_POOL_SIZE`   | no       | Number of worker goroutines (default: 5) |
| `INTEGRATION_TEST`   | no       | Set to 1 to run integration tests        |

**Note on special characters in passwords:** If your database password contains special characters (e.g. `#`), URL-encode them in the connection string. `#` becomes `%23`.
