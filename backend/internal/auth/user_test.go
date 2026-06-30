package auth

import (
	"errors"
	"strings"
	"testing"
)

func buildStore(t *testing.T) UserStore {
	t.Helper()
	users, err := SeedUsers()
	if err != nil {
		t.Fatal(err)
	}
	return NewMemoryStore(users)
}

func TestAuthenticate(t *testing.T) {
	store := buildStore(t)

	t.Run("valid admin", func(t *testing.T) {
		u, err := store.Authenticate("admin@example.com", "admin123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Role != RoleAdmin {
			t.Errorf("role = %q, want admin", u.Role)
		}
		if u.AllowedProviders != nil {
			t.Errorf("admin AllowedProviders = %v, want nil (all)", u.AllowedProviders)
		}
	})

	t.Run("valid viewer restricted to aws", func(t *testing.T) {
		u, err := store.Authenticate("viewer@example.com", "viewer123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(u.AllowedProviders) != 1 || u.AllowedProviders[0] != "aws" {
			t.Errorf("AllowedProviders = %v, want [aws]", u.AllowedProviders)
		}
	})

	t.Run("email is case-insensitive", func(t *testing.T) {
		if _, err := store.Authenticate("ADMIN@Example.com", "admin123"); err != nil {
			t.Fatalf("expected case-insensitive match, got %v", err)
		}
	})

	t.Run("wrong password is rejected", func(t *testing.T) {
		if _, err := store.Authenticate("admin@example.com", "wrongpass"); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("err = %v, want ErrInvalidCredentials", err)
		}
	})

	t.Run("unknown email is rejected with the same error", func(t *testing.T) {
		if _, err := store.Authenticate("ghost@example.com", "whatever"); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("err = %v, want ErrInvalidCredentials", err)
		}
	})

	t.Run("password over 72 bytes is rejected", func(t *testing.T) {
		if _, err := store.Authenticate("admin@example.com", strings.Repeat("a", 73)); !errors.Is(err, ErrPasswordTooLong) {
			t.Fatalf("err = %v, want ErrPasswordTooLong", err)
		}
	})
}
