// Package sanitize validates and normalizes object keys, prefixes, and bucket
// names before they reach a cloud SDK. It implements the rules in
// docs/AUTH-SECURITY-PLAN.md §6 (path-traversal / IDOR defense).
package sanitize

import (
	"errors"
	"net/url"
	"path"
	"regexp"
	"strings"
)

var (
	// ErrPathTraversal indicates the key tried to escape its prefix.
	ErrPathTraversal = errors.New("path traversal detected")
	// ErrInvalidKey indicates the key violates the allow-list rules.
	ErrInvalidKey = errors.New("invalid object key")
	// ErrInvalidBucket indicates the bucket name is malformed.
	ErrInvalidBucket = errors.New("invalid bucket name")
)

const maxKeyLen = 1024

// keyRe requires the key to start with an alphanumeric (no leading "." for
// hidden paths, no leading "-" for CLI flag injection). Slashes are allowed
// because object keys are hierarchical (e.g. 2026/06/payment.log).
var keyRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9./_\-]*$`)

// bucketRe is a conservative subset of the S3 bucket naming rules.
var bucketRe = regexp.MustCompile(`^[a-z0-9][a-z0-9.\-]{1,61}[a-z0-9]$`)

// Key decodes (once), normalizes, and validates an object key. It rejects
// traversal, absolute paths, null bytes, disallowed characters, and over-long
// keys.
func Key(raw string) (string, error) {
	if raw == "" {
		return "", ErrInvalidKey
	}
	decoded, err := url.PathUnescape(raw)
	if err != nil {
		return "", ErrInvalidKey
	}
	decoded = strings.ReplaceAll(decoded, "\\", "/")
	cleaned := path.Clean(decoded)

	if strings.Contains(cleaned, "..") {
		return "", ErrPathTraversal
	}
	if strings.HasPrefix(cleaned, "/") {
		return "", ErrInvalidKey
	}
	if strings.ContainsRune(cleaned, 0) {
		return "", ErrInvalidKey
	}
	if len(cleaned) > maxKeyLen {
		return "", ErrInvalidKey
	}
	if !keyRe.MatchString(cleaned) {
		return "", ErrInvalidKey
	}
	return cleaned, nil
}

// Prefix validates a list prefix. An empty prefix is allowed (lists the whole
// bucket). A trailing slash is permitted and stripped for validation.
func Prefix(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	hadTrailingSlash := strings.HasSuffix(raw, "/")
	trimmed := strings.TrimSuffix(raw, "/")
	if trimmed == "" {
		return "", nil
	}
	cleaned, err := Key(trimmed)
	if err != nil {
		return "", err
	}
	if hadTrailingSlash {
		cleaned += "/"
	}
	return cleaned, nil
}

// Bucket validates a bucket/container name.
func Bucket(raw string) (string, error) {
	if !bucketRe.MatchString(raw) {
		return "", ErrInvalidBucket
	}
	return raw, nil
}
