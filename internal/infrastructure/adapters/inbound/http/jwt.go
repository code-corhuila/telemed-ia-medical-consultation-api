package http

import (
	"context"
	"errors"
	nethttp "net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type claimsKey struct{}

// Claims is the set of values we trust from a validated JWT.
type Claims struct {
	Subject string
	Role    string
}

// FromContext returns the validated claims, or nil if there are none.
func FromContext(ctx context.Context) *Claims {
	if c, ok := ctx.Value(claimsKey{}).(*Claims); ok {
		return c
	}
	return nil
}

// WithClaims attaches claims to the context (used by tests).
func WithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, c)
}

// RequireJWT validates the RS256 bearer token against the public key and
// stores the resulting claims in the request context.
//
// Norm 5.3.7: every service validates the token itself. There is no gateway
// trust assumption here.
func RequireJWT(publicKeyPEM string) func(nethttp.Handler) nethttp.Handler {
	return func(next nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			traceID := traceFrom(r.Context())
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				WriteError(w, nethttp.StatusUnauthorized, "UNAUTHORIZED",
					"missing bearer token", traceID)
				return
			}
			raw := strings.TrimPrefix(auth, "Bearer ")

			key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
			if err != nil {
				WriteError(w, nethttp.StatusInternalServerError, "INTERNAL_ERROR",
					"invalid public key", traceID)
				return
			}

			token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return key, nil
			})
			if err != nil || !token.Valid {
				WriteError(w, nethttp.StatusUnauthorized, "UNAUTHORIZED",
					"invalid token", traceID)
				return
			}

			mapClaims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				WriteError(w, nethttp.StatusUnauthorized, "UNAUTHORIZED",
					"invalid claims", traceID)
				return
			}

			sub, _ := mapClaims["sub"].(string)
			role, _ := mapClaims["role"].(string)
			claims := &Claims{Subject: sub, Role: role}

			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
		})
	}
}
