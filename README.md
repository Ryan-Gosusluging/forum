# Golang Basic Forum

A forum application built with Go, implementing microservices architecture and real-time chat functionality.

## Project Structure

```
.
├── cmd/                    # Application entry points
│   ├── auth-service/      # Authentication service
│   └── forum-service/     # Main forum service
├── internal/              # Private application code
│   ├── auth/             # Authentication logic
│   ├── chat/             # WebSocket chat implementation
│   ├── forum/            # Forum business logic
│   └── storage/          # Database interactions
├── pkg/                   # Public packages
│   ├── config/           # Configuration management
│   ├── logger/           # Logging utilities
│   └── proto/            # gRPC protocol definitions
├── migrations/            # Database migrations
└── docs/                 # API documentation
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Protocol Buffers compiler (protoc)
- protoc-gen-go and protoc-gen-go-grpc plugins

### Installing Protocol Buffers

#### Windows
1. Download the latest release from https://github.com/protocolbuffers/protobuf/releases
2. Extract the zip file
3. Add the bin directory to your PATH

#### Linux
```bash
sudo apt-get install protobuf-compiler
```

### Installing Go plugins
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Getting Started

1. Clone the repository
2. Install dependencies: `go mod download`
3. Generate gRPC code: `./scripts/generate_proto.sh`
4. Set up PostgreSQL database
5. Run migrations: `go run cmd/migrate/main.go`
6. Start services:
   - `go run cmd/auth-service/main.go`
   - `go run cmd/forum-service/main.go`

## Configuration

The application can be configured using environment variables:

- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432)
- `DB_USER`: Database user (default: postgres)
- `DB_PASSWORD`: Database password (default: postgres)
- `DB_NAME`: Database name (default: forum)
- `AUTH_SERVICE_PORT`: Auth service port (default: 50051)
- `FORUM_SERVICE_PORT`: Forum service port (default: 8080)
- `JWT_SECRET`: JWT secret key
- `CHAT_MESSAGE_TTL`: Chat message time-to-live (default: 24h)

## Features

- User authentication and authorization
- Real-time chat with WebSocket
- Forum posts and comments
- Admin panel
- gRPC communication between services
- PostgreSQL database
- Automated database migrations
- Comprehensive testing
- API documentation with Swagger

## API Documentation

API documentation is available at `/swagger/index.html` when running the services.

## Testing

Run tests with:
```bash
go test ./... -cover
``` 