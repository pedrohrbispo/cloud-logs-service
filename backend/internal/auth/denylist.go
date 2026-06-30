package auth

import (
	"context"
	"sync"
	"time"
)

// Denylist is an in-memory set of revoked token IDs (jti) used to make logout
// effective before a token's natural expiry. It is intentionally not durable:
// on restart the list is empty (documented limitation; Redis is the production
// upgrade).
type Denylist struct {
	mu      sync.Mutex
	entries map[string]time.Time // jti -> token expiry
	now     func() time.Time
}

// NewDenylist creates an empty denylist.
func NewDenylist() *Denylist {
	return &Denylist{entries: make(map[string]time.Time), now: time.Now}
}

// Revoke marks a jti as revoked until its token would have expired.
func (d *Denylist) Revoke(jti string, exp time.Time) {
	if jti == "" {
		return
	}
	d.mu.Lock()
	d.entries[jti] = exp
	d.mu.Unlock()
}

// IsRevoked reports whether a jti is currently revoked. Expired entries are
// dropped lazily.
func (d *Denylist) IsRevoked(jti string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	exp, ok := d.entries[jti]
	if !ok {
		return false
	}
	if d.now().After(exp) {
		delete(d.entries, jti)
		return false
	}
	return true
}

// StartCleanup sweeps expired entries on an interval until ctx is cancelled.
func (d *Denylist) StartCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				d.sweep()
			}
		}
	}()
}

func (d *Denylist) sweep() {
	d.mu.Lock()
	now := d.now()
	for jti, exp := range d.entries {
		if now.After(exp) {
			delete(d.entries, jti)
		}
	}
	d.mu.Unlock()
}
