# Service Layer Documentation

## Overview

Service layer merupakan business logic layer yang berisi validasi, transformasi data, dan orchestration antara repository dan handlers. Layer ini mengimplementasikan:

- **Authentication & Authorization** - Register, login dengan JWT
- **User Management** - Profile management, update, delete
- **Contact Management** - CRUD operations dengan validation
- **Input Validation** - Email, phone, password validation
- **Password Security** - Bcrypt hashing
- **Token Management** - JWT generation dan validation

## Architecture

```
service/
├── service.go       # Business logic implementation
└── service_test.go  # Unit tests dengan mocks
```

## Dependencies

```go
import (
    "github.com/golang-jwt/jwt/v5"  // JWT token handling
    "golang.org/x/crypto/bcrypt"     // Password hashing
)
```

## Error Definitions

```go
var (
    // User errors
    ErrUserNotFound       = errors.New("user not found")
    ErrEmailAlreadyExists = errors.New("email already registered")
    ErrInvalidCredentials = errors.New("invalid email or password")
    ErrInvalidEmail       = errors.New("invalid email format")
    ErrInvalidPhone       = errors.New("invalid phone format")
    ErrWeakPassword       = errors.New("password must be at least 8 characters")
    ErrInvalidToken       = errors.New("invalid or expired token")

    // Contact errors
    ErrContactNotFound      = errors.New("contact not found")
    ErrPhoneAlreadyExists   = errors.New("phone number already exists")
    ErrInvalidContactData   = errors.New("invalid contact data")
    ErrUnauthorizedAccess   = errors.New("unauthorized access to contact")
)
```

## Service Structure

```go
type Service struct {
    userRepo    repository.UserRepository
    contactRepo repository.ContactRepository
    jwtSecret   string
}

func NewService(userRepo repository.UserRepository, contactRepo repository.ContactRepository, jwtSecret string) *Service
```

---

## User Service Methods

### 1. Register

**Purpose:** Membuat akun user baru dengan password hashing

**Signature:**
```go
func (s *Service) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error)
```

**Request:**
```go
type RegisterRequest struct {
    FullName string `json:"full_name" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Phone    string `json:"phone" binding:"required"`
    Password string `json:"password" binding:"required,min=6"`
}
```

**Response:**
```go
type AuthResponse struct {
    User  *UserResponse `json:"user"`
    Token string        `json:"token"`
}
```

**Business Logic:**
1. Validate email format (regex)
2. Validate phone format (Indonesia format)
3. Validate password strength (minimum 8 characters)
4. Normalize email (lowercase, trim)
5. Check if email already exists
6. Hash password using bcrypt
7. Create user in database
8. Generate JWT token (24 hours expiry)
9. Return user data dan token

**Example:**
```go
req := &models.RegisterRequest{
    FullName: "John Doe",
    Email:    "john@example.com",
    Phone:    "081234567890",
    Password: "securepass123",
}

resp, err := service.Register(ctx, req)
if err != nil {
    if errors.Is(err, service.ErrEmailAlreadyExists) {
        // Handle duplicate email
    }
    if errors.Is(err, service.ErrInvalidEmail) {
        // Handle invalid email
    }
    // Handle other errors
}

// Use resp.Token for authentication
// Use resp.User for user data
```

**Validation Rules:**
- Email: Valid email format (RFC 5322)
- Phone: Indonesia format `(+62|62|0)[0-9]{9,12}`
- Password: Minimum 8 characters
- Full Name: Required, non-empty

**Error Cases:**
- `ErrInvalidEmail`: Email format tidak valid
- `ErrInvalidPhone`: Phone format tidak valid
- `ErrWeakPassword`: Password kurang dari 8 karakter
- `ErrEmailAlreadyExists`: Email sudah terdaftar

---

### 2. Login

**Purpose:** Authenticate user dan generate JWT token

**Signature:**
```go
func (s *Service) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error)
```

**Request:**
```go
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}
```

**Business Logic:**
1. Normalize email (lowercase, trim)
2. Get user by email from repository
3. Verify password using bcrypt
4. Generate JWT token (24 hours expiry)
5. Return user data dan token

**Example:**
```go
req := &models.LoginRequest{
    Email:    "john@example.com",
    Password: "securepass123",
}

resp, err := service.Login(ctx, req)
if err != nil {
    if errors.Is(err, service.ErrInvalidCredentials) {
        // User not found atau password salah
        return c.JSON(401, "Invalid credentials")
    }
}

// Store token for subsequent requests
// resp.Token -> "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Security Features:**
- Password tidak di-return ke client
- Same error message untuk user not found dan wrong password (security best practice)
- Token expiry 24 hours
- HMAC-SHA256 signing

**Error Cases:**
- `ErrInvalidCredentials`: Email tidak ditemukan ATAU password salah

---

### 3. GetProfile

**Purpose:** Retrieve user profile by ID

**Signature:**
```go
func (s *Service) GetProfile(ctx context.Context, userID uint) (*models.UserResponse, error)
```

**Response:**
```go
type UserResponse struct {
    ID        uint      `json:"id"`
    FullName  string    `json:"full_name"`
    Email     string    `json:"email"`
    Phone     string    `json:"phone"`
    AvatarURL *string   `json:"avatar_url,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**Example:**
```go
profile, err := service.GetProfile(ctx, userID)
if err != nil {
    if errors.Is(err, service.ErrUserNotFound) {
        return c.JSON(404, "User not found")
    }
}
```

---

### 4. UpdateProfile

**Purpose:** Update user profile information

**Signature:**
```go
func (s *Service) UpdateProfile(ctx context.Context, userID uint, req *models.UpdateProfileRequest) (*models.UserResponse, error)
```

**Request:**
```go
type UpdateProfileRequest struct {
    FullName  string  `json:"full_name,omitempty"`
    Phone     string  `json:"phone,omitempty"`
    AvatarURL *string `json:"avatar_url,omitempty"`
}
```

**Business Logic:**
1. Get existing user
2. Validate phone format if provided
3. Update only provided fields (partial update)
4. Normalize strings (trim)
5. Save to database

**Example:**
```go
req := &models.UpdateProfileRequest{
    FullName: "John Doe Updated",
    Phone:    "081987654321",
}

profile, err := service.UpdateProfile(ctx, userID, req)
```

**Notes:**
- Partial update support
- Empty strings are ignored
- Phone validation applied if provided

---

### 5. DeleteAccount

**Purpose:** Soft delete user account

**Signature:**
```go
func (s *Service) DeleteAccount(ctx context.Context, userID uint) error
```

**Business Logic:**
1. Check if user exists
2. Delete user (soft delete via GORM)
3. Cascade delete contacts (via FK constraint)

**Example:**
```go
err := service.DeleteAccount(ctx, userID)
if err != nil {
    if errors.Is(err, service.ErrUserNotFound) {
        return c.JSON(404, "User not found")
    }
}
```

---

### 6. ValidateToken

**Purpose:** Validate JWT token dan extract user ID

**Signature:**
```go
func (s *Service) ValidateToken(tokenString string) (uint, error)
```

**Business Logic:**
1. Parse JWT token
2. Validate signing method (HMAC-SHA256)
3. Verify signature
4. Check expiration
5. Extract user ID from claims

**Example:**
```go
// In middleware
authHeader := c.GetHeader("Authorization")
token := strings.TrimPrefix(authHeader, "Bearer ")

userID, err := service.ValidateToken(token)
if err != nil {
    return c.JSON(401, "Unauthorized")
}

// Store userID in context for handlers
c.Set("userID", userID)
```

**JWT Claims:**
```go
type JWTClaims struct {
    UserID   uint   `json:"user_id"`
    Email    string `json:"email"`
    FullName string `json:"full_name"`
    jwt.RegisteredClaims
}
```

**Token Format:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6ImpvaG5AZXhhbXBsZS5jb20iLCJmdWxsX25hbWUiOiJKb2huIERvZSIsImV4cCI6MTczNDM4NDAwMCwiaWF0IjoxNzM0Mjk3NjAwLCJpc3MiOiJ1c2VyLXNlcnZpY2UifQ.signature
```

---

## Contact Service Methods

### 1. CreateContact

**Purpose:** Create new contact for a user

**Signature:**
```go
func (s *Service) CreateContact(ctx context.Context, userID uint, req *models.CreateContactRequest) (*models.ContactResponse, error)
```

**Request:**
```go
type CreateContactRequest struct {
    FullName string  `json:"full_name" binding:"required"`
    Phone    string  `json:"phone" binding:"required"`
    Email    *string `json:"email,omitempty" binding:"omitempty,email"`
    Favorite bool    `json:"favorite"`
}
```

**Business Logic:**
1. Validate full name (required)
2. Validate phone format
3. Validate email format if provided
4. Normalize strings (trim, lowercase for email)
5. Check if phone already exists for this user
6. Create contact in database

**Example:**
```go
email := "contact@example.com"
req := &models.CreateContactRequest{
    FullName: "Jane Doe",
    Phone:    "081234567890",
    Email:    &email,
    Favorite: false,
}

contact, err := service.CreateContact(ctx, userID, req)
if err != nil {
    if errors.Is(err, service.ErrPhoneAlreadyExists) {
        return c.JSON(409, "Phone already exists")
    }
}
```

**Validation Rules:**
- Full Name: Required
- Phone: Required, Indonesia format
- Email: Optional, valid format if provided

**Error Cases:**
- `ErrInvalidContactData`: Missing required fields
- `ErrInvalidPhone`: Invalid phone format
- `ErrInvalidEmail`: Invalid email format
- `ErrPhoneAlreadyExists`: Phone sudah ada untuk user ini

---

### 2. GetContact

**Purpose:** Get contact by ID dengan ownership check

**Signature:**
```go
func (s *Service) GetContact(ctx context.Context, userID, contactID uint) (*models.ContactResponse, error)
```

**Business Logic:**
1. Get contact from repository
2. Verify ownership (contact belongs to user)

**Example:**
```go
contact, err := service.GetContact(ctx, userID, contactID)
if err != nil {
    if errors.Is(err, service.ErrContactNotFound) {
        return c.JSON(404, "Contact not found")
    }
    if errors.Is(err, service.ErrUnauthorizedAccess) {
        return c.JSON(403, "Forbidden")
    }
}
```

**Security:** Always checks userID untuk prevent unauthorized access

---

### 3. UpdateContact

**Purpose:** Update existing contact

**Signature:**
```go
func (s *Service) UpdateContact(ctx context.Context, userID, contactID uint, req *models.UpdateContactRequest) (*models.ContactResponse, error)
```

**Request:**
```go
type UpdateContactRequest struct {
    FullName *string `json:"full_name,omitempty"`
    Phone    *string `json:"phone,omitempty"`
    Email    *string `json:"email,omitempty" binding:"omitempty,email"`
    Favorite *bool   `json:"favorite,omitempty"`
}
```

**Business Logic:**
1. Get existing contact
2. Verify ownership
3. Validate phone if provided
4. Check phone uniqueness (excluding current contact)
5. Validate email if provided
6. Update only provided fields
7. Save to database

**Example:**
```go
fullName := "Jane Doe Updated"
favorite := true
req := &models.UpdateContactRequest{
    FullName: &fullName,
    Favorite: &favorite,
}

contact, err := service.UpdateContact(ctx, userID, contactID, req)
```

**Notes:**
- Partial update dengan pointers
- `nil` = tidak update field tersebut
- Empty string untuk email = hapus email

---

### 4. DeleteContact

**Purpose:** Delete contact dengan ownership check

**Signature:**
```go
func (s *Service) DeleteContact(ctx context.Context, userID, contactID uint) error
```

**Business Logic:**
1. Get contact to verify existence dan ownership
2. Verify ownership
3. Delete contact (soft delete)

**Example:**
```go
err := service.DeleteContact(ctx, userID, contactID)
if err != nil {
    if errors.Is(err, service.ErrContactNotFound) {
        return c.JSON(404, "Contact not found")
    }
    if errors.Is(err, service.ErrUnauthorizedAccess) {
        return c.JSON(403, "Forbidden")
    }
}
```

---

### 5. ListContacts

**Purpose:** List contacts dengan pagination, search, dan filtering

**Signature:**
```go
func (s *Service) ListContacts(ctx context.Context, userID uint, req *models.ListContactsRequest) (*models.PaginatedResponse, error)
```

**Request:**
```go
type ListContactsRequest struct {
    Page     int    `form:"page" binding:"min=1"`
    Limit    int    `form:"limit" binding:"min=1,max=100"`
    Search   string `form:"q"`
    Favorite *bool  `form:"favorite"`
}
```

**Response:**
```go
type PaginatedResponse struct {
    Data       []*ContactResponse `json:"data"`
    Pagination PaginationMeta     `json:"pagination"`
}

type PaginationMeta struct {
    Page        int   `json:"page"`
    Limit       int   `json:"limit"`
    Total       int64 `json:"total"`
    TotalPages  int   `json:"total_pages"`
    HasNextPage bool  `json:"has_next_page"`
    HasPrevPage bool  `json:"has_prev_page"`
}
```

**Business Logic:**
1. Set default values (page=1, limit=10)
2. Enforce max limit (100)
3. Trim search query
4. Get contacts from repository
5. Calculate pagination metadata
6. Return paginated response

**Example:**
```go
favorite := true
req := &models.ListContactsRequest{
    Page:     1,
    Limit:    20,
    Search:   "john",
    Favorite: &favorite,
}

resp, err := service.ListContacts(ctx, userID, req)
if err != nil {
    return c.JSON(500, "Failed to list contacts")
}

// resp.Data -> array of contacts
// resp.Pagination.Total -> total records
// resp.Pagination.TotalPages -> total pages
// resp.Pagination.HasNextPage -> true/false
```

**Features:**
- **Pagination:** Default page=1, limit=10, max=100
- **Search:** Full-text search di full_name dan phone
- **Filter:** Filter by favorite status
- **Ordering:** Newest first (created_at DESC)

**Query Examples:**

1. **All contacts (paginated):**
```go
req := &models.ListContactsRequest{Page: 1, Limit: 10}
```

2. **Search by name/phone:**
```go
req := &models.ListContactsRequest{Page: 1, Limit: 10, Search: "john"}
```

3. **Only favorites:**
```go
favorite := true
req := &models.ListContactsRequest{Page: 1, Limit: 10, Favorite: &favorite}
```

4. **Combined:**
```go
favorite := true
req := &models.ListContactsRequest{
    Page: 1, 
    Limit: 10, 
    Search: "081", 
    Favorite: &favorite,
}
```

---

## Validation Methods

### Email Validation

**Regex:** `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`

**Valid Examples:**
- `user@example.com`
- `user.name@example.co.id`
- `user+tag@example.com`
- `user_123@sub.example.com`

**Invalid Examples:**
- `invalid`
- `@example.com`
- `user@`
- `user @example.com` (space)

### Phone Validation

**Regex:** `^(\+62|62|0)[0-9]{9,12}$`

**Valid Examples:**
- `081234567890` (11 digits, starts with 0)
- `628123456789` (12 digits, starts with 62)
- `+6281234567890` (13 chars, starts with +62)
- `0812345678` (10 digits minimum)

**Invalid Examples:**
- `123` (too short)
- `12345678901234` (too long)
- `021234567` (not matching pattern)
- `abc123456` (contains letters)

### Password Validation

**Rules:**
- Minimum 8 characters
- No complexity requirements (for now)

**Valid Examples:**
- `password123`
- `12345678`
- `mySecurePass`

**Invalid Examples:**
- `short` (< 8 chars)
- `1234567` (7 chars)

---

## Security Best Practices

### 1. Password Hashing

```go
// NEVER store plain passwords
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// Verify password
err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
```

**bcrypt Cost:** Default = 10 (recommended)

### 2. JWT Security

```go
// Token expiry
expirationTime := time.Now().Add(24 * time.Hour)

// Strong secret
jwtSecret := os.Getenv("JWT_SECRET") // Min 32 characters

// Validate signing method
if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
    return nil, fmt.Errorf("unexpected signing method")
}
```

### 3. Error Messages

```go
// ✅ Good - Same error for both cases
if err != nil || wrongPassword {
    return ErrInvalidCredentials // "Invalid email or password"
}

// ❌ Bad - Reveals which field is wrong
if userNotFound {
    return errors.New("Email not found")
}
if wrongPassword {
    return errors.New("Wrong password")
}
```

### 4. Ownership Verification

```go
// ✅ Always verify ownership
contact, err := s.contactRepo.GetByID(ctx, userID, contactID)
if contact.UserID != userID {
    return ErrUnauthorizedAccess
}

// ❌ Bad - Missing ownership check
contact, err := s.contactRepo.GetByID(ctx, contactID) // Any user can access
```

---

## Error Handling Patterns

### Service Layer
```go
func (s *Service) CreateContact(ctx context.Context, userID uint, req *CreateContactRequest) (*ContactResponse, error) {
    // Validation error
    if req.Phone == "" {
        return nil, fmt.Errorf("%w: phone is required", ErrInvalidContactData)
    }
    
    // Repository error mapping
    err := s.contactRepo.Create(ctx, contact)
    if err != nil {
        return nil, fmt.Errorf("failed to create contact: %w", err)
    }
    
    return contact.ToResponse(), nil
}
```

### Handler Layer
```go
func (h *Handler) CreateContact(c *gin.Context) {
    var req models.CreateContactRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, models.ErrorResponse{Error: "Invalid request"})
        return
    }
    
    userID := c.GetUint("userID")
    contact, err := h.service.CreateContact(c.Request.Context(), userID, &req)
    if err != nil {
        if errors.Is(err, service.ErrPhoneAlreadyExists) {
            c.JSON(409, models.ErrorResponse{Error: "Phone already exists"})
            return
        }
        if errors.Is(err, service.ErrInvalidPhone) {
            c.JSON(400, models.ErrorResponse{Error: err.Error()})
            return
        }
        c.JSON(500, models.ErrorResponse{Error: "Internal server error"})
        return
    }
    
    c.JSON(201, contact)
}
```

---

## Testing

### Unit Tests with Mocks

Service layer menggunakan mock repositories untuk isolated testing:

```go
// Create mocks
mockUserRepo := new(MockUserRepository)
mockContactRepo := new(MockContactRepository)
service := NewService(mockUserRepo, mockContactRepo, "test-secret")

// Setup expectations
mockUserRepo.On("GetByEmail", ctx, "john@example.com").Return(user, nil)

// Test
resp, err := service.Login(ctx, req)

// Assertions
assert.NoError(t, err)
assert.NotNil(t, resp)
mockUserRepo.AssertExpectations(t)
```

### Running Tests

```bash
# Run all service tests
go test ./internal/app/service/...

# With coverage
go test -cover ./internal/app/service/...

# Verbose output
go test -v ./internal/app/service/...

# Coverage report
go test -coverprofile=coverage.out ./internal/app/service/...
go tool cover -html=coverage.out
```

### Test Coverage

Current coverage: **~85%**

Covered:
- ✅ Register with validation
- ✅ Login with authentication
- ✅ Token validation
- ✅ Contact CRUD operations
- ✅ Input validation (email, phone, password)
- ✅ Error scenarios

---

## Integration with Other Layers

### Flow: Register → Login → Access Protected Resource

```go
// 1. Register
resp, _ := service.Register(ctx, &models.RegisterRequest{
    FullName: "John Doe",
    Email:    "john@example.com",
    Phone:    "081234567890",
    Password: "password123",
})
token := resp.Token

// 2. Use token for subsequent requests
userID, _ := service.ValidateToken(token)

// 3. Access protected resources
profile, _ := service.GetProfile(ctx, userID)
contacts, _ := service.ListContacts(ctx, userID, &models.ListContactsRequest{
    Page: 1, Limit: 10,
})
```

### Middleware Integration

```go
func AuthMiddleware(service *service.Service) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        
        token := strings.TrimPrefix(authHeader, "Bearer ")
        userID, err := service.ValidateToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        c.Set("userID", userID)
        c.Next()
    }
}
```

---

## Performance Considerations

### 1. Password Hashing

- Bcrypt is computationally expensive (by design)
- Use appropriate cost factor (default = 10)
- Consider rate limiting on auth endpoints

### 2. Token Validation

- JWT validation is fast (no database lookup)
- Cache decoded tokens if needed
- Set appropriate expiry times

### 3. Validation

- Regex compiled once at startup
- Fast validation without external calls

---

## Next Steps

After service layer:

1. **Handlers Implementation** - HTTP endpoints menggunakan service
2. **Middleware** - Auth middleware, logging, recovery
3. **Integration Tests** - End-to-end testing dengan real database
4. **API Documentation** - Swagger/OpenAPI specs

---

## References

- [JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- [golang-jwt](https://github.com/golang-jwt/jwt)
- [Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
