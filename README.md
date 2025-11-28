# Template API

A Go REST API template using Gin framework with PostgreSQL.

## Features

- RESTful API with Gin framework
- PostgreSQL with GORM ORM
- Structured logging with slog
- Prometheus metrics
- Swagger/OpenAPI documentation
- Docker support with multi-stage builds
- CI/CD with GitHub Actions
- Security scanning (govulncheck, gosec, Trivy)
- Unit tests with table-driven patterns

## Getting Started

### Prerequisites

- Go 1.25+
- PostgreSQL 15+
- [Task](https://taskfile.dev/) (optional, for task runner)

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/GunarsK-templates/template-api.git
   cd template-api
   ```

2. Copy environment file:

   ```bash
   cp .env.example .env
   ```

3. Edit `.env` with your configuration

4. Install dependencies:

   ```bash
   go mod download
   ```

5. Run the service:

   ```bash
   go run cmd/api/main.go
   # or with Task
   task run
   ```

### Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_NAME` | Service identifier | `your-service` |
| `PORT` | HTTP port | `8080` |
| `ENVIRONMENT` | Environment (development/staging/production) | `development` |
| `DB_HOST` | PostgreSQL host | - |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | - |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | - |
| `DB_SSL_MODE` | SSL mode | `disable` |
| `JWT_SECRET` | JWT signing secret (optional) | - |
| `ALLOWED_ORIGINS` | CORS allowed origins (comma-separated) | `localhost:3000` |
| `SWAGGER_HOST` | Swagger host for docs | - |

## Project Structure

```text
.
├── cmd/
│   └── api/
│       └── main.go          # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go        # Main config (combines sub-configs)
│   │   ├── service.go       # Service configuration
│   │   ├── database.go      # Database configuration
│   │   ├── jwt.go           # JWT configuration (optional)
│   │   └── *_test.go        # Unit tests
│   ├── handlers/
│   │   ├── handler.go       # Handler struct and dependencies
│   │   ├── health.go        # Health check endpoint
│   │   ├── errors.go        # Error handling utilities
│   │   └── example.go       # Example CRUD handlers
│   ├── models/
│   │   └── item.go          # Data models
│   ├── repository/
│   │   ├── repository.go    # Repository interface and DB setup
│   │   ├── item.go          # Item repository implementation
│   │   └── errors.go        # Repository errors
│   ├── routes/
│   │   └── routes.go        # Route definitions
│   └── utils/
│       ├── env.go           # Environment variable helpers
│       └── env_test.go      # Unit tests
├── docs/                    # Swagger documentation (generated)
├── Dockerfile
├── Taskfile.yml
├── TESTING.md               # Testing guide
├── go.mod
└── README.md
```

## API Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/health` | Health check | No |
| GET | `/metrics` | Prometheus metrics | No |
| GET | `/api/v1/items` | List all items | No |
| GET | `/api/v1/items/:id` | Get item by ID | No |
| POST | `/api/v1/items` | Create item | Optional |
| PUT | `/api/v1/items/:id` | Update item | Optional |
| DELETE | `/api/v1/items/:id` | Delete item | Optional |

## Development

### Available Tasks

```bash
# Run locally
task run

# Build binary
task build

# Run tests
task test
task test:coverage

# Code quality
task lint
task format
task vet
task tidy
task lint:markdown

# Security
task security:scan
task security:vuln

# Generate Swagger docs
task dev:swagger

# Docker
task docker:build
task docker:run
task docker:stop
task docker:logs

# Run all CI checks
task ci:all

# Install dev tools
task dev:install-tools

# Clean build artifacts
task clean
```

### Generating Swagger Documentation

```bash
task dev:swagger
```

Then access Swagger UI at `http://localhost:8080/swagger/index.html`
(requires `SWAGGER_HOST` to be set).

## Testing

See [TESTING.md](TESTING.md) for testing guide.

```bash
# Run all tests
task test

# Run with coverage
go test -cover ./...

# Run specific tests
go test -v -run TestNewDatabaseConfig ./internal/config/
```

## Docker

### Build

```bash
docker build -t template-api:latest .
```

### Run

```bash
docker run --rm -p 8080:8080 --env-file .env template-api:latest
```

## Customization

### Adding a New Resource

1. Create model in `internal/models/`:

   ```go
   type MyResource struct {
       ID        int64     `json:"id" gorm:"primaryKey"`
       Name      string    `json:"name" gorm:"size:200;not null"`
       CreatedAt time.Time `json:"created_at"`
       UpdatedAt time.Time `json:"updated_at"`
   }
   ```

2. Add repository methods in `internal/repository/`:

   ```go
   // In repository.go interface
   GetAllMyResources(ctx context.Context) ([]models.MyResource, error)
   // ... other methods

   // Create myresource.go with implementations
   ```

3. Add handlers in `internal/handlers/`:

   ```go
   func (h *Handler) GetMyResources(c *gin.Context) { ... }
   ```

4. Add routes in `internal/routes/routes.go`:

   ```go
   myresources := v1.Group("/myresources")
   {
       myresources.GET("", handler.GetMyResources)
       // ...
   }
   ```

5. Regenerate Swagger docs:

   ```bash
   task swagger
   ```

### Adding Authentication

1. Set `JWT_SECRET` in `.env`
2. Uncomment auth middleware in `internal/routes/routes.go`
3. Implement JWT validation middleware

## License

MIT
