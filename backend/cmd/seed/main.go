// Command seed is a one-shot job that ensures each configured provider's
// bucket/container exists and uploads the sample log files. It is idempotent
// and waits for the emulators to become reachable, so it can run as a compose
// init step before the BFF starts. Providers are seeded only when their
// endpoint env var is set.
package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	gcs "cloud.google.com/go/storage"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/aws/aws-sdk-go-v2/aws"
	awscreds "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"google.golang.org/api/option"

	"cloud-log-access/internal/sampledata"
)

const contentType = "text/plain; charset=utf-8"

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if err := seedS3(ctx); err != nil {
		log.Fatalf("seed s3: %v", err)
	}
	if ep := os.Getenv("GCS_INTERNAL_ENDPOINT"); ep != "" {
		if err := seedGCS(ctx, ep); err != nil {
			log.Fatalf("seed gcs: %v", err)
		}
	}
	if ep := os.Getenv("AZURE_INTERNAL_ENDPOINT"); ep != "" {
		if err := seedAzure(ctx, ep); err != nil {
			log.Fatalf("seed azure: %v", err)
		}
	}
	log.Println("seed: done")
}

func seedS3(ctx context.Context) error {
	endpoint := getenv("S3_INTERNAL_ENDPOINT", "http://localhost:4566")
	bucket := firstCSV(getenv("S3_BUCKETS", "production-logs"))

	client := s3.New(s3.Options{
		Region:       getenv("AWS_REGION", "us-east-1"),
		Credentials:  awscreds.NewStaticCredentialsProvider(getenv("AWS_ACCESS_KEY_ID", "test"), getenv("AWS_SECRET_ACCESS_KEY", "test"), ""),
		BaseEndpoint: aws.String(endpoint),
		UsePathStyle: true,
	})

	if err := waitFor(ctx, func() error { _, e := client.ListBuckets(ctx, &s3.ListBucketsInput{}); return e }); err != nil {
		return err
	}

	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	var owned *s3types.BucketAlreadyOwnedByYou
	var exists *s3types.BucketAlreadyExists
	if err != nil && !errors.As(err, &owned) && !errors.As(err, &exists) {
		return err
	}

	for _, f := range sampledata.Logs() {
		if _, err := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(f.Name),
			Body:        bytes.NewReader([]byte(f.Content)),
			ContentType: aws.String(contentType),
		}); err != nil {
			return err
		}
	}
	log.Printf("seed: s3://%s seeded with %d logs", bucket, len(sampledata.Logs()))
	return nil
}

func seedGCS(ctx context.Context, endpoint string) error {
	bucket := firstCSV(getenv("GCS_BUCKETS", "gcs-prod-logs"))

	client, err := gcs.NewClient(ctx, option.WithEndpoint(endpoint), option.WithoutAuthentication())
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	if err := waitFor(ctx, func() error {
		_, e := client.Bucket(bucket).Attrs(ctx)
		if errors.Is(e, gcs.ErrBucketNotExist) {
			return nil // reachable; bucket just doesn't exist yet
		}
		return e
	}); err != nil {
		return err
	}

	if err := client.Bucket(bucket).Create(ctx, "cloud-log-access", nil); err != nil &&
		!strings.Contains(err.Error(), "Conflict") && !strings.Contains(err.Error(), "already") {
		return err
	}

	for _, f := range sampledata.Logs() {
		w := client.Bucket(bucket).Object(f.Name).NewWriter(ctx)
		w.ContentType = contentType
		if _, err := w.Write([]byte(f.Content)); err != nil {
			_ = w.Close()
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}
	}
	log.Printf("seed: gcs://%s seeded with %d logs", bucket, len(sampledata.Logs()))
	return nil
}

func seedAzure(ctx context.Context, endpoint string) error {
	container := firstCSV(getenv("AZURE_CONTAINERS", "app-logs-prod"))
	cred, err := azblob.NewSharedKeyCredential(getenv("AZURE_STORAGE_ACCOUNT", "devstoreaccount1"), os.Getenv("AZURE_STORAGE_KEY"))
	if err != nil {
		return err
	}
	client, err := azblob.NewClientWithSharedKeyCredential(endpoint, cred, nil)
	if err != nil {
		return err
	}

	if err := waitFor(ctx, func() error {
		pager := client.NewListContainersPager(nil)
		_, e := pager.NextPage(ctx)
		return e
	}); err != nil {
		return err
	}

	if _, err := client.CreateContainer(ctx, container, nil); err != nil &&
		!bloberror.HasCode(err, bloberror.ContainerAlreadyExists) {
		return err
	}

	for _, f := range sampledata.Logs() {
		if _, err := client.UploadBuffer(ctx, container, f.Name, []byte(f.Content), &azblob.UploadBufferOptions{
			HTTPHeaders: &blob.HTTPHeaders{BlobContentType: to.Ptr(contentType)},
		}); err != nil {
			return err
		}
	}
	log.Printf("seed: azure://%s seeded with %d logs", container, len(sampledata.Logs()))
	return nil
}

func waitFor(ctx context.Context, probe func() error) error {
	var last error
	for i := 0; i < 30; i++ {
		if last = probe(); last == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return errors.New("emulator not ready: " + last.Error())
}

func firstCSV(csv string) string {
	if i := strings.IndexByte(csv, ','); i >= 0 {
		return strings.TrimSpace(csv[:i])
	}
	return strings.TrimSpace(csv)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
