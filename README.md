# telemed-ia-medical-consultation-api

> Medical-consultation bounded context — service API (Go).

Part of the **TeleMed IA** distributed system — team `telemed-ia`, Grupo 2, Universidad Corhuila, Sistemas Distribuidos 2026-B.

## Purpose

Records the clinical information of an appointment: the professionals attention summary and the post-consultation summary. Owns the entities `Consultation`, `AttentionSummary` and `PostSummary`. Does **not** generate PDFs (that belongs to `document-generation`, per ADR-010) and does **not** consume pre-consultation summaries owned by the intelligent agent.

## Stack

- Go 1.23
- Hexagonal architecture
- PostgreSQL (`pgx/v5`, schema `medical_consultation`)
- JWT RS256 validation (norm 5.3.7)
- Common response envelope, `X-Correlation-Id`, `Idempotency-Key` (norm 5.3.8)

## Layout

```text
cmd/consultation-api/
  main.go (composition root)
internal/
  domain/                 entities + typed errors (no framework imports)
  application/
    dto/                  request/response DTOs
    ports/in/             use case interfaces
    ports/out/            repository + publisher interfaces
    usecase/              use case implementations (fakes in fakes/)
  infrastructure/
    config/               env-driven settings
    adapters/
      inbound/http/       router, handlers, middleware (JWT, correlation, idempotency)
      outbound/persistence/ pgx repositories
      outbound/messaging/   noop event publisher (ADR-011 pending)
```

## Endpoints

| Method | Path | Auth | Idempotency |
|--------|------|------|-------------|
| GET | `/health` | no | — |
| POST | `/api/consultations/{id}/attention` | yes | required |
| GET | `/api/consultations/{id}/attention` | yes | — |
| GET | `/api/post-summaries/{id}` | yes | — |
| POST | `/api/consultations/{id}/post-summary` | yes | required |

## Configuration

Environment variables (see `.env.example`):
- `SERVER_PORT` — HTTP port (default `8080`).
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `DB_SCHEMA` — PostgreSQL connection.
- `JWT_PUBLIC_KEY` — RS256 public key in PEM format (required).
- `HTTP_*_TIMEOUT` — server timeouts.
- `DB_MAX_CONNS`, `DB_MIN_CONNS`, `DB_CONN_MAX_IDLE`, `DB_CONN_MAX_LIFETIME` — pool tuning.

## Local development

```bash
go mod download
go build ./...
go test ./...
go run ./cmd/consultation-api
```

Requires a running PostgreSQL with the `medical_consultation` schema and role created by `telemed-ia-medical-consultation-db`.

### Docker

```bash
cd deploy
docker compose up -d --build
```

The container joins the external platform network so the gateway can reach it by service name (`medical-consultation-api:8080`).

## Governance

Three permanent branches (`develop`, `qa`, `main`). No direct commits. Promotion between permanents uses `git cherry-pick -x` — no merge. Full policy lives in `library-docs`.

## Known technical debt

- `EventPublisherPort` currently resolved by a no-op adapter until ADR-011 (broker) is accepted.
- Idempotency store is in-memory; a persistent store (Redis or a table) is planned in a follow-up.
- Integration tests against a real PostgreSQL (Testcontainers) are not yet implemented.
