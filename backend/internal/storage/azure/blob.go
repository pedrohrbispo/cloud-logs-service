// Package azure is the Azure Blob Storage implementation of storage.Provider,
// targeting Azurite locally.
package azure

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"

	"cloud-log-access/internal/storage"
)

// Config configures the Azure adapter.
type Config struct {
	AccountName      string
	AccountKey       string
	InternalEndpoint string // http://azurite:10000/devstoreaccount1 (server-side)
	PublicEndpoint   string // http://localhost:10000/devstoreaccount1 (SAS host)
}

// Adapter implements storage.Provider for Azure Blob.
type Adapter struct {
	client    *azblob.Client
	cred      *azblob.SharedKeyCredential
	publicURL string
}

// New builds the adapter from a shared-key credential.
func New(cfg Config) (*Adapter, error) {
	cred, err := azblob.NewSharedKeyCredential(cfg.AccountName, cfg.AccountKey)
	if err != nil {
		return nil, fmt.Errorf("azure cred: %w", err)
	}
	client, err := azblob.NewClientWithSharedKeyCredential(cfg.InternalEndpoint, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("azure client: %w", err)
	}
	return &Adapter{
		client:    client,
		cred:      cred,
		publicURL: strings.TrimRight(cfg.PublicEndpoint, "/"),
	}, nil
}

// Name returns the provider id.
func (a *Adapter) Name() string { return "azure" }

// List returns one page of blobs under the prefix, honoring the cursor (marker)
// and returning a NextCursor for the following page.
func (a *Adapter) List(ctx context.Context, p storage.ListParams) (*storage.ListResult, error) {
	opts := &azblob.ListBlobsFlatOptions{}
	if p.Prefix != "" {
		opts.Prefix = &p.Prefix
	}
	if p.Cursor != "" {
		opts.Marker = &p.Cursor
	}
	if p.Limit > 0 {
		lim := p.Limit
		opts.MaxResults = &lim
	}

	pager := a.client.NewListBlobsFlatPager(p.Bucket, opts)
	res := &storage.ListResult{}
	if !pager.More() {
		return res, nil
	}
	page, err := pager.NextPage(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	for _, b := range page.Segment.BlobItems {
		info := storage.ObjectInfo{Key: deref(b.Name)}
		if b.Properties != nil {
			info.Size = deref(b.Properties.ContentLength)
			info.LastModified = derefTime(b.Properties.LastModified)
			info.ContentType = deref(b.Properties.ContentType)
			if b.Properties.ETag != nil {
				info.ETag = strings.Trim(string(*b.Properties.ETag), `"`)
			}
		}
		res.Objects = append(res.Objects, info)
	}
	if page.NextMarker != nil && *page.NextMarker != "" {
		res.NextCursor = *page.NextMarker
	}
	return res, nil
}

// Download streams a blob's body.
func (a *Adapter) Download(ctx context.Context, bucket, key string) (*storage.Object, error) {
	resp, err := a.client.DownloadStream(ctx, bucket, key, nil)
	if err != nil {
		if bloberror.HasCode(err, bloberror.BlobNotFound, bloberror.ContainerNotFound) {
			return nil, storage.ErrNotFound
		}
		return nil, mapErr(err)
	}
	return &storage.Object{
		Body:         resp.Body,
		Size:         deref(resp.ContentLength),
		ContentType:  deref(resp.ContentType),
		LastModified: derefTime(resp.LastModified),
	}, nil
}

// Presign builds a read-only service SAS URL signed against the public host.
// The SAS allows http (Azurite serves http) and backdates StartTime to absorb
// container/host clock skew. For Azure service SAS the host is not part of the
// signature, so pointing the URL at localhost is safe.
func (a *Adapter) Presign(_ context.Context, container, blob string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now().UTC()
	expiry := now.Add(ttl)
	values := sas.BlobSignatureValues{
		// Allow HTTP only when the public endpoint is itself HTTP (Azurite).
		// A real https:// endpoint gets HTTPS-only so the SAS can't be replayed
		// in cleartext.
		Protocol:      a.sasProtocol(),
		StartTime:     now.Add(-5 * time.Minute), // clock-skew tolerance
		ExpiryTime:    expiry,
		Permissions:   (&sas.BlobPermissions{Read: true}).String(),
		ContainerName: container,
		BlobName:      blob,
	}
	qp, err := values.SignWithSharedKey(a.cred)
	if err != nil {
		return "", time.Time{}, mapErr(err)
	}
	return fmt.Sprintf("%s/%s/%s?%s", a.publicURL, container, escapeBlobPath(blob), qp.Encode()), expiry, nil
}

// sasProtocol selects the allowed SAS protocol from the public endpoint scheme.
func (a *Adapter) sasProtocol() sas.Protocol {
	if strings.HasPrefix(a.publicURL, "https://") {
		return sas.ProtocolHTTPS
	}
	return sas.ProtocolHTTPSandHTTP
}

// escapeBlobPath percent-escapes each path segment but keeps '/' literal, the
// idiomatic Azure blob URL form (avoids %2F interop issues with real Azure).
func escapeBlobPath(blob string) string {
	parts := strings.Split(blob, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}

// HealthCheck probes connectivity by listing containers.
func (a *Adapter) HealthCheck(ctx context.Context) error {
	pager := a.client.NewListContainersPager(nil)
	if pager.More() {
		if _, err := pager.NextPage(ctx); err != nil {
			return mapErr(err)
		}
	}
	return nil
}

func mapErr(err error) error {
	if bloberror.HasCode(err, bloberror.BlobNotFound, bloberror.ContainerNotFound) {
		return storage.ErrNotFound
	}
	return fmt.Errorf("%w: %v", storage.ErrUnavailable, err)
}

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func derefTime(p *time.Time) time.Time {
	if p == nil {
		return time.Time{}
	}
	return *p
}
