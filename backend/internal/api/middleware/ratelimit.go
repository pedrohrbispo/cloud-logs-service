package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"

	"cloud-log-access/internal/api/response"
)

// keyByRealIP keys the limiter on the real client IP (see ClientIP) rather than
// r.RemoteAddr, which behind the nginx proxy would collapse every client into a
// single bucket.
func keyByRealIP(r *http.Request) (string, error) {
	return ClientIP(r), nil
}

func rateLimitExceeded(w http.ResponseWriter, r *http.Request) {
	response.Error(w, r, http.StatusTooManyRequests, response.CodeRateLimited, "too many requests")
}

// RateLimit is the global per-IP limiter for the API surface.
func RateLimit(requests int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.Limit(requests, window,
		httprate.WithKeyFuncs(keyByRealIP),
		httprate.WithLimitHandler(rateLimitExceeded),
	)
}

// LoginRateLimit is the stricter per-IP limiter applied only to login, to blunt
// credential stuffing.
func LoginRateLimit(requests int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.Limit(requests, window,
		httprate.WithKeyFuncs(keyByRealIP),
		httprate.WithLimitHandler(rateLimitExceeded),
	)
}
