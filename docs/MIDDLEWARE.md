# Middleware Documentation

## Overview

Middleware components untuk request processing, authentication, timeout handling, dan error recovery.

## Architecture

```
internal/middleware/
├── auth.go            # JWT authentication middleware
├── timeout.go         # Request timeout middleware
├── error_handler.go   # Error recovery & 404 handling
├── secure_headers.go  # Security headers middleware
└── secure_headers_test.go
```

---

## Authentication Middleware

### `AuthMiddleware(service *service.Service) gin.HandlerFunc`

Validates JWT token and sets userID in Gin context.

**Purpose:**
- Extract and validate JWT token from Authorization header
- Set authenticated user ID in request context
- Protect endpoints that require authentication

**Flow:**
1. Extract `Authorization` header
2. Check for `Bearer ` prefix
3. Parse and validate JWT token using service
4. Extract userID from token claims
5. Set userID in Gin context
6. Continue to next handler

**Usage:**
```go
// In routes.go
authMiddleware := middleware.AuthMiddleware(service)

protected := api.Group("/")
protected.Use(authMiddleware)
{
    protected.GET("/me", handler.GetProfile)
    protected.PUT("/me", handler.UpdateProfile)
    protected.GET("/contacts", handler.ListContacts)
}
```

**Request Format:**
```http
GET /api/v1/me HTTP/1.1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Success Response:**
- Sets `userID` in context
- Continues to next handler

**Error Responses:**

**401 - Missing Token:**
```json
{
  "status": 0,
  "status_code": 401,
  "message": "Unauthorized - missing token",
  "data": {}
}
```

**401 - Invalid Format:**
```json
{
  "status": 0,
  "status_code": 401,
  "message": "Unauthorized - invalid token format",
  "data": {}
}
```

**401 - Invalid/Expired Token:**
```json
{
  "status": 0,
  "status_code": 401,
  "message": "Unauthorized - invalid or expired token",
  "data": {}
}
```

**Handler Usage:**
```go
func (h *Handler) GetProfile(c *gin.Context) {
    // Get userID from context (set by AuthMiddleware)
    userID, exists := c.Get("userID")
    if !exists {
        utils.UnauthorizedResponse(c, "Unauthorized")
        return
    }
    
    // Use userID
    profile, err := h.service.GetProfile(c.Request.Context(), userID.(uint))
    // ...
}
```

---

### `CORSMiddleware() gin.HandlerFunc`

Handles Cross-Origin Resource Sharing (CORS) headers.

**Purpose:**
- Allow cross-origin requests from frontend
- Set proper CORS headers
- Handle preflight OPTIONS requests

**Headers Set:**
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization`

**Usage:**
```go
router := gin.Default()
router.Use(middleware.CORSMiddleware())
```

**Production Configuration:**
```go
// Restrict origins in production
router.Use(func(c *gin.Context) {
    origin := c.Request.Header.Get("Origin")
    allowedOrigins := []string{
        "https://yourapp.com",
        "https://www.yourapp.com",
    }
    
    for _, allowed := range allowedOrigins {
        if origin == allowed {
            c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
            break
        }
    }
    
    c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    
    if c.Request.Method == "OPTIONS" {
        c.AbortWithStatus(204)
        return
    }
    
    c.Next()
})
```

---

### `LoggerMiddleware() gin.HandlerFunc`

Logs HTTP requests and responses.

**Purpose:**
- Log incoming requests
- Track request duration
- Monitor API usage

**Logged Information:**
- HTTP method
- Request path
- Status code
- Duration
- Client IP

**Usage:**
```go
router.Use(middleware.LoggerMiddleware())
```

**Console Output:**
```
[GIN] 2025/10/16 - 10:30:45 | 200 | 45.234ms | 127.0.0.1 | GET /api/v1/contacts
[GIN] 2025/10/16 - 10:30:46 | 201 | 123.456ms | 127.0.0.1 | POST /api/v1/auth/register
```

**Production Enhancement:**
```go
// Use structured logging
import "log/slog"

func LoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        
        c.Next()
        
        duration := time.Since(start)
        
        slog.Info("request processed",
            "method", c.Request.Method,
            "path", path,
            "status", c.Writer.Status(),
            "duration_ms", duration.Milliseconds(),
            "client_ip", c.ClientIP(),
        )
    }
}
```

---

## Timeout Middleware

### `TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc`

Times out requests after specified duration.

**Purpose:**
- Prevent long-running requests from blocking resources
- Improve API responsiveness
- Protect against slow queries or external API calls

**Parameters:**
- `timeout` (time.Duration): Maximum request duration

**Usage:**
```go
// Custom timeout
router.Use(middleware.TimeoutMiddleware(10 * time.Second))

// Or use default 30 seconds
router.Use(middleware.DefaultTimeoutMiddleware())
```

**Success:**
- Request completes within timeout
- Normal response returned

**Timeout Response (408):**
```json
{
  "status": 0,
  "status_code": 408,
  "message": "Request timeout - operation took too long",
  "data": {}
}
```

**Example with Context:**
```go
func (h *Handler) SlowOperation(c *gin.Context) {
    // Use c.Request.Context() which has timeout from middleware
    result, err := h.service.LongRunningOperation(c.Request.Context())
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            utils.ErrorResponse(c, 408, "Operation timed out", gin.H{})
            return
        }
        utils.InternalErrorResponse(c, "Operation failed")
        return
    }
    
    utils.OKResponse(c, "Success", result)
}
```

**Service Layer Usage:**
```go
func (s *Service) LongRunningOperation(ctx context.Context) (interface{}, error) {
    // Check context for cancellation/timeout
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // Continue operation
    }
    
    // Database query with context
    var result []Contact
    if err := s.db.WithContext(ctx).Find(&result).Error; err != nil {
        return nil, err
    }
    
    return result, nil
}
```

---

### `DefaultTimeoutMiddleware() gin.HandlerFunc`

Shortcut for 30-second timeout.

**Usage:**
```go
router.Use(middleware.DefaultTimeoutMiddleware())
```

**Recommended Timeouts:**
- **API Gateway**: 30 seconds (default)
- **Database queries**: 5-10 seconds
- **External API calls**: 15-20 seconds
- **File uploads**: 60-300 seconds
- **Report generation**: 60-120 seconds

---

## Error Handler Middleware

### `ErrorHandlerMiddleware() gin.HandlerFunc`

Recovers from panics and returns consistent JSON error responses.

**Purpose:**
- Catch panics in handlers
- Prevent application crashes
- Return consistent error responses
- Log stack traces for debugging

**Usage:**
```go
router.Use(middleware.ErrorHandlerMiddleware())
```

**Response on Panic (500):**
```json
{
  "status": 0,
  "status_code": 500,
  "message": "Internal server error",
  "data": {}
}
```

**Console Output:**
```
PANIC: runtime error: invalid memory address or nil pointer dereference
Stack trace:
goroutine 1 [running]:
runtime/debug.Stack()
    /usr/local/go/src/runtime/debug/stack.go:24 +0x65
...
```

**Example Handler with Panic:**
```go
func (h *Handler) DangerousOperation(c *gin.Context) {
    // This will panic
    var user *User
    name := user.FullName // nil pointer dereference
    
    // ErrorHandlerMiddleware will catch this panic
    // and return 500 error response
}
```

**Production Enhancement:**
```go
// Send alerts on panic
defer func() {
    if err := recover(); err != nil {
        // Log to monitoring service
        sentry.CaptureException(fmt.Errorf("%v", err))
        
        // Send alert
        alerting.SendAlert("Panic in API", fmt.Sprintf("%v", err))
        
        // Return error response
        if !c.Writer.Written() {
            c.JSON(500, gin.H{
                "status": 0,
                "status_code": 500,
                "message": "Internal server error",
                "data": gin.H{},
            })
        }
    }
}()
```

---

### `NotFoundHandler() gin.HandlerFunc`

Handles 404 errors with consistent JSON response.

**Purpose:**
- Return JSON for unknown endpoints
- Maintain consistent API response format

**Usage:**
```go
router.NoRoute(middleware.NotFoundHandler())
```

**Response (404):**
```json
{
  "status": 0,
  "status_code": 404,
  "message": "Endpoint not found",
  "data": {}
}
```

**Example:**
```http
GET /api/v1/unknown-endpoint HTTP/1.1

HTTP/1.1 404 Not Found
Content-Type: application/json

{
  "status": 0,
  "status_code": 404,
  "message": "Endpoint not found",
  "data": {}
}
```

---

### `MethodNotAllowedHandler() gin.HandlerFunc`

Handles 405 errors with consistent JSON response.

**Purpose:**
- Handle requests with unsupported HTTP methods
- Return JSON instead of default HTML error

**Usage:**
```go
router.NoMethod(middleware.MethodNotAllowedHandler())
```

**Response (405):**
```json
{
  "status": 0,
  "status_code": 405,
  "message": "Method not allowed",
  "data": {}
}
```

**Example:**
```http
PATCH /api/v1/contacts/1 HTTP/1.1

HTTP/1.1 405 Method Not Allowed
Content-Type: application/json

{
  "status": 0,
  "status_code": 405,
  "message": "Method not allowed",
  "data": {}
}
```

---

## Complete Middleware Setup

### Recommended Order

```go
func SetupRouter() *gin.Engine {
    router := gin.New() // Use gin.New() instead of gin.Default()
    
    // 1. Recovery - catch panics first
    router.Use(middleware.ErrorHandlerMiddleware())
    
    // 2. Timeout - limit request duration
    router.Use(middleware.DefaultTimeoutMiddleware())
    
    // 3. Logger - log all requests
    router.Use(middleware.LoggerMiddleware())
    
    // 4. CORS - enable cross-origin requests
    router.Use(middleware.CORSMiddleware())
    
    // 5. Custom error handlers
    router.NoRoute(middleware.NotFoundHandler())
    router.NoMethod(middleware.MethodNotAllowedHandler())
    
    // 6. Routes with authentication
    api := router.Group("/api/v1")
    {
        // Public routes
        auth := api.Group("/auth")
        {
            auth.POST("/register", handler.Register)
            auth.POST("/login", handler.Login)
        }
        
        // Protected routes
        authMiddleware := middleware.AuthMiddleware(service)
        protected := api.Group("/")
        protected.Use(authMiddleware)
        {
            protected.GET("/me", handler.GetProfile)
            protected.PUT("/me", handler.UpdateProfile)
            
            contacts := protected.Group("/contacts")
            {
                contacts.GET("", handler.ListContacts)
                contacts.POST("", handler.CreateContact)
                contacts.GET("/:id", handler.GetContact)
                contacts.PUT("/:id", handler.UpdateContact)
                contacts.DELETE("/:id", handler.DeleteContact)
            }
        }
    }
    
    return router
}
```

---

## Testing Middleware

### Test Authentication

```bash
# Without token - should return 401
curl -X GET http://localhost:9001/api/v1/me

# With invalid token - should return 401
curl -X GET http://localhost:9001/api/v1/me \
  -H "Authorization: Bearer invalid_token"

# With valid token - should return 200
curl -X GET http://localhost:9001/api/v1/me \
  -H "Authorization: Bearer <valid_token>"
```

---

### Test Timeout

```go
// Create a slow handler for testing
func (h *Handler) SlowEndpoint(c *gin.Context) {
    time.Sleep(35 * time.Second) // Longer than 30s timeout
    utils.OKResponse(c, "Success", gin.H{})
}

// Test
curl -X GET http://localhost:9001/api/v1/slow
# Should return 408 after 30 seconds
```

---

### Test Error Recovery

```go
// Create a panic handler for testing
func (h *Handler) PanicEndpoint(c *gin.Context) {
    panic("intentional panic for testing")
}

// Test
curl -X GET http://localhost:9001/api/v1/panic
# Should return 500 instead of crashing
```

---

## Best Practices

### 1. Apply Error Handler First

```go
// ✅ Good - catches all panics
router.Use(middleware.ErrorHandlerMiddleware())
router.Use(otherMiddleware...)

// ❌ Bad - panics in other middleware won't be caught
router.Use(otherMiddleware...)
router.Use(middleware.ErrorHandlerMiddleware())
```

---

### 2. Use Context Timeout

```go
// ✅ Good - respects timeout from middleware
func (s *Service) Query(ctx context.Context) error {
    return s.db.WithContext(ctx).Find(&result).Error
}

// ❌ Bad - ignores timeout
func (s *Service) Query(ctx context.Context) error {
    return s.db.Find(&result).Error
}
```

---

### 3. Set Appropriate Timeouts

```go
// ✅ Good - different timeouts for different operations
router.GET("/quick", middleware.TimeoutMiddleware(5*time.Second), handler.Quick)
router.POST("/upload", middleware.TimeoutMiddleware(60*time.Second), handler.Upload)

// ❌ Bad - same timeout for all
router.Use(middleware.TimeoutMiddleware(30*time.Second))
```

---

### 4. Check Authentication in Handlers

```go
// ✅ Good - double check userID exists
userID, exists := c.Get("userID")
if !exists {
    utils.UnauthorizedResponse(c, "Unauthorized")
    return
}

// ❌ Bad - assume userID always exists
userID := c.MustGet("userID") // Will panic if not set
```

---

## Monitoring & Metrics

### Add Metrics Middleware

```go
func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        
        // Record metrics
        metrics.RecordHTTPRequest(
            c.Request.Method,
            c.FullPath(),
            c.Writer.Status(),
            duration,
        )
    }
}
```

---

## References

- [Gin Middleware Documentation](https://gin-gonic.com/docs/examples/custom-middleware/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [HTTP Status Codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
- [CORS Guide](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
