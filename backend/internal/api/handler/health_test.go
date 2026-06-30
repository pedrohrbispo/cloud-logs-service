package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubProbe struct{ ready bool }

func (s stubProbe) Ready(context.Context) bool { return s.ready }

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHealth(nil).Healthz(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
		Meta struct {
			RequestID string `json:"request_id"`
			Timestamp string `json:"timestamp"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if body.Data.Status != "ok" {
		t.Errorf("data.status = %q, want ok", body.Data.Status)
	}
}

func TestReadyz(t *testing.T) {
	t.Run("ready when a provider is reachable", func(t *testing.T) {
		rec := httptest.NewRecorder()
		NewHealth(stubProbe{ready: true}).Readyz(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("503 when no provider is reachable", func(t *testing.T) {
		rec := httptest.NewRecorder()
		NewHealth(stubProbe{ready: false}).Readyz(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", rec.Code)
		}
	})

	t.Run("nil probe is always ready", func(t *testing.T) {
		rec := httptest.NewRecorder()
		NewHealth(nil).Readyz(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})
}
