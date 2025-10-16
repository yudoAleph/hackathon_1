package routes

import (
	"user-service/internal/app"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(router *gin.Engine, handler *app.Handler) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "contact-management-api",
			"version": "1.0.0",
		})
	})

	// API v1 routes
	api := router.Group("/api/v1")
	{
		// Public routes
		api.GET("/ping", handler.Ping)

		// Mobile routes
		mobile := api.Group("/mobile")
		{
			mobile.POST("/users/:id", handler.Get)
		}
	}
}
