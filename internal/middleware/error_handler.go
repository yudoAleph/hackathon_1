package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware recovers from panics and returns consistent JSON error responses
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error and stack trace
				fmt.Printf("PANIC: %v\n", err)
				fmt.Printf("Stack trace:\n%s\n", debug.Stack())

				// Check if headers were already written
				if !c.Writer.Written() {
					// Return consistent error response
					c.JSON(http.StatusInternalServerError, gin.H{
						"status":      0,
						"status_code": http.StatusInternalServerError,
						"message":     "Internal server error",
						"data":        gin.H{},
					})
				}

				// Abort the request
				c.Abort()
			}
		}()

		c.Next()
	}
}

// NotFoundHandler handles 404 errors with consistent JSON response
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"status":      0,
			"status_code": http.StatusNotFound,
			"message":     "Endpoint not found",
			"data":        gin.H{},
		})
	}
}

// MethodNotAllowedHandler handles 405 errors with consistent JSON response
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"status":      0,
			"status_code": http.StatusMethodNotAllowed,
			"message":     "Method not allowed",
			"data":        gin.H{},
		})
	}
}
