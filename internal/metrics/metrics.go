package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HttpDuration mide la latencia (Golden Signal: Latency).
	// Buckets elegidos para el tiempo de respuesta esperado en consultas
	// médicas. El label outcome permite separar success vs error en el
	// dashboard sin duplicar la serie temporal por cada status code.
	HttpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "medical_consultation_http_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5},
	}, []string{"path", "method", "status", "outcome"})

	// HttpRequestsTotal cuenta el tráfico y errores (Golden Signal: Traffic & Errors).
	// El label outcome permite calcular el error rate como
	//   sum(rate(medical_consultation_http_requests_total{outcome="error"}[5m]))
	//   / sum(rate(medical_consultation_http_requests_total[5m]))
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "medical_consultation_http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"path", "method", "status", "outcome"})
)
