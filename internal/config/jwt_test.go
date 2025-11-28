package config

import (
	"os"
	"testing"
	"time"
)

// Test constants
const (
	testJWTSecret = "this-is-a-very-long-secret-key-for-testing-at-least-32-chars"
)

// =============================================================================
// Test Helpers
// =============================================================================

// clearAllJWTEnvVars clears all JWT-related environment variables.
func clearAllJWTEnvVars(t *testing.T) {
	t.Helper()
	vars := []string{"JWT_SECRET", "JWT_ACCESS_EXPIRY", "JWT_REFRESH_EXPIRY"}
	for _, v := range vars {
		t.Setenv(v, "")
		os.Unsetenv(v) //nolint:errcheck // test cleanup
	}
}

// =============================================================================
// NewJWTConfig Tests
// =============================================================================

func TestNewJWTConfig_ReturnsNilWhenSecretNotSet(t *testing.T) {
	clearAllJWTEnvVars(t)

	cfg := NewJWTConfig()

	if cfg != nil {
		t.Error("NewJWTConfig() should return nil when JWT_SECRET is not set")
	}
}

func TestNewJWTConfig_LoadsAllFieldsFromEnv(t *testing.T) {
	clearAllJWTEnvVars(t)

	setEnvForTest(t, "JWT_SECRET", testJWTSecret)
	setEnvForTest(t, "JWT_ACCESS_EXPIRY", "30m")
	setEnvForTest(t, "JWT_REFRESH_EXPIRY", "24h")

	cfg := NewJWTConfig()

	if cfg == nil {
		t.Fatal("NewJWTConfig() returned nil")
	}
	if cfg.Secret != testJWTSecret {
		t.Errorf("Secret = %q, want %q", cfg.Secret, testJWTSecret)
	}
	if cfg.AccessExpiry != 30*time.Minute {
		t.Errorf("AccessExpiry = %v, want %v", cfg.AccessExpiry, 30*time.Minute)
	}
	if cfg.RefreshExpiry != 24*time.Hour {
		t.Errorf("RefreshExpiry = %v, want %v", cfg.RefreshExpiry, 24*time.Hour)
	}
}

func TestNewJWTConfig_UsesDefaultsForOptionalFields(t *testing.T) {
	clearAllJWTEnvVars(t)

	setEnvForTest(t, "JWT_SECRET", testJWTSecret)

	cfg := NewJWTConfig()

	if cfg == nil {
		t.Fatal("NewJWTConfig() returned nil")
	}
	if cfg.AccessExpiry != 15*time.Minute {
		t.Errorf("AccessExpiry default = %v, want %v", cfg.AccessExpiry, 15*time.Minute)
	}
	if cfg.RefreshExpiry != 168*time.Hour {
		t.Errorf("RefreshExpiry default = %v, want %v", cfg.RefreshExpiry, 168*time.Hour)
	}
}

func TestNewJWTConfig_PanicsOnShortSecret(t *testing.T) {
	clearAllJWTEnvVars(t)

	setEnvForTest(t, "JWT_SECRET", "short") // Less than 32 chars

	defer func() {
		if r := recover(); r == nil {
			t.Error("NewJWTConfig() should panic when secret is too short")
		}
	}()

	NewJWTConfig()
}

// =============================================================================
// JWTConfig.HasJWT Tests
// =============================================================================

func TestJWTConfig_HasJWT_TableDriven(t *testing.T) {
	tests := []struct {
		name string
		cfg  *JWTConfig
		want bool
	}{
		{
			name: "returns false for nil config",
			cfg:  nil,
			want: false,
		},
		{
			name: "returns false for empty secret",
			cfg:  &JWTConfig{Secret: ""},
			want: false,
		},
		{
			name: "returns true for valid secret",
			cfg:  &JWTConfig{Secret: "my-secret"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.HasJWT()

			if got != tt.want {
				t.Errorf("HasJWT() = %v, want %v", got, tt.want)
			}
		})
	}
}
