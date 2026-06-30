package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cloud-log-access/internal/auth"
)

const mwSecret = "test-secret-at-least-32-bytes-long-1234567890"

func TestJWTAuth(t *testing.T) {
	iss := auth.NewIssuer(mwSecret, "cloud-log-access-service", "cloud-log-access-service", time.Hour)
	deny := auth.NewDenylist()
	token, claims, err := iss.Issue(&auth.User{ID: "u1", Email: "a@b.com", Role: auth.RoleAdmin})
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, ok := auth.ClaimsFromContext(r.Context()); !ok || c.Email != "a@b.com" {
			t.Error("claims were not injected into the context")
		}
		w.WriteHeader(http.StatusOK)
	})
	handler := JWTAuth(iss, deny)(next)

	do := func(setAuth func(*http.Request)) int {
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		if setAuth != nil {
			setAuth(req)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	t.Run("valid token passes", func(t *testing.T) {
		if code := do(func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+token) }); code != http.StatusOK {
			t.Fatalf("status = %d, want 200", code)
		}
	})
	t.Run("missing header rejected", func(t *testing.T) {
		if code := do(nil); code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", code)
		}
	})
	t.Run("malformed scheme rejected", func(t *testing.T) {
		if code := do(func(r *http.Request) { r.Header.Set("Authorization", token) }); code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", code)
		}
	})
	t.Run("garbage token rejected", func(t *testing.T) {
		if code := do(func(r *http.Request) { r.Header.Set("Authorization", "Bearer not.a.jwt") }); code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", code)
		}
	})
	t.Run("revoked token rejected", func(t *testing.T) {
		deny.Revoke(claims.ID, claims.ExpiresAt.Time)
		if code := do(func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+token) }); code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 after revoke", code)
		}
	})
}
