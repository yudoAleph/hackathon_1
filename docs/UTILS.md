# Utils Documentation

## Overview

Utils package menyediakan utility functions untuk validation, response formatting, dan helper functions yang digunakan di seluruh aplikasi.

## Architecture

```
internal/utils/
├── validation.go       # Email, phone, password validation
├── validation_test.go  # Unit tests untuk validation
└── response.go         # Response helpers untuk konsistensi JSON
```

---

## Validation Functions

### Email Validation

#### `ValidateEmail(email string) bool`

Validates email format berdasarkan RFC 5322.

**Parameters:**
- `email` (string): Email address to validate

**Returns:**
- `bool`: true if valid, false otherwise

**Rules:**
- Not empty
- Length between 5-254 characters
- Contains @ symbol
- Valid domain format

**Example:**
```go
if !utils.ValidateEmail("user@example.com") {
    return errors.New("invalid email format")
}
```

**Valid Emails:**
- `user@example.com`
- `john.doe@mail.example.com`
- `user+tag@example.com`
- `user-name@example.co.id`

**Invalid Emails:**
- `userexample.com` (no @)
- `user@` (no domain)
- `@example.com` (no local part)
- `user @example.com` (contains space)

---

### Phone Validation

#### `ValidatePhone(phone string) bool`

Validates phone number format (international).

**Parameters:**
- `phone` (string): Phone number to validate

**Returns:**
- `bool`: true if valid, false otherwise

**Rules:**
- Not empty
- 10-15 digits (after removing separators)
- Supports international format with country code
- Allows separators: spaces, dashes, dots, parentheses

**Example:**
```go
if !utils.ValidatePhone("+62 812 3456 7890") {
    return errors.New("invalid phone format")
}
```

**Valid Phones:**
- `081234567890`
- `+6281234567890`
- `+62 812 3456 7890`
- `+62-812-3456-7890`
- `+1 (234) 567-8901`

**Invalid Phones:**
- `12345` (too short)
- `12345678901234567` (too long)
- `081234abcd` (contains letters)

---

#### `ValidateIndonesiaPhone(phone string) bool`

Validates Indonesian phone number format (strict).

**Parameters:**
- `phone` (string): Indonesian phone number

**Returns:**
- `bool`: true if valid, false otherwise

**Rules:**
- Must start with: `0`, `62`, or `+62`
- 9-13 digits after prefix
- No separators allowed

**Example:**
```go
if !utils.ValidateIndonesiaPhone("081234567890") {
    return errors.New("invalid Indonesia phone format")
}
```

**Valid Indonesia Phones:**
- `081234567890`
- `6281234567890`
- `+6281234567890`

**Invalid Indonesia Phones:**
- `181234567890` (wrong prefix)
- `08123456` (too short)
- `+12345678901` (not Indonesia)

---

### Password Validation

#### `ValidatePassword(password string) (bool, []string)`

Validates password strength.

**Parameters:**
- `password` (string): Password to validate

**Returns:**
- `bool`: true if valid, false otherwise
- `[]string`: List of validation error messages

**Rules:**
- Minimum 8 characters
- Maximum 128 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one digit

**Example:**
```go
valid, errors := utils.ValidatePassword("Password123")
if !valid {
    return fmt.Errorf("password validation failed: %v", errors)
}
```

**Valid Passwords:**
- `Password123`
- `MyP@ssw0rd`
- `StrongPass1`

**Invalid Passwords:**
- `pass1` (too short, no uppercase)
- `password123` (no uppercase)
- `PASSWORD123` (no lowercase)
- `Password` (no digit)

---

### Full Name Validation

#### `ValidateFullName(name string) (bool, string)`

Validates full name format.

**Parameters:**
- `name` (string): Full name to validate

**Returns:**
- `bool`: true if valid, false otherwise
- `string`: Error message if invalid

**Rules:**
- Not empty
- Minimum 2 characters
- Maximum 100 characters
- Only letters, spaces, and basic punctuation (. ' -)

**Example:**
```go
valid, err := utils.ValidateFullName("John Doe")
if !valid {
    return errors.New(err)
}
```

**Valid Names:**
- `John Doe`
- `O'Brien`
- `Mary-Jane Watson`
- `Dr. Smith`

**Invalid Names:**
- `A` (too short)
- `John123` (contains numbers)
- `John@Doe` (invalid characters)

---

## Sanitization Functions

### `SanitizeEmail(email string) string`

Sanitizes email by trimming whitespace and converting to lowercase.

**Example:**
```go
email := utils.SanitizeEmail("  USER@EXAMPLE.COM  ")
// Result: "user@example.com"
```

---

### `SanitizePhone(phone string) string`

Sanitizes phone number by trimming whitespace.

**Example:**
```go
phone := utils.SanitizePhone("  +62 812 3456 7890  ")
// Result: "+62 812 3456 7890"
```

---

### `NormalizeIndonesiaPhone(phone string) string`

Normalizes Indonesian phone number to +62 format.

**Example:**
```go
phone1 := utils.NormalizeIndonesiaPhone("081234567890")
// Result: "+6281234567890"

phone2 := utils.NormalizeIndonesiaPhone("6281234567890")
// Result: "+6281234567890"

phone3 := utils.NormalizeIndonesiaPhone("0812-3456-7890")
// Result: "+6281234567890"
```

---

## Response Helpers

### Standard Response Format

All API responses use this consistent format:

```go
type StandardResponse struct {
    Status     int         `json:"status"`      // 1 = success, 0 = error
    StatusCode int         `json:"status_code"` // HTTP status code
    Message    string      `json:"message"`     // Human-readable message
    Data       interface{} `json:"data"`        // Response data or error details
}
```

---

### Success Responses

#### `SuccessResponse(c *gin.Context, statusCode int, message string, data interface{})`

Creates a success response.

**Example:**
```go
utils.SuccessResponse(c, 200, "User found", user)
```

**Output:**
```json
{
  "status": 1,
  "status_code": 200,
  "message": "User found",
  "data": { ... }
}
```

---

#### `OKResponse(c *gin.Context, message string, data interface{})`

Shortcut for 200 OK response.

**Example:**
```go
utils.OKResponse(c, "Success", data)
```

---

#### `CreatedResponse(c *gin.Context, message string, data interface{})`

Shortcut for 201 Created response.

**Example:**
```go
utils.CreatedResponse(c, "User created", user)
```

---

#### `NoContentResponse(c *gin.Context)`

Shortcut for 204 No Content response.

**Example:**
```go
utils.NoContentResponse(c)
```

---

### Error Responses

#### `ErrorResponse(c *gin.Context, statusCode int, message string, data interface{})`

Creates an error response.

**Example:**
```go
utils.ErrorResponse(c, 500, "Database error", gin.H{})
```

**Output:**
```json
{
  "status": 0,
  "status_code": 500,
  "message": "Database error",
  "data": {}
}
```

---

#### `BadRequestResponse(c *gin.Context, message string, data interface{})`

Shortcut for 400 Bad Request.

**Example:**
```go
utils.BadRequestResponse(c, "Invalid input", gin.H{"field": "email"})
```

---

#### `UnauthorizedResponse(c *gin.Context, message string)`

Shortcut for 401 Unauthorized.

**Example:**
```go
utils.UnauthorizedResponse(c, "Invalid token")
```

---

#### `ForbiddenResponse(c *gin.Context, message string)`

Shortcut for 403 Forbidden.

**Example:**
```go
utils.ForbiddenResponse(c, "Access denied")
```

---

#### `NotFoundResponse(c *gin.Context, message string)`

Shortcut for 404 Not Found.

**Example:**
```go
utils.NotFoundResponse(c, "User not found")
```

---

#### `ConflictResponse(c *gin.Context, message string, data interface{})`

Shortcut for 409 Conflict.

**Example:**
```go
utils.ConflictResponse(c, "Email already exists", gin.H{"email": email})
```

---

#### `InternalErrorResponse(c *gin.Context, message string)`

Shortcut for 500 Internal Server Error.

**Example:**
```go
utils.InternalErrorResponse(c, "Database connection failed")
```

---

### Validation Error Responses

#### `ValidationErrorResponse(c *gin.Context, field string, messages []string)`

Creates validation error for single field.

**Example:**
```go
utils.ValidationErrorResponse(c, "email", []string{"invalid format", "already exists"})
```

**Output:**
```json
{
  "status": 0,
  "status_code": 400,
  "message": "Validation error",
  "data": {
    "email": ["invalid format", "already exists"]
  }
}
```

---

#### `ValidationErrorsResponse(c *gin.Context, errors map[string][]string)`

Creates validation error for multiple fields.

**Example:**
```go
errors := map[string][]string{
    "email": []string{"invalid format"},
    "password": []string{"too short", "must contain uppercase"},
}
utils.ValidationErrorsResponse(c, errors)
```

**Output:**
```json
{
  "status": 0,
  "status_code": 400,
  "message": "Validation error",
  "data": {
    "email": ["invalid format"],
    "password": ["too short", "must contain uppercase"]
  }
}
```

---

## Usage Examples

### Example 1: Register User with Validation

```go
func (h *Handler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.BadRequestResponse(c, "Invalid request body", gin.H{})
        return
    }

    // Sanitize input
    req.Email = utils.SanitizeEmail(req.Email)
    req.Phone = utils.SanitizePhone(req.Phone)

    // Validate email
    if !utils.ValidateEmail(req.Email) {
        utils.ValidationErrorResponse(c, "email", []string{"invalid format"})
        return
    }

    // Validate phone
    if !utils.ValidateIndonesiaPhone(req.Phone) {
        utils.ValidationErrorResponse(c, "phone", []string{"invalid Indonesia phone format"})
        return
    }

    // Validate password
    valid, errors := utils.ValidatePassword(req.Password)
    if !valid {
        utils.ValidationErrorResponse(c, "password", errors)
        return
    }

    // Validate full name
    valid, errMsg := utils.ValidateFullName(req.FullName)
    if !valid {
        utils.ValidationErrorResponse(c, "full_name", []string{errMsg})
        return
    }

    // Create user
    user, err := h.service.Register(c.Request.Context(), req)
    if err != nil {
        if errors.Is(err, service.ErrEmailExists) {
            utils.ConflictResponse(c, "Email already registered", gin.H{})
            return
        }
        utils.InternalErrorResponse(c, "Failed to register user")
        return
    }

    utils.CreatedResponse(c, "Registration success", user)
}
```

---

### Example 2: Multiple Field Validation

```go
func validateCreateContactRequest(req CreateContactRequest) (bool, map[string][]string) {
    errors := make(map[string][]string)

    // Validate full name
    if valid, errMsg := utils.ValidateFullName(req.FullName); !valid {
        errors["full_name"] = []string{errMsg}
    }

    // Validate phone
    if !utils.ValidatePhone(req.Phone) {
        errors["phone"] = []string{"invalid phone format"}
    }

    // Validate email (optional)
    if req.Email != "" && !utils.ValidateEmail(req.Email) {
        errors["email"] = []string{"invalid email format"}
    }

    return len(errors) == 0, errors
}

func (h *Handler) CreateContact(c *gin.Context) {
    var req CreateContactRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.BadRequestResponse(c, "Invalid request body", gin.H{})
        return
    }

    // Validate all fields
    valid, errors := validateCreateContactRequest(req)
    if !valid {
        utils.ValidationErrorsResponse(c, errors)
        return
    }

    // Create contact...
}
```

---

### Example 3: Normalize Phone Before Storage

```go
func (s *Service) CreateContact(ctx context.Context, userID uint, req CreateContactRequest) (*Contact, error) {
    // Normalize phone to +62 format
    normalizedPhone := utils.NormalizeIndonesiaPhone(req.Phone)
    
    // Sanitize email
    var email *string
    if req.Email != "" {
        sanitized := utils.SanitizeEmail(req.Email)
        email = &sanitized
    }

    contact := &Contact{
        UserID:   userID,
        FullName: req.FullName,
        Phone:    normalizedPhone,
        Email:    email,
    }

    // Save to database...
}
```

---

## Testing

Run validation tests:

```bash
go test ./internal/utils/... -v
```

**Test Coverage:**
- ✅ Email validation (10 test cases)
- ✅ Phone validation (9 test cases)
- ✅ Indonesia phone validation (7 test cases)
- ✅ Password validation (7 test cases)
- ✅ Full name validation (8 test cases)
- ✅ Email sanitization (3 test cases)
- ✅ Phone normalization (5 test cases)

**Total: 49 test cases, all passing**

---

## Best Practices

### 1. Always Sanitize User Input

```go
// ✅ Good
email := utils.SanitizeEmail(req.Email)
if !utils.ValidateEmail(email) {
    return errors.New("invalid email")
}

// ❌ Bad
if !utils.ValidateEmail(req.Email) {
    return errors.New("invalid email")
}
```

---

### 2. Use Consistent Response Format

```go
// ✅ Good - using helper
utils.NotFoundResponse(c, "User not found")

// ❌ Bad - manual response
c.JSON(404, gin.H{"error": "not found"})
```

---

### 3. Validate Before Service Layer

```go
// ✅ Good - validate in handler
func (h *Handler) Register(c *gin.Context) {
    // Validate input first
    if !utils.ValidateEmail(req.Email) {
        utils.ValidationErrorResponse(c, "email", []string{"invalid format"})
        return
    }
    
    // Then call service
    h.service.Register(...)
}

// ❌ Bad - validate in service
func (s *Service) Register(...) {
    // Service should focus on business logic, not input validation
}
```

---

### 4. Normalize Data for Consistency

```go
// ✅ Good - normalize before storage
phone := utils.NormalizeIndonesiaPhone(req.Phone)
// Stored as: +6281234567890

// ❌ Bad - store as-is
phone := req.Phone
// Stored as: 0812-3456-7890 (inconsistent)
```

---

### 5. Return Detailed Validation Errors

```go
// ✅ Good - specific errors
valid, errors := utils.ValidatePassword(password)
if !valid {
    utils.ValidationErrorResponse(c, "password", errors)
}
// Returns: ["must be at least 8 characters", "must contain uppercase"]

// ❌ Bad - generic error
if len(password) < 8 {
    utils.BadRequestResponse(c, "Invalid password", gin.H{})
}
```

---

## Integration with Handlers

Update your handlers to use utils:

```go
import (
    "user-service/internal/utils"
)

// Replace manual responses
// Before:
c.JSON(200, gin.H{"status": 1, "message": "Success", "data": data})

// After:
utils.OKResponse(c, "Success", data)

// Replace manual validation
// Before:
emailRegex := regexp.MustCompile(...)
if !emailRegex.MatchString(email) { ... }

// After:
if !utils.ValidateEmail(email) { ... }
```

---

## References

- [Email Validation RFC 5322](https://tools.ietf.org/html/rfc5322)
- [Phone Number Formats](https://en.wikipedia.org/wiki/E.164)
- [REST API Error Handling Best Practices](https://www.baeldung.com/rest-api-error-handling-best-practices)
