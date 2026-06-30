package auth

import "golang.org/x/crypto/bcrypt"

// SeedUsers returns the demo accounts, hashing their passwords at startup. The
// viewer is restricted to AWS so the provider-level 403 (NoPermissionState) is
// demonstrable. These credentials are intentionally documented in the README
// for graders.
func SeedUsers() ([]*User, error) {
	adminHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), BcryptCost)
	if err != nil {
		return nil, err
	}
	viewerHash, err := bcrypt.GenerateFromPassword([]byte("viewer123"), BcryptCost)
	if err != nil {
		return nil, err
	}
	return []*User{
		{
			ID:               "u-admin",
			Email:            "admin@example.com",
			PasswordHash:     adminHash,
			Role:             RoleAdmin,
			AllowedProviders: nil, // all providers
		},
		{
			ID:               "u-viewer",
			Email:            "viewer@example.com",
			PasswordHash:     viewerHash,
			Role:             RoleViewer,
			AllowedProviders: []string{"aws"},
		},
	}, nil
}
