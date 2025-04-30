package main

import (
	"log"
	"m/backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func createAdminUser() {
	dsn := "host=localhost user=user password=password dbname=sopostavmenya port=5433 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Хэширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin789"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	// Создаем пользователя с email admin@example.com
	adminUser := models.User{
		Email:          "admin@example.com",
		HashedPassword: string(hashedPassword),
	}

	if err := db.Create(&adminUser).Error; err != nil {
		log.Fatalf("failed to create admin user: %v", err)
	}

	log.Println("Admin user created successfully!")
}

func main() {
	createAdminUser()
}
