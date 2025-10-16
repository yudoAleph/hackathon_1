package middleware

import (
	"net/http"
	"strings"

	"user-service/internal/app/service"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token and sets userID in context
func AuthMiddleware(svc *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":      0,
				"status_code": http.StatusUnauthorized,
				"message":     "Unauthorized - missing token",
				"data":        gin.H{},
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":      0,
				"status_code": http.StatusUnauthorized,
				"message":     "Unauthorized - invalid token format",
				"data":        gin.H{},
			})
			c.Abort()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":      0,
				"status_code": http.StatusUnauthorized,
				"message":     "Unauthorized - empty token",
				"data":        gin.H{},
			})
			c.Abort()
			return
		}

		// Validate token
		userID, err := svc.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":      0,
				"status_code": http.StatusUnauthorized,
				"message":     "Unauthorized - invalid or expired token",
				"data":        gin.H{},
			})
			c.Abort()
			return
		}

		// Set userID in context
		c.Set("userID", userID)
		c.Next()
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// LoggerMiddleware logs requests
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Before request
		c.Next()

		// After request
		statusCode := c.Writer.Status()
		if statusCode >= 400 {
			// Log errors (could integrate with proper logger)
			// fmt.Printf("[%s] %s %d\n", c.Request.Method, c.Request.URL.Path, statusCode)
		}
	}
}
