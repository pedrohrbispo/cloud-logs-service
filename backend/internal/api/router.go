// Package api wires the HTTP router, middleware chain, and handlers.
package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"cloud-log-access/internal/api/handler"
	"cloud-log-access/internal/api/middleware"
	"cloud-log-access/internal/auth"
	"cloud-log-access/internal/config"
)

const apiMaxBodyBytes = 64 << 10 // 64 KiB for JSON request bodies

// RouterDeps are the dependencies the router needs.
type RouterDeps struct {
	Logger    *slog.Logger
	Config    *config.Config
	Auth      *handler.Auth
	Storage   *handler.Storage
	Readiness handler.ReadinessProbe
	Validator middleware.TokenValidator
	Denylist  middleware.RevocationChecker
}

// NewRouter builds the HTTP handler with the full middleware chain.
//
// Outer chain (every route): RequestID → Logger → Recoverer → SecurityHeaders → CORS.
// Health probes are public and unthrottled (Docker hits them frequently).
// The /api/v1 surface adds a body cap and a global per-IP rate limit; login is
// additionally throttled; the protected group adds JWTAuth.
func NewRouter(d RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(d.Logger))
	r.Use(middleware.Recoverer(d.Logger))
	r.Use(middleware.SecurityHeaders)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   d.Config.CORSAllowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	health := handler.NewHealth(d.Readiness)
	r.Get("/healthz", health.Healthz)
	r.Get("/readyz", health.Readyz)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.MaxBodyBytes(apiMaxBodyBytes))
		r.Use(middleware.RateLimit(d.Config.RateLimitRPM, time.Minute))

		r.With(middleware.LoginRateLimit(d.Config.LoginRateLimitPM, time.Minute)).
			Post("/auth/login", d.Auth.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(d.Validator, d.Denylist))
			r.Post("/auth/logout", d.Auth.Logout)
			r.Get("/me", d.Auth.Me)

			r.Get("/providers", d.Storage.ListProviders)
			r.Route("/providers/{provider}", func(r chi.Router) {
				r.Use(middleware.RequireProviderAccess)
				r.Get("/logs", d.Storage.ListLogs)
				r.Get("/download", d.Storage.Download)
				r.With(middleware.RequireRole(auth.RoleAdmin)).Post("/presign", d.Storage.Presign)
			})
		})
	})

	return r
}
