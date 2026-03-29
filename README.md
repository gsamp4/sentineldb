# SentinelDB

A personal OSINT and cybersecurity monitoring system built with Go, designed to continuously watch digital assets and alert when exposures or anomalies are detected.

## Motivation

Security monitoring tools are either too expensive, too complex to self-host, or built for enterprise teams. SentinelDB is a lightweight, self-hosted alternative for developers and security-conscious individuals who want visibility over their own digital footprint — without paying for Shodan Monitor, SecurityTrails, or similar SaaS products.

## What It Does

SentinelDB monitors digital assets (IPs, domains, and emails) across multiple threat intelligence sources and sends alerts when something changes or a new exposure is found.

**Shodan integration**
- Detects open ports and exposed services
- Identifies software versions with known CVEs
- Tracks SSL certificate expiration
- Alerts when new ports open or services change since the last scan

**HaveIBeenPwned integration**
- Monitors emails and domains against known data breaches
- Alerts when a new breach is detected containing monitored assets

**Cross-source correlation**
- Combines findings from multiple sources within the same run
- Elevates severity when an asset is exposed on multiple fronts simultaneously (e.g., open database port + email found in breach)

**Telegram notifications**
- Real-time alerts with severity classification (critical, high, medium, low)
- Daily digest summarizing all active findings
- Notifications when findings are resolved

## Architecture

SentinelDB is built around an event-driven, asynchronous processing model using PostgreSQL as the job queue — implementing the Transactional Outbox Pattern to guarantee consistency between run creation and job execution.

```
POST /trigger
      ↓
API inserts run + outbox jobs atomically (same transaction)
      ↓
Worker pool consumes jobs via SELECT FOR UPDATE SKIP LOCKED
      ↓
Each job calls its source (Shodan or HIBP)
      ↓
Results are compared against previous snapshots
      ↓
New findings are persisted and notifications sent via Telegram
```

**Key architectural decisions**

- PostgreSQL as job queue instead of an external broker (RabbitMQ, Pub/Sub) — eliminates infrastructure dependency and guarantees transactional consistency between run creation and job scheduling
- SELECT FOR UPDATE SKIP LOCKED — enables concurrent workers to dequeue jobs without conflicts or duplicate processing
- Snapshot-based change detection — each scan result is stored and compared against the previous one, so alerts are only triggered when something actually changes
- Job chaining — correlation jobs are scheduled after scan jobs complete, enabling multi-source analysis within a single run
- Graceful shutdown — in-flight jobs complete before the process exits

## Tech Stack

- **Language:** Go
- **Web Framework:** Echo
- **Database:** PostgreSQL with GORM
- **Job Queue:** PostgreSQL (Transactional Outbox Pattern)
- **Observability:** Prometheus + Grafana + OpenTelemetry
- **Notifications:** Telegram Bot API
- **External APIs:** Shodan, HaveIBeenPwned
- **Infrastructure:** Docker, Docker Compose

## Project Structure

```
sentineldb/
├── cmd/
│   ├── api/              # API entrypoint
│   └── worker/           # Worker entrypoint
├── internal/
│   └── job/
│       ├── domain/       # Business logic, validation, repository interfaces
│       ├── handlers/     # HTTP handlers and DTOs
│       ├── models/       # GORM models
│       └── routes/       # Route registration
├── pkg/
│   ├── logger/           # Structured logger
│   └── retry/            # Exponential backoff and circuit breaker
├── docker-compose.yml
└── README.md
```

## Database Schema

| Table | Description |
|---|---|
| `assets` | Digital assets registered for monitoring (IPs, domains, emails) |
| `runs` | Execution history — each trigger creates one run |
| `outbox` | Job queue — one job per asset per source within a run |
| `findings` | Actionable results — only what changed or is newly detected |
| `asset_snapshots` | Raw API responses per scan — used for change detection |

## API Routes

| Method | Route | Description |
|---|---|---|
| POST | /api/v1/assets | Register a new asset |
| GET | /api/v1/assets | List all assets |
| GET | /api/v1/assets/:id | Get asset details |
| PUT | /api/v1/assets/:id | Update asset (label, active) |
| DELETE | /api/v1/assets/:id | Soft-delete asset |
| POST | /api/v1/trigger | Start a full scan of all active assets |
| POST | /api/v1/trigger/:id | Start a scan for a specific asset |
| GET | /api/v1/runs | Execution history |
| GET | /api/v1/runs/:id | Run details |
| GET | /api/v1/runs/:id/jobs | Job-level progress within a run |
| GET | /api/v1/findings | All open findings |
| GET | /api/v1/findings/:id | Finding details |
| PATCH | /api/v1/findings/:id/resolve | Mark finding as resolved |
| GET | /api/v1/metrics | Prometheus metrics endpoint |

## Getting Started

```bash
# Clone the repository
git clone https://github.com/yourusername/sentineldb
cd sentineldb

# Copy environment variables
cp .env.example .env

# Start dependencies
docker-compose up -d

# Run the API
go run cmd/api/main.go

# Run the worker (separate terminal)
go run cmd/worker/main.go
```

## Environment Variables

| Variable | Description |
|---|---|
| `SERVER_PORT` | API server port (e.g. 8080) |
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET_KEY` | Secret key for API authentication |
| `SHODAN_API_KEY` | Shodan API key |
| `HIBP_API_KEY` | HaveIBeenPwned API key |
| `TELEGRAM_BOT_TOKEN` | Telegram bot token |
| `TELEGRAM_CHAT_ID` | Telegram chat ID for notifications |

## Learning Goals

This project is being built as a deliberate practice exercise targeting senior backend engineering skills:

- Concurrency and parallelism in Go — worker pool with goroutines, channels, context cancellation
- Transactional Outbox Pattern — guaranteed consistency between API and async processing
- Database internals — SELECT FOR UPDATE SKIP LOCKED, partial indexes, JSONB queries
- Observability — Prometheus metrics, OpenTelemetry distributed tracing, structured logging
- Testing — unit tests with interface-based mocking, integration tests with testcontainers
- Resilience patterns — exponential backoff with jitter, circuit breaker per external source
- Graceful shutdown — in-flight job completion on SIGTERM

## Roadmap

- [ ] Asset CRUD
- [ ] Shodan integration
- [ ] HaveIBeenPwned integration
- [ ] Worker pool with outbox pattern
- [ ] Snapshot-based change detection
- [ ] Findings and severity classification
- [ ] Telegram notifications
- [ ] Cross-source correlation
- [ ] Prometheus metrics
- [ ] OpenTelemetry tracing
- [ ] Grafana dashboard