package utils

import (
	"os"
	"testing"
	"time"
)

// =============================================================================
// Test Helpers
// =============================================================================

// setEnvForTest sets an environment variable and registers cleanup.
func setEnvForTest(t *testing.T, key, value string) {
	t.Helper()
	t.Setenv(key, value)
}

// clearEnvForTest ensures an environment variable is unset.
func clearEnvForTest(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "")
	os.Unsetenv(key) //nolint:errcheck // test cleanup
}

// =============================================================================
// GetEnv Tests
// =============================================================================

func TestGetEnv_ReturnsValueWhenSet(t *testing.T) {
	setEnvForTest(t, "TEST_GET_ENV_VALUE", "custom-value")

	got := GetEnv("TEST_GET_ENV_VALUE", "default")

	if got != "custom-value" {
		t.Errorf("GetEnv() = %q, want %q", got, "custom-value")
	}
}

func TestGetEnv_ReturnsDefaultWhenNotSet(t *testing.T) {
	clearEnvForTest(t, "TEST_GET_ENV_NOT_SET")

	got := GetEnv("TEST_GET_ENV_NOT_SET", "default-value")

	if got != "default-value" {
		t.Errorf("GetEnv() = %q, want %q", got, "default-value")
	}
}

func TestGetEnv_ReturnsDefaultForEmptyString(t *testing.T) {
	setEnvForTest(t, "TEST_GET_ENV_EMPTY", "")

	got := GetEnv("TEST_GET_ENV_EMPTY", "default")

	if got != "default" {
		t.Errorf("GetEnv() = %q, want %q", got, "default")
	}
}

// =============================================================================
// GetEnvRequired Tests
// =============================================================================

func TestGetEnvRequired_ReturnsValueWhenSet(t *testing.T) {
	setEnvForTest(t, "TEST_REQUIRED_SET", "required-value")

	got := GetEnvRequired("TEST_REQUIRED_SET")

	if got != "required-value" {
		t.Errorf("GetEnvRequired() = %q, want %q", got, "required-value")
	}
}

func TestGetEnvRequired_PanicsWhenNotSet(t *testing.T) {
	clearEnvForTest(t, "TEST_REQUIRED_NOT_SET")

	defer func() {
		if r := recover(); r == nil {
			t.Error("GetEnvRequired() should panic for missing env var")
		}
	}()

	GetEnvRequired("TEST_REQUIRED_NOT_SET")
}

// =============================================================================
// GetEnvSlice Tests
// =============================================================================

func TestGetEnvSlice_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		setEnv       bool
		defaultValue []string
		want         []string
	}{
		{
			name:         "returns default when not set",
			setEnv:       false,
			defaultValue: []string{"a", "b"},
			want:         []string{"a", "b"},
		},
		{
			name:         "parses comma-separated values",
			envValue:     "one,two,three",
			setEnv:       true,
			defaultValue: []string{},
			want:         []string{"one", "two", "three"},
		},
		{
			name:         "handles single value",
			envValue:     "single",
			setEnv:       true,
			defaultValue: []string{},
			want:         []string{"single"},
		},
		{
			name:         "returns default for empty string",
			envValue:     "",
			setEnv:       true,
			defaultValue: []string{"default"},
			want:         []string{"default"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_SLICE"
			clearEnvForTest(t, key)

			if tt.setEnv {
				setEnvForTest(t, key, tt.envValue)
			}

			got := GetEnvSlice(key, tt.defaultValue)

			if len(got) != len(tt.want) {
				t.Fatalf("GetEnvSlice() length = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("GetEnvSlice()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// =============================================================================
// GetEnvInt Tests
// =============================================================================

func TestGetEnvInt_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		setEnv       bool
		defaultValue int
		want         int
	}{
		{
			name:         "returns default when not set",
			setEnv:       false,
			defaultValue: 42,
			want:         42,
		},
		{
			name:         "parses valid integer",
			envValue:     "123",
			setEnv:       true,
			defaultValue: 0,
			want:         123,
		},
		{
			name:         "returns default for invalid integer",
			envValue:     "not-a-number",
			setEnv:       true,
			defaultValue: 99,
			want:         99,
		},
		{
			name:         "handles negative numbers",
			envValue:     "-50",
			setEnv:       true,
			defaultValue: 0,
			want:         -50,
		},
		{
			name:         "handles zero",
			envValue:     "0",
			setEnv:       true,
			defaultValue: 100,
			want:         0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_INT"
			clearEnvForTest(t, key)

			if tt.setEnv {
				setEnvForTest(t, key, tt.envValue)
			}

			got := GetEnvInt(key, tt.defaultValue)

			if got != tt.want {
				t.Errorf("GetEnvInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

// =============================================================================
// GetEnvBool Tests
// =============================================================================

func TestGetEnvBool_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		setEnv       bool
		defaultValue bool
		want         bool
	}{
		{
			name:         "returns default when not set",
			setEnv:       false,
			defaultValue: true,
			want:         true,
		},
		{
			name:         "parses true",
			envValue:     "true",
			setEnv:       true,
			defaultValue: false,
			want:         true,
		},
		{
			name:         "parses false",
			envValue:     "false",
			setEnv:       true,
			defaultValue: true,
			want:         false,
		},
		{
			name:         "parses 1 as true",
			envValue:     "1",
			setEnv:       true,
			defaultValue: false,
			want:         true,
		},
		{
			name:         "parses 0 as false",
			envValue:     "0",
			setEnv:       true,
			defaultValue: true,
			want:         false,
		},
		{
			name:         "returns default for invalid bool",
			envValue:     "invalid",
			setEnv:       true,
			defaultValue: true,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_BOOL"
			clearEnvForTest(t, key)

			if tt.setEnv {
				setEnvForTest(t, key, tt.envValue)
			}

			got := GetEnvBool(key, tt.defaultValue)

			if got != tt.want {
				t.Errorf("GetEnvBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// GetEnvDuration Tests
// =============================================================================

func TestGetEnvDuration_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		setEnv       bool
		defaultValue time.Duration
		want         time.Duration
	}{
		{
			name:         "returns default when not set",
			setEnv:       false,
			defaultValue: 5 * time.Minute,
			want:         5 * time.Minute,
		},
		{
			name:         "parses minutes",
			envValue:     "15m",
			setEnv:       true,
			defaultValue: 0,
			want:         15 * time.Minute,
		},
		{
			name:         "parses hours",
			envValue:     "2h",
			setEnv:       true,
			defaultValue: 0,
			want:         2 * time.Hour,
		},
		{
			name:         "parses seconds",
			envValue:     "30s",
			setEnv:       true,
			defaultValue: 0,
			want:         30 * time.Second,
		},
		{
			name:         "parses complex duration",
			envValue:     "1h30m",
			setEnv:       true,
			defaultValue: 0,
			want:         90 * time.Minute,
		},
		{
			name:         "returns default for invalid duration",
			envValue:     "invalid",
			setEnv:       true,
			defaultValue: 10 * time.Second,
			want:         10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_DUR"
			clearEnvForTest(t, key)

			if tt.setEnv {
				setEnvForTest(t, key, tt.envValue)
			}

			got := GetEnvDuration(key, tt.defaultValue)

			if got != tt.want {
				t.Errorf("GetEnvDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
