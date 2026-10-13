package http

import (
	nethttp "net/http"
)

// Router wires the HTTP routes and the middleware chain.
type Router struct {
	mux         *nethttp.ServeMux
	handlers    *Handlers
	jwtKey      string
	idempotency *idempotencyStore
}

// NewRouter builds the router.
func NewRouter(handlers *Handlers, jwtPublicKeyPEM string) *Router {
	return &Router{
		mux:         nethttp.NewServeMux(),
		handlers:    handlers,
		jwtKey:      jwtPublicKeyPEM,
		idempotency: newIdempotencyStore(),
	}
}

// Handler returns the http.Handler ready to be mounted on the server.
//
// /health is intentionally outside the JWT and idempotency middlewares.
// /api/* requires a valid RS256 token and (for mutating verbs) an
// Idempotency-Key header.
func (r *Router) Handler() nethttp.Handler {
	r.mux.HandleFunc("GET /health", r.handlers.Health)

	protected := nethttp.NewServeMux()
	protected.HandleFunc("POST /api/consultations/{id}/attention", r.handlers.RecordAttention)
	protected.HandleFunc("GET /api/consultations/{id}/attention", r.handlers.GetAttention)
	protected.HandleFunc("GET /api/post-summaries/{id}", r.handlers.GetPostSummary)
	protected.HandleFunc("POST /api/consultations/{id}/post-summary", r.handlers.GeneratePostSummary)

	wrapped := Chain(protected,
		RequireJWT(r.jwtKey),
		r.idempotency.RequireIdempotency,
	)

	r.mux.Handle("/api/", wrapped)

	return Chain(r.mux, Correlation)
}
