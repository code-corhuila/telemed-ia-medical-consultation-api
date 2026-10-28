# Observability Implementation

This document explains how the `medical-consultation-api` implements the concepts from the Week 11 Workshop (CP1 to CP4).

## CP1 & CP2: Tracing and Structured Logs

- **Correlation ID:** The `Correlation` middleware reads or generates `X-Correlation-Id`, stores it in the MDC-equivalent (`context.Context`), and echoes it in the response. Every log line includes this ID (via `log.Printf("[%s] ...", trace, ...)`).
- **W3C Trace Context:** The same middleware parses the incoming `traceparent` header (W3C Level 1) and:
  - validates `00-<32hex>-<16hex>-<2hex>`;
  - stores `traceId` and `spanId` in the context (`correlation.TraceIDFromContext` / `correlation.SpanIDFromContext`);
  - echoes `traceparent` in the response so the caller can propagate it to the next hop.
- **Logs:** Handlers use structured logging with `log.Printf` and `slog` from `main.go`, including the correlation ID.

## CP3: Instrumentation Code

- **Metrics package:** `internal/metrics/metrics.go` declares `HttpDuration` (Histogram) and `HttpRequestsTotal` (Counter), both labeled by `path`, `method`, `status`, `outcome`.
- **Metrics Middleware:** `MetricsMiddleware` wraps the router, records latency (`time.Since`), and captures the HTTP status.
  - **Path masking:** UUIDs in the path are replaced by `{id}` (see `sanitizePath`) so Prometheus does not create a new series per consultation id (cardinality explosion).
  - **Outcome:** `outcome="success"` if status < 400, `outcome="error"` if status >= 400.
- **Endpoint:** `/metrics` is exposed (outside the JWT protection) for Prometheus to scrape.

## CP4: Golden Signals and Metric Catalog

| Signal | Metric Name | Type | Labels | PromQL (Dashboard) |
|---|---|---|---|---|
| **Latency** | `medical_consultation_http_duration_seconds` | Histogram | path, method, status, outcome | `histogram_quantile(0.95, sum by (le) (rate(medical_consultation_http_duration_seconds_bucket[5m])))` |
| **Traffic** | `medical_consultation_http_requests_total` | Counter | path, method, status, outcome | `sum(rate(medical_consultation_http_requests_total[5m])) by (path)` |
| **Errors** | `medical_consultation_http_requests_total{outcome="error"}` | Counter | path, method, status, outcome | `sum(rate(medical_consultation_http_requests_total{outcome="error"}[5m])) / sum(rate(medical_consultation_http_requests_total[5m]))` |

**Buckets:** The latency histogram uses `[0.1, 0.25, 0.5, 1, 2.5]` seconds. A medical consultation query/update should not take more than 500ms (0.5s).

**Cardinality:** UUIDs are masked to `{id}`, so the number of series per path stays constant (bounded by the number of routes), regardless of traffic.

## Outbox poller

The `OutboxPoller` runs in a background goroutine and publishes `outbox_event` rows to RabbitMQ with `mandatory=true`. If RabbitMQ is unavailable:

- the poller reconnects with exponential backoff (1s → 60s);
- the API keeps serving traffic;
- failed publishes are retried up to 5 times (`attempts` column), then marked `FAILED` with `last_error`.

Each published message carries:

- `MessageId` = `event_id` (UUID v5, deterministic per summary + event type);
- `CorrelationId` = the request-scoped `X-Correlation-Id`;
- `Type` = `event_type`;
- `Timestamp` = UTC.
