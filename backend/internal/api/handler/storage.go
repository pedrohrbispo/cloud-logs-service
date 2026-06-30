package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	mw "cloud-log-access/internal/api/middleware"
	"cloud-log-access/internal/api/response"
	"cloud-log-access/internal/auth"
	"cloud-log-access/internal/reqctx"
	"cloud-log-access/internal/storage"
	"cloud-log-access/pkg/sanitize"
)

const downloadTimeout = 120 * time.Second

// Storage serves provider/log endpoints over the storage.Service facade.
type Storage struct {
	svc    *storage.Service
	logger *slog.Logger
}

// NewStorage constructs the storage handler.
func NewStorage(svc *storage.Service, logger *slog.Logger) *Storage {
	return &Storage{svc: svc, logger: logger}
}

type providerDTO struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Buckets []string `json:"buckets"`
	Region  string   `json:"region,omitempty"`
	Healthy bool     `json:"healthy"`
	Error   string   `json:"error,omitempty"`
}

type logFileDTO struct {
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified"`
	ETag         string `json:"etag"`
	ContentType  string `json:"content_type"`
}

// ListProviders returns the providers the authenticated user may see, with a
// live health probe each.
func (h *Storage) ListProviders(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	all := h.svc.Providers(r.Context())

	out := make([]providerDTO, 0, len(all))
	for _, p := range all {
		if claims != nil && claims.AllowedProviders != nil && !contains(claims.AllowedProviders, p.ID) {
			continue
		}
		out = append(out, providerDTO{
			ID: p.ID, Name: p.Name, Buckets: p.Buckets,
			Region: p.Region, Healthy: p.Healthy, Error: p.Err,
		})
	}
	response.JSON(w, r, http.StatusOK, map[string]any{"providers": out})
}

// ListLogs lists objects in a bucket for a provider.
func (h *Storage) ListLogs(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	q := r.URL.Query()
	bucket := q.Get("bucket")
	prefix := q.Get("prefix")
	cursor := q.Get("cursor")
	limit := parseLimit(q.Get("limit"))

	res, err := h.svc.List(r.Context(), provider, bucket, prefix, cursor, limit)
	if err != nil {
		h.writeStorageError(w, r, err)
		return
	}

	files := make([]logFileDTO, 0, len(res.Objects))
	for _, o := range res.Objects {
		files = append(files, logFileDTO{
			Key:          o.Key,
			Size:         o.Size,
			LastModified: o.LastModified.UTC().Format(time.RFC3339),
			ETag:         o.ETag,
			ContentType:  o.ContentType,
		})
	}
	response.JSON(w, r, http.StatusOK, map[string]any{
		"files":       files,
		"next_cursor": res.NextCursor,
		"count":       len(files),
	})
}

// Download streams an object through the BFF with the auth boundary intact.
func (h *Storage) Download(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	bucket := r.URL.Query().Get("bucket")
	key := r.URL.Query().Get("key")

	ctx, cancel := context.WithTimeout(r.Context(), downloadTimeout)
	defer cancel()

	obj, err := h.svc.Download(ctx, provider, bucket, key)
	if err != nil {
		h.writeStorageError(w, r, err)
		return
	}
	defer func() { _ = obj.Body.Close() }()

	filename := path.Base(key)
	ct := obj.ContentType
	if ct == "" {
		ct = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if obj.Size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(obj.Size, 10))
	}
	w.WriteHeader(http.StatusOK)

	n, copyErr := io.Copy(w, obj.Body)
	h.auditDownload(r, provider, bucket, key, n, copyErr)
}

type presignRequest struct {
	Key        string `json:"key"`
	TTLSeconds int    `json:"ttl_seconds"`
}

type presignResponse struct {
	URL string `json:"url"`
	// Omitted when the link does not expire (e.g. the GCS emulator media URL).
	ExpiresAt  string `json:"expires_at,omitempty"`
	TTLSeconds int    `json:"ttl_seconds,omitempty"`
}

// Presign issues a temporary access link (admin-only). The server clamps the
// requested TTL to [PRESIGN_MIN_TTL, PRESIGN_MAX_TTL] and reports the effective
// expiry, so the client can show the real "expires in" value. The URL is a
// bearer credential and is NEVER logged.
func (h *Storage) Presign(w http.ResponseWriter, r *http.Request) {
	// Defense-in-depth: middleware already enforces admin, but re-check here.
	claims, _ := auth.ClaimsFromContext(r.Context())
	if claims == nil || claims.Role != auth.RoleAdmin {
		response.Error(w, r, http.StatusForbidden, response.CodeUnauthorized, "admin role required")
		return
	}

	provider := chi.URLParam(r, "provider")
	bucket := r.URL.Query().Get("bucket")

	var req presignRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeValidation, "invalid request body")
		return
	}

	url, expiresAt, err := h.svc.Presign(r.Context(), provider, bucket, req.Key, time.Duration(req.TTLSeconds)*time.Second)
	if err != nil {
		h.writeStorageError(w, r, err)
		return
	}

	// Honest reporting: a zero expiry means the provider's link does not expire
	// (the fake-gcs emulator has no signing key). Do NOT fabricate an expiry.
	resp := presignResponse{URL: url}
	auditExpiry := "non-expiring"
	if !expiresAt.IsZero() {
		resp.ExpiresAt = expiresAt.UTC().Format(time.RFC3339)
		resp.TTLSeconds = int(time.Until(expiresAt).Round(time.Second).Seconds())
		auditExpiry = resp.ExpiresAt
	}

	// Audit WITHOUT the URL (it is a bearer credential).
	h.logger.LogAttrs(r.Context(), slog.LevelInfo, "presign.created",
		slog.String("user_id", claims.Subject),
		slog.String("email", claims.Email),
		slog.String("provider", provider),
		slog.String("bucket", bucket),
		slog.String("key", req.Key),
		slog.String("expires_at", auditExpiry),
		slog.String("ip", mw.ClientIP(r)),
		slog.String("request_id", reqctx.RequestID(r.Context())),
	)

	response.JSON(w, r, http.StatusOK, resp)
}

func (h *Storage) writeStorageError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, storage.ErrInvalidProvider):
		response.Error(w, r, http.StatusBadRequest, response.CodeInvalidProvider, "unknown provider")
	case errors.Is(err, storage.ErrInvalidBucket), errors.Is(err, storage.ErrNoBuckets):
		response.Error(w, r, http.StatusBadRequest, response.CodeInvalidBucket, "unknown or unconfigured bucket")
	case errors.Is(err, sanitize.ErrPathTraversal):
		response.Error(w, r, http.StatusBadRequest, response.CodePathTraversal, "invalid object key")
	case errors.Is(err, sanitize.ErrInvalidKey):
		response.Error(w, r, http.StatusBadRequest, response.CodeInvalidKey, "invalid object key")
	case errors.Is(err, storage.ErrNotFound):
		response.Error(w, r, http.StatusNotFound, response.CodeFileNotFound, "log file not found")
	default:
		h.logger.LogAttrs(r.Context(), slog.LevelError, "storage_error",
			slog.String("request_id", reqctx.RequestID(r.Context())),
			slog.String("err", err.Error()),
		)
		response.Error(w, r, http.StatusServiceUnavailable, response.CodeStorageUnavailable, "storage backend unavailable")
	}
}

func (h *Storage) auditDownload(r *http.Request, provider, bucket, key string, bytes int64, copyErr error) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	email, sub := "", ""
	if claims != nil {
		email, sub = claims.Email, claims.Subject
	}
	status, copyErrStr := "ok", ""
	if copyErr != nil {
		status, copyErrStr = "error", copyErr.Error()
	}
	h.logger.LogAttrs(r.Context(), slog.LevelInfo, "file.download",
		slog.String("user_id", sub),
		slog.String("email", email),
		slog.String("provider", provider),
		slog.String("bucket", bucket),
		slog.String("key", key),
		slog.Int64("bytes", bytes),
		slog.String("status", status),
		slog.String("copy_err", copyErrStr),
		slog.String("ip", mw.ClientIP(r)),
		slog.String("request_id", reqctx.RequestID(r.Context())),
	)
}

func parseLimit(s string) int32 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil || n < 0 {
		return 0
	}
	return int32(n)
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
