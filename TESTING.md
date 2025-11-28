# Testing Guide

## Overview

The template-api uses Go's standard `testing` package for unit tests.

## Quick Commands

```bash
# Run all tests
task test

# Or directly with Go
go test -v ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific test
go test -v -run TestNewDatabaseConfig ./internal/config/

# Run all config tests
go test -v ./internal/config/
```

## Test Files

**`internal/config/database_test.go`** - 8 tests

- DSN formatting (2)
- Environment loading (2)
- Default values (1)
- Required field validation (4) - panics on missing host/user/password/name

**`internal/config/jwt_test.go`** - 5 tests

- Nil when secret not set (1)
- Environment loading (1)
- Default values (1)
- Short secret validation (1)
- HasJWT helper (3 sub-tests)

**`internal/config/service_test.go`** - 5 tests

- Environment loading (1)
- Default values (1)
- AllowedOrigins parsing (4 sub-tests)
- Environment validation (3 valid + 1 invalid)

**`internal/utils/env_test.go`** - 10 tests

- GetEnv (3)
- GetEnvRequired (2)
- GetEnvSlice (4 sub-tests)
- GetEnvInt (5 sub-tests)
- GetEnvBool (6 sub-tests)
- GetEnvDuration (6 sub-tests)

## Key Testing Patterns

**Table-driven tests**: Multiple scenarios with `tests := []struct{...}`

```go
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
        // ... more cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

**Environment cleanup**: Use `t.Cleanup()` for automatic teardown

```go
func setEnvForTest(t *testing.T, key, value string) {
    t.Helper()
    os.Setenv(key, value)
    t.Cleanup(func() { os.Unsetenv(key) })
}
```

**Panic testing**: For validation that should panic

```go
func TestNewDatabaseConfig_PanicsOnMissingHost(t *testing.T) {
    defer func() {
        if r := recover(); r == nil {
            t.Error("should panic when DB_HOST is missing")
        }
    }()

    NewDatabaseConfig()
}
```

**Section markers**: Organize tests by function

```go
// =============================================================================
// NewDatabaseConfig Tests
// =============================================================================
```

## Test Constants

```go
testJWTSecret = "this-is-a-very-long-secret-key-for-testing-at-least-32-chars"
```

## Contributing Tests

1. Follow naming: `Test<FunctionName>_<Scenario>`
2. Organize by function with section markers
3. Use table-driven tests for multiple scenarios
4. Use `t.Helper()` in test helper functions
5. Clean up resources with `t.Cleanup()`
6. Verify: `task ci:all` or `go test -cover ./...`
