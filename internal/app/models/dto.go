package models

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest represents the user registration request payload
type RegisterRequest struct {
	FullName string  `json:"full_name" binding:"required"`
	Email    string  `json:"email" binding:"required,email"`
	Phone    *string `json:"phone,omitempty"` // Optional field
	Password string  `json:"password" binding:"required,min=6"`
}

// UpdateUserRequest represents the update user profile request payload
type UpdateUserRequest struct {
	FullName  string  `json:"full_name" binding:"required"`
	Phone     string  `json:"phone" binding:"required"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// UpdateProfileRequest represents the update profile request payload
type UpdateProfileRequest struct {
	FullName  string  `json:"full_name,omitempty"`
	Phone     string  `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// CreateContactRequest represents the create contact request payload
type CreateContactRequest struct {
	FullName string  `json:"full_name" binding:"required"`
	Phone    string  `json:"phone" binding:"required"`
	Email    *string `json:"email,omitempty" binding:"omitempty,email"`
	Favorite bool    `json:"favorite"`
}

// UpdateContactRequest represents the update contact request payload
type UpdateContactRequest struct {
	FullName *string `json:"full_name,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Email    *string `json:"email,omitempty" binding:"omitempty,email"`
	Favorite *bool   `json:"favorite,omitempty"`
}

// ListContactsRequest represents query parameters for listing contacts
type ListContactsRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	Limit    int    `form:"limit" binding:"min=1,max=100"`
	Search   string `form:"q"`
	Favorite *bool  `form:"favorite"`
}

// Response represents a standard API response
type Response struct {
	Status     int         `json:"status"`
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
	HasNextPage bool  `json:"has_next_page"`
	HasPrevPage bool  `json:"has_prev_page"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Status     int         `json:"status"`
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Error      interface{} `json:"error,omitempty"`
}

// AuthResponse represents authentication response with token
type AuthResponse struct {
	User  *UserResponse `json:"user"`
	Token string        `json:"token"`
}
