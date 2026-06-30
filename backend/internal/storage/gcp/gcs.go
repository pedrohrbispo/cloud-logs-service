// Package gcp is the Google Cloud Storage implementation of storage.Provider,
// targeting fake-gcs-server locally.
package gcp

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	gcs "cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"cloud-log-access/internal/storage"
)

// Config configures the GCS adapter.
type Config struct {
	// InternalEndpoint is the JSON API root used server-side,
	// e.g. http://fake-gcs:4443/storage/v1/
	InternalEndpoint string
	// PublicEndpoint is the host a browser will hit for media URLs,
	// e.g. http://localhost:4443
	PublicEndpoint string
}

// Adapter implements storage.Provider for GCS.
type Adapter struct {
	client     *gcs.Client
	publicHost string
}

// New builds the adapter. WithoutAuthentication + an explicit endpoint targets
// the emulator and skips Google credential discovery.
func New(ctx context.Context, cfg Config) (*Adapter, error) {
	// Production guard: this build only supports the fake-gcs emulator, whose
	// media URLs are unsigned and non-expiring. Real GCS V4 signed URLs require a
	// service-account credential (PrivateKey/SignBytes), which WithoutAuthentication
	// cannot provide — fail fast rather than silently shipping permanent links.
	if cfg.InternalEndpoint == "" {
		return nil, errors.New("gcs: GCS_INTERNAL_ENDPOINT is required (only the fake-gcs emulator is supported; real GCS signing needs a service-account credential)")
	}
	client, err := gcs.NewClient(ctx,
		option.WithEndpoint(cfg.InternalEndpoint),
		option.WithoutAuthentication(),
	)
	if err != nil {
		return nil, fmt.Errorf("gcs client: %w", err)
	}
	return &Adapter{
		client:     client,
		publicHost: strings.TrimRight(cfg.PublicEndpoint, "/"),
	}, nil
}

// Name returns the provider id.
func (a *Adapter) Name() string { return "gcp" }

// List returns one page of objects under the prefix, honoring the cursor and
// returning a NextCursor for the following page.
func (a *Adapter) List(ctx context.Context, p storage.ListParams) (*storage.ListResult, error) {
	it := a.client.Bucket(p.Bucket).Objects(ctx, &gcs.Query{Prefix: p.Prefix})
	pageSize := int(p.Limit)
	if pageSize <= 0 {
		pageSize = 100
	}
	pager := iterator.NewPager(it, pageSize, p.Cursor)

	var attrsList []*gcs.ObjectAttrs
	nextCursor, err := pager.NextPage(&attrsList)
	if err != nil {
		return nil, mapErr(err)
	}
	res := &storage.ListResult{NextCursor: nextCursor}
	for _, attrs := range attrsList {
		res.Objects = append(res.Objects, storage.ObjectInfo{
			Key:          attrs.Name,
			Size:         attrs.Size,
			LastModified: attrs.Updated,
			ETag:         attrs.Etag,
			ContentType:  attrs.ContentType,
		})
	}
	return res, nil
}

// Download streams an object's body.
func (a *Adapter) Download(ctx context.Context, bucket, key string) (*storage.Object, error) {
	r, err := a.client.Bucket(bucket).Object(key).NewReader(ctx)
	if err != nil {
		if errors.Is(err, gcs.ErrObjectNotExist) {
			return nil, storage.ErrNotFound
		}
		return nil, mapErr(err)
	}
	return &storage.Object{
		Body:         r,
		Size:         r.Attrs.Size,
		ContentType:  r.Attrs.ContentType,
		LastModified: r.Attrs.LastModified,
	}, nil
}

// Presign returns the public media URL. fake-gcs-server has no signing key, so
// the URL is UNAUTHENTICATED and does NOT expire — we return a zero expiry so the
// API reports it honestly instead of fabricating an expires_at. (New() already
// fails fast if a non-emulator endpoint is configured.)
func (a *Adapter) Presign(_ context.Context, bucket, key string, _ time.Duration) (string, time.Time, error) {
	return fmt.Sprintf("%s/download/storage/v1/b/%s/o/%s?alt=media",
		a.publicHost, url.PathEscape(bucket), url.PathEscape(key)), time.Time{}, nil
}

// HealthCheck probes connectivity by listing buckets (projectID is ignored by
// the emulator).
func (a *Adapter) HealthCheck(ctx context.Context) error {
	it := a.client.Buckets(ctx, "cloud-log-access")
	_, err := it.Next()
	if err != nil && !errors.Is(err, iterator.Done) {
		return mapErr(err)
	}
	return nil
}

func mapErr(err error) error {
	if errors.Is(err, gcs.ErrObjectNotExist) || errors.Is(err, gcs.ErrBucketNotExist) {
		return storage.ErrNotFound
	}
	return fmt.Errorf("%w: %v", storage.ErrUnavailable, err)
}
