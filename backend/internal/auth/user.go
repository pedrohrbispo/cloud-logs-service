// Package auth implements authentication (credential verification, JWT
// issuing/validation) and the primitives the RBAC middleware enforces.
package auth

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// BcryptCost is the work factor for password hashing (~250ms/verify).
const BcryptCost = 12

// bcryptMaxPasswordBytes is bcrypt's hard input limit. Longer inputs are
// silently truncated by the algorithm, so we reject them instead of hashing.
const bcryptMaxPasswordBytes = 72

// Role is a coarse authorization role.
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleViewer Role = "viewer"
)

// Valid reports whether r is a known role.
func (r Role) Valid() bool { return r == RoleAdmin || r == RoleViewer }

// User is an account in the in-memory store. A nil AllowedProviders means the
// user may access every configured provider (admins).
type User struct {
	ID               string
	Email            string
	PasswordHash     []byte
	Role             Role
	AllowedProviders []string
}

var (
	// ErrInvalidCredentials is returned for any failed login. It deliberately
	// does not distinguish "unknown email" from "wrong password".
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrPasswordTooLong is returned when a password exceeds bcrypt's 72-byte
	// limit; the caller maps it to a 400, not a 401.
	ErrPasswordTooLong = errors.New("password exceeds 72 bytes")
)

// UserStore verifies credentials and looks up users.
type UserStore interface {
	Authenticate(email, password string) (*User, error)
	FindByEmail(email string) (*User, bool)
}

type memoryStore struct {
	byEmail   map[string]*User
	dummyHash []byte // keeps Authenticate constant-time for unknown emails
}

// NewMemoryStore builds an in-memory store and precomputes a dummy bcrypt hash
// so authentication runs bcrypt even for unknown emails (defeats timing-based
// user enumeration).
func NewMemoryStore(users []*User) UserStore {
	m := &memoryStore{byEmail: make(map[string]*User, len(users))}
	for _, u := range users {
		m.byEmail[strings.ToLower(u.Email)] = u
	}
	dummy, err := bcrypt.GenerateFromPassword([]byte("cla-dummy-constant-time-value"), BcryptCost)
	if err != nil {
		panic("auth: failed to precompute dummy hash: " + err.Error())
	}
	m.dummyHash = dummy
	return m
}

func (m *memoryStore) FindByEmail(email string) (*User, bool) {
	u, ok := m.byEmail[strings.ToLower(strings.TrimSpace(email))]
	return u, ok
}

// Authenticate verifies the password against the stored hash. It always runs
// bcrypt (against a dummy hash when the email is unknown) so the response time
// does not reveal whether an account exists.
func (m *memoryStore) Authenticate(email, password string) (*User, error) {
	// Reject over-long passwords up front: this is account-independent input
	// validation, so it leaks nothing about which emails exist.
	if len(password) > bcryptMaxPasswordBytes {
		return nil, ErrPasswordTooLong
	}

	user, found := m.byEmail[strings.ToLower(strings.TrimSpace(email))]
	hash := m.dummyHash
	if found {
		hash = user.PasswordHash
	}

	err := bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil || !found {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}
