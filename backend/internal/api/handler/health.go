// Package handler holds the BFF's HTTP handlers.
package handler

import (
	"context"
	"net/http"

	"cloud-log-access/internal/api/response"
)

// ReadinessProbe reports whether the service can serve storage requests.
type ReadinessProbe interface {
	Ready(ctx context.Context) bool
}

// Health serves liveness and readiness probes.
type Health struct {
	ready ReadinessProbe
}

// NewHealth constructs the health handler. A nil probe makes Readyz always ready.
func NewHealth(ready ReadinessProbe) *Health { return &Health{ready: ready} }

// Healthz is the liveness probe: 200 as long as the process is serving.
func (h *Health) Healthz(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
}

// Readyz is the readiness probe: 503 when no storage provider is reachable, so
// the orchestrator stops routing traffic to a BFF that would 503 every storage
// call. (Per-provider health is exposed granularly via GET /providers.)
func (h *Health) Readyz(w http.ResponseWriter, r *http.Request) {
	if h.ready != nil && !h.ready.Ready(r.Context()) {
		response.Error(w, r, http.StatusServiceUnavailable, response.CodeStorageUnavailable,
			"no storage provider reachable")
		return
	}
	response.JSON(w, r, http.StatusOK, map[string]string{"status": "ready"})
}
