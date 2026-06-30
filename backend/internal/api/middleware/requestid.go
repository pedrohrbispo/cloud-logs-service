// Package middleware holds the BFF's HTTP middleware chain.
package middleware

import (
	"net/http"

	"github.com/google/uuid"

	"cloud-log-access/internal/reqctx"
)

// RequestID assigns each request a UUIDv4 correlation ID (honoring an inbound
// X-Request-ID if present), stores it in the context, and echoes it back in
// the response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(reqctx.HeaderRequestID)
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set(reqctx.HeaderRequestID, id)
		next.ServeHTTP(w, r.WithContext(reqctx.WithRequestID(r.Context(), id)))
	})
}
