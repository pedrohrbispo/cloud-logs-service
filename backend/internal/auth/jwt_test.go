package auth

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"
)

const (
	testSecret   = "test-secret-at-least-32-bytes-long-1234567890"
	testIssuer   = "cloud-log-access-service"
	testAudience = "cloud-log-access-service"
)

func testUser() *User {
	return &User{ID: "u1", Email: "admin@example.com", Role: RoleAdmin}
}

func newTestIssuer() *Issuer {
	return NewIssuer(testSecret, testIssuer, testAudience, time.Hour)
}

func TestIssueValidateRoundTrip(t *testing.T) {
	iss := newTestIssuer()
	token, issued, err := iss.Issue(testUser())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	got, err := iss.Validate(token)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if got.Email != "admin@example.com" || got.Role != RoleAdmin {
		t.Errorf("claims mismatch: %+v", got)
	}
	if got.ID != issued.ID || got.ID == "" {
		t.Errorf("jti mismatch or empty: got %q want %q", got.ID, issued.ID)
	}
}

func b64url(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }

func craftToken(alg, payload, sig string) string {
	return b64url(`{"alg":"`+alg+`","typ":"JWT"}`) + "." + b64url(payload) + "." + sig
}

// TestValidateRejectsAttacks covers the algorithm-confusion, tampering, and
// wrong-secret attack classes from the security plan.
func TestValidateRejectsAttacks(t *testing.T) {
	iss := newTestIssuer()
	valid, _, err := iss.Issue(testUser())
	if err != nil {
		t.Fatal(err)
	}
	// A payload that WOULD pass claim checks if the signature were trusted.
	payload := `{"iss":"cloud-log-access-service","aud":"cloud-log-access-service","role":"admin","exp":9999999999}`

	tests := []struct {
		name  string
		token string
	}{
		{"alg_none", craftToken("none", payload, "")},
		{"alg_none_mixed_case", craftToken("NoNe", payload, "")},
		{"alg_confusion_rs256", craftToken("RS256", payload, "ZmFrZXNpZ25hdHVyZQ")},
		{"alg_confusion_hs512", craftToken("HS512", payload, "ZmFrZXNpZw")},
		{"garbage", "not.a.jwt"},
		{"empty", ""},
		{"tampered_payload", tamperPayload(valid)},
		{"wrong_secret", signWith(t, "another-secret-also-32-bytes-long-987654")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := iss.Validate(tc.token); err == nil {
				t.Fatalf("expected validation error for %q, got nil", tc.name)
			}
		})
	}
}

func tamperPayload(token string) string {
	parts := strings.Split(token, ".")
	p := []byte(parts[1])
	if p[0] == 'A' {
		p[0] = 'B'
	} else {
		p[0] = 'A'
	}
	parts[1] = string(p)
	return strings.Join(parts, ".")
}

func signWith(t *testing.T, secret string) string {
	t.Helper()
	token, _, err := NewIssuer(secret, testIssuer, testAudience, time.Hour).Issue(testUser())
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestValidateExpired(t *testing.T) {
	iss := newTestIssuer()
	iss.now = func() time.Time { return time.Now().Add(-2 * time.Hour) } // exp = now-1h
	token, _, err := iss.Issue(testUser())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := newTestIssuer().Validate(token); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

func TestValidateRejectsWrongIssuerAndAudience(t *testing.T) {
	t.Run("wrong issuer", func(t *testing.T) {
		token, _, _ := NewIssuer(testSecret, "evil-issuer", testAudience, time.Hour).Issue(testUser())
		if _, err := newTestIssuer().Validate(token); err == nil {
			t.Fatal("expected error for wrong issuer")
		}
	})
	t.Run("wrong audience", func(t *testing.T) {
		token, _, _ := NewIssuer(testSecret, testIssuer, "evil-audience", time.Hour).Issue(testUser())
		if _, err := newTestIssuer().Validate(token); err == nil {
			t.Fatal("expected error for wrong audience")
		}
	})
}
