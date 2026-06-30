package sanitize

import (
	"errors"
	"testing"
)

func TestKey(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr error
	}{
		// valid keys
		{"simple", "payment.log", "payment.log", nil},
		{"nested", "2026/06/payment.log", "2026/06/payment.log", nil},
		{"underscores and dashes", "app_logs/worker-1.log", "app_logs/worker-1.log", nil},
		{"collapses inner dot segment", "a/./b.log", "a/b.log", nil},
		{"collapses safe parent within root", "a/b/../c.log", "a/c.log", nil},

		// traversal
		{"dotdot prefix", "../etc/passwd", "", ErrPathTraversal},
		{"encoded dotdot", "%2e%2e%2fetc%2fpasswd", "", ErrPathTraversal},
		{"encoded slash dotdot", "logs%2f..%2f..%2fsecret", "", ErrPathTraversal},
		{"backslash traversal", "..\\..\\windows", "", ErrPathTraversal},
		{"deep traversal", "a/../../../../etc/passwd", "", ErrPathTraversal},

		// invalid
		{"empty", "", "", ErrInvalidKey},
		{"absolute", "/etc/passwd", "", ErrInvalidKey},
		{"encoded null byte", "file%00.log", "", ErrInvalidKey},
		{"leading dash", "-rf.log", "", ErrInvalidKey},
		{"leading dot", ".hidden", "", ErrInvalidKey},
		{"spaces", "my file.log", "", ErrInvalidKey},
		{"tilde", "~/secret", "", ErrInvalidKey},
		{"control char", "file\n.log", "", ErrInvalidKey},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Key(tc.in)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("Key(%q) err = %v, want %v", tc.in, err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Key(%q) unexpected err = %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Key(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestPrefix(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"empty allowed", "", "", false},
		{"simple", "2026", "2026", false},
		{"trailing slash kept", "2026/06/", "2026/06/", false},
		{"traversal rejected", "../x", "", true},
		{"absolute rejected", "/x", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Prefix(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Prefix(%q) expected error", tc.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("Prefix(%q) unexpected err = %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Prefix(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestBucket(t *testing.T) {
	valid := []string{"production-logs", "app-logs-prod", "gcs-prod-logs", "my.bucket"}
	for _, b := range valid {
		if _, err := Bucket(b); err != nil {
			t.Errorf("Bucket(%q) unexpected err = %v", b, err)
		}
	}
	invalid := []string{"", "x", "UPPER", "has space", "-leading", "trailing-", "a_b"}
	for _, b := range invalid {
		if _, err := Bucket(b); err == nil {
			t.Errorf("Bucket(%q) expected error", b)
		}
	}
}
