package repository

import (
	"context"
	"errors"
	"fmt"

	"user-service/internal/app/models"

	"gorm.io/gorm"
)

var (
	// ErrNotFound is returned when a record is not found
	ErrNotFound = errors.New("record not found")
	// ErrDuplicateEmail is returned when email already exists
	ErrDuplicateEmail = errors.New("email already exists")
	// ErrDuplicatePhone is returned when phone already exists in contacts
	ErrDuplicatePhone = errors.New("phone number already exists")
	// ErrInvalidID is returned when ID is invalid
	ErrInvalidID = errors.New("invalid ID")
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *models.User) error
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uint) (*models.User, error)
	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	// Update updates an existing user
	Update(ctx context.Context, user *models.User) error
	// Delete deletes a user by ID
	Delete(ctx context.Context, id uint) error
	// CheckEmailExists checks if email already exists
	CheckEmailExists(ctx context.Context, email string, excludeUserID uint) (bool, error)
}

// ContactRepository defines the interface for contact data operations
type ContactRepository interface {
	// Create creates a new contact
	Create(ctx context.Context, contact *models.Contact) error
	// GetByID retrieves a contact by ID and user ID
	GetByID(ctx context.Context, userID, contactID uint) (*models.Contact, error)
	// Update updates an existing contact
	Update(ctx context.Context, contact *models.Contact) error
	// Delete deletes a contact by ID and user ID
	Delete(ctx context.Context, userID, contactID uint) error
	// List retrieves contacts with pagination and filtering
	List(ctx context.Context, userID uint, req *models.ListContactsRequest) ([]models.Contact, int64, error)
	// CheckPhoneExists checks if phone already exists for a user
	CheckPhoneExists(ctx context.Context, userID uint, phone string, excludeContactID uint) (bool, error)
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || isDuplicateError(err) {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Model(user).Updates(user)
	if result.Error != nil {
		if isDuplicateError(result.Error) {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes a user by ID
func (r *userRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CheckEmailExists checks if email already exists
func (r *userRepository) CheckEmailExists(ctx context.Context, email string, excludeUserID uint) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email)
	if excludeUserID > 0 {
		query = query.Where("id != ?", excludeUserID)
	}
	err := query.Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return count > 0, nil
}

// contactRepository implements ContactRepository interface
type contactRepository struct {
	db *gorm.DB
}

// NewContactRepository creates a new ContactRepository instance
func NewContactRepository(db *gorm.DB) ContactRepository {
	return &contactRepository{db: db}
}

// Create creates a new contact
func (r *contactRepository) Create(ctx context.Context, contact *models.Contact) error {
	if err := r.db.WithContext(ctx).Create(contact).Error; err != nil {
		return fmt.Errorf("failed to create contact: %w", err)
	}
	return nil
}

// GetByID retrieves a contact by ID and user ID
func (r *contactRepository) GetByID(ctx context.Context, userID, contactID uint) (*models.Contact, error) {
	var contact models.Contact
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", contactID, userID).
		First(&contact).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}
	return &contact, nil
}

// Update updates an existing contact
func (r *contactRepository) Update(ctx context.Context, contact *models.Contact) error {
	result := r.db.WithContext(ctx).
		Model(contact).
		Where("id = ? AND user_id = ?", contact.ID, contact.UserID).
		Updates(contact)

	if result.Error != nil {
		return fmt.Errorf("failed to update contact: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes a contact by ID and user ID
func (r *contactRepository) Delete(ctx context.Context, userID, contactID uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", contactID, userID).
		Delete(&models.Contact{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete contact: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// List retrieves contacts with pagination and filtering
func (r *contactRepository) List(ctx context.Context, userID uint, req *models.ListContactsRequest) ([]models.Contact, int64, error) {
	var contacts []models.Contact
	var total int64

	// Build base query
	query := r.db.WithContext(ctx).Model(&models.Contact{}).Where("user_id = ?", userID)

	// Apply search filter
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("full_name LIKE ? OR phone LIKE ?", searchPattern, searchPattern)
	}

	// Apply favorite filter
	if req.Favorite != nil {
		query = query.Where("favorite = ?", *req.Favorite)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count contacts: %w", err)
	}

	// Apply pagination
	offset := (req.Page - 1) * req.Limit
	query = query.Offset(offset).Limit(req.Limit)

	// Order by created_at DESC (newest first)
	query = query.Order("created_at DESC")

	// Execute query
	if err := query.Find(&contacts).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list contacts: %w", err)
	}

	return contacts, total, nil
}

// CheckPhoneExists checks if phone already exists for a user
func (r *contactRepository) CheckPhoneExists(ctx context.Context, userID uint, phone string, excludeContactID uint) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Contact{}).
		Where("user_id = ? AND phone = ?", userID, phone)

	if excludeContactID > 0 {
		query = query.Where("id != ?", excludeContactID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check phone existence: %w", err)
	}
	return count > 0, nil
}

// isDuplicateError checks if error is a duplicate entry error
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	// MySQL duplicate entry error
	return contains(errMsg, "Duplicate entry") ||
		contains(errMsg, "duplicate key") ||
		contains(errMsg, "UNIQUE constraint failed")
}

// contains checks if string contains substring (case-insensitive helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				containsInner(s, substr)))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
