# Observability Implementation

This document explains how the `medical-consultation-api` implements the concepts from the Week 11 Workshop (CP1 to CP4).

## CP1 & CP2: Tracing and Structured Logs
- **Trace Context:** The API already uses the `Correlation` middleware to capture or generate an `X-Correlation-Id`. All logs include this ID to trace the request across services.
- **Logs:** Handlers use structured logging to capture the `trace`, `method`, and `path`.

## CP3: Instrumentation Code
We added a dedicated `metrics` package and an HTTP middleware.
- **Metrics Middleware:** Intercepts every request, measures the duration using `time.Since`, and captures the HTTP status code.
- **Endpoint:** Exposes a standard `/metrics` endpoint (outside the JWT protection) for Prometheus to scrape.

## CP4: Golden Signals and Metric Catalog
| Signal | Metric Name | Type | Labels | PromQL (Dashboard) |
|---|---|---|---|---|
| **Latency** | `medical_consultation_http_duration_seconds` | Histogram | path, method, status | `histogram_quantile(0.95, sum by (le) (rate(medical_consultation_http_duration_seconds_bucket[5m])))` |
| **Traffic** | `medical_consultation_http_requests_total` | Counter | path, method, status | `sum(rate(medical_consultation_http_requests_total[5m])) by (path)` |
| **Errors** | `medical_consultation_http_requests_total` | Counter | path, method, status | `sum(rate(medical_consultation_http_requests_total{status=~"5.."}[5m]))` |

**Buckets:** The latency histogram uses buckets `[0.1, 0.25, 0.5, 1, 2.5]` seconds. A medical consultation query/update should not take more than 500ms (0.5s). 
