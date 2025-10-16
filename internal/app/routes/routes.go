package routes

import (
	"user-service/internal/app/handlers"
	"user-service/internal/app/service"
	"user-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(router *gin.Engine, handler *handlers.Handler, svc *service.Service) {
	// Apply global middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware())

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
		// ========================================
		// PUBLIC ROUTES (No authentication)
		// ========================================

		// Auth endpoints
		auth := api.Group("/auth")
		{
			auth.POST("/register", handler.Register) // POST /api/v1/auth/register
			auth.POST("/login", handler.Login)       // POST /api/v1/auth/login
		}

		// ========================================
		// PROTECTED ROUTES (Require authentication)
		// ========================================

		// Auth middleware
		authMiddleware := middleware.AuthMiddleware(svc)

		// User profile endpoints
		api.GET("/me", authMiddleware, handler.GetProfile)    // GET /api/v1/me
		api.PUT("/me", authMiddleware, handler.UpdateProfile) // PUT /api/v1/me

		// Contact endpoints
		contacts := api.Group("/contacts")
		contacts.Use(authMiddleware)
		{
			contacts.GET("", handler.ListContacts)         // GET /api/v1/contacts?q=&page=1&limit=20
			contacts.POST("", handler.CreateContact)       // POST /api/v1/contacts
			contacts.GET("/:id", handler.GetContact)       // GET /api/v1/contacts/:id
			contacts.PUT("/:id", handler.UpdateContact)    // PUT /api/v1/contacts/:id
			contacts.DELETE("/:id", handler.DeleteContact) // DELETE /api/v1/contacts/:id
		}
	}
}
