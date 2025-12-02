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

Tests are organized by source file, following Go conventions:

| File | Tests | Description |
|------|-------|-------------|
| `handler_test.go` | 1 | Constructor |
| `health_test.go` | 1 | Health check endpoint |
| `contact_message_test.go` | 20 | Contact message CRUD + edge cases |
| `recipient_test.go` | 20 | Recipient CRUD operations |
| `mocks_test.go` | - | Mocks and shared test utilities |

### Test Breakdown

**`handler_test.go`** (1 test)

- Handler constructor initialization

**`health_test.go`** (1 test)

- Health check endpoint success

**`contact_message_test.go`** (20 tests)

- CreateContactMessage: success, spam, publish error, validation, repo error, JSON
- CreateContactMessage edge cases: Unicode, max length, empty honeypot, malicious
- GetContactMessages: success, empty list, repository error
- GetContactMessage: success, invalid ID, not found, repository error
- Context propagation tests
- Invalid ID format validation (table-driven)

**`recipient_test.go`** (20 tests)

- GetRecipients: success, empty list, repository error
- GetRecipient: success, invalid ID, not found, repository error
- CreateRecipient: success, isActive=false, validation, email, repo error, JSON
- UpdateRecipient: success, partial, invalid ID, not found, validation, repo error
- DeleteRecipient: success, invalid ID, not found, repository error
- Context propagation tests
- Invalid ID format validation (table-driven)

**`mocks_test.go`** (shared utilities)

- `mockRepository` - configurable mock implementing Repository interface
- `mockPublisher` - configurable mock implementing Publisher interface
- `setupTestRouter()` - creates Gin router in test mode
- `performRequest()` - executes HTTP request against router
- `createTestContactMessage()` / `createTestContactMessages()` - factory functions
- `createTestRecipient()` / `createTestRecipients()` - factory functions
- `ctxKey` - context key type for propagation tests

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

**Mock Publisher**: Verifies queue integration

```go
mockPub := &mockPublisher{
    publishFunc: func(ctx context.Context, msg interface{}) error {
        // Capture published message for assertions
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

**Table-driven tests**: Multiple scenarios with `tests := []struct{...}`

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
- Queue publish is called with correct message event

### Error Cases

- Repository errors (500)
- Not found errors (404)
- Invalid ID format (400)
- Validation errors (400)
- Invalid email format (400)
- Invalid JSON (400)
- Queue publish failure doesn't fail request

### Edge Cases

- Unicode and special characters (Chinese, emoji, umlauts)
- Max length boundary validation (name: 255, subject: 500, message: 10000)
- Empty honeypot is not detected as spam
- Malicious content stored as-is (SQL injection, XSS, HTML, null bytes)

## API Characteristics

Messaging-api handles contact form and recipient operations:

- **Contact Form**: Public endpoint for submissions with honeypot spam detection
- **Messages**: Admin-only read access (GetAll, GetByID)
- **Recipients**: Admin-only CRUD for email notification recipients
- **Queue Integration**: Messages published to RabbitMQ for async processing

## Spam Protection

The honeypot test (`TestCreateContactMessage_SpamDetected`) verifies:

- Messages with honeypot field filled return 201 (to not alert bots)
- But the message is NOT saved to the database
- And the message is NOT published to the queue

```go
// Honeypot field (website) is filled - indicates spam
body := `{"name":"Bot","email":"bot@spam.com","subject":"Spam","message":"Buy now!","website":"http://spam.com"}`
// Should return success but NOT save
```

## Security Testing

The `TestCreateContactMessage_MaliciousContent` test verifies:

- SQL injection attempts are stored safely (GORM parameterizes queries)
- XSS payloads are stored as-is (escaping happens at render time)
- HTML content is preserved for legitimate use cases
- Control characters and null bytes are handled

## Contributing Tests

1. Add tests to the appropriate file matching the source file being tested
2. Follow naming: `Test<HandlerName>_<Scenario>`
3. Use table-driven tests for multiple scenarios
4. Mock only the repository/publisher methods needed
5. Verify: `go test -cover ./internal/handlers/`
