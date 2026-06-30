package storage

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"cloud-log-access/pkg/sanitize"
)

// fakeProvider is a hand-rolled test double (a 5-method interface does not
// warrant a mocking framework).
type fakeProvider struct {
	listCalls      []ListParams
	downloadKey    string
	healthErr      error
	downloadErr    error
	presignURL     string
	lastPresignTTL time.Duration
}

func (f *fakeProvider) Name() string { return "fake" }

func (f *fakeProvider) List(_ context.Context, p ListParams) (*ListResult, error) {
	f.listCalls = append(f.listCalls, p)
	return &ListResult{Objects: []ObjectInfo{{Key: "a.log", Size: 10}}}, nil
}

func (f *fakeProvider) Download(_ context.Context, _, key string) (*Object, error) {
	if f.downloadErr != nil {
		return nil, f.downloadErr
	}
	f.downloadKey = key
	return &Object{Body: io.NopCloser(strings.NewReader("data")), Size: 4}, nil
}

func (f *fakeProvider) Presign(_ context.Context, _, _ string, ttl time.Duration) (string, time.Time, error) {
	f.lastPresignTTL = ttl
	return f.presignURL, time.Now().Add(ttl), nil
}

func (f *fakeProvider) HealthCheck(_ context.Context) error { return f.healthErr }

func newService(p Provider) *Service {
	s := NewService(time.Minute, 15*time.Minute)
	s.Register("aws", "AWS S3", "us-east-1", []string{"production-logs"}, p)
	return s
}

func TestServiceResolveValidation(t *testing.T) {
	s := newService(&fakeProvider{})

	t.Run("unknown provider", func(t *testing.T) {
		if _, err := s.List(context.Background(), "gcp", "production-logs", "", "", 0); !errors.Is(err, ErrInvalidProvider) {
			t.Fatalf("err = %v, want ErrInvalidProvider", err)
		}
	})
	t.Run("bucket not in allow-list", func(t *testing.T) {
		if _, err := s.List(context.Background(), "aws", "secret-bucket", "", "", 0); !errors.Is(err, ErrInvalidBucket) {
			t.Fatalf("err = %v, want ErrInvalidBucket", err)
		}
	})
}

func TestServiceDownloadSanitizesKey(t *testing.T) {
	fp := &fakeProvider{}
	s := newService(fp)

	if _, err := s.Download(context.Background(), "aws", "production-logs", "../../etc/passwd"); !errors.Is(err, sanitize.ErrPathTraversal) {
		t.Fatalf("err = %v, want ErrPathTraversal", err)
	}

	obj, err := s.Download(context.Background(), "aws", "production-logs", "2026/06/payment.log")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	_ = obj.Body.Close()
	if fp.downloadKey != "2026/06/payment.log" {
		t.Errorf("provider got key %q, want sanitized key", fp.downloadKey)
	}
}

func TestServiceListClampsLimit(t *testing.T) {
	fp := &fakeProvider{}
	s := newService(fp)
	if _, err := s.List(context.Background(), "aws", "production-logs", "", "", 99999); err != nil {
		t.Fatal(err)
	}
	if got := fp.listCalls[0].Limit; got != maxListLimit && got != defaultListLimit {
		t.Errorf("limit = %d, want clamped to <= %d", got, maxListLimit)
	}
}

func TestServicePresignClampsTTL(t *testing.T) {
	fp := &fakeProvider{presignURL: "https://example/x"}
	s := newService(fp)

	if _, _, err := s.Presign(context.Background(), "aws", "production-logs", "a.log", 24*time.Hour); err != nil {
		t.Fatal(err)
	}
	if fp.lastPresignTTL != 15*time.Minute {
		t.Errorf("clamped ttl = %v, want 15m", fp.lastPresignTTL)
	}
	if _, _, err := s.Presign(context.Background(), "aws", "production-logs", "a.log", time.Second); err != nil {
		t.Fatal(err)
	}
	if fp.lastPresignTTL != time.Minute {
		t.Errorf("clamped ttl = %v, want 1m", fp.lastPresignTTL)
	}
}

func TestServiceReady(t *testing.T) {
	healthy := newService(&fakeProvider{})
	if !healthy.Ready(context.Background()) {
		t.Error("expected ready when a provider is healthy")
	}

	down := newService(&fakeProvider{healthErr: errors.New("boom")})
	if down.Ready(context.Background()) {
		t.Error("expected NOT ready when the only provider is down")
	}

	empty := NewService(time.Minute, 15*time.Minute) // no providers registered
	if !empty.Ready(context.Background()) {
		t.Error("expected ready (process-up) when no providers are registered")
	}
}
