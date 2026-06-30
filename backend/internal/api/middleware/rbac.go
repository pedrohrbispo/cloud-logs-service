package middleware

import (
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"

	"cloud-log-access/internal/api/response"
	"cloud-log-access/internal/auth"
)

// RequireRole rejects requests whose authenticated role is not in roles. It
// fails closed: missing claims yield 401, not anonymous passthrough.
func RequireRole(roles ...auth.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.ClaimsFromContext(r.Context())
			if !ok {
				response.Error(w, r, http.StatusUnauthorized, response.CodeTokenMissing,
					"authentication required")
				return
			}
			if !slices.Contains(roles, claims.Role) {
				response.Error(w, r, http.StatusForbidden, response.CodeUnauthorized,
					"insufficient role")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireProviderAccess enforces the allowed_providers claim against the
// {provider} path parameter. A nil allowed_providers means "all providers"
// (admins). Must run after JWTAuth.
func RequireProviderAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.ClaimsFromContext(r.Context())
		if !ok {
			response.Error(w, r, http.StatusUnauthorized, response.CodeTokenMissing,
				"authentication required")
			return
		}
		provider := chi.URLParam(r, "provider")
		if provider != "" && claims.AllowedProviders != nil && !slices.Contains(claims.AllowedProviders, provider) {
			response.Error(w, r, http.StatusForbidden, response.CodeProviderNotAllowed,
				"your role does not permit access to this provider")
			return
		}
		next.ServeHTTP(w, r)
	})
}
