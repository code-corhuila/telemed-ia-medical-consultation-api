package http

import (
	"net/http"
)

// Middleware is a handler decorator.
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares in reverse order so that the first one in the
// list ends up being the outermost wrapper.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
