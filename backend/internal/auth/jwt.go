package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	// ErrTokenInvalid covers any validation failure other than expiry.
	ErrTokenInvalid = errors.New("token invalid")
	// ErrTokenExpired is returned specifically for expired tokens so the API
	// can answer TOKEN_EXPIRED (drives the SPA's re-login flow).
	ErrTokenExpired = errors.New("token expired")
)

// Claims is the JWT payload: the registered claims plus our authorization data.
type Claims struct {
	Email            string   `json:"email"`
	Role             Role     `json:"role"`
	AllowedProviders []string `json:"allowed_providers"`
	jwt.RegisteredClaims
}

// Issuer signs and validates HS256 tokens. The clock is injectable for tests.
type Issuer struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
	now      func() time.Time
}

// NewIssuer constructs an Issuer. The secret length is validated upstream by
// config.Load (>= 32 bytes).
func NewIssuer(secret, issuer, audience string, ttl time.Duration) *Issuer {
	return &Issuer{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
		now:      time.Now,
	}
}

// Issue mints a signed token for the user and returns it with its claims.
func (i *Issuer) Issue(u *User) (string, *Claims, error) {
	now := i.now()
	claims := &Claims{
		Email:            u.Email,
		Role:             u.Role,
		AllowedProviders: u.AllowedProviders,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   u.ID,
			Audience:  jwt.ClaimStrings{i.audience},
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(i.ttl)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.secret)
	if err != nil {
		return "", nil, fmt.Errorf("sign token: %w", err)
	}
	return signed, claims, nil
}

// Validate parses and fully verifies a token string. Two independent guards
// reject non-HMAC algorithms: the key func (method type assertion) and the
// parser option (WithValidMethods). Issuer, audience, and expiry are required.
func (i *Issuer) Validate(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return i.secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(i.issuer),
		jwt.WithAudience(i.audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(5*time.Second),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}
	if !claims.Role.Valid() {
		return nil, fmt.Errorf("%w: unknown role %q", ErrTokenInvalid, claims.Role)
	}
	return claims, nil
}
