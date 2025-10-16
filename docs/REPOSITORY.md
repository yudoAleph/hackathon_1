# Repository Layer Documentation

## Overview

Repository layer bertanggung jawab untuk semua operasi database menggunakan GORM. Layer ini menyediakan abstraksi antara business logic dan database, memastikan separation of concerns yang baik.

## Architecture

```
repository/
├── repository.go       # Interface dan implementasi repository
└── repository_test.go  # Unit tests dengan sqlmock
```

## Custom Errors

Repository layer mendefinisikan error khusus untuk handling yang lebih baik:

```go
var (
    ErrNotFound         = errors.New("record not found")
    ErrDuplicateEmail   = errors.New("email already exists")
    ErrDuplicatePhone   = errors.New("phone number already exists")
    ErrInvalidID        = errors.New("invalid ID")
)
```

## User Repository

### Interface

```go
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id uint) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id uint) error
    CheckEmailExists(ctx context.Context, email string, excludeUserID uint) (bool, error)
}
```

### Operations

#### Create User

```go
user := &models.User{
    FullName: "John Doe",
    Email:    "john@example.com",
    Phone:    "1234567890",
    Password: "hashed_password",
}
err := userRepo.Create(ctx, user)
if err != nil {
    if errors.Is(err, repository.ErrDuplicateEmail) {
        // Handle duplicate email
    }
    // Handle other errors
}
```

**Error Cases:**
- `ErrDuplicateEmail`: Email sudah terdaftar
- Database errors: Wrapped dengan context

#### Get User by ID

```go
user, err := userRepo.GetByID(ctx, 1)
if err != nil {
    if errors.Is(err, repository.ErrNotFound) {
        // User tidak ditemukan
    }
    // Handle other errors
}
```

**Error Cases:**
- `ErrNotFound`: User dengan ID tersebut tidak ada
- Database errors: Wrapped dengan context

#### Get User by Email

```go
user, err := userRepo.GetByEmail(ctx, "john@example.com")
if err != nil {
    if errors.Is(err, repository.ErrNotFound) {
        // User tidak ditemukan
    }
    // Handle other errors
}
```

**Use Case:** Login, email validation

#### Update User

```go
user.FullName = "John Doe Updated"
user.Phone = "9876543210"
err := userRepo.Update(ctx, user)
if err != nil {
    if errors.Is(err, repository.ErrNotFound) {
        // User tidak ada
    }
    if errors.Is(err, repository.ErrDuplicateEmail) {
        // Email baru sudah digunakan
    }
}
```

**Notes:**
- Zero values tidak akan di-update (GORM behavior)
- Menggunakan `Updates()` untuk update multiple fields

#### Delete User

```go
err := userRepo.Delete(ctx, 1)
if err != nil {
    if errors.Is(err, repository.ErrNotFound) {
        // User tidak ada
    }
}
```

**Notes:**
- Soft delete (GORM default)
- Cascade delete contacts (handled by database FK constraints)

#### Check Email Exists

```go
exists, err := userRepo.CheckEmailExists(ctx, "new@example.com", 0)
if err != nil {
    // Handle error
}
if exists {
    // Email sudah digunakan
}

// Untuk update (exclude user saat ini)
exists, err := userRepo.CheckEmailExists(ctx, "new@example.com", userID)
```

**Use Case:** Validasi sebelum create/update

## Contact Repository

### Interface

```go
type ContactRepository interface {
    Create(ctx context.Context, contact *models.Contact) error
    GetByID(ctx context.Context, userID, contactID uint) (*models.Contact, error)
    Update(ctx context.Context, contact *models.Contact) error
    Delete(ctx context.Context, userID, contactID uint) error
    List(ctx context.Context, userID uint, req *models.ListContactsRequest) ([]models.Contact, int64, error)
    CheckPhoneExists(ctx context.Context, userID uint, phone string, excludeContactID uint) (bool, error)
}
```

### Operations

#### Create Contact

```go
email := "contact@example.com"
contact := &models.Contact{
    UserID:   1,
    FullName: "Jane Doe",
    Phone:    "1234567890",
    Email:    &email,
    Favorite: false,
}
err := contactRepo.Create(ctx, contact)
```

**Notes:**
- Email adalah optional (*string)
- UserID wajib ada dan valid (FK constraint)

#### Get Contact by ID

```go
contact, err := contactRepo.GetByID(ctx, userID, contactID)
if err != nil {
    if errors.Is(err, repository.ErrNotFound) {
        // Contact tidak ditemukan atau bukan milik user
    }
}
```

**Security:** Selalu verify userID untuk memastikan user hanya akses contact miliknya

#### Update Contact

```go
contact.FullName = "Jane Doe Updated"
contact.Favorite = true
err := contactRepo.Update(ctx, contact)
if err != nil {
    if errors.Is(err, repository.ErrNotFound) {
        // Contact tidak ada
    }
}
```

**Notes:**
- Where clause include userID untuk security
- Zero values tidak akan di-update

#### Delete Contact

```go
err := contactRepo.Delete(ctx, userID, contactID)
if err != nil {
    if errors.Is(err, repository.ErrNotFound) {
        // Contact tidak ada
    }
}
```

**Security:** Verify userID sebelum delete

#### List Contacts with Pagination

```go
favorite := true
req := &models.ListContactsRequest{
    Page:     1,
    Limit:    10,
    Search:   "john",           // Optional: search in name & phone
    Favorite: &favorite,        // Optional: filter by favorite
}

contacts, total, err := contactRepo.List(ctx, userID, req)
if err != nil {
    // Handle error
}

// Calculate pagination metadata
totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))
hasNextPage := req.Page < totalPages
hasPrevPage := req.Page > 1
```

**Features:**
- **Pagination:** Page dan Limit
- **Search:** Full-text search di `full_name` dan `phone` (LIKE pattern)
- **Filter:** Filter by favorite status (optional)
- **Ordering:** Default order by `created_at DESC` (newest first)

**Query Examples:**

1. **Simple pagination:**
```go
req := &models.ListContactsRequest{
    Page:  1,
    Limit: 20,
}
```

2. **Search contacts:**
```go
req := &models.ListContactsRequest{
    Page:   1,
    Limit:  10,
    Search: "089", // Find by phone prefix
}
```

3. **Filter favorites:**
```go
favorite := true
req := &models.ListContactsRequest{
    Page:     1,
    Limit:    10,
    Favorite: &favorite,
}
```

4. **Combined search + filter:**
```go
favorite := true
req := &models.ListContactsRequest{
    Page:     1,
    Limit:    10,
    Search:   "john",
    Favorite: &favorite,
}
```

**Performance Notes:**
- Index pada `user_id` untuk fast filtering
- Index pada `full_name` dan `phone` untuk search optimization
- LIMIT/OFFSET pagination (simple but may have performance issues on large datasets)

#### Check Phone Exists

```go
exists, err := contactRepo.CheckPhoneExists(ctx, userID, "1234567890", 0)
if err != nil {
    // Handle error
}
if exists {
    // Phone sudah ada di contacts user ini
}

// Untuk update (exclude contact saat ini)
exists, err := contactRepo.CheckPhoneExists(ctx, userID, "1234567890", contactID)
```

**Use Case:** Validasi duplikat phone per user

## Error Handling Patterns

### Service Layer Example

```go
user, err := r.userRepo.GetByID(ctx, userID)
if err != nil {
    if errors.Is(err, repository.ErrNotFound) {
        return nil, &ServiceError{
            Code:    "USER_NOT_FOUND",
            Message: "User not found",
            Status:  http.StatusNotFound,
        }
    }
    return nil, &ServiceError{
        Code:    "DATABASE_ERROR",
        Message: "Failed to get user",
        Status:  http.StatusInternalServerError,
        Err:     err,
    }
}
```

### Handler Layer Example

```go
err := s.service.CreateUser(ctx, req)
if err != nil {
    if errors.Is(err, repository.ErrDuplicateEmail) {
        return c.JSON(http.StatusConflict, models.ErrorResponse{
            Error: "Email already registered",
        })
    }
    return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
        Error: "Failed to create user",
    })
}
```

## Testing

Repository layer menggunakan `go-sqlmock` untuk unit testing tanpa database real.

### Setup Test

```go
func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
    db, mock, err := sqlmock.New()
    // ... setup GORM with mock DB
    return gormDB, mock, cleanup
}
```

### Example Test

```go
func TestUserRepository_GetByID(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := NewUserRepository(db)
    
    // Setup mock expectations
    rows := sqlmock.NewRows([]string{"id", "full_name", "email"}).
        AddRow(1, "John Doe", "john@example.com")
    
    mock.ExpectQuery("SELECT \\* FROM `users`").
        WithArgs(1).
        WillReturnRows(rows)
    
    // Execute
    user, err := repo.GetByID(context.Background(), 1)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "john@example.com", user.Email)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

### Running Tests

```bash
# Run all repository tests
go test ./internal/app/repository/...

# With coverage
go test -cover ./internal/app/repository/...

# Verbose output
go test -v ./internal/app/repository/...
```

## Best Practices

### 1. Always Use Context

```go
// ✅ Good
func (r *userRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).First(&user, id).Error
    return &user, err
}

// ❌ Bad - no context
func (r *userRepository) GetByID(id uint) (*models.User, error) {
    var user models.User
    err := r.db.First(&user, id).Error
    return &user, err
}
```

**Benefits:**
- Timeout control
- Cancellation support
- Request tracing

### 2. Wrap Errors with Context

```go
// ✅ Good
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// ❌ Bad - losing context
if err != nil {
    return err
}
```

### 3. Use Transactions for Multiple Operations

```go
func (r *userRepository) CreateUserWithProfile(ctx context.Context, user *models.User, profile *Profile) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        profile.UserID = user.ID
        if err := tx.Create(profile).Error; err != nil {
            return err
        }
        return nil
    })
}
```

### 4. Specific Error Checks

```go
// ✅ Good - specific error types
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, ErrNotFound
}

// ❌ Bad - checking error string
if err.Error() == "record not found" {
    return nil, ErrNotFound
}
```

### 5. Security in Contact Operations

```go
// ✅ Good - always verify userID
func (r *contactRepository) GetByID(ctx context.Context, userID, contactID uint) (*models.Contact, error) {
    var contact models.Contact
    err := r.db.WithContext(ctx).
        Where("id = ? AND user_id = ?", contactID, userID).
        First(&contact).Error
    // ...
}

// ❌ Bad - missing userID check
func (r *contactRepository) GetByID(ctx context.Context, contactID uint) (*models.Contact, error) {
    var contact models.Contact
    err := r.db.WithContext(ctx).First(&contact, contactID).Error
    // User bisa akses contact milik user lain!
}
```

### 6. Pagination Performance

```go
// ✅ Good - count before limit/offset
query := r.db.Model(&models.Contact{}).Where("user_id = ?", userID)
query.Count(&total)
query.Offset(offset).Limit(limit).Find(&contacts)

// ❌ Bad - count after limit
query := r.db.Model(&models.Contact{}).Where("user_id = ?", userID)
query.Offset(offset).Limit(limit).Find(&contacts)
query.Count(&total) // Wrong total!
```

## Database Indexes

Untuk performa optimal, pastikan indexes berikut ada:

```sql
-- Users
CREATE INDEX idx_users_email ON users(email);

-- Contacts
CREATE INDEX idx_contacts_user_id ON contacts(user_id);
CREATE INDEX idx_contacts_user_id_created_at ON contacts(user_id, created_at);
CREATE INDEX idx_contacts_full_name ON contacts(full_name);
CREATE INDEX idx_contacts_phone ON contacts(phone);
CREATE INDEX idx_contacts_favorite ON contacts(favorite);
```

## Migration to Repository Pattern

Jika mengubah dari direct DB access ke repository pattern:

### Before (Direct DB)
```go
func (s *Service) GetUser(id uint) (*models.User, error) {
    var user models.User
    err := s.db.First(&user, id).Error
    return &user, err
}
```

### After (Repository)
```go
func (s *Service) GetUser(ctx context.Context, id uint) (*models.User, error) {
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, repository.ErrNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return user, nil
}
```

**Benefits:**
- Easier to test (mock repository interface)
- Better error handling
- Separation of concerns
- Reusable queries

## Next Steps

Setelah repository layer selesai, implement:

1. **Service Layer** - Business logic dan validasi
2. **Handlers** - HTTP request handling
3. **Middleware** - Authentication, logging, rate limiting
4. **Integration Tests** - Test dengan real database

## References

- [GORM Documentation](https://gorm.io/docs/)
- [Go sqlmock](https://github.com/DATA-DOG/go-sqlmock)
- [Repository Pattern](https://martinfowler.com/eaaCatalog/repository.html)
