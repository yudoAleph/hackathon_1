package main

import (
	"fmt"
	"log"
	"user-service/configs"
	"user-service/internal/app/handlers"
	"user-service/internal/app/routes"
	"user-service/internal/logger"
	"user-service/pkg/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title Contact Management API
// @version 1.0
// @description API untuk manajemen kontak dengan health check
// @termsOfService http://swagger.io/terms/

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Load .env file
	_ = godotenv.Load("configs/.env")

	// Load configuration
	cfg := configs.LoadConfig()

	// Initialize logger
	logConfig := logger.Config{
		Level:      "info",
		OutputPath: "logs/app.log",
	}
	if err := logger.Init(logConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	logger.Info("Starting Contact Management API",
		"port", cfg.Port,
		"environment", cfg.DBName,
	)

	// Build MySQL DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// Initialize database with MySQL
	database, err := db.NewSQLConnection(dsn)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		log.Fatalf("failed to initialize database: %v", err)
	}

	logger.Info("Database connected successfully")

	// Run auto migrations
	db.AutoMigrate(database)

	logger.Info("Database migrations completed")

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New() // Use gin.New() instead of gin.Default()

	// Add logger middleware FIRST
	router.Use(logger.LoggingMiddleware())

	// Initialize handler
	handler := handlers.NewHandler(cfg, database)

	// Setup routes (pass handler's service)
	routes.SetupRoutes(router, handler, handler.GetService())

	// Start server on port 9001
	logger.Info("Server starting", "port", "9001")
	log.Printf("Starting server on port 9001...")
	if err := router.Run(":9001"); err != nil {
		logger.Error("Failed to start server", "error", err)
		log.Fatalf("failed to start server: %v", err)
	}
}
