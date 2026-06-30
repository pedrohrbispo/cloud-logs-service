package aws

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"cloud-log-access/internal/storage"
)

// TestS3Integration exercises the adapter against a live S3 emulator. It is
// skipped unless RUN_INTEGRATION=1 (the B5 CI smoke job sets it after starting
// LocalStack). Run locally with:
//
//	RUN_INTEGRATION=1 go test ./internal/storage/aws
func TestS3Integration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("integration test: set RUN_INTEGRATION=1 with LocalStack running")
	}
	endpoint := envOr("S3_INTERNAL_ENDPOINT", "http://localhost:4566")
	const bucket = "integ-test-bucket"
	ctx := context.Background()

	raw := s3.New(s3.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("test", "test", ""),
		BaseEndpoint: aws.String(endpoint),
		UsePathStyle: true,
	})
	_, _ = raw.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	if _, err := raw.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String("dir/test.log"),
		Body:        bytes.NewReader([]byte("hello-integration")),
		ContentType: aws.String("text/plain"),
	}); err != nil {
		t.Fatalf("put fixture: %v", err)
	}

	a := New(Config{
		Region:           "us-east-1",
		InternalEndpoint: endpoint,
		PublicEndpoint:   endpoint,
		AccessKeyID:      "test",
		SecretAccessKey:  "test",
		UsePathStyle:     true,
	})

	if err := a.HealthCheck(ctx); err != nil {
		t.Fatalf("health: %v", err)
	}

	res, err := a.List(ctx, storage.ListParams{Bucket: bucket, Limit: 100})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var found bool
	for _, o := range res.Objects {
		if o.Key == "dir/test.log" {
			found = true
			if o.Size != int64(len("hello-integration")) {
				t.Errorf("size = %d, want %d", o.Size, len("hello-integration"))
			}
		}
	}
	if !found {
		t.Fatal("uploaded object not returned by List")
	}

	obj, err := a.Download(ctx, bucket, "dir/test.log")
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	defer func() { _ = obj.Body.Close() }()
	body, _ := io.ReadAll(obj.Body)
	if string(body) != "hello-integration" {
		t.Errorf("body = %q, want hello-integration", string(body))
	}

	url, expiresAt, err := a.Presign(ctx, bucket, "dir/test.log", time.Minute)
	if err != nil {
		t.Fatalf("presign: %v", err)
	}
	if expiresAt.IsZero() {
		t.Error("expected a non-zero expiry for an S3 presigned URL")
	}
	if !strings.Contains(url, bucket) || !strings.Contains(url, "X-Amz-Signature") {
		t.Errorf("presigned URL looks malformed: %s", url)
	}

	// missing key maps to ErrNotFound
	if _, err := a.Download(ctx, bucket, "does-not-exist.log"); err == nil {
		t.Error("expected error for missing key")
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
