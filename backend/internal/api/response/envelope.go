// Package response writes the BFF's uniform success and error envelopes.
// Success: {"data": ..., "meta": {request_id, timestamp}}
// Error:   {"error": {code, message, request_id}}
package response

import (
	"encoding/json"
	"net/http"
	"time"

	"cloud-log-access/internal/reqctx"
)

// Meta is attached to every successful response.
type Meta struct {
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
}

type okEnvelope struct {
	Data any  `json:"data"`
	Meta Meta `json:"meta"`
}

// JSON writes a success envelope with the given status and payload.
func JSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	write(w, status, okEnvelope{
		Data: data,
		Meta: Meta{
			RequestID: reqctx.RequestID(r.Context()),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// write encodes v as JSON with HTML escaping disabled. The API is consumed as
// JSON (never embedded in HTML), and disabling escaping keeps URLs in payloads
// (e.g. presigned links with & query separators) clean and literal.
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}
