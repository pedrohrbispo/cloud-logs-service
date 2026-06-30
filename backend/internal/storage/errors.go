package storage

import "errors"

var (
	// ErrInvalidProvider: the provider id is not registered.
	ErrInvalidProvider = errors.New("invalid provider")
	// ErrInvalidBucket: the bucket is not in the provider's allow-list.
	ErrInvalidBucket = errors.New("invalid bucket")
	// ErrNoBuckets: the provider has no buckets configured (operator error).
	ErrNoBuckets = errors.New("no buckets configured for provider")
	// ErrNotFound: the requested object does not exist.
	ErrNotFound = errors.New("object not found")
	// ErrUnavailable: the storage backend is unreachable or errored.
	ErrUnavailable = errors.New("storage backend unavailable")
)
