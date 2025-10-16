package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"user-service/internal/app/models"
	"user-service/internal/app/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

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
	ErrContactNotFound    = errors.New("contact not found")
	ErrPhoneAlreadyExists = errors.New("phone number already exists")
	ErrInvalidContactData = errors.New("invalid contact data")
	ErrUnauthorizedAccess = errors.New("unauthorized access to contact")
)

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Phone validation regex (International format)
// Supports: +6281234567890, +1234567890, 081234567890, etc.
// Minimum 10 characters total (including country code), maximum 16
var phoneRegex = regexp.MustCompile(`^\+?[0-9]{10,15}$`)

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	jwt.RegisteredClaims
}

type Service struct {
	userRepo    repository.UserRepository
	contactRepo repository.ContactRepository
	jwtSecret   string
}

func NewService(userRepo repository.UserRepository, contactRepo repository.ContactRepository, jwtSecret string) *Service {
	return &Service{
		userRepo:    userRepo,
		contactRepo: contactRepo,
		jwtSecret:   jwtSecret,
	}
}

// ============================================================================
// USER SERVICE METHODS
// ============================================================================

// Register creates a new user account with hashed password
func (s *Service) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Validate input
	if err := s.validateEmail(req.Email); err != nil {
		return nil, err
	}

	// Validate phone only if provided
	if req.Phone != nil && *req.Phone != "" {
		if err := s.validatePhone(*req.Phone); err != nil {
			return nil, err
		}
	}

	if err := s.validatePassword(req.Password); err != nil {
		return nil, err
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)

	// Normalize phone if provided
	if req.Phone != nil {
		trimmed := strings.TrimSpace(*req.Phone)
		req.Phone = &trimmed
	}

	// Check if email already exists
	exists, err := s.userRepo.CheckEmailExists(ctx, req.Email, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		FullName: req.FullName,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: hashedPassword,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// Login authenticates a user and returns JWT token
func (s *Service) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if err := s.verifyPassword(user.Password, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// GetProfile retrieves user profile by ID
func (s *Service) GetProfile(ctx context.Context, userID uint) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user.ToResponse(), nil
}

// UpdateProfile updates user profile information
func (s *Service) UpdateProfile(ctx context.Context, userID uint, req *models.UpdateProfileRequest) (*models.UserResponse, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields if provided
	if req.FullName != "" {
		user.FullName = strings.TrimSpace(req.FullName)
	}

	if req.Phone != "" {
		if err := s.validatePhone(req.Phone); err != nil {
			return nil, err
		}
		trimmed := strings.TrimSpace(req.Phone)
		user.Phone = &trimmed
	}

	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}

	// Update in database
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user.ToResponse(), nil
}

// DeleteAccount deletes user account
func (s *Service) DeleteAccount(ctx context.Context, userID uint) error {
	// Check if user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Delete user (cascade delete contacts via FK)
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ValidateToken validates JWT token and returns user ID
func (s *Service) ValidateToken(tokenString string) (uint, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return 0, ErrInvalidToken
	}

	return claims.UserID, nil
}

// ============================================================================
// CONTACT SERVICE METHODS
// ============================================================================

// CreateContact creates a new contact for a user
func (s *Service) CreateContact(ctx context.Context, userID uint, req *models.CreateContactRequest) (*models.ContactResponse, error) {
	// Validate input
	if req.FullName == "" {
		return nil, fmt.Errorf("%w: full name is required", ErrInvalidContactData)
	}
	if req.Phone == "" {
		return nil, fmt.Errorf("%w: phone is required", ErrInvalidContactData)
	}
	if err := s.validatePhone(req.Phone); err != nil {
		return nil, err
	}

	// Validate email if provided
	if req.Email != nil && *req.Email != "" {
		if err := s.validateEmail(*req.Email); err != nil {
			return nil, err
		}
		normalized := strings.ToLower(strings.TrimSpace(*req.Email))
		req.Email = &normalized
	}

	// Normalize fields
	req.FullName = strings.TrimSpace(req.FullName)
	req.Phone = strings.TrimSpace(req.Phone)

	// Check if phone already exists for this user
	exists, err := s.contactRepo.CheckPhoneExists(ctx, userID, req.Phone, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check phone: %w", err)
	}
	if exists {
		return nil, ErrPhoneAlreadyExists
	}

	// Create contact
	contact := &models.Contact{
		UserID:   userID,
		FullName: req.FullName,
		Phone:    req.Phone,
		Email:    req.Email,
		Favorite: false,
	}

	if err := s.contactRepo.Create(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	return contact.ToResponse(), nil
}

// GetContact retrieves a contact by ID
func (s *Service) GetContact(ctx context.Context, userID, contactID uint) (*models.ContactResponse, error) {
	contact, err := s.contactRepo.GetByID(ctx, userID, contactID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrContactNotFound
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	// Verify ownership
	if contact.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	return contact.ToResponse(), nil
}

// UpdateContact updates an existing contact
func (s *Service) UpdateContact(ctx context.Context, userID, contactID uint, req *models.UpdateContactRequest) (*models.ContactResponse, error) {
	// Get existing contact
	contact, err := s.contactRepo.GetByID(ctx, userID, contactID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrContactNotFound
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	// Verify ownership
	if contact.UserID != userID {
		return nil, ErrUnauthorizedAccess
	}

	// Update fields if provided
	if req.FullName != nil {
		contact.FullName = strings.TrimSpace(*req.FullName)
	}

	if req.Phone != nil {
		if err := s.validatePhone(*req.Phone); err != nil {
			return nil, err
		}
		phone := strings.TrimSpace(*req.Phone)

		// Check if new phone already exists (excluding current contact)
		exists, err := s.contactRepo.CheckPhoneExists(ctx, userID, phone, contactID)
		if err != nil {
			return nil, fmt.Errorf("failed to check phone: %w", err)
		}
		if exists {
			return nil, ErrPhoneAlreadyExists
		}
		contact.Phone = phone
	}

	if req.Email != nil {
		if *req.Email != "" {
			if err := s.validateEmail(*req.Email); err != nil {
				return nil, err
			}
			normalized := strings.ToLower(strings.TrimSpace(*req.Email))
			contact.Email = &normalized
		} else {
			contact.Email = nil
		}
	}

	if req.Favorite != nil {
		contact.Favorite = *req.Favorite
	}

	// Update in database
	if err := s.contactRepo.Update(ctx, contact); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrContactNotFound
		}
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	return contact.ToResponse(), nil
}

// DeleteContact deletes a contact
func (s *Service) DeleteContact(ctx context.Context, userID, contactID uint) error {
	// Check if contact exists and belongs to user
	contact, err := s.contactRepo.GetByID(ctx, userID, contactID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrContactNotFound
		}
		return fmt.Errorf("failed to get contact: %w", err)
	}

	// Verify ownership
	if contact.UserID != userID {
		return ErrUnauthorizedAccess
	}

	// Delete contact
	if err := s.contactRepo.Delete(ctx, userID, contactID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrContactNotFound
		}
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	return nil
}

// ListContacts retrieves contacts with pagination and filtering
func (s *Service) ListContacts(ctx context.Context, userID uint, req *models.ListContactsRequest) (*models.PaginatedResponse, error) {
	// Set default values
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100 // Max limit
	}

	// Trim search query
	if req.Search != "" {
		req.Search = strings.TrimSpace(req.Search)
	}

	// Get contacts from repository
	contacts, total, err := s.contactRepo.List(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}

	// Convert to response format
	contactResponses := make([]*models.ContactResponse, len(contacts))
	for i, contact := range contacts {
		contactResponses[i] = contact.ToResponse()
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return &models.PaginatedResponse{
		Data: contactResponses,
		Pagination: models.PaginationMeta{
			Page:        req.Page,
			Limit:       req.Limit,
			Total:       total,
			TotalPages:  totalPages,
			HasNextPage: req.Page < totalPages,
			HasPrevPage: req.Page > 1,
		},
	}, nil
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// validateEmail validates email format
func (s *Service) validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidEmail)
	}
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// validatePhone validates phone format (International)
// Supports formats: +6281234567890, +1234567890, 081234567890, etc.
func (s *Service) validatePhone(phone string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return fmt.Errorf("%w: phone is required", ErrInvalidPhone)
	}
	if !phoneRegex.MatchString(phone) {
		return ErrInvalidPhone
	}
	return nil
}

// validatePassword validates password strength
func (s *Service) validatePassword(password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}
	return nil
}

// hashPassword hashes a password using bcrypt
func (s *Service) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// verifyPassword verifies a password against a hash
func (s *Service) verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// generateToken generates a JWT token for a user
func (s *Service) generateToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours

	claims := &JWTClaims{
		UserID:   user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "user-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
