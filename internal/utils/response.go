package utils

import (
	"github.com/gin-gonic/gin"
)

// StandardResponse represents the standard API response format
type StandardResponse struct {
	Status     int         `json:"status"`      // 1 for success, 0 for error
	StatusCode int         `json:"status_code"` // HTTP status code
	Message    string      `json:"message"`     // Human-readable message
	Data       interface{} `json:"data"`        // Response data or error details
}

// SuccessResponse creates a success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	if data == nil {
		data = gin.H{}
	}

	c.JSON(statusCode, StandardResponse{
		Status:     1,
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	})
}

// ErrorResponse creates an error response
func ErrorResponse(c *gin.Context, statusCode int, message string, data interface{}) {
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

// ValidationErrorResponse creates a validation error response
func ValidationErrorResponse(c *gin.Context, field string, messages []string) {
	c.JSON(400, StandardResponse{
		Status:     0,
		StatusCode: 400,
		Message:    "Validation error",
		Data: gin.H{
			field: messages,
		},
	})
}

// ValidationErrorsResponse creates a validation error response with multiple fields
func ValidationErrorsResponse(c *gin.Context, errors map[string][]string) {
	c.JSON(400, StandardResponse{
		Status:     0,
		StatusCode: 400,
		Message:    "Validation error",
		Data:       errors,
	})
}

// UnauthorizedResponse creates an unauthorized error response
func UnauthorizedResponse(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}

	ErrorResponse(c, 401, message, gin.H{})
}

// ForbiddenResponse creates a forbidden error response
func ForbiddenResponse(c *gin.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}

	ErrorResponse(c, 403, message, gin.H{})
}

// NotFoundResponse creates a not found error response
func NotFoundResponse(c *gin.Context, message string) {
	if message == "" {
		message = "Resource not found"
	}

	ErrorResponse(c, 404, message, gin.H{})
}

// ConflictResponse creates a conflict error response
func ConflictResponse(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = "Resource conflict"
	}

	if data == nil {
		data = gin.H{}
	}

	ErrorResponse(c, 409, message, data)
}

// InternalErrorResponse creates an internal server error response
func InternalErrorResponse(c *gin.Context, message string) {
	if message == "" {
		message = "Internal server error"
	}

	ErrorResponse(c, 500, message, gin.H{})
}

// BadRequestResponse creates a bad request error response
func BadRequestResponse(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = "Bad request"
	}

	if data == nil {
		data = gin.H{}
	}

	ErrorResponse(c, 400, message, data)
}

// CreatedResponse creates a created success response (201)
func CreatedResponse(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = "Resource created successfully"
	}

	SuccessResponse(c, 201, message, data)
}

// OKResponse creates an OK success response (200)
func OKResponse(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = "Success"
	}

	SuccessResponse(c, 200, message, data)
}

// NoContentResponse creates a no content success response (204)
func NoContentResponse(c *gin.Context) {
	c.JSON(204, StandardResponse{
		Status:     1,
		StatusCode: 204,
		Message:    "Success",
		Data:       gin.H{},
	})
}
