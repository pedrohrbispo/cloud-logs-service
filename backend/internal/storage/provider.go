// Package storage defines the provider-agnostic object-storage abstraction and
// the Service facade that all handlers call. The facade centralizes
// authorization-adjacent validation (provider/bucket allow-list, key
// sanitization, presign TTL clamping) so each cloud adapter stays thin.
package storage

import (
	"context"
	"io"
	"time"
)

// ObjectInfo is a listed object's metadata.
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
	ContentType  string
}

// Object is a downloadable object. The caller MUST Close Body.
type Object struct {
	Body         io.ReadCloser
	Size         int64
	ContentType  string
	ETag         string
	LastModified time.Time
}

// ListParams are the inputs to a List call (already validated by the Service).
type ListParams struct {
	Bucket string
	Prefix string
	Cursor string
	Limit  int32
}

// ListResult is a page of objects plus an opaque cursor for the next page.
type ListResult struct {
	Objects    []ObjectInfo
	NextCursor string
}

// Provider is one cloud backend (S3, GCS, Azure Blob). Implementations are thin
// adapters over the official SDK; validation lives in the Service.
type Provider interface {
	Name() string
	List(ctx context.Context, p ListParams) (*ListResult, error)
	Download(ctx context.Context, bucket, key string) (*Object, error)
	// Presign returns a temporary URL and the time it actually expires. A zero
	// expiresAt means the link does not expire (e.g. the fake-gcs emulator,
	// which cannot sign URLs) — callers must report that honestly, not fabricate
	// an expiry.
	Presign(ctx context.Context, bucket, key string, ttl time.Duration) (url string, expiresAt time.Time, err error)
	HealthCheck(ctx context.Context) error
}

// ProviderInfo is the public description of a registered provider.
type ProviderInfo struct {
	ID      string
	Name    string
	Buckets []string
	Region  string
	Healthy bool
	Err     string
}
