package http

import (
	nethttp "net/http"
	"sync"
)

// IdempotencyKey is the header used by the client (norm 5.3.8).
const IdempotencyKey = "Idempotency-Key"

// idempotencyStore is a minimal in-process cache of already-seen keys.
//
// NOTE: this implementation is intentionally simple and in-memory. It will
// be replaced by a persistent store (Redis or a table) once the broker and
// the cache layer are agreed by the team. Its purpose today is to enforce
// the contract: a POST/PUT/PATCH without an Idempotency-Key must be
// rejected.
type idempotencyStore struct {
	mu   sync.Mutex
	seen map[string]bool
}

func newIdempotencyStore() *idempotencyStore {
	return &idempotencyStore{seen: map[string]bool{}}
}

func (s *idempotencyStore) has(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seen[key]
}

func (s *idempotencyStore) remember(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen[key] = true
}

// RequireIdempotency rejects mutating requests that do not carry an
// Idempotency-Key header. It does NOT replay stored responses yet (that is
// a follow-up when a persistent store is added).
func (s *idempotencyStore) RequireIdempotency(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if r.Method != nethttp.MethodPost &&
			r.Method != nethttp.MethodPut &&
			r.Method != nethttp.MethodPatch {
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get(IdempotencyKey)
		trace := traceFrom(r.Context())
		if key == "" {
			WriteError(w, nethttp.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY",
				"Idempotency-Key header is required", trace)
			return
		}
		s.remember(key)
		next.ServeHTTP(w, r)
	})
}
