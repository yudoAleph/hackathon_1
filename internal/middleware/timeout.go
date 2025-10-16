package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TimeoutMiddleware creates a middleware that times out requests after the specified duration
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace the request context with the timeout context
		c.Request = c.Request.WithContext(ctx)

		// Channel to signal when the request is done
		finished := make(chan struct{})

		// Run the request in a goroutine
		go func() {
			c.Next()
			close(finished)
		}()

		// Wait for either the request to finish or the timeout
		select {
		case <-finished:
			// Request completed successfully
			return
		case <-ctx.Done():
			// Timeout occurred
			if ctx.Err() == context.DeadlineExceeded {
				c.JSON(http.StatusRequestTimeout, gin.H{
					"status":      0,
					"status_code": http.StatusRequestTimeout,
					"message":     "Request timeout - operation took too long",
					"data":        gin.H{},
				})
				c.Abort()
			}
		}
	}
}

// DefaultTimeoutMiddleware creates a middleware with 30 seconds timeout
func DefaultTimeoutMiddleware() gin.HandlerFunc {
	return TimeoutMiddleware(30 * time.Second)
}
