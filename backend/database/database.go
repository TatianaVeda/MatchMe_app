package database

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"backend/models"
)

// DB global variable to hold the database connection
var DB *gorm.DB

// ConnectDB initializes the database connection and handles migrations
func ConnectDB() *gorm.DB {
	// Use environment variables for database configuration
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=match_me port=5432 sslmode=disable"
	}

	// Connect to PostgreSQL database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Auto-migrate models
	err = db.AutoMigrate(&models.User{}, &models.Chat{}, &models.ConnectionRequest{}, &models.DismissedUser{}, &models.Recommendation{})
	if err != nil {
		log.Printf("❌ Error migrating database: %v", err)
	} else {
		log.Println("✅ Database migration successful!")
	}

	// Set the global DB variable
	DB = db
	log.Println("✅ Database connected successfully!")

	// Return the database instance for further use
	return db
}