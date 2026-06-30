package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	mw "cloud-log-access/internal/api/middleware"
	"cloud-log-access/internal/api/response"
	"cloud-log-access/internal/auth"
	"cloud-log-access/internal/reqctx"
)

const loginMaxBodyBytes = 4 << 10 // 4 KiB

// tokenIssuer signs a token for an authenticated user.
type tokenIssuer interface {
	Issue(*auth.User) (string, *auth.Claims, error)
}

// revoker revokes a token id until its expiry.
type revoker interface {
	Revoke(jti string, exp time.Time)
}

// Auth handles login, logout, and the current-user endpoint.
type Auth struct {
	users    auth.UserStore
	issuer   tokenIssuer
	denylist revoker
	expiry   time.Duration
	logger   *slog.Logger
}

// NewAuth constructs the auth handler.
func NewAuth(users auth.UserStore, issuer tokenIssuer, denylist revoker, expiry time.Duration, logger *slog.Logger) *Auth {
	return &Auth{users: users, issuer: issuer, denylist: denylist, expiry: expiry, logger: logger}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userDTO struct {
	ID               string   `json:"id"`
	Email            string   `json:"email"`
	Role             string   `json:"role"`
	AllowedProviders []string `json:"allowed_providers"`
}

type loginResponse struct {
	Token     string  `json:"token"`
	TokenType string  `json:"token_type"`
	ExpiresIn int     `json:"expires_in"`
	User      userDTO `json:"user"`
}

// Login verifies credentials and returns a signed JWT plus the user (including
// allowed_providers, so the SPA needs no client-side token decoding).
func (a *Auth) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, loginMaxBodyBytes)

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeValidation, "invalid request body")
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" || !strings.Contains(req.Email, "@") {
		response.Error(w, r, http.StatusBadRequest, response.CodeValidation, "email and password are required")
		return
	}

	user, err := a.users.Authenticate(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrPasswordTooLong) {
			response.Error(w, r, http.StatusBadRequest, response.CodeValidation, "password must be at most 72 bytes")
			return
		}
		a.logAttempt(r, req.Email, "failure")
		// Identical response for unknown email and wrong password (no enumeration).
		response.Error(w, r, http.StatusUnauthorized, response.CodeInvalidCredentials, "invalid credentials")
		return
	}

	token, _, err := a.issuer.Issue(user)
	if err != nil {
		a.logger.LogAttrs(r.Context(), slog.LevelError, "token_issue_failed",
			slog.String("request_id", reqctx.RequestID(r.Context())), slog.String("err", err.Error()))
		response.Error(w, r, http.StatusInternalServerError, response.CodeInternalError, "could not issue token")
		return
	}

	a.logAttempt(r, user.Email, "success")
	response.JSON(w, r, http.StatusOK, loginResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: int(a.expiry.Seconds()),
		User: userDTO{
			ID:               user.ID,
			Email:            user.Email,
			Role:             string(user.Role),
			AllowedProviders: user.AllowedProviders,
		},
	})
}

// Logout revokes the current token (jti denylist) so it cannot be replayed.
func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, r, http.StatusUnauthorized, response.CodeTokenMissing, "authentication required")
		return
	}
	if claims.ExpiresAt != nil {
		a.denylist.Revoke(claims.ID, claims.ExpiresAt.Time)
	}
	a.logger.LogAttrs(r.Context(), slog.LevelInfo, "auth.logout",
		slog.String("user_id", claims.Subject),
		slog.String("email", claims.Email),
		slog.String("ip", mw.ClientIP(r)),
		slog.String("request_id", reqctx.RequestID(r.Context())),
	)
	response.JSON(w, r, http.StatusOK, map[string]string{"status": "logged_out"})
}

// Me returns the current authenticated user (used by the SPA to hydrate state).
func (a *Auth) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, r, http.StatusUnauthorized, response.CodeTokenMissing, "authentication required")
		return
	}
	response.JSON(w, r, http.StatusOK, userDTO{
		ID:               claims.Subject,
		Email:            claims.Email,
		Role:             string(claims.Role),
		AllowedProviders: claims.AllowedProviders,
	})
}

// logAttempt emits a structured login-attempt audit line. It never logs the
// password or the issued token.
func (a *Auth) logAttempt(r *http.Request, email, outcome string) {
	a.logger.LogAttrs(r.Context(), slog.LevelInfo, "auth.login.attempt",
		slog.String("outcome", outcome),
		slog.String("email", email),
		slog.String("ip", mw.ClientIP(r)),
		slog.String("request_id", reqctx.RequestID(r.Context())),
	)
}
