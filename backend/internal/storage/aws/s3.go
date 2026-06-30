// Package aws is the AWS S3 implementation of storage.Provider, targeting
// LocalStack locally. It uses two clients: an internal one (compose hostname)
// for list/download, and a public one (localhost) used ONLY to sign URLs, so a
// browser on the host can open them.
package aws

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithy "github.com/aws/smithy-go"

	"cloud-log-access/internal/storage"
)

// Config configures the S3 adapter.
type Config struct {
	Region           string
	InternalEndpoint string // e.g. http://localstack:4566 (server-side)
	PublicEndpoint   string // e.g. http://localhost:4566 (used for signing)
	AccessKeyID      string
	SecretAccessKey  string
	UsePathStyle     bool
}

// Adapter implements storage.Provider for S3.
type Adapter struct {
	client  *s3.Client
	presign *s3.PresignClient
	region  string
}

// New builds the adapter. Static credentials + an explicit endpoint mean the
// SDK never probes EC2 instance metadata (IMDS).
func New(cfg Config) *Adapter {
	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")

	internal := s3.New(s3.Options{
		Region:       cfg.Region,
		Credentials:  creds,
		BaseEndpoint: aws.String(cfg.InternalEndpoint),
		UsePathStyle: cfg.UsePathStyle,
	})
	public := s3.New(s3.Options{
		Region:       cfg.Region,
		Credentials:  creds,
		BaseEndpoint: aws.String(cfg.PublicEndpoint),
		UsePathStyle: cfg.UsePathStyle,
	})

	return &Adapter{
		client:  internal,
		presign: s3.NewPresignClient(public),
		region:  cfg.Region,
	}
}

// Name returns the provider id.
func (a *Adapter) Name() string { return "aws" }

// List returns a page of objects.
func (a *Adapter) List(ctx context.Context, p storage.ListParams) (*storage.ListResult, error) {
	in := &s3.ListObjectsV2Input{Bucket: aws.String(p.Bucket)}
	if p.Prefix != "" {
		in.Prefix = aws.String(p.Prefix)
	}
	if p.Limit > 0 {
		in.MaxKeys = aws.Int32(p.Limit)
	}
	if p.Cursor != "" {
		in.ContinuationToken = aws.String(p.Cursor)
	}

	out, err := a.client.ListObjectsV2(ctx, in)
	if err != nil {
		return nil, mapErr(err)
	}

	res := &storage.ListResult{Objects: make([]storage.ObjectInfo, 0, len(out.Contents))}
	for _, o := range out.Contents {
		res.Objects = append(res.Objects, storage.ObjectInfo{
			Key:          aws.ToString(o.Key),
			Size:         aws.ToInt64(o.Size),
			LastModified: aws.ToTime(o.LastModified),
			ETag:         strings.Trim(aws.ToString(o.ETag), `"`),
		})
	}
	if aws.ToBool(out.IsTruncated) {
		res.NextCursor = aws.ToString(out.NextContinuationToken)
	}
	return res, nil
}

// Download streams an object's body. The caller must Close it.
func (a *Adapter) Download(ctx context.Context, bucket, key string) (*storage.Object, error) {
	out, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &storage.Object{
		Body:         out.Body,
		Size:         aws.ToInt64(out.ContentLength),
		ContentType:  aws.ToString(out.ContentType),
		ETag:         strings.Trim(aws.ToString(out.ETag), `"`),
		LastModified: aws.ToTime(out.LastModified),
	}, nil
}

// Presign returns a time-limited GET URL signed against the public endpoint.
// Content-Disposition is set as a SIGNED query parameter so it survives.
func (a *Adapter) Presign(ctx context.Context, bucket, key string, ttl time.Duration) (string, time.Time, error) {
	req, err := a.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(bucket),
		Key:                        aws.String(key),
		ResponseContentDisposition: aws.String(`attachment; filename="` + path.Base(key) + `"`),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", time.Time{}, mapErr(err)
	}
	return req.URL, time.Now().Add(ttl), nil
}

// HealthCheck probes connectivity to the backend.
func (a *Adapter) HealthCheck(ctx context.Context) error {
	if _, err := a.client.ListBuckets(ctx, &s3.ListBucketsInput{}); err != nil {
		return mapErr(err)
	}
	return nil
}

// mapErr translates SDK errors into storage sentinels.
func mapErr(err error) error {
	var noKey *types.NoSuchKey
	var noBucket *types.NoSuchBucket
	var notFound *types.NotFound
	if errors.As(err, &noKey) || errors.As(err, &noBucket) || errors.As(err, &notFound) {
		return storage.ErrNotFound
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchKey", "NoSuchBucket", "NotFound", "404":
			return storage.ErrNotFound
		}
	}
	return fmt.Errorf("%w: %v", storage.ErrUnavailable, err)
}
