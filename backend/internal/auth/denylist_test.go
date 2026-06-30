package auth

import (
	"testing"
	"time"
)

func TestDenylist(t *testing.T) {
	d := NewDenylist()
	future := time.Now().Add(time.Hour)

	if d.IsRevoked("jti-1") {
		t.Fatal("empty denylist must not report a token as revoked")
	}

	d.Revoke("jti-1", future)
	if !d.IsRevoked("jti-1") {
		t.Fatal("jti-1 should be revoked")
	}

	// An entry whose token has already expired is treated as not revoked.
	d.Revoke("jti-expired", time.Now().Add(-time.Minute))
	if d.IsRevoked("jti-expired") {
		t.Fatal("an expired revocation must not be active")
	}

	// Revoking an empty jti is a no-op.
	d.Revoke("", future)
	if d.IsRevoked("") {
		t.Fatal("empty jti must never be considered revoked")
	}
}
