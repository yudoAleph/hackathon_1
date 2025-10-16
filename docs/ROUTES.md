# API Routes Documentation

## Server Configuration

- **Port**: 9001
- **Framework**: Gin (Go)
- **Base URL**: `http://localhost:9001`

## Available Endpoints

### Health Check
Check if the service is running and healthy.

**Endpoint**: `GET /health`

**Response**:
```json
{
  "status": "healthy",
  "service": "contact-management-api",
  "version": "1.0.0"
}
```

**Example**:
```bash
curl http://localhost:9001/health
```

---

### Ping
Simple ping endpoint to test API connectivity.

**Endpoint**: `GET /api/v1/ping`

**Response**:
```json
{
  "message": "pong"
}
```

**Example**:
```bash
curl http://localhost:9001/api/v1/ping
```

---

### Get User by ID
Retrieve user information by ID.

**Endpoint**: `POST /api/v1/mobile/users/:id`

**Parameters**:
- `id` (path parameter): User ID

**Response**:
```json
{
  "id": "user-id",
  "email": "user@example.com",
  "name": "User Name"
}
```

**Example**:
```bash
curl -X POST http://localhost:9001/api/v1/mobile/users/123
```

---

## Adding New Routes

To add new routes, edit the file: `internal/app/routes/routes.go`

### Example:

```go
// In routes.go
func SetupRoutes(router *gin.Engine, handler *app.Handler) {
    // Add your new route here
    router.GET("/your-endpoint", handler.YourHandler)
}
```

Then implement the handler in `internal/app/handler.go`:

```go
func (h *Handler) YourHandler(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "Your response"
    })
}
```

## Testing

Run the server:
```bash
make run
# or
go run ./cmd/server/main.go
```

Test with curl:
```bash
# Health check
curl http://localhost:9001/health

# Ping
curl http://localhost:9001/api/v1/ping
```
