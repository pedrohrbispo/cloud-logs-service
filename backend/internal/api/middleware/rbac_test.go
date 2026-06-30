package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"cloud-log-access/internal/auth"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
}

func injectClaims(c *auth.Claims) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(auth.WithClaims(r.Context(), c)))
		})
	}
}

func TestRequireRole(t *testing.T) {
	run := func(c *auth.Claims) int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/presign", nil)
		injectClaims(c)(RequireRole(auth.RoleAdmin)(okHandler())).ServeHTTP(rec, req)
		return rec.Code
	}

	if code := run(&auth.Claims{Role: auth.RoleAdmin}); code != http.StatusOK {
		t.Errorf("admin: status = %d, want 200", code)
	}
	if code := run(&auth.Claims{Role: auth.RoleViewer}); code != http.StatusForbidden {
		t.Errorf("viewer: status = %d, want 403", code)
	}
}

func TestRequireProviderAccess(t *testing.T) {
	mount := func(c *auth.Claims) http.Handler {
		r := chi.NewRouter()
		r.Route("/providers/{provider}", func(r chi.Router) {
			r.Use(injectClaims(c))
			r.Use(RequireProviderAccess)
			r.Get("/logs", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
		})
		return r
	}

	viewer := &auth.Claims{Role: auth.RoleViewer, AllowedProviders: []string{"aws"}}
	admin := &auth.Claims{Role: auth.RoleAdmin, AllowedProviders: nil}

	cases := []struct {
		name   string
		claims *auth.Claims
		path   string
		want   int
	}{
		{"viewer aws allowed", viewer, "/providers/aws/logs", http.StatusOK},
		{"viewer gcp forbidden", viewer, "/providers/gcp/logs", http.StatusForbidden},
		{"viewer azure forbidden", viewer, "/providers/azure/logs", http.StatusForbidden},
		{"admin gcp allowed", admin, "/providers/gcp/logs", http.StatusOK},
		{"admin azure allowed", admin, "/providers/azure/logs", http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			mount(tc.claims).ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}
