// Package reqctx carries per-request values (request ID, authenticated claims)
// through the context. It depends on nothing else internal, so both the
// middleware and response packages can use it without an import cycle.
package reqctx

import "context"

type ctxKey int

const (
	requestIDKey ctxKey = iota
)

// HeaderRequestID is the canonical request-correlation header.
const HeaderRequestID = "X-Request-ID"

// WithRequestID returns a copy of ctx carrying the request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID returns the request ID stored in ctx, or "" if none.
func RequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
