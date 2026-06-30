package response

import (
	"net/http"

	"cloud-log-access/internal/reqctx"
)

// Error codes returned in the error envelope. Kept as a closed catalog so the
// frontend can switch on them and so no raw internal detail leaks to clients.
const (
	CodeInternalError      = "INTERNAL_ERROR"
	CodeNotFound           = "NOT_FOUND"
	CodeValidation         = "VALIDATION_ERROR"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeTokenExpired       = "TOKEN_EXPIRED"
	CodeTokenInvalid       = "TOKEN_INVALID"
	CodeTokenMissing       = "TOKEN_MISSING"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeProviderNotAllowed = "PROVIDER_NOT_ALLOWED"
	CodeInvalidProvider    = "INVALID_PROVIDER"
	CodeInvalidBucket      = "INVALID_BUCKET"
	CodeFileNotFound       = "FILE_NOT_FOUND"
	CodePathTraversal      = "PATH_TRAVERSAL"
	CodeInvalidKey         = "INVALID_KEY"
	CodeTTLOutOfRange      = "TTL_OUT_OF_RANGE"
	CodeRateLimited        = "RATE_LIMITED"
	CodeStorageUnavailable = "STORAGE_UNAVAILABLE"
)

type errorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

// Error writes an error envelope. The message must be a safe, non-sensitive
// string; never interpolate raw user input or internal error details.
func Error(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	write(w, status, errorEnvelope{
		Error: errorBody{
			Code:      code,
			Message:   message,
			RequestID: reqctx.RequestID(r.Context()),
		},
	})
}
