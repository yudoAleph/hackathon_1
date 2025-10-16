package service

import (
	"context"
	"testing"

	"user-service/internal/app/models"
	"user-service/internal/app/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) CheckEmailExists(ctx context.Context, email string, excludeUserID uint) (bool, error) {
	args := m.Called(ctx, email, excludeUserID)
	return args.Bool(0), args.Error(1)
}

// MockContactRepository is a mock implementation of ContactRepository
type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) Create(ctx context.Context, contact *models.Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *MockContactRepository) GetByID(ctx context.Context, userID, contactID uint) (*models.Contact, error) {
	args := m.Called(ctx, userID, contactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Contact), args.Error(1)
}

func (m *MockContactRepository) Update(ctx context.Context, contact *models.Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *MockContactRepository) Delete(ctx context.Context, userID, contactID uint) error {
	args := m.Called(ctx, userID, contactID)
	return args.Error(0)
}

func (m *MockContactRepository) List(ctx context.Context, userID uint, req *models.ListContactsRequest) ([]models.Contact, int64, error) {
	args := m.Called(ctx, userID, req)
	return args.Get(0).([]models.Contact), args.Get(1).(int64), args.Error(2)
}

func (m *MockContactRepository) CheckPhoneExists(ctx context.Context, userID uint, phone string, excludeContactID uint) (bool, error) {
	args := m.Called(ctx, userID, phone, excludeContactID)
	return args.Bool(0), args.Error(1)
}

// ============================================================================
// USER SERVICE TESTS
// ============================================================================

func TestService_Register(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockContactRepo := new(MockContactRepository)
	service := NewService(mockUserRepo, mockContactRepo, "test-secret")

	t.Run("successful registration", func(t *testing.T) {
		ctx := context.Background()
		req := &models.RegisterRequest{
			FullName: "John Doe",
			Email:    "john@example.com",
			Phone:    "081234567890",
			Password: "password123",
		}

		mockUserRepo.On("CheckEmailExists", ctx, "john@example.com", uint(0)).Return(false, nil).Once()
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil).Once()

		resp, err := service.Register(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Token)
		assert.Equal(t, "john@example.com", resp.User.Email)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		ctx := context.Background()
		req := &models.RegisterRequest{
			FullName: "Jane Doe",
			Email:    "existing@example.com",
			Phone:    "081234567890",
			Password: "password123",
		}

		mockUserRepo.On("CheckEmailExists", ctx, "existing@example.com", uint(0)).Return(true, nil).Once()

		resp, err := service.Register(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrEmailAlreadyExists)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("invalid email format", func(t *testing.T) {
		ctx := context.Background()
		req := &models.RegisterRequest{
			FullName: "John Doe",
			Email:    "invalid-email",
			Phone:    "081234567890",
			Password: "password123",
		}

		resp, err := service.Register(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrInvalidEmail)
	})

	t.Run("invalid phone format", func(t *testing.T) {
		ctx := context.Background()
		req := &models.RegisterRequest{
			FullName: "John Doe",
			Email:    "john@example.com",
			Phone:    "123", // Too short
			Password: "password123",
		}

		resp, err := service.Register(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrInvalidPhone)
	})

	t.Run("weak password", func(t *testing.T) {
		ctx := context.Background()
		req := &models.RegisterRequest{
			FullName: "John Doe",
			Email:    "john@example.com",
			Phone:    "081234567890",
			Password: "123", // Too short
		}

		resp, err := service.Register(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrWeakPassword)
	})
}

func TestService_Login(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockContactRepo := new(MockContactRepository)
	service := NewService(mockUserRepo, mockContactRepo, "test-secret")

	t.Run("successful login", func(t *testing.T) {
		ctx := context.Background()
		req := &models.LoginRequest{
			Email:    "john@example.com",
			Password: "password123",
		}

		// Hash the password for comparison
		hashedPassword, _ := service.hashPassword("password123")
		user := &models.User{
			ID:       1,
			FullName: "John Doe",
			Email:    "john@example.com",
			Password: hashedPassword,
		}

		mockUserRepo.On("GetByEmail", ctx, "john@example.com").Return(user, nil).Once()

		resp, err := service.Login(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Token)
		assert.Equal(t, "john@example.com", resp.User.Email)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		ctx := context.Background()
		req := &models.LoginRequest{
			Email:    "notfound@example.com",
			Password: "password123",
		}

		mockUserRepo.On("GetByEmail", ctx, "notfound@example.com").Return(nil, repository.ErrNotFound).Once()

		resp, err := service.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("wrong password", func(t *testing.T) {
		ctx := context.Background()
		req := &models.LoginRequest{
			Email:    "john@example.com",
			Password: "wrongpassword",
		}

		hashedPassword, _ := service.hashPassword("password123")
		user := &models.User{
			ID:       1,
			Email:    "john@example.com",
			Password: hashedPassword,
		}

		mockUserRepo.On("GetByEmail", ctx, "john@example.com").Return(user, nil).Once()

		resp, err := service.Login(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestService_GetProfile(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockContactRepo := new(MockContactRepository)
	service := NewService(mockUserRepo, mockContactRepo, "test-secret")

	t.Run("successful get profile", func(t *testing.T) {
		ctx := context.Background()
		user := &models.User{
			ID:       1,
			FullName: "John Doe",
			Email:    "john@example.com",
			Phone:    "081234567890",
		}

		mockUserRepo.On("GetByID", ctx, uint(1)).Return(user, nil).Once()

		profile, err := service.GetProfile(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "john@example.com", profile.Email)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		ctx := context.Background()

		mockUserRepo.On("GetByID", ctx, uint(999)).Return(nil, repository.ErrNotFound).Once()

		profile, err := service.GetProfile(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.ErrorIs(t, err, ErrUserNotFound)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestService_ValidateToken(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockContactRepo := new(MockContactRepository)
	service := NewService(mockUserRepo, mockContactRepo, "test-secret")

	t.Run("valid token", func(t *testing.T) {
		user := &models.User{
			ID:       1,
			FullName: "John Doe",
			Email:    "john@example.com",
		}

		token, err := service.generateToken(user)
		assert.NoError(t, err)

		userID, err := service.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), userID)
	})

	t.Run("invalid token", func(t *testing.T) {
		userID, err := service.ValidateToken("invalid-token")
		assert.Error(t, err)
		assert.Equal(t, uint(0), userID)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
}

// ============================================================================
// CONTACT SERVICE TESTS
// ============================================================================

func TestService_CreateContact(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockContactRepo := new(MockContactRepository)
	service := NewService(mockUserRepo, mockContactRepo, "test-secret")

	t.Run("successful create contact", func(t *testing.T) {
		ctx := context.Background()
		email := "contact@example.com"
		req := &models.CreateContactRequest{
			FullName: "Jane Doe",
			Phone:    "081234567890",
			Email:    &email,
		}

		mockContactRepo.On("CheckPhoneExists", ctx, uint(1), "081234567890", uint(0)).Return(false, nil).Once()
		mockContactRepo.On("Create", ctx, mock.AnythingOfType("*models.Contact")).Return(nil).Once()

		resp, err := service.CreateContact(ctx, 1, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Jane Doe", resp.FullName)
		mockContactRepo.AssertExpectations(t)
	})

	t.Run("phone already exists", func(t *testing.T) {
		ctx := context.Background()
		req := &models.CreateContactRequest{
			FullName: "Jane Doe",
			Phone:    "081234567890",
		}

		mockContactRepo.On("CheckPhoneExists", ctx, uint(1), "081234567890", uint(0)).Return(true, nil).Once()

		resp, err := service.CreateContact(ctx, 1, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrPhoneAlreadyExists)
		mockContactRepo.AssertExpectations(t)
	})

	t.Run("invalid email format", func(t *testing.T) {
		ctx := context.Background()
		invalidEmail := "invalid-email"
		req := &models.CreateContactRequest{
			FullName: "Jane Doe",
			Phone:    "081234567890",
			Email:    &invalidEmail,
		}

		resp, err := service.CreateContact(ctx, 1, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrInvalidEmail)
	})
}

func TestService_ListContacts(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockContactRepo := new(MockContactRepository)
	service := NewService(mockUserRepo, mockContactRepo, "test-secret")

	t.Run("successful list contacts", func(t *testing.T) {
		ctx := context.Background()
		req := &models.ListContactsRequest{
			Page:  1,
			Limit: 10,
		}

		contacts := []models.Contact{
			{ID: 1, UserID: 1, FullName: "Contact 1", Phone: "081111111111"},
			{ID: 2, UserID: 1, FullName: "Contact 2", Phone: "082222222222"},
		}

		mockContactRepo.On("List", ctx, uint(1), req).Return(contacts, int64(2), nil).Once()

		resp, err := service.ListContacts(ctx, 1, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Data, 2)
		assert.Equal(t, int64(2), resp.Pagination.Total)
		assert.Equal(t, 1, resp.Pagination.Page)
		mockContactRepo.AssertExpectations(t)
	})

	t.Run("pagination defaults", func(t *testing.T) {
		ctx := context.Background()
		req := &models.ListContactsRequest{
			Page:  0, // Should default to 1
			Limit: 0, // Should default to 10
		}

		contacts := []models.Contact{}

		mockContactRepo.On("List", ctx, uint(1), mock.MatchedBy(func(r *models.ListContactsRequest) bool {
			return r.Page == 1 && r.Limit == 10
		})).Return(contacts, int64(0), nil).Once()

		resp, err := service.ListContacts(ctx, 1, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockContactRepo.AssertExpectations(t)
	})
}

func TestService_Validation(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockContactRepo := new(MockContactRepository)
	service := NewService(mockUserRepo, mockContactRepo, "test-secret")

	t.Run("validate email", func(t *testing.T) {
		// Valid emails
		assert.NoError(t, service.validateEmail("user@example.com"))
		assert.NoError(t, service.validateEmail("user.name@example.co.id"))
		assert.NoError(t, service.validateEmail("user+tag@example.com"))

		// Invalid emails
		assert.Error(t, service.validateEmail(""))
		assert.Error(t, service.validateEmail("invalid"))
		assert.Error(t, service.validateEmail("@example.com"))
		assert.Error(t, service.validateEmail("user@"))
	})

	t.Run("validate phone", func(t *testing.T) {
		// Valid phones (Indonesia)
		assert.NoError(t, service.validatePhone("081234567890"))
		assert.NoError(t, service.validatePhone("628123456789"))
		assert.NoError(t, service.validatePhone("+6281234567890"))
		assert.NoError(t, service.validatePhone("0812345678"))

		// Invalid phones
		assert.Error(t, service.validatePhone(""))
		assert.Error(t, service.validatePhone("123"))
		assert.Error(t, service.validatePhone("12345678901234")) // Too long
		assert.Error(t, service.validatePhone("abc123456"))
	})

	t.Run("validate password", func(t *testing.T) {
		// Valid passwords
		assert.NoError(t, service.validatePassword("password123"))
		assert.NoError(t, service.validatePassword("12345678"))

		// Invalid passwords
		assert.Error(t, service.validatePassword(""))
		assert.Error(t, service.validatePassword("short"))
		assert.Error(t, service.validatePassword("1234567")) // 7 chars
	})
}
