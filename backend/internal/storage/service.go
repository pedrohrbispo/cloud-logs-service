package storage

import (
	"context"
	"sort"
	"sync"
	"time"

	"cloud-log-access/pkg/sanitize"
)

const (
	defaultListLimit = 100
	maxListLimit     = 1000
	healthTimeout    = 3 * time.Second
)

// Service is the facade every handler calls. It owns the provider registry, the
// per-provider bucket allow-list, key sanitization, and the presign TTL clamp.
type Service struct {
	providers map[string]Provider
	allow     map[string]map[string]struct{}
	display   map[string]string
	region    map[string]string
	minTTL    time.Duration
	maxTTL    time.Duration
}

// NewService creates an empty Service. Register providers before serving.
func NewService(minTTL, maxTTL time.Duration) *Service {
	return &Service{
		providers: map[string]Provider{},
		allow:     map[string]map[string]struct{}{},
		display:   map[string]string{},
		region:    map[string]string{},
		minTTL:    minTTL,
		maxTTL:    maxTTL,
	}
}

// Register adds a provider with its display name, region, and allowed buckets.
// Must be called during startup (before the server accepts requests).
func (s *Service) Register(id, name, region string, buckets []string, p Provider) {
	set := make(map[string]struct{}, len(buckets))
	for _, b := range buckets {
		set[b] = struct{}{}
	}
	s.providers[id] = p
	s.display[id] = name
	s.region[id] = region
	s.allow[id] = set
}

func (s *Service) resolve(provider, bucket string) (Provider, error) {
	p, ok := s.providers[provider]
	if !ok {
		return nil, ErrInvalidProvider
	}
	allowed, ok := s.allow[provider]
	if !ok || len(allowed) == 0 {
		return nil, ErrNoBuckets
	}
	if _, ok := allowed[bucket]; !ok {
		return nil, ErrInvalidBucket
	}
	return p, nil
}

// List validates inputs and lists objects in the bucket.
func (s *Service) List(ctx context.Context, provider, bucket, prefix, cursor string, limit int32) (*ListResult, error) {
	p, err := s.resolve(provider, bucket)
	if err != nil {
		return nil, err
	}
	cleanPrefix, err := sanitize.Prefix(prefix)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > maxListLimit {
		limit = defaultListLimit
	}
	return p.List(ctx, ListParams{Bucket: bucket, Prefix: cleanPrefix, Cursor: cursor, Limit: limit})
}

// Download validates inputs (including key sanitization) and streams the object.
func (s *Service) Download(ctx context.Context, provider, bucket, key string) (*Object, error) {
	p, err := s.resolve(provider, bucket)
	if err != nil {
		return nil, err
	}
	cleanKey, err := sanitize.Key(key)
	if err != nil {
		return nil, err
	}
	return p.Download(ctx, bucket, cleanKey)
}

// Presign validates inputs, clamps the TTL to [minTTL, maxTTL], and returns a
// temporary URL plus the time it actually expires (zero = non-expiring, which
// the caller must report honestly).
func (s *Service) Presign(ctx context.Context, provider, bucket, key string, ttl time.Duration) (string, time.Time, error) {
	p, err := s.resolve(provider, bucket)
	if err != nil {
		return "", time.Time{}, err
	}
	cleanKey, err := sanitize.Key(key)
	if err != nil {
		return "", time.Time{}, err
	}
	url, expiresAt, err := p.Presign(ctx, bucket, cleanKey, s.clampTTL(ttl))
	if err != nil {
		return "", time.Time{}, err
	}
	return url, expiresAt, nil
}

func (s *Service) clampTTL(ttl time.Duration) time.Duration {
	switch {
	case ttl < s.minTTL:
		return s.minTTL
	case ttl > s.maxTTL:
		return s.maxTTL
	default:
		return ttl
	}
}

func (s *Service) bucketList(id string) []string {
	out := make([]string, 0, len(s.allow[id]))
	for b := range s.allow[id] {
		out = append(out, b)
	}
	sort.Strings(out)
	return out
}

// Providers returns every registered provider with a concurrent health probe.
func (s *Service) Providers(ctx context.Context) []ProviderInfo {
	ids := make([]string, 0, len(s.providers))
	for id := range s.providers {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	infos := make([]ProviderInfo, len(ids))
	var wg sync.WaitGroup
	for i, id := range ids {
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			info := ProviderInfo{
				ID:      id,
				Name:    s.display[id],
				Region:  s.region[id],
				Buckets: s.bucketList(id),
			}
			hctx, cancel := context.WithTimeout(ctx, healthTimeout)
			defer cancel()
			if err := s.providers[id].HealthCheck(hctx); err != nil {
				info.Err = "unreachable"
			} else {
				info.Healthy = true
			}
			infos[i] = info
		}(i, id)
	}
	wg.Wait()
	return infos
}

// Ready reports whether the BFF can serve storage requests: true if there are
// no providers (nothing to be unready about) or at least one registered provider
// is reachable. A single degraded cloud does NOT make the whole BFF unready — the
// per-provider Healthy/Error in Providers() is the granular signal for the UI.
func (s *Service) Ready(ctx context.Context) bool {
	if len(s.providers) == 0 {
		return true
	}
	for _, info := range s.Providers(ctx) {
		if info.Healthy {
			return true
		}
	}
	return false
}
