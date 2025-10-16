package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"user-service/configs"
	"user-service/internal/app/models"
	"user-service/internal/app/repository"
	"user-service/internal/app/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db      *gorm.DB
	service *service.Service
}

func NewHandler(cfg configs.Config, db *gorm.DB) *Handler {
	userRepo := repository.NewUserRepository(db)
	contactRepo := repository.NewContactRepository(db)
	svc := service.NewService(userRepo, contactRepo, cfg.JWTSecret)
	return &Handler{db: db, service: svc}
}

// GetService returns the service instance (for middleware)
func (h *Handler) GetService() *service.Service {
	return h.service
}

// StandardResponse represents the standard API response format
type StandardResponse struct {
	Status     int         `json:"status"`
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

// TokenData represents the token structure in response
type TokenData struct {
	AccessToken string `json:"access_token"`
}

// AuthResponseData represents the auth response data structure
type AuthResponseData struct {
	ID        uint       `json:"id"`
	FullName  string     `json:"full_name"`
	Email     string     `json:"email"`
	Phone     *string    `json:"phone,omitempty"` // Optional field
	AvatarURL *string    `json:"avatar_url,omitempty"`
	Token     *TokenData `json:"token,omitempty"`
}

// ContactsListData represents contacts list response data
type ContactsListData struct {
	Count    int                       `json:"count"`
	Page     int                       `json:"page"`
	Limit    int                       `json:"limit"`
	Contacts []*models.ContactResponse `json:"contacts"`
}

// successResponse helper function
func (h *Handler) successResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, StandardResponse{
		Status:     1,
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	})
}

// errorResponse helper function
func (h *Handler) errorResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	c.JSON(statusCode, StandardResponse{
		Status:     0,
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	})
}

// validationErrorResponse helper function
func (h *Handler) validationErrorResponse(c *gin.Context, field string, messages []string) {
	c.JSON(http.StatusBadRequest, StandardResponse{
		Status:     0,
		StatusCode: http.StatusBadRequest,
		Message:    "Validation error",
		Data:       gin.H{field: messages},
	})
}

// Ping health check endpoint
func (h *Handler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// ============================================================================
// AUTH HANDLERS
// ============================================================================

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request body", gin.H{})
		return
	}

	// Call service
	authResp, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		// Handle specific errors
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			h.errorResponse(c, http.StatusConflict, "Email already registered", gin.H{})
			return
		}
		if errors.Is(err, service.ErrInvalidEmail) {
			h.validationErrorResponse(c, "email", []string{"invalid format"})
			return
		}
		if errors.Is(err, service.ErrInvalidPhone) {
			h.validationErrorResponse(c, "phone", []string{"invalid format"})
			return
		}
		if errors.Is(err, service.ErrWeakPassword) {
			h.validationErrorResponse(c, "password", []string{"must be at least 8 characters"})
			return
		}
		// Log the actual error for debugging
		c.Error(fmt.Errorf("registration failed: %w", err))
		h.errorResponse(c, http.StatusInternalServerError, "Internal server error", gin.H{})
		return
	}

	// Format response
	data := AuthResponseData{
		ID:        authResp.User.ID,
		FullName:  authResp.User.FullName,
		Email:     authResp.User.Email,
		Phone:     authResp.User.Phone,
		AvatarURL: authResp.User.AvatarURL,
		Token: &TokenData{
			AccessToken: authResp.Token,
		},
	}

	h.successResponse(c, http.StatusCreated, "Registration success", data)
}

// Login handles user authentication
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request body", gin.H{})
		return
	}

	// Call service
	authResp, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			h.errorResponse(c, http.StatusUnauthorized, "Invalid email or password", gin.H{})
			return
		}
		h.errorResponse(c, http.StatusInternalServerError, "Internal server error", gin.H{})
		return
	}

	// Format response
	data := AuthResponseData{
		ID:        authResp.User.ID,
		FullName:  authResp.User.FullName,
		Email:     authResp.User.Email,
		Phone:     authResp.User.Phone,
		AvatarURL: authResp.User.AvatarURL,
		Token: &TokenData{
			AccessToken: authResp.Token,
		},
	}

	h.successResponse(c, http.StatusOK, "Login success", data)
}

// ============================================================================
// USER PROFILE HANDLERS
// ============================================================================

// GetProfile retrieves the logged-in user's profile
func (h *Handler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.errorResponse(c, http.StatusUnauthorized, "Unauthorized - invalid or expired token", gin.H{})
		return
	}

	profile, err := h.service.GetProfile(c.Request.Context(), userID.(uint))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.errorResponse(c, http.StatusNotFound, "User not found", gin.H{})
			return
		}
		h.errorResponse(c, http.StatusInternalServerError, "Internal server error", gin.H{})
		return
	}

	// Format response (without token)
	data := AuthResponseData{
		ID:        profile.ID,
		FullName:  profile.FullName,
		Email:     profile.Email,
		Phone:     profile.Phone,
		AvatarURL: profile.AvatarURL,
	}

	h.successResponse(c, http.StatusOK, "Profile loaded successfully", data)
}

// UpdateProfile updates the logged-in user's profile
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.errorResponse(c, http.StatusUnauthorized, "Unauthorized", gin.H{})
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request body", gin.H{})
		return
	}

	// Validate full_name if provided
	if req.FullName != "" && strings.TrimSpace(req.FullName) == "" {
		h.validationErrorResponse(c, "full_name", []string{"must not be empty"})
		return
	}

	profile, err := h.service.UpdateProfile(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.errorResponse(c, http.StatusNotFound, "User not found", gin.H{})
			return
		}
		if errors.Is(err, service.ErrInvalidPhone) {
			h.validationErrorResponse(c, "phone", []string{"invalid format"})
			return
		}
		h.errorResponse(c, http.StatusInternalServerError, "Internal server error", gin.H{})
		return
	}

	// Format response
	data := AuthResponseData{
		ID:        profile.ID,
		FullName:  profile.FullName,
		Email:     profile.Email,
		Phone:     profile.Phone,
		AvatarURL: profile.AvatarURL,
	}

	h.successResponse(c, http.StatusOK, "Profile updated successfully", data)
}

// ============================================================================
// CONTACT HANDLERS
// ============================================================================

// ListContacts retrieves contacts with search and pagination
func (h *Handler) ListContacts(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.errorResponse(c, http.StatusUnauthorized, "Unauthorized", gin.H{})
		return
	}

	// Parse query parameters
	var req models.ListContactsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid query parameters", gin.H{})
		return
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	// Get search query from 'q' parameter
	req.Search = c.Query("q")

	resp, err := h.service.ListContacts(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Internal server error", gin.H{})
		return
	}

	// Format response
	data := ContactsListData{
		Count:    int(resp.Pagination.Total),
		Page:     resp.Pagination.Page,
		Limit:    resp.Pagination.Limit,
		Contacts: resp.Data.([]*models.ContactResponse),
	}

	h.successResponse(c, http.StatusOK, "Contacts loaded successfully", data)
}

// CreateContact creates a new contact
func (h *Handler) CreateContact(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.errorResponse(c, http.StatusUnauthorized, "Unauthorized", gin.H{})
		return
	}

	var req models.CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request body", gin.H{})
		return
	}

	contact, err := h.service.CreateContact(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		if errors.Is(err, service.ErrPhoneAlreadyExists) {
			h.errorResponse(c, http.StatusConflict, "Contact phone already exists", gin.H{
				"phone": []string{req.Phone},
			})
			return
		}
		if errors.Is(err, service.ErrInvalidPhone) {
			h.validationErrorResponse(c, "phone", []string{"invalid format"})
			return
		}
		if errors.Is(err, service.ErrInvalidEmail) {
			h.validationErrorResponse(c, "email", []string{"invalid format"})
			return
		}
		if errors.Is(err, service.ErrInvalidContactData) {
			h.errorResponse(c, http.StatusBadRequest, err.Error(), gin.H{})
			return
		}
		h.errorResponse(c, http.StatusInternalServerError, "Internal server error", gin.H{})
		return
	}

	h.successResponse(c, http.StatusCreated, "Contact created successfully", contact)
}

// GetContact retrieves a contact by ID
func (h *Handler) GetContact(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.errorResponse(c, http.StatusUnauthorized, "Unauthorized", gin.H{})
		return
	}

	contactID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid contact ID", gin.H{})
		return
	}

	contact, err := h.service.GetContact(c.Request.Context(), userID.(uint), uint(contactID))
	if err != nil {
		if errors.Is(err, service.ErrContactNotFound) {
			h.errorResponse(c, http.StatusNotFound, "Contact not found", gin.H{})
			return
		}
		if errors.Is(err, service.ErrUnauthorizedAccess) {
			h.errorResponse(c, http.StatusForbidden, "Forbidden", gin.H{})
			return
		}
		h.errorResponse(c, http.StatusInternalServerError, "Internal server error", gin.H{})
		return
	}

	h.successResponse(c, http.StatusOK, "Contact detail loaded", contact)
}

// UpdateContact updates an existing contact
func (h *Handler) UpdateContact(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.errorResponse(c, http.StatusUnauthorized, "Unauthorized", gin.H{})
		return
	}

	contactID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid contact ID", gin.H{})
		return
	}

	var req models.UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request body", gin.H{})
		return
	}

	contact, err := h.service.UpdateContact(c.Request.Context(), userID.(uint), uint(contactID), &req)
	if err != nil {
		if errors.Is(err, service.ErrContactNotFound) {
			h.errorResponse(c, http.StatusNotFound, "Contact not found", gin.H{})
			return
		}
		if errors.Is(err, service.ErrPhoneAlreadyExists) {
			h.errorResponse(c, http.StatusConflict, "Phone number already exists", gin.H{})
			return
		}
		if errors.Is(err, service.ErrInvalidPhone) {
			h.validationErrorResponse(c, "phone", []string{"invalid format"})
			return
		}
		if errors.Is(err, service.ErrInvalidEmail) {
			h.validationErrorResponse(c, "email", []string{"invalid format"})
			return
		}
		if errors.Is(err, service.ErrUnauthorizedAccess) {
			h.errorResponse(c, http.StatusForbidden, "Forbidden", gin.H{})
			return
		}
		h.errorResponse(c, http.StatusInternalServerError, "Internal server error", gin.H{})
		return
	}

	h.successResponse(c, http.StatusOK, "Contact updated successfully", contact)
}

// DeleteContact deletes a contact
func (h *Handler) DeleteContact(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.errorResponse(c, http.StatusUnauthorized, "Unauthorized", gin.H{})
		return
	}

	contactID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid contact ID", gin.H{})
		return
	}

	err = h.service.DeleteContact(c.Request.Context(), userID.(uint), uint(contactID))
	if err != nil {
		if errors.Is(err, service.ErrContactNotFound) {
			h.errorResponse(c, http.StatusNotFound, "Contact not found", gin.H{})
			return
		}
		if errors.Is(err, service.ErrUnauthorizedAccess) {
			h.errorResponse(c, http.StatusForbidden, "Forbidden", gin.H{})
			return
		}
		h.errorResponse(c, http.StatusInternalServerError, "Internal server error", gin.H{})
		return
	}

	h.successResponse(c, http.StatusOK, "Contact deleted successfully", gin.H{})
}
