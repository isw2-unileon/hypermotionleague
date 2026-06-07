package config

import "testing"

// Load must reject an empty or too-short JWT_SECRET (weak HS256 secrets are
// brute-forceable, enabling token forgery) and accept a sufficiently long one.
func TestLoad_JWTSecretStrength(t *testing.T) {
	cases := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{"empty", "", true},
		{"too short", "too-short-secret", true},                        // 16 chars
		{"exactly minimum", "0123456789abcdef0123456789abcdef", false}, // 32 chars
		{"strong", "0123456789abcdef0123456789abcdef0123456789abcdef", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET", tc.secret)

			cfg, err := Load()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for secret %q (len %d), got nil", tc.secret, len(tc.secret))
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error for secret of len %d, got %v", len(tc.secret), err)
			}
			if cfg.JWTSecret != tc.secret {
				t.Fatalf("JWTSecret = %q, want %q", cfg.JWTSecret, tc.secret)
			}
		})
	}
}
