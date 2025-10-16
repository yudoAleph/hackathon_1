package main

import (
	"log"
	"user-service/configs"
	"user-service/internal/app"
	"user-service/internal/app/routes"
	"user-service/pkg/db"

	"github.com/gin-gonic/gin"
)

// @title Contact Management API
// @version 1.0
// @description API untuk manajemen kontak dengan health check
// @termsOfService http://swagger.io/terms/

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Load configuration
	cfg := configs.LoadConfig()

	// Initialize database
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Run auto migrations
	db.AutoMigrate(database)

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Initialize handler
	handler := app.NewHandler(cfg, database)

	// Setup routes
	routes.SetupRoutes(router, handler)

	// Start server on port 9001
	log.Printf("Starting server on port 9001...")
	if err := router.Run(":9001"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
