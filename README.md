# Messaging API

![CI](https://github.com/GunarsK-portfolio/messaging-api/workflows/CI/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/GunarsK-portfolio/messaging-api)](https://goreportcard.com/report/github.com/GunarsK-portfolio/messaging-api)
[![codecov](https://codecov.io/gh/GunarsK-portfolio/messaging-api/graph/badge.svg)](https://codecov.io/gh/GunarsK-portfolio/messaging-api)
[![CodeRabbit](https://img.shields.io/coderabbit/prs/github/GunarsK-portfolio/messaging-api?label=CodeRabbit&color=2ea44f)](https://coderabbit.ai)

RESTful API for contact form submissions and recipient management.

## Features

- Public contact form submission endpoint with honeypot spam detection
- Admin-only contact message viewing
- Recipient management (CRUD) for email notifications
- JWT authentication for protected endpoints
- RESTful API with Swagger documentation
- Rate limiting via Traefik

## Tech Stack

- **Language**: Go 1.25.3
- **Framework**: Gin
- **Database**: PostgreSQL (GORM)
- **Common**: portfolio-common library (shared database utilities, auth middleware)
- **Auth**: JWT validation via auth-service
- **Documentation**: Swagger/OpenAPI

## Prerequisites

- Go 1.25+
- Node.js 22+ and npm 11+
- PostgreSQL (or use Docker Compose)
- auth-service running (for protected endpoints)

## Project Structure

```text
messaging-api/
├── cmd/
│   └── api/              # Application entrypoint
├── internal/
│   ├── config/           # Configuration
│   ├── handlers/         # HTTP handlers
│   ├── repository/       # Data access layer
│   └── routes/           # Route definitions
└── docs/                 # Swagger documentation
```

## Quick Start

### Using Docker Compose

```bash
docker-compose up -d
```

### Local Development

1. Copy environment file:

```bash
cp .env.example .env
```

1. Update `.env` with your configuration:

```env
PORT=8086
DB_HOST=localhost
DB_PORT=5432
DB_USER=portfolio_messaging
DB_PASSWORD=portfolio_messaging_dev_pass
DB_NAME=portfolio
JWT_SECRET=your-secret-key
ALLOWED_ORIGINS=http://localhost:3000
```

1. Start infrastructure (if not running):

```bash
# From infrastructure directory
docker-compose up -d postgres flyway
```

1. Run the service:

```bash
go run cmd/api/main.go
```

## Available Commands

Using Task:

```bash
# Development
task dev:swagger         # Generate Swagger documentation
task dev:install-tools   # Install dev tools (golangci-lint, govulncheck, etc.)

# Build and run
task build               # Build binary
task test                # Run tests
task test:coverage       # Run tests with coverage report
task clean               # Clean build artifacts

# Code quality
task format              # Format code with gofmt
task tidy                # Tidy and verify go.mod
task lint                # Run golangci-lint
task vet                 # Run go vet

# Security
task security:scan       # Run gosec security scanner
task security:vuln       # Check for vulnerabilities with govulncheck

# Docker
task docker:build        # Build Docker image
task docker:run          # Run service in Docker container
task docker:stop         # Stop running Docker container
task docker:logs         # View Docker container logs

# CI/CD
task ci:all              # Run all CI checks
```

Using Go directly:

```bash
go run cmd/api/main.go                          # Run
go build -o bin/messaging-api cmd/api/main.go   # Build
go test ./...                                   # Test
```

## API Endpoints

Base URL: `http://localhost:8086/api/v1`

### Health Check

- `GET /health` - Service health status

### Public Endpoints

No authentication required.

#### Contact

- `POST /contact` - Submit a contact message

### Protected Endpoints

All endpoints below require JWT authentication via
`Authorization: Bearer <token>` header.

#### Messages

- `GET /messages` - List all contact messages
- `GET /messages/:id` - Get contact message by ID

#### Recipients

- `GET /recipients` - List all recipients
- `GET /recipients/:id` - Get recipient by ID
- `POST /recipients` - Create recipient
- `PUT /recipients/:id` - Update recipient
- `DELETE /recipients/:id` - Delete recipient

## Swagger Documentation

When running, Swagger UI is available at:

- `http://localhost:8086/swagger/index.html`

(Requires `SWAGGER_HOST` environment variable to be set)

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8086` |
| `DB_HOST` | PostgreSQL host | - |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | - |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | - |
| `DB_SSLMODE` | PostgreSQL SSL mode | `disable` |
| `JWT_SECRET` | JWT signing secret | - |
| `ALLOWED_ORIGINS` | CORS allowed origins (comma-separated) | - |
| `SWAGGER_HOST` | Swagger host for docs | - |

## Authentication

This API validates JWT tokens issued by auth-service using the
portfolio-common auth middleware. The middleware:

1. Extracts token from `Authorization: Bearer <token>` header
2. Validates token signature and expiry
3. Injects user information into request context

The `/contact` endpoint is public (no auth required) to allow anonymous
contact form submissions. All other endpoints require authentication.

## Spam Protection

The contact form includes honeypot field detection. Messages with non-empty
honeypot fields are silently accepted but not saved, preventing bots from
knowing they've been detected.

## Integration

- **Public website**: Submits contact forms via `/contact` endpoint
- **Admin panel**: Views messages and manages recipients via protected endpoints
- **Future**: Message delivery worker will process pending messages

## License

[MIT](LICENSE)
