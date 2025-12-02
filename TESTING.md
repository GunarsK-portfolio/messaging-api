# Testing Guide

## Overview

The messaging-api uses Go's standard `testing` package with httptest for handler
unit tests. This service handles contact form submissions and recipient management.

## Quick Commands

```bash
# Run all tests
go test ./internal/handlers/

# Run with coverage
go test -cover ./internal/handlers/

# Generate coverage report
go test -coverprofile=coverage.out ./internal/handlers/
go tool cover -html=coverage.out -o coverage.html

# Run specific test
go test -v -run TestCreateContactMessage_Success ./internal/handlers/

# Run all Contact Message tests
go test -v -run ContactMessage ./internal/handlers/

# Run all Recipient tests
go test -v -run Recipient ./internal/handlers/
```

## Test Files

**`handlers/`** - 42 tests

| Category | Tests | Coverage |
|----------|-------|----------|
| Constructor | 1 | Handler initialization |
| Health Check | 1 | HealthCheck endpoint |
| Contact Messages | 20 | Create, GetAll, GetByID + edge cases |
| Recipients | 20 | GetAll, GetByID, Create, Update, Delete + errors |

Tests are split by source file: `handler_test.go`, `health_test.go`,
`contact_message_test.go`, `recipient_test.go`, `mocks_test.go`.

## Key Testing Patterns

**Mock Repository**: Function fields allow per-test behavior customization

```go
mockRepo := &mockRepository{
    createContactMessageFunc: func(
        ctx context.Context,
        msg *models.ContactMessage,
    ) error {
        return nil
    },
}
```

**HTTP Testing**: Uses `httptest.ResponseRecorder` with Gin router

```go
body := `{"name":"John","email":"john@example.com","subject":"Test","message":"Hello"}`
w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))
if w.Code != http.StatusCreated { ... }
```

**Test Helpers**: Factory functions for consistent test data

```go
msg := createTestContactMessage()
recipient := createTestRecipient()
```

## Test Categories

### Success Cases

- Returns expected data
- Sets correct HTTP status
- Sets Location header on create
- Spam detection returns success but doesn't save

### Error Cases

- Repository errors (500)
- Not found errors (404)
- Invalid ID format (400)
- Validation errors (400)
- Invalid email format (400)
- Invalid JSON (400)

### Edge Cases

- Unicode and special characters
- Max length boundary validation
- Malicious content (SQL injection, XSS, HTML)

## API Characteristics

Messaging-api handles contact form and recipient operations:

- **Contact Form**: Public endpoint for submissions with honeypot spam detection
- **Messages**: Admin-only read access (GetAll, GetByID)
- **Recipients**: Admin-only CRUD for email notification recipients

## Spam Protection

The honeypot test (`TestCreateContactMessage_SpamDetected`) verifies:

- Messages with honeypot field filled return 201 (to not alert bots)
- But the message is NOT saved to the database

```go
// Honeypot field (website) is filled - indicates spam
body := `{"name":"Bot","email":"bot@spam.com","subject":"Spam","message":"Buy now!","website":"http://spam.com"}`
// Should return success but NOT save
```

## Contributing Tests

1. Follow naming: `Test<HandlerName>_<Scenario>`
2. Organize by entity with section markers
3. Mock only the repository methods needed
4. Verify: `go test -cover ./internal/handlers/`
