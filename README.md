# Social BE API

A Go-based backend API for social volunteer management system.

## Features

- RESTful API with Gin framework
- PostgreSQL database with GORM
- JWT authentication
- Redis caching
- OpenTelemetry tracing
- Prometheus metrics
- Database migrations
- Docker containerization
- Comprehensive testing

## API Documentation

The API provides multiple documentation interfaces for better developer experience:

### Swagger UI
Interactive API documentation with testing capabilities.
- **URL**: `http://localhost:8080/swagger/index.html`
- Features: Try out API endpoints directly from the browser

### Apidog Integration
Enhanced API design and testing with Apidog.
- **Documentation Page**: `http://localhost:8080/docs`
- **OpenAPI JSON**: `http://localhost:8080/docs/openapi.json`
- **OpenAPI YAML**: `http://localhost:8080/docs/openapi.yaml`

#### How to Import into Apidog:
1. Open Apidog application
2. Click "Import" or "New Project from OpenAPI"
3. Paste the JSON or YAML URL from above
4. Click "Import"

**Note**: The OpenAPI specification auto-refreshes on every build.

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.22+ (for local development)

### Running with Docker
```bash
# Clone the repository
git clone <repository-url>
cd social-be

# Start all services
./run.sh
```

This will:
1. Run unit tests
2. Build the Docker image
3. Start all services (app, postgres, redis, prometheus, grafana, jaeger)

### Local Development
```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Generate swagger docs
go run github.com/swaggo/swag/cmd/swag@v1.16.2 init -g cmd/main.go

# Run the application
go run cmd/main.go
```

## API Endpoints

### Authentication
- `POST /api/register` - Register new user
- `POST /api/login` - Login with email and password
- `POST /api/refresh` - Refresh access token

### Protected Endpoints (require Bearer token)
- `GET /api/users` - Get all users
- `GET /api/users/:id` - Get user by ID
- `GET /api/volunteers` - Get all volunteers
- `GET /api/volunteers/:id` - Get volunteer by ID
- `POST /api/volunteers` - Create volunteer
- `PUT /api/volunteers/:id` - Update volunteer
- `DELETE /api/volunteers/:id` - Delete volunteer

## Architecture

```
cmd/
├── main.go                 # Application entry point

internal/
├── config/                 # Configuration management
├── domain/                 # Business logic layers
│   ├── auth/              # Authentication domain
│   ├── user/              # User management
│   ├── volunteer/         # Volunteer management
│   └── ...                # Other domains
├── middleware/            # HTTP middlewares
└── pkg/                   # Shared packages
    ├── apperror/          # Error handling
    ├── cache/             # Redis cache
    ├── logger/            # Logging
    ├── response/          # HTTP responses
    ├── validation/        # Request validation
    └── ...

docs/                       # Generated API documentation
migrations/                 # Database migrations
```

## Testing

Run all tests:
```bash
go test ./...
```

Tests are automatically run during Docker build to ensure code quality.

## Monitoring

- **Metrics**: `http://localhost:8080/metrics` (Prometheus format)
- **Health Checks**:
  - `http://localhost:8080/health/live` - Liveness probe
  - `http://localhost:8080/health/ready` - Readiness probe
- **Tracing**: Jaeger UI at `http://localhost:16686`
- **Monitoring**: Grafana at `http://localhost:3000`

## Development

### Adding New API Endpoints
1. Create domain layer (entity, dto, repository, service, handler)
2. Add routes in `cmd/main.go`
3. Update swagger comments
4. Run `go run github.com/swaggo/swag/cmd/swag@v1.16.2 init -g cmd/main.go` to regenerate docs
5. Add tests

### Database Migrations
```bash
# Create new migration
migrate create -ext sql -dir migrations -seq add_new_table

# Apply migrations (handled automatically in Docker)
migrate -path migrations -database "postgres://..." up
```

## License

[Add your license here]