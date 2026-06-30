package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"cloud-log-access/internal/auth"
	"cloud-log-access/internal/storage"
)

// fakeProv is a storage.Provider double for handler tests.
type fakeProv struct {
	url       string
	expiresAt time.Time
	lastTTL   time.Duration
}

func (f *fakeProv) Name() string { return "fake" }
func (f *fakeProv) List(context.Context, storage.ListParams) (*storage.ListResult, error) {
	return &storage.ListResult{}, nil
}
func (f *fakeProv) Download(context.Context, string, string) (*storage.Object, error) {
	return &storage.Object{Body: io.NopCloser(bytes.NewReader(nil))}, nil
}
func (f *fakeProv) Presign(_ context.Context, _, _ string, ttl time.Duration) (string, time.Time, error) {
	f.lastTTL = ttl
	return f.url, f.expiresAt, nil
}
func (f *fakeProv) HealthCheck(context.Context) error { return nil }

func newStorageHandler(p storage.Provider) *Storage {
	svc := storage.NewService(time.Minute, 15*time.Minute)
	svc.Register("aws", "AWS S3", "us-east-1", []string{"production-logs"}, p)
	return NewStorage(svc, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func callPresign(h *Storage, claims *auth.Claims, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/providers/aws/presign?bucket=production-logs", bytes.NewBufferString(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider", "aws")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	if claims != nil {
		ctx = auth.WithClaims(ctx, claims)
	}
	rec := httptest.NewRecorder()
	h.Presign(rec, req.WithContext(ctx))
	return rec
}

func TestPresignHandler(t *testing.T) {
	admin := &auth.Claims{Email: "a@b.com", Role: auth.RoleAdmin}
	viewer := &auth.Claims{Email: "v@b.com", Role: auth.RoleViewer}

	t.Run("viewer is forbidden (defense-in-depth)", func(t *testing.T) {
		h := newStorageHandler(&fakeProv{url: "https://x/y", expiresAt: time.Now().Add(time.Minute)})
		if code := callPresign(h, viewer, `{"key":"payment.log","ttl_seconds":300}`).Code; code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", code)
		}
	})

	t.Run("missing claims is forbidden", func(t *testing.T) {
		h := newStorageHandler(&fakeProv{})
		if code := callPresign(h, nil, `{"key":"x","ttl_seconds":60}`).Code; code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", code)
		}
	})

	t.Run("invalid body is 400", func(t *testing.T) {
		h := newStorageHandler(&fakeProv{})
		if code := callPresign(h, admin, `not json`).Code; code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", code)
		}
	})

	t.Run("admin gets url + expiry, ttl clamped", func(t *testing.T) {
		fp := &fakeProv{url: "https://x/y?sig=abc", expiresAt: time.Now().Add(15 * time.Minute)}
		h := newStorageHandler(fp)
		rec := callPresign(h, admin, `{"key":"payment.log","ttl_seconds":86400}`) // 24h
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if fp.lastTTL != 15*time.Minute {
			t.Errorf("clamped ttl = %v, want 15m", fp.lastTTL)
		}
		var body struct {
			Data struct {
				URL       string `json:"url"`
				ExpiresAt string `json:"expires_at"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Data.URL == "" || body.Data.ExpiresAt == "" {
			t.Errorf("missing url/expires_at: %s", rec.Body.String())
		}
	})

	t.Run("non-expiring provider omits expires_at (honest reporting)", func(t *testing.T) {
		// zero expiry mimics the fake-gcs media URL.
		h := newStorageHandler(&fakeProv{url: "http://localhost:4443/x", expiresAt: time.Time{}})
		rec := callPresign(h, admin, `{"key":"x","ttl_seconds":300}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if bytes.Contains(rec.Body.Bytes(), []byte("expires_at")) {
			t.Errorf("non-expiring response must omit expires_at, got: %s", rec.Body.String())
		}
	})
}
