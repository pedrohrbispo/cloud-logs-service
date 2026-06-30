package middleware

import (
	"errors"
	"net/http"
	"strings"

	"cloud-log-access/internal/api/response"
	"cloud-log-access/internal/auth"
)

// TokenValidator validates a bearer token and returns its claims.
type TokenValidator interface {
	Validate(token string) (*auth.Claims, error)
}

// RevocationChecker reports whether a token (by jti) has been revoked.
type RevocationChecker interface {
	IsRevoked(jti string) bool
}

const bearerPrefix = "Bearer "

// JWTAuth validates the Authorization bearer token, rejects revoked tokens, and
// injects the claims into the request context. It must run before RBAC.
func JWTAuth(validator TokenValidator, denylist RevocationChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hdr := r.Header.Get("Authorization")
			if len(hdr) <= len(bearerPrefix) || !strings.EqualFold(hdr[:len(bearerPrefix)], bearerPrefix) {
				response.Error(w, r, http.StatusUnauthorized, response.CodeTokenMissing,
					"missing or malformed authorization header")
				return
			}

			claims, err := validator.Validate(strings.TrimSpace(hdr[len(bearerPrefix):]))
			if err != nil {
				code, msg := response.CodeTokenInvalid, "invalid token"
				if errors.Is(err, auth.ErrTokenExpired) {
					code, msg = response.CodeTokenExpired, "token expired"
				}
				response.Error(w, r, http.StatusUnauthorized, code, msg)
				return
			}

			if denylist.IsRevoked(claims.ID) {
				response.Error(w, r, http.StatusUnauthorized, response.CodeTokenInvalid,
					"token has been revoked")
				return
			}

			next.ServeHTTP(w, r.WithContext(auth.WithClaims(r.Context(), claims)))
		})
	}
}
