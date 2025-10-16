# Logger Documentation

## Overview

Structured JSON logging system untuk Kibana/ELK Stack dengan logging middleware yang mencatat setiap HTTP request dan response dalam format yang mudah di-query.

## Architecture

```
internal/logger/
├── logger.go          # Core logger dengan structured logging
├── middleware.go      # Gin middleware untuk HTTP logging
└── logger_test.go     # Unit tests (9 test suites)
```

---

## Features

✅ **Structured JSON Logging** - Format JSON untuk Kibana/Elasticsearch
✅ **Request/Response Logging** - Capture semua HTTP requests dan responses
✅ **Correlation ID** - UUID unik untuk request tracking
✅ **Sensitive Data Masking** - Password dan data sensitif di-redact
✅ **Performance Tracking** - Latency monitoring per request
✅ **User Context** - Track user_id dari authenticated requests
✅ **Error Tracking** - Error type dan message logging
✅ **File + Console Output** - Log ke file dan stdout
✅ **Log Levels** - Debug, Info, Warn, Error
✅ **Response Body Capture** - Record API responses untuk audit

---

## Configuration

### Initialize Logger

```go
import "user-service/internal/logger"

func main() {
    // Configure logger
    config := logger.Config{
        Level:      "info",        // debug, info, warn, error
        OutputPath: "logs/app.log", // log file path
    }
    
    // Initialize
    if err := logger.Init(config); err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }
    defer logger.Close()
    
    logger.Info("Application started", "version", "1.0.0")
}
```

### Environment Variables

Add to `.env`:
```env
LOG_LEVEL=info
LOG_OUTPUT=logs/app.log
```

---

## Log Format

### Standard Fields

All logs include these fields:

```json
{
  "timestamp": "2025-10-16T10:30:45.123456+07:00",
  "level": "INFO",
  "msg": "GET /api/v1/contacts - ✓ (45ms)",
  
  // HTTP Request fields
  "method": "GET",
  "path": "/api/v1/contacts",
  "status": 200,
  "latency_ms": 45,
  "client_ip": "127.0.0.1",
  "user_agent": "Mozilla/5.0...",
  
  // Request/Response data
  "request_body": "{\"query\":\"john\"}",
  "response_body": "{\"status\":1,\"data\":[]}",
  
  // Context fields
  "user_id": 123,
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  
  // Error fields (when applicable)
  "error_type": "database_error",
  "error_message": "Connection timeout"
}
```

---

## HTTP Request Logging

### Automatic Logging Middleware

Add to your Gin router:

```go
import (
    "github.com/gin-gonic/gin"
    "user-service/internal/logger"
)

func main() {
    router := gin.New() // Use gin.New() instead of gin.Default()
    
    // Add logging middleware FIRST
    router.Use(logger.LoggingMiddleware())
    
    // Then add other middleware
    router.Use(otherMiddleware...)
    
    // Setup routes...
}
```

### What Gets Logged

**On Every Request:**
- ✅ HTTP method, path, status code
- ✅ Request latency in milliseconds
- ✅ Client IP address and User-Agent
- ✅ Request body (sanitized for auth endpoints)
- ✅ Response body (truncated if too large)
- ✅ User ID (if authenticated)
- ✅ Unique correlation ID
- ✅ Error information (if error occurred)

---

## Log Examples

### Success Request

```json
{
  "timestamp": "2025-10-16T10:30:45.587667+07:00",
  "level": "INFO",
  "msg": "GET /api/v1/contacts - ✓ (45ms)",
  "method": "GET",
  "path": "/api/v1/contacts",
  "status": 200,
  "latency_ms": 45,
  "client_ip": "127.0.0.1",
  "user_agent": "PostmanRuntime/7.29.2",
  "user_id": 1,
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "request_body": "",
  "response_body": "{\"status\":1,\"status_code\":200,\"message\":\"Contacts loaded successfully\",\"data\":{\"count\":4,\"page\":1,\"limit\":20,\"contacts\":[...]}}"
}
```

### Authentication Request (Password Redacted)

```json
{
  "timestamp": "2025-10-16T10:31:20.123456+07:00",
  "level": "INFO",
  "msg": "POST /api/v1/auth/login - ✓ (89ms)",
  "method": "POST",
  "path": "/api/v1/auth/login",
  "status": 200,
  "latency_ms": 89,
  "client_ip": "127.0.0.1",
  "user_agent": "Mozilla/5.0",
  "correlation_id": "660e8400-e29b-41d4-a716-446655440001",
  "request_body": "{\"email\":\"user@example.com\",\"password\":\"***REDACTED***\"}",
  "response_body": "{\"status\":1,\"data\":{\"id\":1,\"token\":{\"access_token\":\"eyJ...\"}}}"
}
```

### Error Request

```json
{
  "timestamp": "2025-10-16T10:32:15.789012+07:00",
  "level": "ERROR",
  "msg": "POST /api/v1/contacts - ✗ (123ms)",
  "method": "POST",
  "path": "/api/v1/contacts",
  "status": 500,
  "latency_ms": 123,
  "client_ip": "127.0.0.1",
  "user_agent": "curl/7.68.0",
  "user_id": 1,
  "correlation_id": "770e8400-e29b-41d4-a716-446655440002",
  "request_body": "{\"full_name\":\"John Doe\",\"phone\":\"081234567890\"}",
  "response_body": "{\"status\":0,\"status_code\":500,\"message\":\"Internal server error\",\"data\":{}}",
  "error_type": "server_error",
  "error_message": "Internal server error"
}
```

### Client Error (4xx)

```json
{
  "timestamp": "2025-10-16T10:33:00.456789+07:00",
  "level": "WARN",
  "msg": "GET /api/v1/contacts/999 - ⚠ (15ms)",
  "method": "GET",
  "path": "/api/v1/contacts/999",
  "status": 404,
  "latency_ms": 15,
  "client_ip": "127.0.0.1",
  "user_agent": "PostmanRuntime/7.29.2",
  "user_id": 1,
  "correlation_id": "880e8400-e29b-41d4-a716-446655440003",
  "response_body": "{\"status\":0,\"status_code\":404,\"message\":\"Contact not found\",\"data\":{}}",
  "error_type": "client_error",
  "error_message": "Contact not found"
}
```

---

## Manual Logging

### Basic Logging

```go
import "user-service/internal/logger"

// Info level
logger.Info("User logged in", "user_id", 123, "email", "user@example.com")

// Error level
logger.Error("Database connection failed", "error", err, "retry_count", 3)

// Warning level
logger.Warn("API rate limit approaching", "remaining", 10, "limit", 100)

// Debug level (only in debug mode)
logger.Debug("Cache miss", "key", "user:123", "ttl", 3600)
```

### Structured Logging with Fields

```go
logger.Info("Order created",
    "order_id", 12345,
    "user_id", 67,
    "amount", 150.00,
    "currency", "USD",
    "items", 3,
)

// Output:
// {"timestamp":"...","level":"INFO","msg":"Order created","order_id":12345,"user_id":67,"amount":150,"currency":"USD","items":3}
```

### Logger with Context Fields

```go
// Create logger with persistent fields
contextLogger := logger.WithFields(map[string]interface{}{
    "service": "payment",
    "version": "1.0.0",
    "environment": "production",
})

// All logs from this logger include context fields
contextLogger.Info("Processing payment", "amount", 100.00)
contextLogger.Error("Payment failed", "error", err)

// Output includes service, version, environment in every log
```

---

## Custom Log Entry

For specialized logging needs:

```go
import "user-service/internal/logger"

entry := logger.LogEntry{
    Level:         "info",
    Method:        "POST",
    Path:          "/api/v1/orders",
    Status:        201,
    Latency:       156,
    ClientIP:      "192.168.1.100",
    UserAgent:     "Mobile App v2.1.0",
    UserID:        &userID,
    CorrelationID: correlationID,
    Message:       "Order created successfully",
    AdditionalData: map[string]interface{}{
        "order_id": 12345,
        "total": 250.00,
    },
}

logger.LogHTTPRequest(entry)
```

---

## Correlation ID Tracking

### Generate Correlation ID

```go
correlationID := logger.GenerateCorrelationID()
// Returns: "550e8400-e29b-41d4-a716-446655440000"
```

### Use in Request Context

The middleware automatically:
1. Generates correlation ID for each request
2. Sets it in Gin context: `c.Set("correlation_id", correlationID)`
3. Includes it in all logs for that request

**Access in handler:**
```go
func (h *Handler) GetProfile(c *gin.Context) {
    correlationID, _ := c.Get("correlation_id")
    
    // Use for distributed tracing
    logger.Info("Fetching user profile",
        "correlation_id", correlationID,
        "user_id", userID,
    )
}
```

---

## Sensitive Data Handling

### Automatic Password Redaction

For endpoints containing `/auth/`:
- Password fields are automatically replaced with `***REDACTED***`
- Original password never appears in logs

**Example:**
```json
// Request body before logging:
{"email":"user@example.com","password":"MySecret123"}

// In logs:
"request_body": "{\"email\":\"user@example.com\",\"password\":\"***REDACTED***\"}"
```

### Manual Sanitization

For custom sensitive data:

```go
func sanitizeData(data map[string]interface{}) map[string]interface{} {
    sensitiveFields := []string{"password", "token", "secret", "apiKey"}
    
    for _, field := range sensitiveFields {
        if _, exists := data[field]; exists {
            data[field] = "***REDACTED***"
        }
    }
    
    return data
}
```

---

## Response Body Truncation

### Automatic Truncation

- Request bodies limited to **500 characters** in logs
- Response bodies limited to **500 characters** in logs
- Larger content is truncated with `...` suffix

**Example:**
```json
"response_body": "{\"status\":1,\"data\":[{\"id\":1,\"name\":\"John\"},{\"id\":2,\"name\":\"Jane\"}... (truncated)"
```

### Disable for Specific Routes

```go
// In handler, set flag to skip body logging
func (h *Handler) UploadFile(c *gin.Context) {
    c.Set("skip_body_logging", true)
    // Handle large file upload...
}
```

---

## Error Logging

### Automatic Error Detection

Middleware automatically detects errors based on:

**Status Code:**
- `4xx`: Level = WARN, Error Type = "client_error"
- `5xx`: Level = ERROR, Error Type = "server_error"

**Gin Errors:**
- Checks `c.Errors` for any handler errors
- Sets Error Type = "handler_error"

### Error Message Extraction

Automatically extracts error messages from response:

```json
// Response body:
{
  "status": 0,
  "message": "User not found"
}

// Log entry:
{
  "error_message": "User not found"
}
```

---

## Kibana Integration

### Index Pattern

Create Kibana index pattern:
```
logs-app-*
```

### Useful Queries

**Find all errors:**
```
level: ERROR
```

**Find requests by user:**
```
user_id: 123
```

**Find slow requests (>1000ms):**
```
latency_ms: >1000
```

**Trace request by correlation ID:**
```
correlation_id: "550e8400-e29b-41d4-a716-446655440000"
```

**Find authentication failures:**
```
path: "/api/v1/auth/*" AND status: 401
```

**Find database errors:**
```
error_type: "database_error"
```

### Dashboard Examples

**Request Volume:**
```
Visualization: Line chart
Y-axis: Count
X-axis: @timestamp
Split series: status
```

**Response Time Percentiles:**
```
Visualization: Line chart
Y-axis: Percentiles of latency_ms (50th, 95th, 99th)
X-axis: @timestamp
```

**Error Rate:**
```
Visualization: Metric
Aggregation: Count where level: ERROR
```

**Top Users by Request Count:**
```
Visualization: Pie chart
Slice by: user_id
Size: 10
```

---

## Log File Management

### Location

Default: `logs/app.log`

### Directory Structure

```
logs/
├── app.log              # Current log file
├── app-2025-10-15.log   # Rotated log (if rotation enabled)
└── app-2025-10-14.log
```

### Log Rotation

**Manual Rotation (recommended for production):**

Use `logrotate` on Linux:

```bash
# /etc/logrotate.d/hackathon-api
/path/to/hackathon_1/logs/app.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0644 app app
    sharedscripts
    postrotate
        kill -USR1 $(cat /var/run/app.pid)
    endscript
}
```

---

## Performance

### Benchmarks

```
BenchmarkLogHTTPRequest-8    50000    24567 ns/op    4096 B/op    42 allocs/op
```

**Throughput:** ~40,000 log entries/second

### Optimization Tips

1. **Disable debug logs in production:**
   ```go
   config.Level = "info"  // Not "debug"
   ```

2. **Limit body sizes:**
   - Already implemented (500 chars limit)

3. **Async logging (if needed):**
   ```go
   go logger.LogHTTPRequest(entry)
   ```

4. **Skip logging for health checks:**
   ```go
   if c.Request.URL.Path == "/health" {
       c.Next()
       return
   }
   ```

---

## Testing

### Run Tests

```bash
# Test logger package
go test ./internal/logger/... -v

# Test with coverage
go test ./internal/logger/... -cover

# Benchmark
go test ./internal/logger/... -bench=.
```

### Test Results

```
PASS: TestInit (3 subtests)
PASS: TestGenerateCorrelationID
PASS: TestLogHTTPRequest (3 subtests)
PASS: TestLoggerHelpers (4 subtests)
PASS: TestWithFields
PASS: TestSanitizeRequestBody (3 subtests)
PASS: TestExtractErrorMessage (4 subtests)
PASS: TestGenerateLogMessage (2 subtests)
PASS: TestLimitString (2 subtests)

Total: 22 test cases, all passing
Coverage: ~85%
```

---

## Best Practices

### 1. Use Structured Logging

```go
// ✅ Good - structured
logger.Info("User created",
    "user_id", user.ID,
    "email", user.Email,
    "role", user.Role,
)

// ❌ Bad - string formatting
logger.Info(fmt.Sprintf("User %d created with email %s", user.ID, user.Email))
```

### 2. Log Contextual Information

```go
// ✅ Good - includes context
logger.Error("Failed to update user",
    "user_id", userID,
    "error", err,
    "retry_count", retryCount,
    "correlation_id", correlationID,
)

// ❌ Bad - missing context
logger.Error("Failed to update user", "error", err)
```

### 3. Use Appropriate Log Levels

```go
logger.Debug("Cache lookup", "key", cacheKey)           // Debug: detailed debugging
logger.Info("User logged in", "user_id", userID)        // Info: important events
logger.Warn("Rate limit approaching", "remaining", 10)  // Warn: concerning but not error
logger.Error("Database query failed", "error", err)     // Error: actual errors
```

### 4. Don't Log Sensitive Data

```go
// ✅ Good - password masked
logger.Info("Login attempt", "email", email)

// ❌ Bad - logging password
logger.Info("Login attempt", "email", email, "password", password)
```

### 5. Include Correlation IDs

```go
// ✅ Good - traceable across services
logger.Info("Calling external API",
    "correlation_id", correlationID,
    "api", "payment-service",
    "endpoint", "/api/process",
)
```

---

## Troubleshooting

### Log File Not Created

**Problem:** `logs/app.log` not created

**Solution:**
```bash
# Create logs directory
mkdir -p logs
chmod 755 logs
```

### Permission Denied

**Problem:** Cannot write to log file

**Solution:**
```bash
# Fix permissions
chmod 666 logs/app.log
```

### Logs Not Appearing

**Problem:** Logs not showing in file

**Solution:**
```go
// Ensure logger is initialized
logger.Init(config)

// Ensure Close() is called
defer logger.Close()

// Check log level
config.Level = "debug" // Try debug level
```

### Large Log Files

**Problem:** Log file growing too large

**Solution:**
- Implement log rotation (see Log Rotation section)
- Increase truncation limits
- Filter out health check logs

---

## Migration from Gin Default Logger

### Before

```go
router := gin.Default() // Includes default logger
```

### After

```go
router := gin.New()                       // No default middleware
router.Use(logger.LoggingMiddleware())    // Our structured logger
router.Use(gin.Recovery())                // Still need recovery
```

---

## References

- [Structured Logging Best Practices](https://www.thoughtworks.com/insights/blog/structured-logging)
- [Kibana Query Language (KQL)](https://www.elastic.co/guide/en/kibana/current/kuery-query.html)
- [Go slog Package](https://pkg.go.dev/log/slog)
- [Logrotate Configuration](https://linux.die.net/man/8/logrotate)
