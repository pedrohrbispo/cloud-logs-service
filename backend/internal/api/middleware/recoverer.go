package middleware

import (
	"log/slog"
	"net/http"

	"cloud-log-access/internal/api/response"
	"cloud-log-access/internal/reqctx"
)

// Recoverer converts a panic into a 500 error envelope and logs it with the
// request ID. The raw panic value and stack are never sent to the client.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.LogAttrs(r.Context(), slog.LevelError, "panic_recovered",
						slog.Any("panic", rec),
						slog.String("path", r.URL.Path),
						slog.String("request_id", reqctx.RequestID(r.Context())),
					)
					response.Error(w, r, http.StatusInternalServerError,
						response.CodeInternalError, "an internal error occurred")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
