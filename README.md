# Task Management REST API

A scalable Task Management Service built with Go and Fiber v2, featuring JWT authentication, background workers, and clean architecture.

## Features

- ✅ RESTful API with CRUD operations
- 🔐 JWT-based authentication & authorization
- 👥 User and Admin role-based access control
- ⚡ Concurrent background workers for auto-completion
- 🗄️ PostgreSQL persistence with repository pattern
- 🏗️ Clean architecture (handlers, services, repositories)
- 🐳 Docker and Docker Compose support
- 📊 Pagination and filtering
- 🛡️ Input validation and error handling
- 🔄 Graceful shutdown with context

## Architecture

```
task-management-api/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── domain/         # Domain models
│   ├── handler/        # HTTP handlers
│   ├── middleware/     # Authentication middleware
│   ├── repository/     # Data access layer
│   ├── service/        # Business logic
│   └── util/           # Utility functions
├── pkg/database/       # Database connection
├── Dockerfile
└── docker-compose.yml
```

## API Endpoints

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/auth/register` | Register new user | No |
| POST | `/auth/login` | Login user | No |

### Tasks

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/tasks` | Create task | Yes |
| GET | `/tasks` | List tasks (filtered) | Yes |
| GET | `/tasks/:id` | Get task by ID | Yes |
| PUT | `/tasks/:id` | Update task | Yes |
| DELETE | `/tasks/:id` | Delete task | Yes |

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone <repo-url>
cd task-management-api

# Start services
docker-compose up -d

# Check logs
docker-compose logs -f app
```

The API will be available at `http://localhost:3000`

### Manual Setup

#### Prerequisites

- Go 1.21 or higher
- PostgreSQL 15+

#### Installation

```bash
# Install dependencies
go mod download

# Set up environment variables
cp .env.example .env
# Edit .env with your database credentials

# Run the application
go run cmd/server/main.go
```

## Configuration

Environment variables (`.env`):

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=taskdb
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-secret-key-change-this
JWT_EXPIRY_HOURS=24

# Server
SERVER_PORT=3000

# Worker
AUTO_COMPLETE_MINUTES=5
```

## Usage Examples

### 1. Register a User

```bash
curl -X POST http://localhost:3000/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### 2. Register an Admin

```bash
curl -X POST http://localhost:3000/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123",
    "role": "admin"
  }'
```

### 3. Login

```bash
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

Response:
```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "role": "user",
      "created_at": "2025-01-22T10:00:00Z",
      "updated_at": "2025-01-22T10:00:00Z"
    }
  }
}
```

### 4. Create a Task

```bash
curl -X POST http://localhost:3000/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Complete project documentation",
    "description": "Write comprehensive API documentation"
  }'
```

### 5. List Tasks

```bash
# List all tasks (user sees only their tasks)
curl http://localhost:3000/tasks \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Filter by status
curl "http://localhost:3000/tasks?status=pending" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Pagination
curl "http://localhost:3000/tasks?limit=10&offset=0" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 6. Get Task by ID

```bash
curl http://localhost:3000/tasks/TASK_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 7. Update Task

```bash
curl -X PUT http://localhost:3000/tasks/TASK_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Updated title",
    "status": "in_progress"
  }'
```

### 8. Delete Task

```bash
curl -X DELETE http://localhost:3000/tasks/TASK_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Authorization Rules

- **Regular Users**: Can only access their own tasks
- **Admin Users**: Can access all tasks across all users

## Background Worker

The application includes a concurrent background worker that:

- Automatically marks tasks as `completed` after X minutes (configurable via `AUTO_COMPLETE_MINUTES`)
- Uses goroutines and channels for concurrent processing
- Implements thread-safe access to shared resources
- Runs periodic scans to catch any missed tasks
- Does not block API requests

### Worker Features

- **Task Queue**: Buffered channel for task IDs
- **Multiple Workers**: 5 concurrent worker goroutines
- **Scanner**: Periodic background scanner (1-minute interval)
- **Auto-completion Logic**:
  - Only completes tasks in `pending` or `in_progress` status
  - Skips tasks already completed or deleted
  - Thread-safe with sync.Map for tracking processed tasks

## Error Handling

The API returns consistent JSON error responses:

```json
{
  "error": "Bad Request",
  "message": "Detailed error message"
}
```

### HTTP Status Codes

- `200` - Success
- `201` - Created
- `204` - No Content (delete success)
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error

## Testing

### Health Check

```bash
curl http://localhost:3000/health
```

### Complete Workflow Test

```bash
# 1. Register
TOKEN=$(curl -s -X POST http://localhost:3000/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test123"}' \
  | jq -r '.data.token')

# 2. Login
TOKEN=$(curl -s -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test123"}' \
  | jq -r '.data.token')

# 3. Create task
TASK_ID=$(curl -s -X POST http://localhost:3000/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Test Task","description":"Test"}' \
  | jq -r '.data.id')

# 4. List tasks
curl -s http://localhost:3000/tasks \
  -H "Authorization: Bearer $TOKEN" | jq

# 5. Update task
curl -s -X PUT http://localhost:3000/tasks/$TASK_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"status":"completed"}' | jq

# 6. Delete task
curl -s -X DELETE http://localhost:3000/tasks/$TASK_ID \
  -H "Authorization: Bearer $TOKEN"
```

## Project Structure Details

### Domain Models

- **User**: ID, Email, Password (hashed), Role, Timestamps
- **Task**: ID, UserID, Title, Description, Status, Timestamps

### Task Statuses

- `pending` - Task is not started
- `in_progress` - Task is being worked on
- `completed` - Task is finished

### Database Schema

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
```

## Security Features

- Password hashing with bcrypt
- JWT token-based authentication
- Token expiration
- Role-based authorization
- SQL injection protection via parameterized queries
- CORS configuration

## Graceful Shutdown

The application supports graceful shutdown:

- Handles OS signals (SIGINT, SIGTERM)
- Cancels background worker context
- Waits for in-flight requests (10s timeout)
- Closes database connections

## Development

### Running Locally

```bash
# Install Air for hot reload
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

### Building

```bash
# Build binary
go build -o bin/server cmd/server/main.go

# Run binary
./bin/server
```

## Production Considerations

- Change `JWT_SECRET` to a strong random value
- Use environment variables for sensitive config
- Enable HTTPS/TLS
- Implement rate limiting
- Add request logging
- Monitor worker queue depth
- Set up database backups
- Use connection pooling
- Add metrics and monitoring

## License

MIT