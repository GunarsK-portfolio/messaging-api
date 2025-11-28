package config

import (
	"os"
	"testing"
)

// =============================================================================
// Test Helpers
// =============================================================================

// clearAllServiceEnvVars clears all service-related environment variables.
func clearAllServiceEnvVars(t *testing.T) {
	t.Helper()
	vars := []string{"SERVICE_NAME", "PORT", "ENVIRONMENT", "ALLOWED_ORIGINS", "SWAGGER_HOST"}
	for _, v := range vars {
		t.Setenv(v, "")
		os.Unsetenv(v) //nolint:errcheck // test cleanup
	}
}

// =============================================================================
// NewServiceConfig Tests
// =============================================================================

func TestNewServiceConfig_LoadsAllFieldsFromEnv(t *testing.T) {
	clearAllServiceEnvVars(t)

	setEnvForTest(t, "SERVICE_NAME", "test-service")
	setEnvForTest(t, "PORT", "9090")
	setEnvForTest(t, "ENVIRONMENT", "production")
	setEnvForTest(t, "ALLOWED_ORIGINS", "https://example.com,https://api.example.com")
	setEnvForTest(t, "SWAGGER_HOST", "api.example.com")

	cfg := NewServiceConfig()

	if cfg.Name != "test-service" {
		t.Errorf("Name = %q, want %q", cfg.Name, "test-service")
	}
	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9090")
	}
	if cfg.Environment != "production" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "production")
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Errorf("AllowedOrigins length = %d, want %d", len(cfg.AllowedOrigins), 2)
	}
	if cfg.SwaggerHost != "api.example.com" {
		t.Errorf("SwaggerHost = %q, want %q", cfg.SwaggerHost, "api.example.com")
	}
}

func TestNewServiceConfig_UsesDefaultsForOptionalFields(t *testing.T) {
	clearAllServiceEnvVars(t)

	// ALLOWED_ORIGINS has a default, so no env var needed
	setEnvForTest(t, "ALLOWED_ORIGINS", "http://localhost:3000")

	cfg := NewServiceConfig()

	if cfg.Name != "your-service" {
		t.Errorf("Name default = %q, want %q", cfg.Name, "your-service")
	}
	if cfg.Port != "8080" {
		t.Errorf("Port default = %q, want %q", cfg.Port, "8080")
	}
	if cfg.Environment != "development" {
		t.Errorf("Environment default = %q, want %q", cfg.Environment, "development")
	}
	if cfg.SwaggerHost != "" {
		t.Errorf("SwaggerHost default = %q, want empty string", cfg.SwaggerHost)
	}
}

// =============================================================================
// NewServiceConfig AllowedOrigins Parsing Tests
// =============================================================================

func TestNewServiceConfig_ParsesAllowedOrigins_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     []string
	}{
		{
			name:     "single origin",
			envValue: "http://localhost:3000",
			want:     []string{"http://localhost:3000"},
		},
		{
			name:     "multiple origins",
			envValue: "http://localhost:3000,https://example.com",
			want:     []string{"http://localhost:3000", "https://example.com"},
		},
		{
			name:     "trims whitespace",
			envValue: "http://localhost:3000 , https://example.com ",
			want:     []string{"http://localhost:3000", "https://example.com"},
		},
		{
			name:     "handles three origins",
			envValue: "http://a.com,http://b.com,http://c.com",
			want:     []string{"http://a.com", "http://b.com", "http://c.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearAllServiceEnvVars(t)
			setEnvForTest(t, "ALLOWED_ORIGINS", tt.envValue)

			cfg := NewServiceConfig()

			if len(cfg.AllowedOrigins) != len(tt.want) {
				t.Fatalf("AllowedOrigins length = %d, want %d", len(cfg.AllowedOrigins), len(tt.want))
			}
			for i := range cfg.AllowedOrigins {
				if cfg.AllowedOrigins[i] != tt.want[i] {
					t.Errorf("AllowedOrigins[%d] = %q, want %q", i, cfg.AllowedOrigins[i], tt.want[i])
				}
			}
		})
	}
}

// =============================================================================
// NewServiceConfig Validation Tests
// =============================================================================

func TestNewServiceConfig_ValidatesEnvironment_TableDriven(t *testing.T) {
	validEnvironments := []string{"development", "staging", "production"}

	for _, env := range validEnvironments {
		t.Run("valid_"+env, func(t *testing.T) {
			clearAllServiceEnvVars(t)
			setEnvForTest(t, "ALLOWED_ORIGINS", "http://localhost:3000")
			setEnvForTest(t, "ENVIRONMENT", env)

			// Should not panic
			cfg := NewServiceConfig()

			if cfg.Environment != env {
				t.Errorf("Environment = %q, want %q", cfg.Environment, env)
			}
		})
	}
}

func TestNewServiceConfig_PanicsOnInvalidEnvironment(t *testing.T) {
	clearAllServiceEnvVars(t)

	setEnvForTest(t, "ALLOWED_ORIGINS", "http://localhost:3000")
	setEnvForTest(t, "ENVIRONMENT", "invalid-env")

	defer func() {
		if r := recover(); r == nil {
			t.Error("NewServiceConfig() should panic for invalid environment")
		}
	}()

	NewServiceConfig()
}
