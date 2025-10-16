# Models Documentation

## Overview

Models package contains all data structures and entities used in the application, including database models (GORM entities) and DTOs (Data Transfer Objects) for API requests and responses.

## Database Models

### User Model

Represents a user in the system with authentication credentials and profile information.

```go
type User struct {
    ID        uint      // Primary key, auto-increment
    FullName  string    // User's full name (required, indexed)
    Email     string    // User's email (required, unique, indexed)
    Phone     string    // User's phone number (required, indexed)
    Password  string    // Hashed password (not exposed in JSON)
    AvatarURL *string   // Optional avatar URL
    CreatedAt time.Time // Auto-managed creation timestamp
    UpdatedAt time.Time // Auto-managed update timestamp
    
    // Relations
    Contacts []Contact // One-to-many relationship with contacts
}
```

**Database Table:** `users`

**Indexes:**
- `idx_users_full_name` - Full name lookup
- `idx_users_email` - Email lookup (unique)
- `idx_users_phone` - Phone number lookup
- `idx_users_created_at` - Time-based queries

### Contact Model

Represents a contact entry belonging to a user.

```go
type Contact struct {
    ID        uint      // Primary key, auto-increment
    UserID    uint      // Foreign key to users table
    FullName  string    // Contact's full name (required, indexed)
    Phone     string    // Contact's phone number (required, indexed)
    Email     *string   // Optional email address (indexed)
    Favorite  bool      // Favorite flag (default: false, indexed)
    CreatedAt time.Time // Auto-managed creation timestamp
    UpdatedAt time.Time // Auto-managed update timestamp
    
    // Relations
    User User // Many-to-one relationship with user
}
```

**Database Table:** `contacts`

**Indexes:**
- `idx_contacts_user_id` - User-based queries
- `idx_contacts_full_name` - Name search
- `idx_contacts_phone` - Phone search
- `idx_contacts_email` - Email search
- `idx_contacts_favorite` - Favorite filter
- `idx_contacts_created_at` - Time-based queries
- `idx_contacts_user_favorite` - Composite index for user + favorite queries
- `idx_contacts_user_created` - Composite index for user + created_at queries

**Foreign Keys:**
- `fk_contacts_user_id` - References `users(id)` with CASCADE delete

## Response Models

### UserResponse

Clean user data for API responses (excludes password).

```go
type UserResponse struct {
    ID        uint
    FullName  string
    Email     string
    Phone     string
    AvatarURL *string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### ContactResponse

Clean contact data for API responses.

```go
type ContactResponse struct {
    ID        uint
    UserID    uint
    FullName  string
    Phone     string
    Email     *string
    Favorite  bool
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

## Request DTOs

### Authentication

#### RegisterRequest
```go
type RegisterRequest struct {
    FullName string // Required
    Email    string // Required, valid email
    Phone    string // Required
    Password string // Required, min 6 characters
}
```

#### LoginRequest
```go
type LoginRequest struct {
    Email    string // Required, valid email
    Password string // Required, min 6 characters
}
```

### User Profile

#### UpdateUserRequest
```go
type UpdateUserRequest struct {
    FullName  string  // Required
    Phone     string  // Required
    AvatarURL *string // Optional
}
```

### Contacts

#### CreateContactRequest
```go
type CreateContactRequest struct {
    FullName string  // Required
    Phone    string  // Required
    Email    *string // Optional, valid email if provided
    Favorite bool    // Default: false
}
```

#### UpdateContactRequest
```go
type UpdateContactRequest struct {
    FullName string  // Required
    Phone    string  // Required
    Email    *string // Optional, valid email if provided
    Favorite bool
}
```

#### ListContactsRequest
```go
type ListContactsRequest struct {
    Page     int    // Min: 1
    Limit    int    // Min: 1, Max: 100
    Search   string // Search query for full_name or phone
    Favorite *bool  // Filter by favorite status
}
```

## Standard API Responses

### Response
Standard success response:
```go
type Response struct {
    Status     int         // 1 for success, 0 for error
    StatusCode int         // HTTP status code
    Message    string      // Human-readable message
    Data       interface{} // Response data
}
```

### PaginatedResponse
Response with pagination:
```go
type PaginatedResponse struct {
    Status     int
    StatusCode int
    Message    string
    Data       interface{}
    Pagination PaginationMeta
}

type PaginationMeta struct {
    Page       int
    Limit      int
    Total      int64
    TotalPages int
}
```

### ErrorResponse
Error response:
```go
type ErrorResponse struct {
    Status     int         // Always 0 for errors
    StatusCode int         // HTTP error code
    Message    string      // Error message
    Error      interface{} // Error details
}
```

### AuthResponse
Authentication response with token:
```go
type AuthResponse struct {
    User  UserResponse
    Token string // JWT token
}
```

## Usage Examples

### Creating a User
```go
user := models.User{
    FullName: "John Doe",
    Email:    "john@example.com",
    Phone:    "+1234567890",
    Password: hashedPassword, // Should be hashed
}
db.Create(&user)
```

### Creating a Contact
```go
email := "contact@example.com"
contact := models.Contact{
    UserID:   user.ID,
    FullName: "Jane Smith",
    Phone:    "+0987654321",
    Email:    &email,
    Favorite: false,
}
db.Create(&contact)
```

### Converting to Response
```go
// User to UserResponse
userResponse := user.ToResponse()

// Contact to ContactResponse
contactResponse := contact.ToResponse()
```

### Querying with Relations
```go
// Load user with contacts
var user models.User
db.Preload("Contacts").First(&user, userID)

// Load contact with user
var contact models.Contact
db.Preload("User").First(&contact, contactID)
```

## Validation Tags

Models use Gin's binding tags for validation:

- `required` - Field cannot be empty
- `email` - Must be valid email format
- `min=X` - Minimum length/value
- `max=X` - Maximum length/value
- `omitempty` - Field is optional

## JSON Tags

- Field names use `snake_case` in JSON
- Password field is excluded with `json:"-"`
- Optional fields use `json:"field,omitempty"`

## GORM Tags

Common GORM tags used:

- `primaryKey` - Primary key field
- `autoIncrement` - Auto-incrementing value
- `not null` - Required field
- `uniqueIndex` - Unique constraint with index
- `index` - Create index for performance
- `foreignKey` - Define foreign key
- `constraint:OnDelete:CASCADE` - Cascade delete
- `autoCreateTime` - Auto-set on creation
- `autoUpdateTime` - Auto-update on modification

## Migration Compatibility

These models are designed to match the database schema defined in:
- `internal/app/migrations/migrations.go`
- Migration 001: `create_users_table`
- Migration 002: `create_contacts_table`

The GORM tags and field types correspond exactly to the SQL CREATE TABLE statements in the migrations.
