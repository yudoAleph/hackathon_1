package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"user-service/configs"
	"user-service/internal/app/migrations"
	"user-service/pkg/db"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	_ = godotenv.Load("configs/.env")

	// Parse command flags
	command := flag.String("command", "up", "Migration command: up, down, or status")
	flag.Parse()

	// Load configuration
	cfg := configs.LoadConfig()

	// Build DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// Initialize database connection
	database, err := db.NewSQLConnection(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get SQL DB connection for migrations
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("Failed to get SQL DB: %v", err)
	}

	// Create migration runner
	runner := migrations.NewRunner(sqlDB)

	// Execute command
	switch *command {
	case "up":
		fmt.Println("📦 Running database migrations...")
		if err := runner.MigrateUp(); err != nil {
			log.Fatalf("❌ Migration failed: %v", err)
		}
		fmt.Println("✅ All migrations completed successfully!")

	case "down":
		fmt.Println("⏪ Rolling back last migration...")
		if err := runner.MigrateDown(); err != nil {
			log.Fatalf("❌ Rollback failed: %v", err)
		}
		fmt.Println("✅ Rollback completed successfully!")

	case "status":
		fmt.Println("📊 Checking migration status...")
		if err := runner.Status(); err != nil {
			log.Fatalf("❌ Status check failed: %v", err)
		}

	default:
		log.Fatalf("Unknown command: %s. Use 'up', 'down', or 'status'", *command)
	}

	os.Exit(0)
}
