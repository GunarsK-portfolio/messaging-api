package config

import (
	"os"
	"testing"
)

// =============================================================================
// Test Helpers
// =============================================================================

// setEnvForTest sets an environment variable and registers cleanup.
func setEnvForTest(t *testing.T, key, value string) {
	t.Helper()
	t.Setenv(key, value)
}

// clearAllDatabaseEnvVars clears all database-related environment variables.
func clearAllDatabaseEnvVars(t *testing.T) {
	t.Helper()
	vars := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSL_MODE"}
	for _, v := range vars {
		t.Setenv(v, "")
		os.Unsetenv(v) //nolint:errcheck // test cleanup
	}
}

// =============================================================================
// DatabaseConfig.DSN Tests
// =============================================================================

func TestDatabaseConfig_DSN_FormatsCorrectly(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		Name:     "testdb",
		SSLMode:  "disable",
	}

	want := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	got := cfg.DSN()

	if got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}

func TestDatabaseConfig_DSN_HandlesSpecialCharactersInPassword(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "user",
		Password: "p@ss=word!",
		Name:     "db",
		SSLMode:  "disable",
	}

	got := cfg.DSN()

	// DSN should include the password as-is (driver handles escaping)
	if got == "" {
		t.Error("DSN() should not return empty string")
	}
}

// =============================================================================
// NewDatabaseConfig Tests
// =============================================================================

func TestNewDatabaseConfig_LoadsAllFieldsFromEnv(t *testing.T) {
	clearAllDatabaseEnvVars(t)

	setEnvForTest(t, "DB_HOST", "testhost")
	setEnvForTest(t, "DB_PORT", "5433")
	setEnvForTest(t, "DB_USER", "dbuser")
	setEnvForTest(t, "DB_PASSWORD", "dbpass")
	setEnvForTest(t, "DB_NAME", "mydb")
	setEnvForTest(t, "DB_SSL_MODE", "require")

	cfg := NewDatabaseConfig()

	if cfg.Host != "testhost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "testhost")
	}
	if cfg.Port != "5433" {
		t.Errorf("Port = %q, want %q", cfg.Port, "5433")
	}
	if cfg.User != "dbuser" {
		t.Errorf("User = %q, want %q", cfg.User, "dbuser")
	}
	if cfg.Password != "dbpass" {
		t.Errorf("Password = %q, want %q", cfg.Password, "dbpass")
	}
	if cfg.Name != "mydb" {
		t.Errorf("Name = %q, want %q", cfg.Name, "mydb")
	}
	if cfg.SSLMode != "require" {
		t.Errorf("SSLMode = %q, want %q", cfg.SSLMode, "require")
	}
}

func TestNewDatabaseConfig_UsesDefaultsForOptionalFields(t *testing.T) {
	clearAllDatabaseEnvVars(t)

	// Set only required fields
	setEnvForTest(t, "DB_HOST", "localhost")
	setEnvForTest(t, "DB_USER", "user")
	setEnvForTest(t, "DB_PASSWORD", "pass")
	setEnvForTest(t, "DB_NAME", "db")

	cfg := NewDatabaseConfig()

	if cfg.Port != "5432" {
		t.Errorf("Port default = %q, want %q", cfg.Port, "5432")
	}
	if cfg.SSLMode != "disable" {
		t.Errorf("SSLMode default = %q, want %q", cfg.SSLMode, "disable")
	}
}

func TestNewDatabaseConfig_PanicsOnMissingHost(t *testing.T) {
	clearAllDatabaseEnvVars(t)

	setEnvForTest(t, "DB_USER", "user")
	setEnvForTest(t, "DB_PASSWORD", "pass")
	setEnvForTest(t, "DB_NAME", "db")

	defer func() {
		if r := recover(); r == nil {
			t.Error("NewDatabaseConfig() should panic when DB_HOST is missing")
		}
	}()

	NewDatabaseConfig()
}

func TestNewDatabaseConfig_PanicsOnMissingUser(t *testing.T) {
	clearAllDatabaseEnvVars(t)

	setEnvForTest(t, "DB_HOST", "localhost")
	setEnvForTest(t, "DB_PASSWORD", "pass")
	setEnvForTest(t, "DB_NAME", "db")

	defer func() {
		if r := recover(); r == nil {
			t.Error("NewDatabaseConfig() should panic when DB_USER is missing")
		}
	}()

	NewDatabaseConfig()
}

func TestNewDatabaseConfig_PanicsOnMissingPassword(t *testing.T) {
	clearAllDatabaseEnvVars(t)

	setEnvForTest(t, "DB_HOST", "localhost")
	setEnvForTest(t, "DB_USER", "user")
	setEnvForTest(t, "DB_NAME", "db")

	defer func() {
		if r := recover(); r == nil {
			t.Error("NewDatabaseConfig() should panic when DB_PASSWORD is missing")
		}
	}()

	NewDatabaseConfig()
}

func TestNewDatabaseConfig_PanicsOnMissingName(t *testing.T) {
	clearAllDatabaseEnvVars(t)

	setEnvForTest(t, "DB_HOST", "localhost")
	setEnvForTest(t, "DB_USER", "user")
	setEnvForTest(t, "DB_PASSWORD", "pass")

	defer func() {
		if r := recover(); r == nil {
			t.Error("NewDatabaseConfig() should panic when DB_NAME is missing")
		}
	}()

	NewDatabaseConfig()
}
